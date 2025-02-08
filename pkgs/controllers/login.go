package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// loginRequest is a representation of a valid request to the login route
// a login request contains an email and password
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse is a representation of a valid response to the login route
type loginResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
}

// Login handles the login of a user
// it verifies that the password hash matches the password provided for a given email
// if it matches, it generates a refresh token as as http only cookie
// it also provides an access token in the response
func (lc *Controller) Login(w http.ResponseWriter, r *http.Request) {
	//Checks if the request method is GET
	if r.Method == http.MethodGet {

		//Gets the refresh token from the request
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			http.Error(w, "No refresh token provided", http.StatusBadRequest)
			return
		}

		//Refreshes the refresh token
		authResp, err := lc.auth.RefreshAuth(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
			return
		}

		loginResp := loginResponse{
			Message:     "Token refreshed",
			AccessToken: authResp.AccessToken,
		}

		//Sets the access token in the response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(loginResp)
		return
	} else if r.Method != http.MethodPost {
		//Checks if the request method is POST
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Decodes the request body into a loginRequest
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if _, err := r.Cookie("refresh_token"); err == nil {
		http.Error(w, "Already logged in", http.StatusBadRequest)
		return
	}

	//Gets the user from the database
	user, err := (*lc.db).GetByEmail(req.Email)
	if err != nil {
		// Log the actual error for debugging
		log.Printf("Login error for email %s: %v", req.Email, err)

		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	//Checks if the password hash matches the password provided
	if !lc.hasher.Compare(user.HashedPassword, req.Password) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	//Gets the auth response with access token and refresh token
	authResp, err := lc.auth.Authenticate(r.Context(), user.ID.String(), w)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	//Creates a login response with the access token
	loginResp := loginResponse{
		Message:     "Login successful",
		AccessToken: authResp.AccessToken,
	}

	//Sets the access token in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResp)
}
