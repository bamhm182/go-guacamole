package guacamole

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// ── Authentication ─────────────────────────────────────────────────────────────

func TestAuthenticate_success(t *testing.T) {
	srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPost)
		assertPath(t, r, "/api/tokens")
		assertHeader(t, r, "Content-Type", "application/x-www-form-urlencoded")

		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.FormValue("username") != "admin" {
			t.Errorf("username: got %q, want %q", r.FormValue("username"), "admin")
		}
		if r.FormValue("password") != "secret" {
			t.Errorf("password: got %q, want %q", r.FormValue("password"), "secret")
		}

		writeJSON(t, w, AuthResponse{
			AuthToken:  "mytoken",
			DataSource: "mysql",
		})
	})
	// Reset auth state so Authenticate actually performs the request
	srv.authToken = ""
	srv.dataSource = ""

	if err := srv.Authenticate(context.Background(), "admin", "secret"); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if srv.authToken != "mytoken" {
		t.Errorf("authToken: got %q, want %q", srv.authToken, "mytoken")
	}
	if srv.dataSource != "mysql" {
		t.Errorf("dataSource: got %q, want %q", srv.dataSource, "mysql")
	}
}

func TestAuthenticate_error(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeAPIError(t, w, http.StatusForbidden, ErrTypePermissionDenied, "Invalid credentials.")
	})
	c.authToken = ""
	c.dataSource = ""

	err := c.Authenticate(context.Background(), "bad", "creds")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsPermissionDenied(err) {
		t.Errorf("IsPermissionDenied: got false, want true (err=%v)", err)
	}
}

func TestLogout(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodDelete)
		assertPath(t, r, "/api/session")
		assertHeader(t, r, "Guacamole-Token", "test-token")
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.Logout(context.Background()); err != nil {
		t.Fatalf("Logout: %v", err)
	}
}

// ── Error handling ─────────────────────────────────────────────────────────────

func TestIsNotFound_through_wrapped_error(t *testing.T) {
	// IsNotFound must work even when the APIError is wrapped by fmt.Errorf %w.
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeAPIError(t, w, http.StatusNotFound, ErrTypeNotFound, `Not found: "99"`)
	})
	_, err := c.GetConnection(context.Background(), "99")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound: got false, want true (err=%v)", err)
	}
	if IsPermissionDenied(err) {
		t.Error("IsPermissionDenied: got true, want false")
	}
}

func TestIsPermissionDenied_through_wrapped_error(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeAPIError(t, w, http.StatusForbidden, ErrTypePermissionDenied, "Permission Denied.")
	})
	_, err := c.GetUser(context.Background(), "bob")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsPermissionDenied(err) {
		t.Errorf("IsPermissionDenied: got false, want true (err=%v)", err)
	}
	if IsNotFound(err) {
		t.Error("IsNotFound: got true, want false")
	}
}

func TestAPIError_error_message(t *testing.T) {
	e := &APIError{HTTPStatus: 404, Type: ErrTypeNotFound, Message: `Not found: "1"`}
	got := e.Error()
	if !strings.Contains(got, "404") {
		t.Errorf("error message %q missing HTTP status", got)
	}
	if !strings.Contains(got, ErrTypeNotFound) {
		t.Errorf("error message %q missing type", got)
	}
}

func TestIsNotFound_nil_error(t *testing.T) {
	if IsNotFound(nil) {
		t.Error("IsNotFound(nil): got true, want false")
	}
}

func TestIsPermissionDenied_nil_error(t *testing.T) {
	if IsPermissionDenied(nil) {
		t.Error("IsPermissionDenied(nil): got true, want false")
	}
}

// ── Auth token header ──────────────────────────────────────────────────────────

func TestAuthTokenSentOnRequests(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertHeader(t, r, "Guacamole-Token", "test-token")
		writeJSON(t, w, map[string]Connection{})
	})
	if _, err := c.ListConnections(context.Background()); err != nil {
		t.Fatalf("ListConnections: %v", err)
	}
}

// ── URL encoding ───────────────────────────────────────────────────────────────

func TestDataPath_url_encodes_special_chars(t *testing.T) {
	cases := []struct {
		segment string
		want    string
	}{
		{"normal", "normal"},
		{"with space", "with%20space"},
		{"user@domain.com", "user%40domain.com"},
		{"group/name", "group%2Fname"},
	}
	c := &Client{dataSource: "postgresql"}
	for _, tc := range cases {
		path := c.dataPath("users", tc.segment)
		encoded := url.PathEscape(tc.segment)
		if !strings.Contains(path, encoded) {
			t.Errorf("dataPath(%q): got %q, want it to contain %q", tc.segment, path, encoded)
		}
	}
}

func TestGetUser_special_chars_url_encoded(t *testing.T) {
	const username = "bob@example.com"
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		want := fmt.Sprintf("/api/session/data/postgresql/users/%s", url.PathEscape(username))
		assertPath(t, r, want)
		writeJSON(t, w, User{Username: username})
	})
	u, err := c.GetUser(context.Background(), username)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.Username != username {
		t.Errorf("Username: got %q, want %q", u.Username, username)
	}
}

// ── JSON body content type ─────────────────────────────────────────────────────

func TestPostSetsContentTypeJSON(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type: got %q, want %q", ct, "application/json")
		}
		writeJSON(t, w, User{Username: "u"})
	})
	_, _ = c.CreateUser(context.Background(), User{Username: "u"})
}

// ── Non-2xx without JSON body ──────────────────────────────────────────────────

func TestParseError_non_json_body(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("Service Unavailable")); err != nil {
			t.Errorf("write: %v", err)
		}
	})
	_, err := c.ListUsers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	// Should still return an *APIError even if the body isn't JSON
	if !isAPIError(err, &apiErr) {
		t.Errorf("expected *APIError in chain, got %T: %v", err, err)
	}
	if apiErr.HTTPStatus != http.StatusServiceUnavailable {
		t.Errorf("HTTPStatus: got %d, want %d", apiErr.HTTPStatus, http.StatusServiceUnavailable)
	}
}

// isAPIError walks the error chain to find an *APIError.
func isAPIError(err error, target **APIError) bool {
	for err != nil {
		if e, ok := err.(*APIError); ok {
			*target = e
			return true
		}
		type unwrapper interface{ Unwrap() error }
		if u, ok := err.(unwrapper); ok {
			err = u.Unwrap()
		} else {
			break
		}
	}
	return false
}
