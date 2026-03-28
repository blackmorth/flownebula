package internal

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type counter struct{ v atomic.Uint64 }

func (c *counter) Add(n uint64) { c.v.Add(n) }
func (c *counter) Load() uint64 { return c.v.Load() }

type gauge struct{ v atomic.Int64 }

func (g *gauge) Set(v int64) { g.v.Store(v) }
func (g *gauge) Load() int64 { return g.v.Load() }

// Histogram buckets in milliseconds.
var defaultLatencyBuckets = []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2000, 5000}

type histogram struct {
	mu      sync.Mutex
	buckets []float64
	counts  []uint64
	sumMs   float64
	total   uint64
}

func newHistogram(buckets []float64) *histogram {
	cp := append([]float64(nil), buckets...)
	sort.Float64s(cp)
	return &histogram{buckets: cp, counts: make([]uint64, len(cp)+1)}
}

func (h *histogram) Observe(d time.Duration) {
	ms := float64(d.Milliseconds())
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sumMs += ms
	h.total++
	for i, b := range h.buckets {
		if ms <= b {
			h.counts[i]++
			return
		}
	}
	h.counts[len(h.counts)-1]++
}

type AgentMetricsRegistry struct {
	IngestReceived     counter
	IngestAccepted     counter
	ValidateProcessed  counter
	ValidationErrors   counter
	AggregateProcessed counter
	ProbeEventPackets  counter
	ProbeNamePackets   counter
	ProbeSessionEnds   counter
	Drops              counter
	QueueDepth         gauge
	FlushLatency       *histogram
	WALQueued          counter
	WALReplayed        counter
	RetryAttempts      counter
	SendFailures       counter
}

var AgentMetrics = &AgentMetricsRegistry{FlushLatency: newHistogram(defaultLatencyBuckets)}

func StartMetricsServer(addr string) {
	if strings.TrimSpace(addr) == "" {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP nebula_ingest_rate_total Total packets received by ingest.\n")
		fmt.Fprintf(w, "# TYPE nebula_ingest_rate_total counter\nnebula_ingest_rate_total %d\n", AgentMetrics.IngestReceived.Load())
		fmt.Fprintf(w, "# HELP nebula_ingest_accepted_total Total packets accepted into queue.\n")
		fmt.Fprintf(w, "# TYPE nebula_ingest_accepted_total counter\nnebula_ingest_accepted_total %d\n", AgentMetrics.IngestAccepted.Load())
		fmt.Fprintf(w, "# HELP nebula_validation_processed_total Total packets processed by validation stage.\n")
		fmt.Fprintf(w, "# TYPE nebula_validation_processed_total counter\nnebula_validation_processed_total %d\n", AgentMetrics.ValidateProcessed.Load())
		fmt.Fprintf(w, "# HELP nebula_validation_errors_total Total packets rejected by validation.\n")
		fmt.Fprintf(w, "# TYPE nebula_validation_errors_total counter\nnebula_validation_errors_total %d\n", AgentMetrics.ValidationErrors.Load())
		fmt.Fprintf(w, "# HELP nebula_aggregate_processed_total Total events aggregated into sessions.\n")
		fmt.Fprintf(w, "# TYPE nebula_aggregate_processed_total counter\nnebula_aggregate_processed_total %d\n", AgentMetrics.AggregateProcessed.Load())
		fmt.Fprintf(w, "# TYPE nebula_probe_event_packets_total counter\nnebula_probe_event_packets_total %d\n", AgentMetrics.ProbeEventPackets.Load())
		fmt.Fprintf(w, "# TYPE nebula_probe_name_packets_total counter\nnebula_probe_name_packets_total %d\n", AgentMetrics.ProbeNamePackets.Load())
		fmt.Fprintf(w, "# TYPE nebula_probe_session_end_total counter\nnebula_probe_session_end_total %d\n", AgentMetrics.ProbeSessionEnds.Load())
		fmt.Fprintf(w, "# HELP nebula_drops_total Total packets dropped by backpressure policy.\n")
		fmt.Fprintf(w, "# TYPE nebula_drops_total counter\nnebula_drops_total %d\n", AgentMetrics.Drops.Load())
		fmt.Fprintf(w, "# HELP nebula_queue_depth Current ingest queue depth.\n")
		fmt.Fprintf(w, "# TYPE nebula_queue_depth gauge\nnebula_queue_depth %d\n", AgentMetrics.QueueDepth.Load())
		fmt.Fprintf(w, "# HELP nebula_wal_queued_total Sessions queued to WAL.\n")
		fmt.Fprintf(w, "# TYPE nebula_wal_queued_total counter\nnebula_wal_queued_total %d\n", AgentMetrics.WALQueued.Load())
		fmt.Fprintf(w, "# HELP nebula_wal_replayed_total Sessions replayed from WAL.\n")
		fmt.Fprintf(w, "# TYPE nebula_wal_replayed_total counter\nnebula_wal_replayed_total %d\n", AgentMetrics.WALReplayed.Load())
		fmt.Fprintf(w, "# HELP nebula_retry_attempts_total Total retry attempts for session upload.\n")
		fmt.Fprintf(w, "# TYPE nebula_retry_attempts_total counter\nnebula_retry_attempts_total %d\n", AgentMetrics.RetryAttempts.Load())
		fmt.Fprintf(w, "# HELP nebula_send_failures_total Total final send failures.\n")
		fmt.Fprintf(w, "# TYPE nebula_send_failures_total counter\nnebula_send_failures_total %d\n", AgentMetrics.SendFailures.Load())

		AgentMetrics.FlushLatency.mu.Lock()
		defer AgentMetrics.FlushLatency.mu.Unlock()
		fmt.Fprintf(w, "# HELP nebula_flush_latency_ms Session flush latency in ms.\n")
		fmt.Fprintf(w, "# TYPE nebula_flush_latency_ms histogram\n")
		cum := uint64(0)
		for i, b := range AgentMetrics.FlushLatency.buckets {
			cum += AgentMetrics.FlushLatency.counts[i]
			fmt.Fprintf(w, "nebula_flush_latency_ms_bucket{le=\"%g\"} %d\n", b, cum)
		}
		cum += AgentMetrics.FlushLatency.counts[len(AgentMetrics.FlushLatency.counts)-1]
		fmt.Fprintf(w, "nebula_flush_latency_ms_bucket{le=\"+Inf\"} %d\n", cum)
		fmt.Fprintf(w, "nebula_flush_latency_ms_sum %g\n", AgentMetrics.FlushLatency.sumMs)
		fmt.Fprintf(w, "nebula_flush_latency_ms_count %d\n", AgentMetrics.FlushLatency.total)
	})

	go func() {
		_ = http.ListenAndServe(addr, mux)
	}()
}
