package sessions

import "time"

type Session struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	AgentID   string    `json:"agent_id"`
	CreatedAt time.Time `json:"created_at"`
}
