package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The rewrite in vercel.json is the only thing standing between a request and
// the route it was meant for, and it cannot be exercised until the thing is
// deployed. This checks our side of that contract: the path arrives as a query
// parameter, and every route answers as it would on the standalone server.
func TestHandlerServesEveryRouteThroughTheRewrite(t *testing.T) {
	tests := []struct {
		name   string
		method string
		target string
		body   string
		status int
	}{
		{"health", http.MethodGet, "/api/index?path=/healthz", "", http.StatusOK},
		{"words", http.MethodGet, "/api/index?path=/api/words", "", http.StatusOK},
		{
			"suggest", http.MethodPost, "/api/index?path=/api/suggest",
			`{"mode":"easy","history":[]}`, http.StatusOK,
		},
		{
			"rate", http.MethodPost, "/api/index?path=/api/rate",
			`{"mode":"easy","history":[],"played":"crane"}`, http.StatusOK,
		},
		{
			"an impossible position still reports itself",
			http.MethodPost, "/api/index?path=/api/suggest",
			`{"mode":"easy","history":[{"guess":"aalii","pattern":"ggggg"}]}`,
			http.StatusUnprocessableEntity,
		},
		{"an unrouted path is not found", http.MethodGet, "/api/index?path=/api/nope", "", http.StatusNotFound},
		{"the wrong method is not allowed", http.MethodGet, "/api/index?path=/api/suggest", "", http.StatusMethodNotAllowed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			Handler(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tc.status, rec.Body.String())
			}
		})
	}
}

// Should the platform ever deliver the path it was asked for rather than the
// one it rewrote to, nothing here needs to change. The query is a fallback,
// not a requirement.
func TestHandlerServesAPlainPathToo(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/words", nil)
	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body struct {
		Allowed     []string `json:"allowed"`
		AnswerCount int      `json:"answerCount"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if len(body.Allowed) == 0 || body.AnswerCount == 0 {
		t.Errorf("got %d allowed and %d answers, want the whole corpus", len(body.Allowed), body.AnswerCount)
	}
}
