// Package ingest workers.go
package ingest

import (
	"flownebula/agent/internal/aggregator"
	"runtime"
)

func StartWorkersRing(workers int, ring *Ring, agg *aggregator.AggRing) {
	for i := 0; i < workers; i++ {
		go func(id int) {
			shard := &ring.shards[id]
			aggShardID := id

			const batch = 64

			for {
				if shard.read == shard.write {
					runtime.Gosched()
					continue
				}

				for n := 0; n < batch && shard.read != shard.write; n++ {
					ev := shard.buffer[shard.read]
					shard.read = (shard.read + 1) & shard.mask

					_ = agg.Push(aggShardID, ev) // si false → drop, on pourra compter plus tard
				}
			}
		}(i)
	}
}
