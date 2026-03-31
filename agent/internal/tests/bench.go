package tests

import (
	"encoding/binary"
	"flownebula/agent/internal/aggregator"
	"flownebula/agent/internal/ingest"
	"flownebula/agent/internal/protocol"
	"log"
	"net"
	"time"
)

// RunBenchmark mesure le débit maximal du pipeline interne
// (ring + workers + SessionFast.addEvent) sans socket.
func RunBenchmark(workers int) {
	log.Printf("Starting internal benchmark with %d workers...", workers)

	ring := ingest.NewRing(workers, 1<<20)
	agg := aggregator.NewAggRing(workers, 1<<20)

	ingest.StartWorkersRing(workers, ring, agg)
	aggregator.StartAggWorkers(workers, agg)

	ev := protocol.Event{
		SessionID: 1,
		FuncID:    42,
	}

	start := time.Now()
	total := 0

	var ts uint64 = 0
	var kind uint8 = protocol.EventEnter

	for {
		ev.Kind = kind
		ev.TSNs = ts

		if ring.Push(ev) {
			total++
			ts++

			if kind == protocol.EventEnter {
				kind = protocol.EventExit
			} else {
				kind = protocol.EventEnter
			}
		}

		if total >= 50_000_000 {
			d := time.Since(start)
			mps := float64(total) / d.Seconds() / 1e6

			log.Printf("Processed 50M events in %v (%.2f M events/s)", d, mps)

			total = 0
			start = time.Now()
		}
	}
}

func RunStreamBenchmark(path string, totalEvents int) {
	log.Printf("Starting STREAM benchmark to %s ...", path)

	conn, err := net.Dial("unix", path)
	if err != nil {
		log.Fatalf("stream bench: cannot connect: %v", err)
	}
	defer conn.Close()

	batchSize := 200
	buf := make([]byte, protocol.EventSize*batchSize)

	var sessionID uint64 = 1
	var funcID uint32 = 42
	var depth uint32 = 0
	var kind uint8 = protocol.EventEnter
	var funcType uint8 = 0
	var flags uint8 = 0
	var argCount uint8 = 0
	var hasException uint8 = 0
	var jitFlag uint8 = 0

	start := time.Now()
	sent := 0
	var tsNs uint64 = 0
	var unixNs uint64 = 0

	for {
		for i := 0; i < batchSize; i++ {
			off := i * protocol.EventSize

			binary.LittleEndian.PutUint64(buf[off+0:], sessionID)
			buf[off+8] = kind
			binary.LittleEndian.PutUint64(buf[off+16:], tsNs)
			binary.LittleEndian.PutUint32(buf[off+24:], depth)
			binary.LittleEndian.PutUint32(buf[off+28:], funcID)
			buf[off+32] = funcType
			buf[off+33] = flags
			buf[off+34] = argCount
			buf[off+35] = hasException
			buf[off+36] = jitFlag
			binary.LittleEndian.PutUint64(buf[off+40:], unixNs)

			tsNs++
			unixNs++
		}

		if _, err := conn.Write(buf); err != nil {
			log.Fatalf("stream bench: write error: %v", err)
		}

		sent += batchSize
		if sent >= totalEvents {
			d := time.Since(start)
			mps := float64(sent) / d.Seconds() / 1e6
			log.Printf("STREAM: Sent %d events in %v (%.2f M events/s)", sent, d, mps)
			sent = 0
			start = time.Now()
		}
	}
}
