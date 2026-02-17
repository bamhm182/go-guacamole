package guacamole

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a Client pointed at a mock HTTP server. The client is
// pre-populated with a test auth token and data source so individual tests do
// not need to exercise the authentication flow. The server is automatically
// closed when the test finishes.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Client{
		baseURL:    srv.URL,
		httpClient: srv.Client(),
		authToken:  "test-token",
		dataSource: "postgresql",
	}
}

// writeJSON serialises v as JSON and writes it to w with a 200 status.
func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("writeJSON: %v", err)
	}
}

// mustReadJSON decodes the request body into v, failing the test on error.
func mustReadJSON(t *testing.T, r *http.Request, v any) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
}

// assertMethod fails the test if r.Method != want.
func assertMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.Method != want {
		t.Errorf("HTTP method: got %q, want %q", r.Method, want)
	}
}

// assertPath fails the test if the request URL path != want.
func assertPath(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.URL.Path != want {
		t.Errorf("URL path: got %q, want %q", r.URL.Path, want)
	}
}

// assertHeader fails the test if the named header != want.
func assertHeader(t *testing.T, r *http.Request, name, want string) {
	t.Helper()
	if got := r.Header.Get(name); got != want {
		t.Errorf("header %q: got %q, want %q", name, got, want)
	}
}

// writeAPIError writes a Guacamole-style JSON error with the given HTTP status.
func writeAPIError(t *testing.T, w http.ResponseWriter, status int, errType, message string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body, _ := json.Marshal(map[string]any{
		"message": message,
		"type":    errType,
	})
	if _, err := w.Write(body); err != nil {
		t.Errorf("writeAPIError: %v", err)
	}
}
