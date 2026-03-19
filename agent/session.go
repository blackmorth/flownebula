package agent

import (
	"fmt"
	"log"
	"os"
	"time"
)

func GetFuncName(id uint32) string {
	if id == 0 {
		return "main()"
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

	s = &Session{
		ID:       id,
		Root:     NewNode(0, 0),
		stack:    []*Node{},
		LastSeen: time.Now(),
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
		node.Metrics["wt"] += int64(ev.Inclusive)
		node.Metrics["ewt"] += int64(ev.Exclusive)
		node.Metrics["cpu"] += int64(ev.CPUTime)
		node.Metrics["mu"] += ev.MemDelta
		node.Metrics["io"] += int64(ev.IOWait)
		node.Metrics["nw"] += int64(ev.Network)
		if int64(ev.PeakMemory) > node.Metrics["pmu"] {
			node.Metrics["pmu"] = int64(ev.PeakMemory)
		}
	}
}

func (s *Session) Print() {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("\n=== Session %016x ===\n", s.ID)

	jsonData, err := s.ExportToDetailedJSON()
	if err == nil {
		filename := fmt.Sprintf("session_%016x.json", s.ID)
		_ = os.WriteFile(filename, jsonData, 0644)
		_ = os.WriteFile("session.json", jsonData, 0644)
		log.Printf("Exported session %016x to %s", s.ID, filename)
	} else {
		log.Printf("Failed to export session %016x: %v", s.ID, err)
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
