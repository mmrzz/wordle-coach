// Package api holds the HTTP handlers for the wordle-coach API.
package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// Health reports that the service is up.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// The status is already sent, so a failure here can only be logged.
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("health: encoding response: %v", err)
	}
}
