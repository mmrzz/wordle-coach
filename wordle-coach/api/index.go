// Package handler runs the coach as a serverless function.
//
// Vercel builds each file under api/ into its own function and calls the
// exported Handler. One file serves every route: a file per endpoint would
// mean several exported symbols of the same name in one directory, which the
// rest of the module could no longer be built alongside.
//
// The build happens inside a synthesized module of Vercel's own, with this
// package relabelled handler/api and the real one pulled in beside it. That is
// why the shared packages sit under pkg/ and not internal/: from a module that
// is not ours, an internal package is not importable at all.
package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/mmrzz/wordle-coach/pkg/api"
	"github.com/mmrzz/wordle-coach/pkg/data"
)

// load parses the word lists once per instance rather than once per request.
// A warm function then answers from memory, and a cold one pays about 10ms.
var load = sync.OnceValues(func() (*http.ServeMux, error) {
	set, err := data.Load()
	if err != nil {
		return nil, err
	}
	return api.NewMux(set), nil
})

// Handler serves every route the coach has.
//
// No CORS here: the page and the API are served from one origin, so there is
// no cross-origin request to allow. That is a property of this deployment and
// not of the API, which is why the middleware stays with the standalone
// server rather than moving into pkg/api.
func Handler(w http.ResponseWriter, r *http.Request) {
	mux, err := load()
	if err != nil {
		// The lists are embedded in the binary, so this cannot be a transient
		// failure: it is a corrupt build, and every request will hit it.
		log.Printf("handler: loading the word lists: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"the word lists could not be loaded","code":"unavailable"}`))
		return
	}

	// Every route is rewritten to this one function, so the path the mux
	// matches on is passed along explicitly rather than assumed to survive.
	if path := r.URL.Query().Get("path"); path != "" {
		r.URL.Path = path
	}

	mux.ServeHTTP(w, r)
}
