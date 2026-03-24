package sessions

import "time"
import "gorm.io/datatypes"

type Session struct {
	ID             int64          `json:"id" gorm:"primaryKey"`
	UserID         int64          `json:"user_id"`
	AgentID        string         `json:"agent_id"`
	AgentSessionID string         `json:"agent_session_id"`
	Payload        datatypes.JSON `json:"payload"`
	CreatedAt      time.Time      `json:"created_at"`
}
