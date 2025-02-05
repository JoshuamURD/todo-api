package main

import (
	"encoding/json"
	"joshuamURD/go-auth-api/pkgs/controllers"
	"net/http"

	_ "modernc.org/sqlite" // Import with blank identifier to register the driver
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/register", controllers.Register)

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}

	server.ListenAndServe()
}

// jsonResponse writes the given data as JSON.
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
