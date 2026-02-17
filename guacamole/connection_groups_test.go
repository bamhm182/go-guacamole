package guacamole

import (
	"context"
	"net/http"
	"testing"
)

func TestListConnectionGroups(t *testing.T) {
	want := map[string]ConnectionGroup{
		"1": {Identifier: "1", Name: "Servers", Type: ConnectionGroupTypeOrganizational},
	}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/connectionGroups")
		writeJSON(t, w, want)
	})
	got, err := c.ListConnectionGroups(context.Background())
	if err != nil {
		t.Fatalf("ListConnectionGroups: %v", err)
	}
	if got["1"].Name != "Servers" {
		t.Errorf(`got["1"].Name: got %q, want "Servers"`, got["1"].Name)
	}
}

func TestCreateConnectionGroup(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPost)
		assertPath(t, r, "/api/session/data/postgresql/connectionGroups")
		var body ConnectionGroup
		mustReadJSON(t, r, &body)
		if body.Type != ConnectionGroupTypeOrganizational {
			t.Errorf("body.Type: got %q, want %q", body.Type, ConnectionGroupTypeOrganizational)
		}
		writeJSON(t, w, ConnectionGroup{Identifier: "3", Name: body.Name, Type: body.Type})
	})
	cg, err := c.CreateConnectionGroup(context.Background(), ConnectionGroup{
		Name:             "DC East",
		Type:             ConnectionGroupTypeOrganizational,
		ParentIdentifier: RootConnectionGroupIdentifier,
	})
	if err != nil {
		t.Fatalf("CreateConnectionGroup: %v", err)
	}
	if cg.Identifier != "3" {
		t.Errorf("Identifier: got %q, want %q", cg.Identifier, "3")
	}
}

func TestGetConnectionGroup(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/connectionGroups/2")
		writeJSON(t, w, ConnectionGroup{Identifier: "2", Name: "DC West", Type: ConnectionGroupTypeBalancing})
	})
	cg, err := c.GetConnectionGroup(context.Background(), "2")
	if err != nil {
		t.Fatalf("GetConnectionGroup: %v", err)
	}
	if cg.Type != ConnectionGroupTypeBalancing {
		t.Errorf("Type: got %q, want %q", cg.Type, ConnectionGroupTypeBalancing)
	}
}

func TestGetConnectionGroupTree_ROOT(t *testing.T) {
	tree := ConnectionGroup{
		Name:       "ROOT",
		Identifier: RootConnectionGroupIdentifier,
		ChildConnectionGroups: []ConnectionGroup{
			{Identifier: "1", Name: "Servers"},
		},
		ChildConnections: []Connection{
			{Identifier: "5", Name: "jumphost", Protocol: "ssh"},
		},
	}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/connectionGroups/ROOT/tree")
		writeJSON(t, w, tree)
	})
	got, err := c.GetConnectionGroupTree(context.Background(), RootConnectionGroupIdentifier)
	if err != nil {
		t.Fatalf("GetConnectionGroupTree: %v", err)
	}
	if len(got.ChildConnectionGroups) != 1 {
		t.Errorf("ChildConnectionGroups: got %d, want 1", len(got.ChildConnectionGroups))
	}
	if len(got.ChildConnections) != 1 {
		t.Errorf("ChildConnections: got %d, want 1", len(got.ChildConnections))
	}
}

func TestGetConnectionGroupTree_subtree(t *testing.T) {
	// GetConnectionGroupTree must accept an arbitrary group ID, not just ROOT.
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertPath(t, r, "/api/session/data/postgresql/connectionGroups/7/tree")
		writeJSON(t, w, ConnectionGroup{Identifier: "7", Name: "DC East"})
	})
	got, err := c.GetConnectionGroupTree(context.Background(), "7")
	if err != nil {
		t.Fatalf("GetConnectionGroupTree(7): %v", err)
	}
	if got.Identifier != "7" {
		t.Errorf("Identifier: got %q, want %q", got.Identifier, "7")
	}
}

func TestUpdateConnectionGroup(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPut)
		assertPath(t, r, "/api/session/data/postgresql/connectionGroups/4")
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateConnectionGroup(context.Background(), "4", ConnectionGroup{Name: "Updated", Type: ConnectionGroupTypeOrganizational})
	if err != nil {
		t.Fatalf("UpdateConnectionGroup: %v", err)
	}
}

func TestDeleteConnectionGroup(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodDelete)
		assertPath(t, r, "/api/session/data/postgresql/connectionGroups/4")
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteConnectionGroup(context.Background(), "4"); err != nil {
		t.Fatalf("DeleteConnectionGroup: %v", err)
	}
}
