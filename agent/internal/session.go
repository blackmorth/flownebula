package internal

import (
	"fmt"
	"log"
	"os"
	"time"
)

func nsToUS(v uint64) int64 {
	// Probe timings are emitted in nanoseconds. The UI and exported payload
	// expect microseconds for wall/cpu timing dimensions.
	return int64((v + 500) / 1000)
}

func GetFuncName(id uint32) string {
	if id == 0 {
		return "{invalid}"
	}
	FuncNamesMu.RLock()
	defer FuncNamesMu.RUnlock()
	if name, ok := FuncNames[id]; ok {
		return name
	}
	return fmt.Sprintf("func_%d", id)
}

func NewNode(id uint64, funcID uint32) *Node {
	return &Node{
		ID:       id,
		FuncID:   funcID,
		Metrics:  make(map[string]int64),
		Children: make(map[uint32]*Node),
	}
}

func GetSession(id uint64) *Session {
	shard := GetShard(id)
	shard.Mu.RLock()
	s, ok := shard.Sessions[id]
	shard.Mu.RUnlock()

	if ok {
		s.mu.Lock()
		s.LastSeen = time.Now()
		s.mu.Unlock()
		return s
	}

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	// Double-check after acquiring the write lock
	if s, ok = shard.Sessions[id]; ok {
		s.mu.Lock()
		s.LastSeen = time.Now()
		s.mu.Unlock()
		return s
	}

	now := time.Now()
	s = &Session{
		ID:          id,
		Root:        NewNode(0, 1),
		stack:       []*Node{},
		LastSeen:    now,
		LastEventAt: now,
		Protocol:    ProtocolVersion,
	}
	shard.Sessions[id] = s
	return s
}

func (s *Session) Push(n *Node) {
	s.stack = append(s.stack, n)
}

func (s *Session) Pop() *Node {
	if len(s.stack) == 0 {
		return s.Root
	}
	n := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return n
}

func (s *Session) Top() *Node {
	if len(s.stack) == 0 {
		return s.Root
	}
	return s.stack[len(s.stack)-1]
}

func (s *Session) AddEvent(ev CallEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if !s.LastEventAt.IsZero() {
		delta := now.Sub(s.LastEventAt).Nanoseconds()
		if delta > 0 {
			if s.AvgInterEventNanos == 0 {
				s.AvgInterEventNanos = float64(delta)
			} else {
				const alpha = 0.2
				s.AvgInterEventNanos = alpha*float64(delta) + (1-alpha)*s.AvgInterEventNanos
			}
		}
	}
	s.LastEventAt = now
	s.LastSeen = now

	switch ev.Type {
	case EventEnter: // enter
		parent := s.Top()
		if parent == nil {
			return
		}
		if parent.Children == nil {
			parent.Children = make(map[uint32]*Node)
		}
		child, ok := parent.Children[ev.FuncID]
		if !ok {
			s.nodeCount++
			child = NewNode(s.nodeCount, ev.FuncID)
			parent.Children[ev.FuncID] = child
		}
		s.Push(child)

	case EventExit: // exit
		if len(s.stack) == 0 {
			return
		}

		node := s.Top()
		if node.FuncID != ev.FuncID {
			found := false
			for i := len(s.stack) - 1; i >= 0; i-- {
				if s.stack[i].FuncID == ev.FuncID {
					s.stack = s.stack[:i+1]
					node = s.stack[i]
					found = true
					break
				}
			}
			if !found {
				return
			}
		}

		_ = s.Pop()
		node.Metrics["ct"]++
		node.Metrics["wt"] += nsToUS(ev.Inclusive)
		node.Metrics["ewt"] += nsToUS(ev.Exclusive)
		node.Metrics["cpu"] += nsToUS(ev.CPUTime)
		node.Metrics["mu"] += ev.MemDelta
		node.Metrics["io"] += nsToUS(ev.IOWait)
		node.Metrics["nw"] += nsToUS(ev.Network)
		if int64(ev.PeakMemory) > node.Metrics["pmu"] {
			node.Metrics["pmu"] = int64(ev.PeakMemory)
		}
	case EventSessionEnd:
		s.Dropped = ev.IOWait
		s.Protocol = int(ev.Network)
		s.Closed = true
		return
	}
}

func (s *Session) Print() {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("\n=== Session %016x ===\n", s.ID)

	jsonData, err := s.ExportToDetailedJSON()
	if err != nil {
		log.Printf("Failed to export session %016x: %v", s.ID, err)
		return
	}

	filename := fmt.Sprintf("session_%016x.json", s.ID)
	_ = os.WriteFile(filename, jsonData, 0644)
	_ = os.WriteFile("session.json", jsonData, 0644)
	log.Printf("Exported session %016x to %s", s.ID, filename)

	if GlobalSender != nil {
		if err := GlobalSender.SendSession(jsonData); err != nil {
			log.Printf("Failed to send session %016x: %v", s.ID, err)
		} else {
			log.Printf("Session %016x sent to server", s.ID)
		}
	}
}

func CleanupSessions() {
	for range time.Tick(30 * time.Second) {
		for i := 0; i < NumShards; i++ {
			shard := SessionShards[i]
			shard.Mu.Lock()
			for id, s := range shard.Sessions {
				if time.Since(s.LastSeen) > 2*time.Minute {
					delete(shard.Sessions, id)
				}
			}
			shard.Mu.Unlock()
		}
	}
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Closed = true
}

func ExportSessionsLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		exportReadySessions()
	}
}

func exportReadySessions() {
	now := time.Now()
	baseIdleTimeout := 5 * time.Second // minimum
	maxIdleTimeout := 60 * time.Second

	for i := 0; i < NumShards; i++ {
		shard := SessionShards[i]
		shard.Mu.Lock()
		for id, s := range shard.Sessions {
			s.mu.Lock()

			// session déjà exportée → on peut la supprimer
			if s.Exported {
				s.mu.Unlock()
				delete(shard.Sessions, id)
				continue
			}

			adaptiveIdle := baseIdleTimeout
			if s.AvgInterEventNanos > 0 {
				candidate := time.Duration(2 * s.AvgInterEventNanos)
				if candidate > adaptiveIdle {
					adaptiveIdle = candidate
				}
			}
			if adaptiveIdle > maxIdleTimeout {
				adaptiveIdle = maxIdleTimeout
			}

			// session fermée OU inactive depuis un moment
			if s.Closed || now.Sub(s.LastSeen) > adaptiveIdle {
				s.mu.Unlock() // on libère avant Print() qui relock
				flushStart := time.Now()
				s.Print()
				recordFlushLatency(flushStart)

				s.mu.Lock()
				s.Exported = true
				s.mu.Unlock()

				delete(shard.Sessions, id)
				continue
			}

			s.mu.Unlock()
		}
		shard.Mu.Unlock()
	}
}
