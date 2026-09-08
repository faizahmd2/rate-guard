package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const databasePath = "./data/rateguard.db"

func NewDB() (*sql.DB, error) {
	dir := filepath.Dir(databasePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf(
			"create sqlite directory: %w",
			err,
		)
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf(
			"open sqlite: %w",
			err,
		)
	}

	if err := db.Ping(); err != nil {
		db.Close()

		return nil, fmt.Errorf(
			"ping sqlite: %w",
			err,
		)
	}

	if err := migrate(db); err != nil {
		db.Close()

		return nil, fmt.Errorf(
			"migrate sqlite: %w",
			err,
		)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS rate_limit_rules (
			id TEXT PRIMARY KEY,
			service TEXT NOT NULL,
			resource TEXT NOT NULL,
			algorithm TEXT NOT NULL,
			config TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,

			UNIQUE(service, resource)
		);

		CREATE TABLE IF NOT EXISTS admin_users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS api_tokens (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			client TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			last_used_at DATETIME
		);
	`)

	if err != nil {
		return fmt.Errorf("create sqlite tables: %w", err)
	}

	return nil
}
