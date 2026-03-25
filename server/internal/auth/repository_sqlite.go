package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
)

type sqliteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) UserRepository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) Create(user *User) error {
	var usersCount int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&usersCount); err != nil {
		return err
	}
	if usersCount == 0 {
		user.Roles = []string{"ROLE_ADMIN", "ROLE_USER"}
	}

	rolesJSON, _ := json.Marshal(user.Roles)

	res, err := r.db.Exec(`
        INSERT INTO users (email, password, roles, agent_token, agent_enabled)
        VALUES (?, ?, ?, ?, ?)
    `, user.Email, user.Password, string(rolesJSON), user.AgentToken, user.AgentEnabled)

	if err != nil {
		return err
	}

	id, _ := res.LastInsertId()
	user.ID = id
	return nil
}

func (r *sqliteRepo) FindByEmail(email string) (*User, error) {
	row := r.db.QueryRow(`
        SELECT id, email, password, roles, agent_token, agent_enabled
        FROM users WHERE email = ?
    `, email)

	var u User
	var rolesJSON string

	if err := row.Scan(&u.ID, &u.Email, &u.Password, &rolesJSON, &u.AgentToken, &u.AgentEnabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	json.Unmarshal([]byte(rolesJSON), &u.Roles)
	return &u, nil
}

func (r *sqliteRepo) FindByAgentToken(token string) (*User, error) {
	row := r.db.QueryRow(`
        SELECT id, email, password, roles, agent_token, agent_enabled
        FROM users WHERE agent_token = ? AND agent_enabled = TRUE
    `, token)

	var u User
	var rolesJSON string

	if err := row.Scan(&u.ID, &u.Email, &u.Password, &rolesJSON, &u.AgentToken, &u.AgentEnabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found or agent not enabled")
		}
		return nil, err
	}

	json.Unmarshal([]byte(rolesJSON), &u.Roles)
	return &u, nil
}

func (r *sqliteRepo) FindByID(id int64) (*User, error) {
	row := r.db.QueryRow(`
        SELECT id, email, password, roles, agent_token, agent_enabled
        FROM users WHERE id = ?
    `, id)

	var u User
	var rolesJSON string

	if err := row.Scan(&u.ID, &u.Email, &u.Password, &rolesJSON, &u.AgentToken, &u.AgentEnabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	json.Unmarshal([]byte(rolesJSON), &u.Roles)
	return &u, nil
}

func (r *sqliteRepo) FindAll() ([]*User, error) {
	rows, err := r.db.Query(`
        SELECT id, email, roles, agent_enabled
        FROM users ORDER BY id ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		var u User
		var rolesJSON string

		if err := rows.Scan(&u.ID, &u.Email, &rolesJSON, &u.AgentEnabled); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(rolesJSON), &u.Roles)
		users = append(users, &u)
	}

	return users, nil
}

func (r *sqliteRepo) UpdateAgentEnabled(id int64, enabled bool) error {
	_, err := r.db.Exec(`
        UPDATE users SET agent_enabled = ? WHERE id = ?
    `, enabled, id)
	return err
}

func (r *sqliteRepo) UpdateAgentToken(id int64, token string) error {
	_, err := r.db.Exec(`
        UPDATE users SET agent_token = ? WHERE id = ?
    `, token, id)
	return err
}

func (r *sqliteRepo) UpdateRoles(id int64, roles []string) error {
	rolesJSON, _ := json.Marshal(roles)
	_, err := r.db.Exec(`
        UPDATE users SET roles = ? WHERE id = ?
    `, string(rolesJSON), id)
	return err
}
