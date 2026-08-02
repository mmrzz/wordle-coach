package main

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the status code,
// which the standard ResponseWriter does not expose for reading.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// parseOrigins turns a comma-separated origin list into a lookup set.
func parseOrigins(list string) map[string]bool {
	origins := make(map[string]bool)
	for o := range strings.SplitSeq(list, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins[o] = true
		}
	}
	return origins
}

// cors builds middleware allowing browser requests from the given origins
// and answering preflights.
func cors(allowed map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The response differs per origin, so caches must not share it.
			w.Header().Add("Vary", "Origin")

			if origin := r.Header.Get("Origin"); allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Max-Age", "300")
			}

			// A preflight is answered here and never reaches the route.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// logging logs the method, path, status and duration of every request.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Default to 200: a handler that never calls WriteHeader still sends 200.
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}
