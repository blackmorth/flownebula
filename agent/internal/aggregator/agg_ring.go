package aggregator

import (
	"flownebula/agent/internal/protocol"
)

type AggShard struct {
	buffer []protocol.Event
	mask   uint64
	read   uint64
	write  uint64
}

type AggRing struct {
	shards []AggShard
}

func NewAggRing(shards int, size uint64) *AggRing {
	r := &AggRing{
		shards: make([]AggShard, shards),
	}

	// size DOIT être une puissance de 2 (comme ton Ring)
	for i := 0; i < shards; i++ {
		r.shards[i] = AggShard{
			buffer: make([]protocol.Event, size),
			mask:   size - 1,
			read:   0,
			write:  0,
		}
	}

	return r
}

func (r *AggRing) Push(shardID int, ev protocol.Event) bool {
	shard := &r.shards[shardID]

	w := shard.write
	nw := (w + 1) & shard.mask

	if nw == shard.read {
		return false // full
	}

	shard.buffer[w] = ev
	shard.write = nw
	return true
}

func (s *AggShard) Pop(ev *protocol.Event) bool {
	if s.read == s.write {
		return false
	}

	*ev = s.buffer[s.read]
	s.read = (s.read + 1) & s.mask
	return true
}
