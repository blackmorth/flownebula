package metrics

import "time"

type Metric struct {
	ID           int64     `json:"id"`
	SessionID    int64     `json:"session_id"`
	CPUUsage     float64   `json:"cpu_usage"`
	RAMUsage     float64   `json:"ram_usage"`
	LoadAvg      float64   `json:"load_avg"`
	ProcessCount int       `json:"process_count"`
	CreatedAt    time.Time `json:"created_at"`
}
