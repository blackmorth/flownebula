//go:build windows
// +build windows

package sampler

import "fmt"

// Windows stub
func StartCPUSampler(pid int, freq uint64) (int, error) {
	return -1, fmt.Errorf("perf_event_open not supported on Windows")
}

func ReadSamples(fd int, out chan<- Sample) {
	// no-op
}
