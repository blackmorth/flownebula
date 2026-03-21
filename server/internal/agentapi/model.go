package agentapi

type HeartbeatRequest struct {
	AgentID string `json:"agent_id"`
	Version string `json:"version"`
}

type HeartbeatResponse struct {
	Status        string `json:"status"`
	SessionID     int64  `json:"session_id"`
	CheckInterval int    `json:"check_interval"`
}
