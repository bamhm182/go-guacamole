package guacamole

import (
	"context"
	"fmt"
)

// System permission constants.
const (
	SystemPermissionCreateUser           = "CREATE_USER"
	SystemPermissionCreateUserGroup      = "CREATE_USER_GROUP"
	SystemPermissionCreateConnection     = "CREATE_CONNECTION"
	SystemPermissionCreateConnectionGroup = "CREATE_CONNECTION_GROUP"
	SystemPermissionCreateSharingProfile = "CREATE_SHARING_PROFILE"
	SystemPermissionAdminister           = "ADMINISTER"
)

// Object permission constants. These apply to connections, connection groups,
// sharing profiles, users, and user groups.
const (
	PermissionRead       = "READ"
	PermissionUpdate     = "UPDATE"
	PermissionDelete     = "DELETE"
	PermissionAdminister = "ADMINISTER"
)

// ListUsers returns all users visible to the authenticated user, keyed by
// username.
func (c *Client) ListUsers(ctx context.Context) (map[string]User, error) {
	var result map[string]User
	if err := c.get(ctx, c.dataPath("users"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: list users: %w", err)
	}
	return result, nil
}

// CreateUser creates a new user and returns the created resource. The Password
// field of the returned User will be empty (the API does not echo passwords).
func (c *Client) CreateUser(ctx context.Context, user User) (*User, error) {
	var result User
	if err := c.post(ctx, c.dataPath("users"), user, &result); err != nil {
		return nil, fmt.Errorf("guacamole: create user: %w", err)
	}
	return &result, nil
}

// GetUser retrieves the user with the given username.
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	var result User
	if err := c.get(ctx, c.dataPath("users", username), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get user %s: %w", username, err)
	}
	return &result, nil
}

// UpdateUser replaces the user identified by username with the supplied User.
// To change a user's password, include the new password in the Password field.
// To leave the password unchanged, omit it (empty string).
func (c *Client) UpdateUser(ctx context.Context, username string, user User) error {
	if err := c.put(ctx, c.dataPath("users", username), user); err != nil {
		return fmt.Errorf("guacamole: update user %s: %w", username, err)
	}
	return nil
}

// DeleteUser permanently removes the user with the given username.
func (c *Client) DeleteUser(ctx context.Context, username string) error {
	if err := c.delete(ctx, c.dataPath("users", username)); err != nil {
		return fmt.Errorf("guacamole: delete user %s: %w", username, err)
	}
	return nil
}

// ── Permissions ───────────────────────────────────────────────────────────────

