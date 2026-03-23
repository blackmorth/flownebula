package profiles

import "time"

type SessionProfile struct {
	SessionID int64     `json:"session_id"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}
