package main

import (
	"flag"
	"flownebula/agent/internal/aggregator"
	"flownebula/agent/internal/export"
	"flownebula/agent/internal/ingest"
	"flownebula/agent/internal/metrics"
	"flownebula/agent/internal/server"
	"flownebula/agent/internal/tests"
)

func main() {
	benchFlag := flag.Bool("bench", false, "Run internal throughput benchmark")
	benchSocketFlag := flag.Bool("bench-socket", false, "Run internal socket throughput benchmark")
	workers := flag.Int("workers", 8, "Number of worker goroutines")
	sockPath := flag.String("sock", "/var/run/nebula.sock", "Unix socket path")
	metricsAddr := flag.String("metrics-addr", ":9108", "Prometheus metrics address")
	serverURL := flag.String("server-url", "", "Backend server URL for exporting sessions")
	agentToken := flag.String("agent-token", "", "Authentication token for the backend")

	flag.Parse()

	// Mode benchmark interne
	if *benchFlag {
		tests.RunBenchmark(*workers)
		return
	}
	if *benchSocketFlag {
		tests.RunStreamBenchmark("/var/run/nebula.sock", 50_000_000)
		return
	}

	// Metrics
	metrics.StartMetricsServer(*metricsAddr)

	// Exporter backend
	if *serverURL != "" {
		export.SetServerURL(*serverURL)
	}
	if *agentToken != "" {
		export.SetAgentToken(*agentToken)
	}

	ring := ingest.NewRing(*workers, 1<<20)
	agg := aggregator.NewAggRing(*workers, 1<<20)

	go server.StartUnixStreamServer(*sockPath, ring, *workers)

	ingest.StartWorkersRing(*workers, ring, agg)
	aggregator.StartAggWorkers(*workers, agg)

	select {}
}
