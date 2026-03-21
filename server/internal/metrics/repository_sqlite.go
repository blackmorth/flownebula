package metrics

import (
	"database/sql"
	"time"
)

type sqliteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) Repository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) Insert(m *Metric) error {
	now := time.Now()

	res, err := r.db.Exec(
		`INSERT INTO metrics (session_id, cpu_usage, ram_usage, load_avg, process_count, created_at)
         VALUES (?, ?, ?, ?, ?, ?)`,
		m.SessionID, m.CPUUsage, m.RAMUsage, m.LoadAvg, m.ProcessCount, now,
	)
	if err != nil {
		return err
	}

	id, _ := res.LastInsertId()
	m.ID = id
	m.CreatedAt = now

	return nil
}

func (r *sqliteRepo) ListBySession(sessionID int64) ([]*Metric, error) {
	rows, err := r.db.Query(
		`SELECT id, session_id, cpu_usage, ram_usage, load_avg, process_count, created_at
         FROM metrics WHERE session_id = ? ORDER BY id DESC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Metric
	for rows.Next() {
		var m Metric
		if err := rows.Scan(&m.ID, &m.SessionID, &m.CPUUsage, &m.RAMUsage, &m.LoadAvg, &m.ProcessCount, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &m)
	}

	return out, nil
}
