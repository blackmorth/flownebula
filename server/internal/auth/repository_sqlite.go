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
	rolesJSON, _ := json.Marshal(user.Roles)

	res, err := r.db.Exec(`
        INSERT INTO users (email, password, roles)
        VALUES (?, ?, ?)
    `, user.Email, user.Password, string(rolesJSON))
	if err != nil {
		return err
	}

	id, _ := res.LastInsertId()
	user.ID = id
	return nil
}

func (r *sqliteRepo) FindByEmail(email string) (*User, error) {
	row := r.db.QueryRow(`
        SELECT id, email, password, roles
        FROM users WHERE email = ?
    `, email)

	var u User
	var rolesJSON string

	if err := row.Scan(&u.ID, &u.Email, &u.Password, &rolesJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	json.Unmarshal([]byte(rolesJSON), &u.Roles)
	return &u, nil
}

func (r *sqliteRepo) FindByID(id int64) (*User, error) {
	row := r.db.QueryRow(`
        SELECT id, email, password, roles
        FROM users WHERE id = ?
    `, id)

	var u User
	var rolesJSON string

	if err := row.Scan(&u.ID, &u.Email, &u.Password, &rolesJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	json.Unmarshal([]byte(rolesJSON), &u.Roles)
	return &u, nil
}
