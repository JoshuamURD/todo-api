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

func (lc *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := (*lc.db).GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if !lc.hasher.Compare(user.HashedPassword, req.Password) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(loginResponse{Token: "token"})
}
