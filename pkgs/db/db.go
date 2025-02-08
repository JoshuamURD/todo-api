package db

import (
	"database/sql"
	"fmt"
	"joshuamURD/go-auth-api/pkgs/models"
	"log"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// SQLiteRepository is a wrapper around the sql.DB type.
type SQLiteRepository struct {
	db *sql.DB
}

// Database is an interface that defines the methods for the SQLiteRepository.
type Database interface {
	GetAll() ([]models.User, error)
	Create(models.User) (int, error)
	GetByEmail(string) (models.User, error)
	GetByID(uuid.UUID) (models.User, error)
}

// TableCreator defines the interface for table creation
type TableCreator interface {
	CreateTable(db *sql.DB) error
}

// SQLiteTableCreator implements TableCreator for SQLite
type SQLiteTableCreator struct{}

func (s SQLiteTableCreator) CreateTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		verified BOOLEAN NOT NULL,
		failed_attempts INTEGER NOT NULL,
		locked BOOLEAN NOT NULL,
		hashed_password TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

// NewSQLiteRepository creates a new SQLiteRepository.
func NewSQLiteRepository(path string, creator TableCreator) *SQLiteRepository {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)                 // Limit max open connections
	db.SetMaxIdleConns(25)                 // Set max idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Set max lifetime for connections

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	if err := creator.CreateTable(db); err != nil {
		log.Fatal("Failed to create table:", err)
	}

	repo := &SQLiteRepository{db: db}

	// Migrate timestamps to RFC3339 format
	if err := repo.MigrateTimestamps(); err != nil {
		log.Printf("Warning: Failed to migrate timestamps: %v", err)
	}

	return repo
}

// Close closes the database connection
func (d *SQLiteRepository) Close() error {
	return d.db.Close()
}

// getItems retrieves all items from the database.
func (d *SQLiteRepository) GetAll() ([]models.User, error) {
	rows, err := d.db.Query("SELECT id, email, verified, failed_attempts, locked, hashed_password, created_at, updated_at FROM users")
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Verified,
			&user.FailedAttempts,
			&user.Locked,
			&user.HashedPassword,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Parse the time strings
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing created_at time: %w", err)
		}
		user.CreatedAt = createdAt

		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing updated_at time: %w", err)
		}
		user.UpdatedAt = updatedAt

		users = append(users, user)
	}

	return users, nil
}

// addItem inserts a new item into the database.
func (d *SQLiteRepository) Create(user models.User) (int, error) {
	// Format the timestamps in RFC3339 format
	createdAt := user.CreatedAt.Format(time.RFC3339)
	updatedAt := user.UpdatedAt.Format(time.RFC3339)

	result, err := d.db.Exec(
		"INSERT INTO users (id, email, verified, failed_attempts, locked, hashed_password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		user.ID,
		user.Email,
		user.Verified,
		user.FailedAttempts,
		user.Locked,
		user.HashedPassword,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("error creating user: %w", err)
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (d *SQLiteRepository) GetByEmail(email string) (models.User, error) {
	var user models.User
	var createdAtStr, updatedAtStr string

	row := d.db.QueryRow("SELECT id, email, verified, failed_attempts, locked, hashed_password, created_at, updated_at FROM users WHERE email = ?", email)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Verified,
		&user.FailedAttempts,
		&user.Locked,
		&user.HashedPassword,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return user, fmt.Errorf("user not found with email: %s", email)
	}
	if err != nil {
		return user, fmt.Errorf("database error: %w", err)
	}

	// Try parsing with RFC3339 first
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		// If that fails, try parsing the current format
		createdAt, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAtStr)
		if err != nil {
			return user, fmt.Errorf("error parsing created_at time: %w", err)
		}
	}
	user.CreatedAt = createdAt

	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		// If that fails, try parsing the current format
		updatedAt, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", updatedAtStr)
		if err != nil {
			return user, fmt.Errorf("error parsing updated_at time: %w", err)
		}
	}
	user.UpdatedAt = updatedAt

	return user, nil
}

func (d *SQLiteRepository) GetByID(id uuid.UUID) (models.User, error) {
	var user models.User
	var createdAtStr, updatedAtStr string

	row := d.db.QueryRow("SELECT id, email, verified, failed_attempts, locked, hashed_password, created_at, updated_at FROM users WHERE id = ?", id)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Verified,
		&user.FailedAttempts,
		&user.Locked,
		&user.HashedPassword,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return user, fmt.Errorf("user not found with id: %s", id)
	}
	if err != nil {
		return user, fmt.Errorf("database error: %w", err)
	}

	// Parse the time strings
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return user, fmt.Errorf("error parsing created_at time: %w", err)
	}
	user.CreatedAt = createdAt

	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return user, fmt.Errorf("error parsing updated_at time: %w", err)
	}
	user.UpdatedAt = updatedAt

	return user, nil
}

// MigrateTimestamps updates all existing timestamps to RFC3339 format
func (d *SQLiteRepository) MigrateTimestamps() error {
	rows, err := d.db.Query("SELECT id, created_at, updated_at FROM users")
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&id, &createdAtStr, &updatedAtStr); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse and reformat created_at
		createdAt, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAtStr)
		if err == nil {
			createdAtStr = createdAt.Format(time.RFC3339)
		}

		// Parse and reformat updated_at
		updatedAt, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", updatedAtStr)
		if err == nil {
			updatedAtStr = updatedAt.Format(time.RFC3339)
		}

		// Update the record
		_, err = d.db.Exec(
			"UPDATE users SET created_at = ?, updated_at = ? WHERE id = ?",
			createdAtStr,
			updatedAtStr,
			id,
		)
		if err != nil {
			return fmt.Errorf("failed to update timestamps for user %s: %w", id, err)
		}
	}

	return nil
}