// GetUserPermissions returns the explicit permissions granted directly to the
// user. This does not include permissions inherited via group membership; use
// GetUserEffectivePermissions for the full resolved set.
func (c *Client) GetUserPermissions(ctx context.Context, username string) (*Permissions, error) {
	var result Permissions
	if err := c.get(ctx, c.dataPath("users", username, "permissions"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get user permissions %s: %w", username, err)
	}
	return &result, nil
}

// GetUserEffectivePermissions returns the full resolved permission set for the
// user, including permissions inherited from group memberships.
func (c *Client) GetUserEffectivePermissions(ctx context.Context, username string) (*Permissions, error) {
	var result Permissions
	if err := c.get(ctx, c.dataPath("users", username, "effectivePermissions"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get user effective permissions %s: %w", username, err)
	}
	return &result, nil
}

// UpdateUserPermissions applies the given JSON Patch operations to the user's
// permissions. Use AddUserConnectionPermission, AddUserSystemPermission, and
// the other patch helpers to construct the operations slice.
func (c *Client) UpdateUserPermissions(ctx context.Context, username string, ops []PatchOperation) error {
	if err := c.patch(ctx, c.dataPath("users", username, "permissions"), ops); err != nil {
		return fmt.Errorf("guacamole: update user permissions %s: %w", username, err)
	}
	return nil
}

// ── Group membership ──────────────────────────────────────────────────────────

// GetUserGroups returns the identifiers of the user groups that the given user
// is a direct member of.
func (c *Client) GetUserGroups(ctx context.Context, username string) ([]string, error) {
	var result []string
	if err := c.get(ctx, c.dataPath("users", username, "userGroups"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get user groups for %s: %w", username, err)
	}
	return result, nil
}

// UpdateUserGroups applies the given JSON Patch operations to the user's group
// membership list.
func (c *Client) UpdateUserGroups(ctx context.Context, username string, ops []PatchOperation) error {
	if err := c.patch(ctx, c.dataPath("users", username, "userGroups"), ops); err != nil {
		return fmt.Errorf("guacamole: update user groups for %s: %w", username, err)
	}
	return nil
}

// ── Patch helpers ─────────────────────────────────────────────────────────────

// AddConnectionPermission returns a PatchOperation that grants the given
// permission on a connection to a user or group.
func AddConnectionPermission(connectionID, permission string) PatchOperation {
	return PatchOperation{Op: "add", Path: "/connectionPermissions/" + connectionID, Value: permission}
}

// RemoveConnectionPermission returns a PatchOperation that revokes the given
// permission on a connection from a user or group.
func RemoveConnectionPermission(connectionID, permission string) PatchOperation {
	return PatchOperation{Op: "remove", Path: "/connectionPermissions/" + connectionID, Value: permission}
}

// AddConnectionGroupPermission returns a PatchOperation that grants the given
// permission on a connection group.
func AddConnectionGroupPermission(groupID, permission string) PatchOperation {
	return PatchOperation{Op: "add", Path: "/connectionGroupPermissions/" + groupID, Value: permission}
}

// RemoveConnectionGroupPermission returns a PatchOperation that revokes the
// given permission on a connection group.
func RemoveConnectionGroupPermission(groupID, permission string) PatchOperation {
	return PatchOperation{Op: "remove", Path: "/connectionGroupPermissions/" + groupID, Value: permission}
}

// AddSharingProfilePermission returns a PatchOperation that grants the given
// permission on a sharing profile.
func AddSharingProfilePermission(profileID, permission string) PatchOperation {
	return PatchOperation{Op: "add", Path: "/sharingProfilePermissions/" + profileID, Value: permission}
}

// RemoveSharingProfilePermission returns a PatchOperation that revokes the
// given permission on a sharing profile.
func RemoveSharingProfilePermission(profileID, permission string) PatchOperation {
	return PatchOperation{Op: "remove", Path: "/sharingProfilePermissions/" + profileID, Value: permission}
}

// AddUserPermission returns a PatchOperation that grants the given permission
// on a user account (e.g. READ, UPDATE, ADMINISTER).
func AddUserPermission(targetUsername, permission string) PatchOperation {
	return PatchOperation{Op: "add", Path: "/userPermissions/" + targetUsername, Value: permission}
}

// RemoveUserPermission returns a PatchOperation that revokes the given
// permission on a user account.
func RemoveUserPermission(targetUsername, permission string) PatchOperation {
	return PatchOperation{Op: "remove", Path: "/userPermissions/" + targetUsername, Value: permission}
}

// AddUserGroupPermission returns a PatchOperation that grants the given
// permission on a user group.
func AddUserGroupPermission(groupID, permission string) PatchOperation {
	return PatchOperation{Op: "add", Path: "/userGroupPermissions/" + groupID, Value: permission}
}

// RemoveUserGroupPermission returns a PatchOperation that revokes the given
// permission on a user group.
func RemoveUserGroupPermission(groupID, permission string) PatchOperation {
	return PatchOperation{Op: "remove", Path: "/userGroupPermissions/" + groupID, Value: permission}
}

// AddSystemPermission returns a PatchOperation that grants the given system
// permission (e.g. SystemPermissionCreateConnection).
func AddSystemPermission(permission string) PatchOperation {
	return PatchOperation{Op: "add", Path: "/systemPermissions", Value: permission}
}

// RemoveSystemPermission returns a PatchOperation that revokes the given
// system permission.
func RemoveSystemPermission(permission string) PatchOperation {
	return PatchOperation{Op: "remove", Path: "/systemPermissions", Value: permission}
}

// AddGroupMembership returns a PatchOperation that adds a user or group to a
// membership list (the path "/" used by userGroups and memberUsers endpoints).
func AddGroupMembership(identifier string) PatchOperation {
	return PatchOperation{Op: "add", Path: "/", Value: identifier}
}

// RemoveGroupMembership returns a PatchOperation that removes a user or group
// from a membership list.
func RemoveGroupMembership(identifier string) PatchOperation {
	return PatchOperation{Op: "remove", Path: "/", Value: identifier}
}
