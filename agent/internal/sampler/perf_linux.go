//go:build linux
// +build linux

package sampler

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	perfTypeHardware    = 0
	perfCountHWCPU      = 0
	perfSampleIP        = 1 << 0
	perfSampleCallchain = 1 << 1

	// -1 en Go = ^uintptr(0)
	perfFlagFDNoGroup = ^uintptr(0)
	perfFlagAllCPU    = ^uintptr(0)
)

type perfEventAttr struct {
	Type       uint32
	Size       uint32
	Config     uint64
	SampleFreq uint64
	SampleType uint64
	ReadFormat uint64
	Flags      uint64
}

type Sample struct {
	IPs []uint64
}

func StartCPUSampler(pid int, freq uint64) (int, error) {
	attr := perfEventAttr{
		Type:       perfTypeHardware,
		Size:       uint32(unsafe.Sizeof(perfEventAttr{})),
		Config:     perfCountHWCPU,
		SampleFreq: freq,
		SampleType: perfSampleIP | perfSampleCallchain,
		Flags:      1 << 0, // disabled = 1
	}

	fd, _, errno := syscall.Syscall6(
		syscall.SYS_PERF_EVENT_OPEN,
		uintptr(unsafe.Pointer(&attr)),
		uintptr(pid),
		perfFlagAllCPU,
		perfFlagFDNoGroup,
		0,
		0,
	)

	if errno != 0 {
		return -1, fmt.Errorf("perf_event_open: %v", errno)
	}

	// enable
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, 0x2400, 0); errno != 0 {
		syscall.Close(int(fd))
		return -1, fmt.Errorf("ioctl ENABLE: %v", errno)
	}

	return int(fd), nil
}

func ReadSamples(fd int, out chan<- Sample) {
	buf := make([]byte, 4096)

	for {
		n, err := syscall.Read(fd, buf)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			fmt.Fprintf(os.Stderr, "perf read error: %v\n", err)
			return
		}
		if n <= 0 {
			continue
		}

		s := parseSample(buf[:n])
		if len(s.IPs) > 0 {
			out <- s
		}
	}
}

func parseSample(data []byte) Sample {
	if len(data) < 8 {
		return Sample{}
	}

	size := *(*uint32)(unsafe.Pointer(&data[4]))
	if int(size) > len(data) {
		return Sample{}
	}

	body := data[8:size]
	if len(body) < 8 {
		return Sample{}
	}

	nr := *(*uint64)(unsafe.Pointer(&body[0]))
	if nr == 0 {
		return Sample{}
	}

	if len(body) < int(8+nr*8) {
		return Sample{}
	}

	ips := make([]uint64, 0, nr)
	for i := 0; i < int(nr); i++ {
		ip := *(*uint64)(unsafe.Pointer(&body[8+i*8]))
		ips = append(ips, ip)
	}

	return Sample{IPs: ips}
}
