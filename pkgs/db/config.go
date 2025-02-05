package db

import (
	"sync"
)

var (
	instance *SQLiteRepository
	once     sync.Once
)

// Config holds database configuration
type Config struct {
	Path string
}

// Initialize sets up the database connection with the given configuration
func Initialize(config Config) error {
	var err error
	once.Do(func() {
		creator := SQLiteTableCreator{}
		instance = NewSQLiteRepository(config.Path, creator)
	})
	return err
}

// GetInstance returns the database instance
func GetInstance() Database {
	if instance == nil {
		panic("Database not initialized. Call Initialize first")
	}
	return instance
}

// Close closes the database connection
func Close() error {
	if instance != nil {
		return instance.db.Close()
	}
	return nil
}
