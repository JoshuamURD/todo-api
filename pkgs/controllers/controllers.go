package controllers

import (
	"encoding/json"
	"net/http"
)

// jsonResponse writes the given data as JSON.
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
