package profiles

type Repository interface {
	Create(userID int64, agentID string, payload string) (*SessionProfile, error)
	Get(sessionID int64) (*SessionProfile, error)
}
