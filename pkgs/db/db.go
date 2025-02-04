package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// DB is a wrapper around the sql.DB type.
type DB struct {
	db *sql.DB
}

type Database interface {
	GetAll() ([]Client, error)
	Add(Client) (int, error)
}

func NewDB(path string) *DB {

	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	createTable(db)
	defer db.Close()

	return &DB{db: db}
}

func createTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		phone TEXT NOT NULL
	);`
	if _, err := db.Exec(query); err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

// getItems retrieves all items from the database.
func getClients() ([]Client, error) {
	rows, err := db.Query("SELECT id, name, email, phone FROM clients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var client Client
		if err := rows.Scan(&client.ID, &client.Name, &client.Email, &client.Phone); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, nil
}

// addItem inserts a new item into the database.
func addClient(name string) (int, error) {
	result, err := db.Exec("INSERT INTO clients (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}
