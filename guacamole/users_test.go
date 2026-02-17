package guacamole

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListUsers(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/users")
		writeJSON(t, w, map[string]User{
			"alice": {Username: "alice"},
			"bob":   {Username: "bob"},
		})
	})
	got, err := c.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len: got %d, want 2", len(got))
	}
}

func TestCreateUser(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPost)
		assertPath(t, r, "/api/session/data/postgresql/users")

		// Verify the attributes field is present and is an object, even when
		// the caller did not set any attributes (nil NullableStringMap → {}).
		var raw map[string]json.RawMessage
		mustReadJSON(t, r, &raw)
		attr, ok := raw["attributes"]
		if !ok {
			t.Error(`"attributes" missing from request body - Guacamole will return HTTP 500`)
		} else if string(attr) != "{}" {
			t.Errorf("attributes: got %s, want {}", attr)
		}

		var body User
		_ = json.Unmarshal(raw["username"], &body.Username)
		writeJSON(t, w, User{Username: body.Username})
	})
	user, err := c.CreateUser(context.Background(), User{Username: "alice", Password: "s3cr3t"})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("Username: got %q, want %q", user.Username, "alice")
	}
}

func TestGetUser(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/users/alice")
		writeJSON(t, w, User{
			Username:   "alice",
			Attributes: NullableStringMap{"guac-full-name": "Alice Smith", "expired": ""},
		})
	})
	u, err := c.GetUser(context.Background(), "alice")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.Attributes["guac-full-name"] != "Alice Smith" {
		t.Errorf(`Attributes["guac-full-name"]: got %q, want "Alice Smith"`, u.Attributes["guac-full-name"])
	}
}

func TestGetUser_null_attributes_normalised(t *testing.T) {
	// Guacamole returns null for unset attribute values. NullableStringMap must
	// convert these to empty strings so callers get a consistent type.
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Return raw JSON with null values, bypassing Go struct serialisation.
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"alice","attributes":{"guac-full-name":null,"expired":null}}`))
	})
	u, err := c.GetUser(context.Background(), "alice")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.Attributes["guac-full-name"] != "" {
		t.Errorf(`null attribute: got %q, want ""`, u.Attributes["guac-full-name"])
	}
}

func TestUpdateUser(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPut)
		assertPath(t, r, "/api/session/data/postgresql/users/alice")
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateUser(context.Background(), "alice", User{Username: "alice", Password: "newpass"})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodDelete)
		assertPath(t, r, "/api/session/data/postgresql/users/alice")
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteUser(context.Background(), "alice"); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
}

// ── Permissions ───────────────────────────────────────────────────────────────

func TestGetUserPermissions(t *testing.T) {
	want := Permissions{
		ConnectionPermissions: map[string][]string{"1": {PermissionRead}},
		SystemPermissions:     []string{SystemPermissionCreateConnection},
	}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/users/alice/permissions")
		writeJSON(t, w, want)
	})
	got, err := c.GetUserPermissions(context.Background(), "alice")
	if err != nil {
		t.Fatalf("GetUserPermissions: %v", err)
	}
	if len(got.ConnectionPermissions["1"]) != 1 || got.ConnectionPermissions["1"][0] != PermissionRead {
		t.Errorf("ConnectionPermissions[1]: got %v, want [READ]", got.ConnectionPermissions["1"])
	}
}

func TestUpdateUserPermissions(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPatch)
		assertPath(t, r, "/api/session/data/postgresql/users/alice/permissions")
		var ops []PatchOperation
		mustReadJSON(t, r, &ops)
		if len(ops) != 1 {
			t.Errorf("ops: got %d, want 1", len(ops))
		}
		if ops[0].Op != "add" || ops[0].Path != "/connectionPermissions/5" || ops[0].Value != PermissionRead {
			t.Errorf("op: got %+v, want add /connectionPermissions/5 READ", ops[0])
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateUserPermissions(context.Background(), "alice", []PatchOperation{
		AddConnectionPermission("5", PermissionRead),
	})
	if err != nil {
		t.Fatalf("UpdateUserPermissions: %v", err)
	}
}

// ── Group membership ──────────────────────────────────────────────────────────

func TestGetUserGroups(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/users/alice/userGroups")
		writeJSON(t, w, []string{"admins", "developers"})
	})
	groups, err := c.GetUserGroups(context.Background(), "alice")
	if err != nil {
		t.Fatalf("GetUserGroups: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("len: got %d, want 2", len(groups))
	}
}

func TestUpdateUserGroups(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPatch)
		assertPath(t, r, "/api/session/data/postgresql/users/alice/userGroups")
		var ops []PatchOperation
		mustReadJSON(t, r, &ops)
		if ops[0].Path != "/" || ops[0].Value != "admins" {
			t.Errorf("op: got %+v, want add / admins", ops[0])
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateUserGroups(context.Background(), "alice", []PatchOperation{
		AddGroupMembership("admins"),
	})
	if err != nil {
		t.Fatalf("UpdateUserGroups: %v", err)
	}
}

// ── Patch helpers ─────────────────────────────────────────────────────────────

func TestPatchHelpers(t *testing.T) {
	cases := []struct {
		name string
		got  PatchOperation
		want PatchOperation
	}{
		{
			"AddConnectionPermission",
			AddConnectionPermission("3", PermissionRead),
			PatchOperation{Op: "add", Path: "/connectionPermissions/3", Value: "READ"},
		},
		{
			"RemoveConnectionPermission",
			RemoveConnectionPermission("3", PermissionRead),
			PatchOperation{Op: "remove", Path: "/connectionPermissions/3", Value: "READ"},
		},
		{
			"AddConnectionGroupPermission",
			AddConnectionGroupPermission("5", PermissionAdminister),
			PatchOperation{Op: "add", Path: "/connectionGroupPermissions/5", Value: "ADMINISTER"},
		},
		{
			"AddSharingProfilePermission",
			AddSharingProfilePermission("2", PermissionRead),
			PatchOperation{Op: "add", Path: "/sharingProfilePermissions/2", Value: "READ"},
		},
		{
			"AddUserPermission",
			AddUserPermission("bob", PermissionUpdate),
			PatchOperation{Op: "add", Path: "/userPermissions/bob", Value: "UPDATE"},
		},
		{
			"AddUserGroupPermission",
			AddUserGroupPermission("devs", PermissionRead),
			PatchOperation{Op: "add", Path: "/userGroupPermissions/devs", Value: "READ"},
		},
		{
			"AddSystemPermission",
			AddSystemPermission(SystemPermissionCreateUser),
			PatchOperation{Op: "add", Path: "/systemPermissions", Value: "CREATE_USER"},
		},
		{
			"RemoveSystemPermission",
			RemoveSystemPermission(SystemPermissionAdminister),
			PatchOperation{Op: "remove", Path: "/systemPermissions", Value: "ADMINISTER"},
		},
		{
			"AddGroupMembership",
			AddGroupMembership("admins"),
			PatchOperation{Op: "add", Path: "/", Value: "admins"},
		},
		{
			"RemoveGroupMembership",
			RemoveGroupMembership("admins"),
			PatchOperation{Op: "remove", Path: "/", Value: "admins"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("got %+v, want %+v", tc.got, tc.want)
			}
		})
	}
}
