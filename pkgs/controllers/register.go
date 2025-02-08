package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"joshuamURD/go-auth-api/pkgs/models"

	"github.com/google/uuid"
)

// registerRequest is a struct that contains the email and password of the user
type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles the registration of a new user
func (rc *Controller) Register(w http.ResponseWriter, r *http.Request) {
	//Ensures that the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Check if already logged in
	_, err := r.Cookie("refresh_token")
	if err == nil {
		http.Error(w, "Already logged in", http.StatusUnauthorized)
		return
	}

	//Decodes the request body into a registerRequest struct and checks for errors
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//Checks if the email and password are empty and returns an error if they are
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	//Hashes the password
	hashedPassword, err := rc.hasher.Hash(req.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	//Creates a new user with the email and hashed password
	user := models.User{
		ID:             uuid.New(),
		Email:          req.Email,
		HashedPassword: hashedPassword,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Verified:       false,
		FailedAttempts: 0,
		Locked:         false,
	}

	//Creates the user in the database
	if _, err := (*rc.db).Create(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Get auth response with access token
	authResp, err := rc.auth.Authenticate(r.Context(), user.ID.String(), w)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	//Writes a success message to the response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "User registered successfully",
		"access_token": authResp.AccessToken,
	})
}
