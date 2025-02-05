package db

import (
	"database/sql"
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

	return &SQLiteRepository{db: db}
}

// Close closes the database connection
func (d *SQLiteRepository) Close() error {
	return d.db.Close()
}

// getItems retrieves all items from the database.
func (d *SQLiteRepository) GetAll() ([]models.User, error) {
	var users []models.User
	rows, err := d.db.Query("SELECT id, email, verified, failed_attempts, locked, hashed_password, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Verified, &user.FailedAttempts, &user.Locked, &user.HashedPassword, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// addItem inserts a new item into the database.
func (d *SQLiteRepository) Create(user models.User) (int, error) {
	result, err := d.db.Exec("INSERT INTO users (id, email, verified, failed_attempts, locked, hashed_password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", user.ID, user.Email, user.Verified, user.FailedAttempts, user.Locked, user.HashedPassword, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (d *SQLiteRepository) GetByEmail(email string) (models.User, error) {
	var user models.User
	row := d.db.QueryRow("SELECT id, email, verified, failed_attempts, locked, hashed_password, created_at, updated_at FROM users WHERE email = ?", email)
	err := row.Scan(&user.ID, &user.Email, &user.Verified, &user.FailedAttempts, &user.Locked, &user.HashedPassword, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (d *SQLiteRepository) GetByID(id uuid.UUID) (models.User, error) {
	var user models.User
	row := d.db.QueryRow("SELECT id, email, verified, failed_attempts, locked, hashed_password, created_at, updated_at FROM users WHERE id = ?", id)
	err := row.Scan(&user.ID, &user.Email, &user.Verified, &user.FailedAttempts, &user.Locked, &user.HashedPassword, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}
