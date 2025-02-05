package controllers

import (
	"encoding/json"
	"net/http"

	"joshuamURD/go-auth-api/pkgs/db"
	"joshuamURD/go-auth-api/pkgs/hash"
	"joshuamURD/go-auth-api/pkgs/models"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterController struct {
	hasher hash.Hasher
	db     *db.Database
}

func NewRegisterController(hasher hash.Hasher, db *db.Database) *RegisterController {
	return &RegisterController{
		hasher: hasher,
		db:     db,
	}
}

// Register handles the registration of a new user
func (rc *RegisterController) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := rc.hasher.Hash(req.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	if err := rc.db.Create(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
	})
}
