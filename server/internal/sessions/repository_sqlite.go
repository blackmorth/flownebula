package sessions

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/datatypes"
)

type sqliteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) Repository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) Create(session *Session) error {
	now := time.Now()
	session.CreatedAt = now

	res, err := r.db.Exec(
		`INSERT INTO sessions (user_id, agent_id, agent_session_id, payload, created_at)
         VALUES (?, ?, ?, ?, ?)`,
		session.UserID,
		session.AgentID,
		session.AgentSessionID,
		string(session.Payload), // datatypes.JSON → string
		session.CreatedAt,
	)
	if err != nil {
		return err
	}

	id, _ := res.LastInsertId()
	session.ID = id
	return nil
}

func (r *sqliteRepo) List() ([]*Session, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, agent_id, agent_session_id, payload, created_at
         FROM sessions ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Session
	for rows.Next() {
		var s Session
		var payloadStr string

		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.AgentID,
			&s.AgentSessionID,
			&payloadStr,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}

		s.Payload = datatypes.JSON([]byte(payloadStr))
		out = append(out, &s)
	}

	return out, nil
}

func (r *sqliteRepo) Get(id int64) (*Session, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, agent_id, agent_session_id, payload, created_at
         FROM sessions WHERE id = ?`,
		id,
	)

	var s Session
	var payloadStr string

	if err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.AgentID,
		&s.AgentSessionID,
		&payloadStr,
		&s.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	s.Payload = datatypes.JSON([]byte(payloadStr))
	return &s, nil
}
