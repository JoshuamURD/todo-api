package controllers

import (
	"encoding/json"
	"net/http"
)

// loginRequest is a representation of a valid request to the login route
// a login request contains an email and password
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse is a representation of a valid response to the login route
type loginResponse struct {
	message     string `json:"message"`
	accessToken string `json:"access_token"`
}

// Login handles the login of a user
// it verifies that the password hash matches the password provided for a given email
// if it matches, it generates a refresh token as as http only cookie
// it also provides an access token in the response
func (lc *Controller) Login(w http.ResponseWriter, r *http.Request) {
	//Ensures that the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	println(r.Cookie("refresh_token"))

	user, err := (*c.db).GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if !c.hasher.Compare(user.HashedPassword, req.Password) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Get auth response with access token
	authResp, err := c.auth.Authenticate(r.Context(), user.ID.String(), w)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Return access token in response body
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResp)
}
