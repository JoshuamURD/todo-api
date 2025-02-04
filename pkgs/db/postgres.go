package db

import "database/sql"

// PostgresTableCreator implements TableCreator for PostgreSQL
type PostgresTableCreator struct{}

func (p PostgresTableCreator) CreateTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY,
        email TEXT NOT NULL UNIQUE,
        verified BOOLEAN NOT NULL,
        failed_attempts INTEGER NOT NULL,
        locked BOOLEAN NOT NULL,
        hashed_password TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
    );`
	_, err := db.Exec(query)
	return err
}
