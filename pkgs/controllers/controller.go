package controllers

import (
	"joshuamURD/go-auth-api/pkgs/auth"
	"joshuamURD/go-auth-api/pkgs/db"
	"joshuamURD/go-auth-api/pkgs/hash"
)

// Controller is a struct that contains the hasher, database, and middleware
// a hasher is used to hash the password
// a database is used to store the user data
// an auth service is used to authenticate the user
type Controller struct {
	hasher hash.Hasher
	db     *db.Database
	auth   auth.AuthService
}

// NewRegisterController creates a new RegisterController
// It takes a hasher and a database instance and returns a pointer to a RegisterController
func NewController(hasher hash.Hasher, db *db.Database, auth auth.AuthService) *Controller {
	return &Controller{
		hasher: hasher,
		db:     db,
		auth:   auth,
	}
}
