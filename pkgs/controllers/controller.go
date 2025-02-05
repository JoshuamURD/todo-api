package controllers

import (
	"joshuamURD/go-auth-api/pkgs/db"
	"joshuamURD/go-auth-api/pkgs/hash"
)

type Controller struct {
	hasher hash.Hasher
	db     *db.Database
}

// NewRegisterController creates a new RegisterController
// It takes a hasher and a database instance and returns a pointer to a RegisterController
func NewController(hasher hash.Hasher, db *db.Database) *Controller {
	return &Controller{
		hasher: hasher,
		db:     db,
	}
}
