// Package api holds the HTTP handlers for the wordle-coach API.
//
// This and its siblings live under pkg/ rather than internal/ deliberately.
// The serverless entrypoint in api/ is compiled inside a module of the host's
// own making, and an internal package cannot be imported from outside the
// module that owns it: moving these back would build fine here and fail only
// on deploy.
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
