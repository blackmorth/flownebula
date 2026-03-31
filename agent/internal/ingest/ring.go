package ingest

import (
	"flownebula/agent/internal/protocol"
)

type Shard struct {
	buffer []protocol.Event
	mask   int
	read   int
	write  int
}

type Ring struct {
	shards []Shard
}

func NewRing(workers int, size int) *Ring {
	r := &Ring{
		shards: make([]Shard, workers),
	}

	for i := 0; i < workers; i++ {
		r.shards[i] = Shard{
			buffer: make([]protocol.Event, size),
			mask:   size - 1,
		}
	}

	return r
}

func (r *Ring) Push(ev protocol.Event) bool {
	workerID := ev.SessionID % uint64(len(r.shards))
	shard := &r.shards[workerID]

	next := (shard.write + 1) & shard.mask
	if next == shard.read {
		return false
	}

	shard.buffer[shard.write] = ev
	shard.write = next
	return true
}
