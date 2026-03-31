// agg_workers.go
package aggregator

import (
	"flownebula/agent/internal/protocol"
	"runtime"
)

func StartAggWorkers(workers int, agg *AggRing) {
	for i := 0; i < workers; i++ {
		go func(id int) {
			shard := &agg.shards[id]
			var ev protocol.Event

			const batch = 64

			for {
				if shard.read == shard.write {
					runtime.Gosched()
					continue
				}

				for n := 0; n < batch && shard.Pop(&ev); n++ {
					s := GetSession(ev.SessionID)
					s.Fast.addEvent(ev)
				}
			}
		}(i)
	}
}
