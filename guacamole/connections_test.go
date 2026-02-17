package guacamole

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListConnections(t *testing.T) {
	want := map[string]Connection{
		"1": {Identifier: "1", Name: "My SSH", Protocol: "ssh"},
		"2": {Identifier: "2", Name: "My RDP", Protocol: "rdp"},
	}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/connections")
		writeJSON(t, w, want)
	})
	got, err := c.ListConnections(context.Background())
	if err != nil {
		t.Fatalf("ListConnections: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len: got %d, want 2", len(got))
	}
	if got["1"].Name != "My SSH" {
		t.Errorf(`got["1"].Name: got %q, want "My SSH"`, got["1"].Name)
	}
}

func TestCreateConnection(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPost)
		assertPath(t, r, "/api/session/data/postgresql/connections")

		var body Connection
		mustReadJSON(t, r, &body)
		if body.Name != "My SSH" {
			t.Errorf("body.Name: got %q, want %q", body.Name, "My SSH")
		}
		if body.Protocol != "ssh" {
			t.Errorf("body.Protocol: got %q, want %q", body.Protocol, "ssh")
		}

		// Verify attributes field is always present (never omitted or null)
		var raw map[string]json.RawMessage
		data, _ := json.Marshal(body)
		_ = json.Unmarshal(data, &raw)
		if _, ok := raw["attributes"]; !ok {
			t.Error(`"attributes" missing from request body - Guacamole will return HTTP 500`)
		}

		writeJSON(t, w, Connection{Identifier: "5", Name: body.Name, Protocol: body.Protocol})
	})

	conn, err := c.CreateConnection(context.Background(), Connection{
		Name:     "My SSH",
		Protocol: "ssh",
		Parameters: map[string]string{"hostname": "10.0.0.1", "port": "22"},
	})
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	if conn.Identifier != "5" {
		t.Errorf("Identifier: got %q, want %q", conn.Identifier, "5")
	}
}

func TestCreateConnection_nil_attributes_serialized_as_empty_object(t *testing.T) {
	// Regression test: nil Attributes must marshal as {} not be omitted.
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		mustReadJSON(t, r, &raw)
		attr, ok := raw["attributes"]
		if !ok {
			t.Error(`"attributes" key missing from request body`)
		} else if string(attr) != "{}" {
			t.Errorf("attributes: got %s, want {}", attr)
		}
		writeJSON(t, w, Connection{Identifier: "1"})
	})
	_, _ = c.CreateConnection(context.Background(), Connection{Name: "x", Protocol: "ssh"})
}

func TestGetConnection(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/connections/42")
		writeJSON(t, w, Connection{Identifier: "42", Name: "found", Protocol: "rdp"})
	})
	conn, err := c.GetConnection(context.Background(), "42")
	if err != nil {
		t.Fatalf("GetConnection: %v", err)
	}
	if conn.Protocol != "rdp" {
		t.Errorf("Protocol: got %q, want %q", conn.Protocol, "rdp")
	}
}

func TestGetConnection_not_found(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeAPIError(t, w, http.StatusNotFound, ErrTypeNotFound, `Not found: "999"`)
	})
	_, err := c.GetConnection(context.Background(), "999")
	if !IsNotFound(err) {
		t.Errorf("IsNotFound: got false, want true (err=%v)", err)
	}
}

func TestGetConnectionParameters(t *testing.T) {
	want := map[string]string{"hostname": "10.0.0.1", "port": "22", "username": "admin"}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/connections/7/parameters")
		writeJSON(t, w, want)
	})
	got, err := c.GetConnectionParameters(context.Background(), "7")
	if err != nil {
		t.Fatalf("GetConnectionParameters: %v", err)
	}
	if got["hostname"] != "10.0.0.1" {
		t.Errorf(`hostname: got %q, want "10.0.0.1"`, got["hostname"])
	}
}

func TestUpdateConnection(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPut)
		assertPath(t, r, "/api/session/data/postgresql/connections/3")
		var body Connection
		mustReadJSON(t, r, &body)
		if body.Name != "Updated" {
			t.Errorf("body.Name: got %q, want %q", body.Name, "Updated")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateConnection(context.Background(), "3", Connection{Name: "Updated", Protocol: "ssh"})
	if err != nil {
		t.Fatalf("UpdateConnection: %v", err)
	}
}

func TestDeleteConnection(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodDelete)
		assertPath(t, r, "/api/session/data/postgresql/connections/9")
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteConnection(context.Background(), "9"); err != nil {
		t.Fatalf("DeleteConnection: %v", err)
	}
}
