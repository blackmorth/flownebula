package internal

import (
	"encoding/binary"
	"log"
	"net"
	"sync"
)

type Packet struct {
	Data []byte
	N    int
}

// --------------------------------------------------
// Buffer pool (zero alloc réseau)
// --------------------------------------------------

var BufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 65507)
	},
}

const NumShards = 32

func StartWorkers(workers int, eventChan <-chan Packet) {
	for i := 0; i < workers; i++ {
		go func() {
			for p := range eventChan {
				ProcessPacket(p)
			}
		}()
	}
}

func ProcessPacket(p Packet) {
	data := p.Data[:p.N]
	defer BufferPool.Put(p.Data)

	// Special type for function names (255)
	if len(data) >= NameHeaderSize && data[SessionIDSize] == EventFuncName {
		funcID := binary.LittleEndian.Uint32(data[9:13])
		nameLen := int(binary.LittleEndian.Uint32(data[13:17]))
		if NameHeaderSize+nameLen <= len(data) {
			name := string(data[NameHeaderSize : NameHeaderSize+nameLen])
			FuncNamesMu.Lock()
			FuncNames[funcID] = name
			FuncNamesMu.Unlock()
		}
		return
	}

	for j := 0; j+EventSize <= len(data); j += EventSize {
		d := data[j : j+EventSize]
		sessID := binary.LittleEndian.Uint64(d[:SessionIDSize])

		ev := CallEvent{
			Type:       d[8],
			FuncID:     binary.LittleEndian.Uint32(d[9:13]),
			Inclusive:  binary.LittleEndian.Uint64(d[13:21]),
			Exclusive:  binary.LittleEndian.Uint64(d[21:29]),
			CPUTime:    binary.LittleEndian.Uint64(d[29:37]),
			MemDelta:   int64(binary.LittleEndian.Uint64(d[37:45])),
			PeakMemory: binary.LittleEndian.Uint64(d[45:53]),
			IOWait:     binary.LittleEndian.Uint64(d[53:61]),
			Network:    binary.LittleEndian.Uint64(d[61:69]),
		}
		s := GetSession(sessID)
		s.AddEvent(ev)
	}
}

func ListenUnixgram(conn *net.UnixConn, eventChan chan<- Packet) {
	for {
		buf := BufferPool.Get().([]byte)
		n, _, err := conn.ReadFromUnix(buf)
		log.Printf("Received %d bytes", n)
		log.Printf("Raw packet: %x", buf[:n])
		if err != nil {
			log.Printf("ReadFromUnix error: %v", err)
			BufferPool.Put(buf)
			continue
		}
		if n < 17 {
			BufferPool.Put(buf)
			continue
		}
		select {
		case eventChan <- Packet{Data: buf, N: n}:
		default:
			BufferPool.Put(buf)
		}
	}
}
