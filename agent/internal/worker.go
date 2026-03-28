package internal

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"
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

var (
	ErrPacketTooShort       = errors.New("packet shorter than minimum header")
	ErrInvalidNamePacket    = errors.New("invalid function-name packet")
	ErrInvalidEventPayload  = errors.New("event payload is not aligned to event size")
	ErrUnknownProtocolEvent = errors.New("unknown event type")
	ErrInvalidSessionEnd    = errors.New("invalid session-end packet")
	ErrUnsupportedProtocol  = errors.New("unsupported protocol version")
)

func StartWorkers(workers int, eventChan <-chan Packet) {
	if workers <= 1 {
		go func() {
			for p := range eventChan {
				ProcessPacket(p)
			}
		}()
		return
	}

	workerQueues := make([]chan Packet, workers)
	for i := 0; i < workers; i++ {
		workerQueues[i] = make(chan Packet, 1024)
		go func(queue <-chan Packet) {
			for p := range queue {
				ProcessPacket(p)
			}
		}(workerQueues[i])
	}

	// Keep packet order per session by routing all packets for a given session ID
	// to the same worker queue.
	go func() {
		var rr uint64
		for p := range eventChan {
			idx := 0
			if p.N >= SessionIDSize {
				sessID := binary.LittleEndian.Uint64(p.Data[:SessionIDSize])
				idx = int(sessID % uint64(workers))
			} else {
				idx = int(rr % uint64(workers))
				rr++
			}
			workerQueues[idx] <- p
		}
	}()
}

func ProcessPacket(p Packet) {
	data := p.Data[:p.N]
	defer BufferPool.Put(p.Data)

	if err := validatePacket(data); err != nil {
		AgentMetrics.ValidationErrors.Add(1)
		return
	}

	AgentMetrics.ValidateProcessed.Add(1)

	// Special type for function names (255)
	if len(data) >= NameHeaderSize && data[SessionIDSize] == EventFuncName {
		AgentMetrics.ProbeNamePackets.Add(1)
		funcID := binary.LittleEndian.Uint32(data[9:13])
		nameLen := int(binary.LittleEndian.Uint32(data[13:17]))
		name := string(data[NameHeaderSize : NameHeaderSize+nameLen])
		FuncNamesMu.Lock()
		FuncNames[funcID] = name
		FuncNamesMu.Unlock()
		AgentMetrics.AggregateProcessed.Add(1)
		return
	}

	for j := 0; j+EventSize <= len(data); j += EventSize {
		d := data[j : j+EventSize]
		AgentMetrics.ProbeEventPackets.Add(1)
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
		AgentMetrics.AggregateProcessed.Add(1)
	}
}

func validatePacket(data []byte) error {
	if len(data) < SessionIDSize+1 {
		return ErrPacketTooShort
	}

	typeByte := data[SessionIDSize]
	switch typeByte {
	case EventFuncName:
		if len(data) < NameHeaderSize {
			return ErrInvalidNamePacket
		}
		nameLen := int(binary.LittleEndian.Uint32(data[13:17]))
		if NameHeaderSize+nameLen != len(data) {
			return ErrInvalidNamePacket
		}
		return nil
	case EventEnter, EventExit, EventSessionEnd:
		if len(data)%EventSize != 0 {
			return ErrInvalidEventPayload
		}
		for j := 0; j+EventSize <= len(data); j += EventSize {
			e := data[j+SessionIDSize]
			if e != EventEnter && e != EventExit && e != EventSessionEnd {
				return ErrUnknownProtocolEvent
			}
			if e == EventSessionEnd {
				funcID := binary.LittleEndian.Uint32(data[j+9 : j+13])
				if funcID != 0 {
					return ErrInvalidSessionEnd
				}
				version := binary.LittleEndian.Uint64(data[j+61 : j+69])
				if version != ProtocolVersion {
					return ErrUnsupportedProtocol
				}
				AgentMetrics.ProbeSessionEnds.Add(1)
			}
		}
		return nil
	default:
		return ErrUnknownProtocolEvent
	}
}

func enqueuePacketWithPolicy(eventChan chan<- Packet, pkt Packet, dropNewest bool) bool {
	if !dropNewest {
		eventChan <- pkt
		return true
	}
	select {
	case eventChan <- pkt:
		return true
	default:
		return false
	}
}

func ListenUnixgram(conn *net.UnixConn, eventChan chan<- Packet, dropNewestWhenFull bool, highWatermark float64) {
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

		AgentMetrics.IngestReceived.Add(1)
		AgentMetrics.QueueDepth.Set(int64(len(eventChan)))

		adaptiveDrop := dropNewestWhenFull
		if highWatermark > 0 {
			capCh := cap(eventChan)
			if capCh > 0 {
				fillRatio := float64(len(eventChan)) / float64(capCh)
				if fillRatio >= highWatermark {
					adaptiveDrop = true
				}
			}
		}

		if ok := enqueuePacketWithPolicy(eventChan, Packet{Data: buf, N: n}, adaptiveDrop); !ok {
			BufferPool.Put(buf)
			AgentMetrics.Drops.Add(1)
			continue
		}
		AgentMetrics.IngestAccepted.Add(1)
		AgentMetrics.QueueDepth.Set(int64(len(eventChan)))
	}
}

func recordFlushLatency(start time.Time) {
	AgentMetrics.FlushLatency.Observe(time.Since(start))
}
