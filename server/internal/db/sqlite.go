package db

import (
	"database/sql"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func Open(path string) *sql.DB {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("failed to open sqlite database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping sqlite database: %v", err)
	}

	return db
}

func Migrate(db *sql.DB) {
	userTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        agent_token TEXT UNIQUE NOT NULL,
        agent_enabled BOOLEAN DEFAULT FALSE,
        roles TEXT DEFAULT '["ROLE_USER"]'
    );`

	sessionTable := `
    CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		agent_id TEXT,
		agent_session_id TEXT,
		service TEXT,
		endpoint TEXT,
		release TEXT,
		tags TEXT,
		payload TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_created_at ON sessions(user_id, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_agent_session_id ON sessions(agent_session_id);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_service_endpoint ON sessions(service, endpoint);`,
	}

	if _, err := db.Exec(userTable); err != nil {
		log.Fatalf("failed to migrate users table: %v", err)
	}

	if _, err := db.Exec(sessionTable); err != nil {
		log.Fatalf("failed to migrate sessions table: %v", err)
	}

	alterStatements := []string{
		`ALTER TABLE sessions ADD COLUMN agent_id TEXT`,
		`ALTER TABLE sessions ADD COLUMN agent_session_id TEXT`,
		`ALTER TABLE sessions ADD COLUMN service TEXT`,
		`ALTER TABLE sessions ADD COLUMN endpoint TEXT`,
		`ALTER TABLE sessions ADD COLUMN release TEXT`,
		`ALTER TABLE sessions ADD COLUMN tags TEXT`,
		`ALTER TABLE sessions ADD COLUMN payload TEXT NOT NULL DEFAULT '{}'`,
	}
	for _, stmt := range alterStatements {
		if _, err := db.Exec(stmt); err != nil {
			// Ignore duplicate column errors for existing databases.
			// modernc/sqlite returns generic errors as text.
			if !strings.Contains(err.Error(), "duplicate column name") {
				log.Fatalf("failed to alter sessions table: %v", err)
			}
		}
	}

	if _, err := db.Exec(`UPDATE sessions SET payload = '{}' WHERE payload IS NULL`); err != nil {
		log.Fatalf("failed to backfill sessions payload: %v", err)
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			log.Fatalf("failed to create sessions index: %v", err)
		}
	}
}

func ApplyRetention(db *sql.DB, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		return 0, nil
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	res, err := db.Exec(`DELETE FROM sessions WHERE created_at < ?`, cutoff)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}
