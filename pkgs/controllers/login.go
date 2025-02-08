package controllers

import (
	"encoding/json"
	"net/http"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
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
