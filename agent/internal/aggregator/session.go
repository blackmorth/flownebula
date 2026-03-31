package aggregator

import (
	"flownebula/agent/internal/protocol"
	"sync"
	"time"
)

// --- Types de base ----------------------------------------------------------

type Node struct {
	ID       uint64
	FuncID   uint32
	Metrics  map[string]int64
	Children map[uint32]*Node
}

type Session struct {
	ID                  uint64
	Root                *Node
	stack               []*Node
	LastSeen            time.Time
	LastEventAt         time.Time
	AvgInterEventNanos  float64
	Mu                  sync.Mutex
	nodeCount           uint64
	Closed              bool
	Exported            bool
	Protocol            int
	Dropped             uint64
	FlushErrors         uint64
	BufferHighWatermark uint64
	FirstEventUnixNanos uint64
	LastEventUnixNanos  uint64
	Fast                *SessionFast
}

// --- Sharding ---------------------------------------------------------------

const NumShards = 32

var sessions = make(map[uint64]*Session)

type SessionShard struct {
	Mu       sync.RWMutex
	Sessions map[uint64]*Session
}

var SessionShards [NumShards]*SessionShard

func init() {
	for i := 0; i < NumShards; i++ {
		SessionShards[i] = &SessionShard{
			Sessions: make(map[uint64]*Session),
		}
	}
}

func shardFor(id uint64) *SessionShard {
	return SessionShards[id%NumShards]
}

// --- Session lifecycle ------------------------------------------------------

func newNode(id uint64, funcID uint32) *Node {
	return &Node{
		ID:       id,
		FuncID:   funcID,
		Metrics:  make(map[string]int64),
		Children: make(map[uint32]*Node),
	}
}

func GetSession(id uint64) *Session {
	s := sessions[id]
	if s != nil {
		return s
	}

	s = &Session{
		ID:   id,
		Fast: newSessionFast(),
	}

	sessions[id] = s
	return s
}

func (s *Session) touch() {
	s.Mu.Lock()
	s.LastSeen = time.Now()
	s.Mu.Unlock()
}

// --- Stack helpers ----------------------------------------------------------

func (s *Session) push(n *Node) {
	s.stack = append(s.stack, n)
}

func (s *Session) pop() *Node {
	if len(s.stack) == 0 {
		return s.Root
	}
	n := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return n
}

func (s *Session) top() *Node {
	if len(s.stack) == 0 {
		return s.Root
	}
	return s.stack[len(s.stack)-1]
}

// --- Event handling ---------------------------------------------------------

func (s *Session) AddEvent(ev protocol.Event) {
	s.Fast.addEvent(ev)
}
