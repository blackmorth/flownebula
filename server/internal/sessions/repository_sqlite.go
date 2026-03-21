package sessions

import (
	"database/sql"
	"errors"
	"time"
)

type sqliteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) Repository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) Create(userID int64, agentID string) (*Session, error) {
	now := time.Now()

	res, err := r.db.Exec(
		"INSERT INTO sessions (user_id, agent_id, created_at) VALUES (?, ?, ?)",
		userID, agentID, now,
	)
	if err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()

	return &Session{
		ID:        id,
		UserID:    userID,
		AgentID:   agentID,
		CreatedAt: now,
	}, nil
}

func (r *sqliteRepo) List() ([]*Session, error) {
	rows, err := r.db.Query("SELECT id, user_id, agent_id, created_at FROM sessions ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.UserID, &s.AgentID, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}

	return out, nil
}

func (r *sqliteRepo) Get(id int64) (*Session, error) {
	row := r.db.QueryRow(
		"SELECT id, user_id, agent_id, created_at FROM sessions WHERE id = ?",
		id,
	)

	var s Session
	if err := row.Scan(&s.ID, &s.UserID, &s.AgentID, &s.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return &s, nil
}
