package internal

import (
	"encoding/binary"
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

// --------------------------------------------------
// Sharded workers (lock-free ingestion)
// --------------------------------------------------

const NumShards = 32
const ChanSize = 4096

type shardWorker struct {
	ch       chan Packet
	sessions map[uint64]*Session
}

var workers [NumShards]*shardWorker

// InitWorkers doit être appelé au démarrage
func InitWorkers() {
	for i := 0; i < NumShards; i++ {
		w := &shardWorker{
			ch:       make(chan Packet, ChanSize),
			sessions: make(map[uint64]*Session),
		}
		workers[i] = w

		go w.run()
	}
}

// --------------------------------------------------
// Dispatcher (O(1), no lock)
// --------------------------------------------------

func dispatchPacket(p Packet) {
	data := p.Data[:p.N]

	if len(data) < 8 {
		BufferPool.Put(p.Data)
		return
	}

	sessID := binary.LittleEndian.Uint64(data[:8])
	idx := sessID % NumShards

	select {
	case workers[idx].ch <- p:
	default:
		// backpressure → drop
		BufferPool.Put(p.Data)
	}
}

// --------------------------------------------------
// Worker loop
// --------------------------------------------------

func (w *shardWorker) run() {
	for p := range w.ch {
		w.processPacket(p)
	}
}

// --------------------------------------------------
// Core processing (hot path)
// --------------------------------------------------

func (w *shardWorker) processPacket(p Packet) {
	data := p.Data[:p.N]
	defer BufferPool.Put(p.Data)

	// ---- FUNC NAME ----
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

	// ---- EVENTS ----
	for j := 0; j+EventSize <= len(data); j += EventSize {
		_ = data[j+EventSize-1] // bounds check hint

		base := j

		sessID := binary.LittleEndian.Uint64(data[base : base+8])

		s := w.sessions[sessID]
		if s == nil {
			s = newSession()
			w.sessions[sessID] = s
		}

		s.AddEvent(CallEvent{
			Type:       data[base+8],
			FuncID:     binary.LittleEndian.Uint32(data[base+9 : base+13]),
			Inclusive:  binary.LittleEndian.Uint64(data[base+13 : base+21]),
			Exclusive:  binary.LittleEndian.Uint64(data[base+21 : base+29]),
			CPUTime:    binary.LittleEndian.Uint64(data[base+29 : base+37]),
			MemDelta:   int64(binary.LittleEndian.Uint64(data[base+37 : base+45])),
			PeakMemory: binary.LittleEndian.Uint64(data[base+45 : base+53]),
			IOWait:     binary.LittleEndian.Uint64(data[base+53 : base+61]),
			Network:    binary.LittleEndian.Uint64(data[base+61 : base+69]),
		})
	}
}

// --------------------------------------------------
// Listener (entrée réseau)
// --------------------------------------------------

func ListenUnixgram(conn *net.UnixConn) {
	for {
		buf := BufferPool.Get().([]byte)

		n, _, err := conn.ReadFromUnix(buf)
		if err != nil {
			BufferPool.Put(buf)
			continue
		}

		if n < 17 {
			BufferPool.Put(buf)
			continue
		}

		dispatchPacket(Packet{
			Data: buf,
			N:    n,
		})
	}
}
