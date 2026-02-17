package guacamole

import (
	"context"
	"net/http"
	"testing"
)

func TestListUserGroups(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/userGroups")
		writeJSON(t, w, map[string]UserGroup{
			"admins": {Identifier: "admins"},
			"devs":   {Identifier: "devs"},
		})
	})
	got, err := c.ListUserGroups(context.Background())
	if err != nil {
		t.Fatalf("ListUserGroups: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len: got %d, want 2", len(got))
	}
}

func TestCreateUserGroup(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPost)
		assertPath(t, r, "/api/session/data/postgresql/userGroups")
		var body UserGroup
		mustReadJSON(t, r, &body)
		if body.Identifier != "admins" {
			t.Errorf("Identifier: got %q, want %q", body.Identifier, "admins")
		}
		writeJSON(t, w, UserGroup{Identifier: body.Identifier})
	})
	ug, err := c.CreateUserGroup(context.Background(), UserGroup{Identifier: "admins"})
	if err != nil {
		t.Fatalf("CreateUserGroup: %v", err)
	}
	if ug.Identifier != "admins" {
		t.Errorf("Identifier: got %q, want %q", ug.Identifier, "admins")
	}
}

func TestGetUserGroup(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/admins")
		writeJSON(t, w, UserGroup{Identifier: "admins", Disabled: false})
	})
	ug, err := c.GetUserGroup(context.Background(), "admins")
	if err != nil {
		t.Fatalf("GetUserGroup: %v", err)
	}
	if ug.Identifier != "admins" {
		t.Errorf("Identifier: got %q, want %q", ug.Identifier, "admins")
	}
}

func TestUpdateUserGroup(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPut)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/admins")
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateUserGroup(context.Background(), "admins", UserGroup{Identifier: "admins"})
	if err != nil {
		t.Fatalf("UpdateUserGroup: %v", err)
	}
}

func TestDeleteUserGroup(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodDelete)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/admins")
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteUserGroup(context.Background(), "admins"); err != nil {
		t.Fatalf("DeleteUserGroup: %v", err)
	}
}

// ── Permissions ───────────────────────────────────────────────────────────────

func TestGetUserGroupPermissions(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/admins/permissions")
		writeJSON(t, w, Permissions{
			SystemPermissions: []string{SystemPermissionAdminister},
		})
	})
	perms, err := c.GetUserGroupPermissions(context.Background(), "admins")
	if err != nil {
		t.Fatalf("GetUserGroupPermissions: %v", err)
	}
	if len(perms.SystemPermissions) != 1 || perms.SystemPermissions[0] != SystemPermissionAdminister {
		t.Errorf("SystemPermissions: got %v, want [ADMINISTER]", perms.SystemPermissions)
	}
}

func TestUpdateUserGroupPermissions(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPatch)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/admins/permissions")
		var ops []PatchOperation
		mustReadJSON(t, r, &ops)
		if ops[0].Path != "/systemPermissions" || ops[0].Value != SystemPermissionAdminister {
			t.Errorf("op: got %+v", ops[0])
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateUserGroupPermissions(context.Background(), "admins", []PatchOperation{
		AddSystemPermission(SystemPermissionAdminister),
	})
	if err != nil {
		t.Fatalf("UpdateUserGroupPermissions: %v", err)
	}
}

// ── Member management ─────────────────────────────────────────────────────────

func TestGetUserGroupMemberUsers(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/admins/memberUsers")
		writeJSON(t, w, []string{"alice", "bob"})
	})
	users, err := c.GetUserGroupMemberUsers(context.Background(), "admins")
	if err != nil {
		t.Fatalf("GetUserGroupMemberUsers: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("len: got %d, want 2", len(users))
	}
}

func TestUpdateUserGroupMemberUsers(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPatch)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/admins/memberUsers")
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateUserGroupMemberUsers(context.Background(), "admins", []PatchOperation{
		AddGroupMembership("alice"),
	})
	if err != nil {
		t.Fatalf("UpdateUserGroupMemberUsers: %v", err)
	}
}

func TestGetUserGroupMemberGroups(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/all-staff/memberUserGroups")
		writeJSON(t, w, []string{"admins", "devs"})
	})
	groups, err := c.GetUserGroupMemberGroups(context.Background(), "all-staff")
	if err != nil {
		t.Fatalf("GetUserGroupMemberGroups: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("len: got %d, want 2", len(groups))
	}
}

func TestUpdateUserGroupMemberGroups(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPatch)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/all-staff/memberUserGroups")
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateUserGroupMemberGroups(context.Background(), "all-staff", []PatchOperation{
		AddGroupMembership("devs"),
	})
	if err != nil {
		t.Fatalf("UpdateUserGroupMemberGroups: %v", err)
	}
}

func TestGetUserGroupParentGroups(t *testing.T) {
	// Verifies the /userGroups endpoint (which groups does this group belong to),
	// distinct from /memberUserGroups (which groups are inside this group).
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/devs/userGroups")
		writeJSON(t, w, []string{"all-staff"})
	})
	parents, err := c.GetUserGroupParentGroups(context.Background(), "devs")
	if err != nil {
		t.Fatalf("GetUserGroupParentGroups: %v", err)
	}
	if len(parents) != 1 || parents[0] != "all-staff" {
		t.Errorf("parents: got %v, want [all-staff]", parents)
	}
}

func TestUpdateUserGroupParentGroups(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPatch)
		assertPath(t, r, "/api/session/data/postgresql/userGroups/devs/userGroups")
		var ops []PatchOperation
		mustReadJSON(t, r, &ops)
		if ops[0].Value != "all-staff" {
			t.Errorf("op value: got %q, want %q", ops[0].Value, "all-staff")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateUserGroupParentGroups(context.Background(), "devs", []PatchOperation{
		AddGroupMembership("all-staff"),
	})
	if err != nil {
		t.Fatalf("UpdateUserGroupParentGroups: %v", err)
	}
}
