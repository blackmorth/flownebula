package internal

import (
	"sync"
	"time"
)

type CallEvent struct {
	Type       uint8
	FuncID     uint32
	Inclusive  uint64
	Exclusive  uint64
	CPUTime    uint64
	MemDelta   int64
	PeakMemory uint64
	IOWait     uint64
	Network    uint64
}

const (
	EventEnter      = 0
	EventExit       = 1
	EventFuncName   = 255
	EventSessionEnd = 254
	ProtocolVersion = 1
	SessionIDSize   = 8
	NameHeaderSize  = 17 // 8 (SID) + 1 (Type) + 4 (ID) + 4 (Len)
	EventSize       = 8 + 1 + 4 + 8 + 8 + 8 + 8 + 8 + 8 + 8
)

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
	mu                  sync.Mutex
	nodeCount           uint64
	Closed              bool
	Exported            bool
	Protocol            int
	Dropped             uint64
	FlushErrors         uint64
	BufferHighWatermark uint64
}

type DetailedJSON struct {
	AgentSessionID      string               `json:"agent_session_id"`
	Dimensions          map[string]Dimension `json:"dimensions"`
	Root                string               `json:"root"`
	Nodes               map[string]JSONNode  `json:"nodes"`
	Edges               map[string]JSONEdge  `json:"edges"`
	Comparison          bool                 `json:"comparison"`
	Peaks               Peaks                `json:"peaks"`
	Language            string               `json:"language"`
	Protocol            int                  `json:"protocol_version"`
	DroppedEvents       uint64               `json:"dropped_events,omitempty"`
	FlushErrors         uint64               `json:"flush_errors,omitempty"`
	BufferHighWatermark uint64               `json:"buffer_high_watermark,omitempty"`
}

type JSONEdge struct {
	EdgeID string           `json:"edgeId"`
	Caller string           `json:"caller"`
	Callee string           `json:"callee"`
	Cost   map[string]int64 `json:"cost"`
}

type Peaks struct {
	Inclusive map[string]int64 `json:"inclusive"`
	Exclusive map[string]int64 `json:"exclusive"`
}

type Dimension struct {
	Dim     string `json:"dim"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

type JSONNode struct {
	NodeID              string             `json:"nodeId"`
	Name                string             `json:"name,omitempty"`
	CalledClass         string             `json:"called_class,omitempty"`
	Metrics             []string           `json:"metrics"`
	InclusiveCost       map[string]int64   `json:"inclusive_cost"`
	ExclusiveCost       map[string]int64   `json:"exclusive_cost"`
	InclusivePercentage map[string]float64 `json:"inclusive_percentage"`
	ExclusivePercentage map[string]float64 `json:"exclusive_percentage"`
}

var (
	FuncNames   = make(map[uint32]string)
	FuncNamesMu sync.RWMutex
)

type SessionShard struct {
	Sessions map[uint64]*Session
	Mu       sync.RWMutex
}

var SessionShards [NumShards]*SessionShard

func init() {
	for i := 0; i < NumShards; i++ {
		SessionShards[i] = &SessionShard{
			Sessions: make(map[uint64]*Session),
		}
	}
}

func GetShard(id uint64) *SessionShard {
	return SessionShards[id%NumShards]
}
