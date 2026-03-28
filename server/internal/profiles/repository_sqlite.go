package profiles

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

func (r *sqliteRepo) Create(userID int64, agentID string, payload string) (*SessionProfile, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	res, err := tx.Exec(
		"INSERT INTO sessions (user_id, agent_id, payload, created_at) VALUES (?, ?, ?, ?)",
		userID, agentID, "{}", now,
	)
	if err != nil {
		return nil, err
	}

	sessionID, _ := res.LastInsertId()

	_, err = tx.Exec(
		"INSERT INTO session_profiles (session_id, payload) VALUES (?, ?)",
		sessionID, payload,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &SessionProfile{
		SessionID: sessionID,
		Payload:   payload,
		CreatedAt: now,
	}, nil
}

func (r *sqliteRepo) Get(sessionID int64) (*SessionProfile, error) {
	row := r.db.QueryRow(
		"SELECT session_id, payload, created_at FROM session_profiles WHERE session_id = ?",
		sessionID,
	)

	var p SessionProfile
	if err := row.Scan(&p.SessionID, &p.Payload, &p.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("profile not found")
		}
		return nil, err
	}

	return &p, nil
}
