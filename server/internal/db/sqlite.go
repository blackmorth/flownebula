package db

import (
	"database/sql"
	"log"
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
        agent_id TEXT NOT NULL,
        created_at DATETIME NOT NULL,
        FOREIGN KEY(user_id) REFERENCES users(id)
    );`

	metricTable := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		cpu_usage REAL NOT NULL,
		ram_usage REAL NOT NULL,
		load_avg REAL NOT NULL,
		process_count INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY(session_id) REFERENCES sessions(id)
	);`

	sessionProfileTable := `
	CREATE TABLE IF NOT EXISTS session_profiles (
		session_id INTEGER PRIMARY KEY,
		payload TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);`

	if _, err := db.Exec(userTable); err != nil {
		log.Fatalf("failed to migrate users table: %v", err)
	}

	if _, err := db.Exec(sessionTable); err != nil {
		log.Fatalf("failed to migrate sessions table: %v", err)
	}

	if _, err := db.Exec(metricTable); err != nil {
		log.Fatalf("failed to migrate metrics table: %v", err)
	}

	if _, err := db.Exec(sessionProfileTable); err != nil {
		log.Fatalf("failed to migrate session_profiles table: %v", err)
	}
}
