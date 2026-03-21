package metrics

type Repository interface {
	Insert(m *Metric) error
	ListBySession(sessionID int64) ([]*Metric, error)
}
