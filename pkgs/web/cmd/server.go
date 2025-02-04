package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite" // Import with blank identifier to register the driver
)

// CSRF token settings.
const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
)

// Item represents a record in our simple table.
type Client struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func main() {

	// Setup routes.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/clients", clientsHandler)
	// Serve static files (the React production build)
	mux.Handle("/", http.FileServer(http.Dir("./build")))

	// Wrap the multiplexer with our CSRF middleware.
	csrfHandler := csrfMiddleware(mux)

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", csrfHandler))
}

// clientsHandler handles GET and POST requests for clients.
func clientsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		clients, err := getClients()
		if err != nil {
			http.Error(w, "Failed to get clients", http.StatusInternalServerError)
			return
		}
		jsonResponse(w, clients)
	case "POST":
		var newClient Client
		if err := json.NewDecoder(r.Body).Decode(&newClient); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		id, err := addClient(newClient.Name)
		if err != nil {
			http.Error(w, "Failed to add client", http.StatusInternalServerError)
			return
		}
		newClient.ID = id
		jsonResponse(w, newClient)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// jsonResponse writes the given data as JSON.
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// csrfMiddleware implements a simple CSRF protection mechanism.
func csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For unsafe methods, check for a valid CSRF token.
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			cookie, err := r.Cookie(csrfCookieName)
			if err != nil || cookie.Value == "" {
				http.Error(w, "CSRF token missing", http.StatusForbidden)
				return
			}
			token := r.Header.Get(csrfHeaderName)
			if token == "" || token != cookie.Value {
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}
		} else if r.Method == http.MethodGet {
			// For GET requests, if the token is not already set, generate one.
			if _, err := r.Cookie(csrfCookieName); err != nil {
				token, err := generateCSRFToken()
				if err == nil {
					cookie := &http.Cookie{
						Name:  csrfCookieName,
						Value: token,
						Path:  "/",
						// For this demo, the cookie is accessible by JavaScript
						// (so the React app can read it). In production you may want to adjust
						// the security (e.g. using SameSite, Secure, etc.).
						HttpOnly: false,
						Expires:  time.Now().Add(24 * time.Hour),
					}
					http.SetCookie(w, cookie)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// generateCSRFToken returns a random 32-character hex string.
func generateCSRFToken() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
