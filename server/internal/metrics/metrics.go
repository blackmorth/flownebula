package metrics

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

type counter struct{ v atomic.Uint64 }

func (c *counter) Add(n uint64) { c.v.Add(n) }
func (c *counter) Load() uint64 { return c.v.Load() }

type ServerMetricsRegistry struct {
	HTTPRequestsTotal      counter
	HTTPUnauthorizedTotal  counter
	AgentUploadsTotal      counter
	AgentUploadErrorsTotal counter
	ProtocolRejectsTotal   counter
}

var ServerMetrics = &ServerMetricsRegistry{}

func StartMetricsServer(addr string) {
	if strings.TrimSpace(addr) == "" {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# TYPE nebula_server_http_requests_total counter\nnebula_server_http_requests_total %d\n", ServerMetrics.HTTPRequestsTotal.Load())
		fmt.Fprintf(w, "# TYPE nebula_server_http_unauthorized_total counter\nnebula_server_http_unauthorized_total %d\n", ServerMetrics.HTTPUnauthorizedTotal.Load())
		fmt.Fprintf(w, "# TYPE nebula_server_agent_uploads_total counter\nnebula_server_agent_uploads_total %d\n", ServerMetrics.AgentUploadsTotal.Load())
		fmt.Fprintf(w, "# TYPE nebula_server_agent_upload_errors_total counter\nnebula_server_agent_upload_errors_total %d\n", ServerMetrics.AgentUploadErrorsTotal.Load())
		fmt.Fprintf(w, "# TYPE nebula_server_protocol_rejects_total counter\nnebula_server_protocol_rejects_total %d\n", ServerMetrics.ProtocolRejectsTotal.Load())
	})
	go func() {
		_ = http.ListenAndServe(addr, mux)
	}()
}
