package main

import (
	"encoding/json"
	"net/http"

	_ "modernc.org/sqlite" // Import with blank identifier to register the driver
)

func main() {

}

// jsonResponse writes the given data as JSON.
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
