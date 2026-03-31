package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricEventsIngested = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nebula_agent_events_ingested_total",
			Help: "Total number of events ingested from the socket",
		},
	)

	metricEventsDropped = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nebula_agent_events_dropped_total",
			Help: "Total number of events dropped due to full ring",
		},
	)

	MetricSessionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "nebula_agent_sessions_active",
			Help: "Number of active sessions in memory",
		},
	)
)

func init() {
	prometheus.MustRegister(
		metricEventsIngested,
		metricEventsDropped,
		MetricSessionsActive,
	)
}

// StartMetricsServer démarre un serveur HTTP Prometheus sur addr (ex: ":9108").
// Si addr est vide, il ne fait rien.
func StartMetricsServer(addr string) {
	if addr == "" {
		return
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(addr, nil)
	}()
}
