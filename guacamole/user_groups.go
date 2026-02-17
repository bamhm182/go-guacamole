package guacamole

import (
	"context"
	"fmt"
)

// ListUserGroups returns all user groups visible to the authenticated user,
// keyed by identifier.
func (c *Client) ListUserGroups(ctx context.Context) (map[string]UserGroup, error) {
	var result map[string]UserGroup
	if err := c.get(ctx, c.dataPath("userGroups"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: list user groups: %w", err)
	}
	return result, nil
}

// CreateUserGroup creates a new user group and returns the created resource.
func (c *Client) CreateUserGroup(ctx context.Context, group UserGroup) (*UserGroup, error) {
	var result UserGroup
	if err := c.post(ctx, c.dataPath("userGroups"), group, &result); err != nil {
		return nil, fmt.Errorf("guacamole: create user group: %w", err)
	}
	return &result, nil
}

// GetUserGroup retrieves the user group with the given identifier.
func (c *Client) GetUserGroup(ctx context.Context, id string) (*UserGroup, error) {
	var result UserGroup
	if err := c.get(ctx, c.dataPath("userGroups", id), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get user group %s: %w", id, err)
	}
	return &result, nil
}

// UpdateUserGroup replaces the user group identified by id with the supplied
// UserGroup. The identifier field within group is ignored; id is used.
func (c *Client) UpdateUserGroup(ctx context.Context, id string, group UserGroup) error {
	if err := c.put(ctx, c.dataPath("userGroups", id), group); err != nil {
		return fmt.Errorf("guacamole: update user group %s: %w", id, err)
	}
	return nil
}

// DeleteUserGroup permanently removes the user group with the given
// identifier.
func (c *Client) DeleteUserGroup(ctx context.Context, id string) error {
	if err := c.delete(ctx, c.dataPath("userGroups", id)); err != nil {
		return fmt.Errorf("guacamole: delete user group %s: %w", id, err)
	}
	return nil
}

// ── Permissions ───────────────────────────────────────────────────────────────

// GetUserGroupPermissions returns the explicit permissions granted to the user
// group. These permissions apply to all members of the group.
func (c *Client) GetUserGroupPermissions(ctx context.Context, id string) (*Permissions, error) {
	var result Permissions
	if err := c.get(ctx, c.dataPath("userGroups", id, "permissions"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get user group permissions %s: %w", id, err)
	}
	return &result, nil
}

// UpdateUserGroupPermissions applies the given JSON Patch operations to the
// user group's permissions.
func (c *Client) UpdateUserGroupPermissions(ctx context.Context, id string, ops []PatchOperation) error {
	if err := c.patch(ctx, c.dataPath("userGroups", id, "permissions"), ops); err != nil {
		return fmt.Errorf("guacamole: update user group permissions %s: %w", id, err)
	}
	return nil
}

// ── Member management ─────────────────────────────────────────────────────────

// GetUserGroupMemberUsers returns the usernames of individual users who are
// direct members of the given user group.
func (c *Client) GetUserGroupMemberUsers(ctx context.Context, id string) ([]string, error) {
	var result []string
	if err := c.get(ctx, c.dataPath("userGroups", id, "memberUsers"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get member users of group %s: %w", id, err)
	}
	return result, nil
}

// UpdateUserGroupMemberUsers applies the given JSON Patch operations to the
// user group's member user list. Use AddGroupMembership / RemoveGroupMembership
// to construct the operations.
func (c *Client) UpdateUserGroupMemberUsers(ctx context.Context, id string, ops []PatchOperation) error {
	if err := c.patch(ctx, c.dataPath("userGroups", id, "memberUsers"), ops); err != nil {
		return fmt.Errorf("guacamole: update member users of group %s: %w", id, err)
	}
	return nil
}

// GetUserGroupMemberGroups returns the identifiers of child user groups that
// are nested within the given user group.
func (c *Client) GetUserGroupMemberGroups(ctx context.Context, id string) ([]string, error) {
	var result []string
	if err := c.get(ctx, c.dataPath("userGroups", id, "memberUserGroups"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get member groups of group %s: %w", id, err)
	}
	return result, nil
}

// UpdateUserGroupMemberGroups applies the given JSON Patch operations to the
// user group's nested-group membership list.
func (c *Client) UpdateUserGroupMemberGroups(ctx context.Context, id string, ops []PatchOperation) error {
	if err := c.patch(ctx, c.dataPath("userGroups", id, "memberUserGroups"), ops); err != nil {
		return fmt.Errorf("guacamole: update member groups of group %s: %w", id, err)
	}
	return nil
}

// ── Parent group membership ───────────────────────────────────────────────────
// These are the mirror of the member endpoints above. While memberUserGroups
// manages which groups are *inside* a given group, the userGroups endpoint
// manages which groups the given group itself *belongs to*.

// GetUserGroupParentGroups returns the identifiers of the groups that the
// given user group is a direct member of.
func (c *Client) GetUserGroupParentGroups(ctx context.Context, id string) ([]string, error) {
	var result []string
	if err := c.get(ctx, c.dataPath("userGroups", id, "userGroups"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get parent groups of group %s: %w", id, err)
	}
	return result, nil
}

// UpdateUserGroupParentGroups applies the given JSON Patch operations to the
// set of groups that the given user group belongs to.
func (c *Client) UpdateUserGroupParentGroups(ctx context.Context, id string, ops []PatchOperation) error {
	if err := c.patch(ctx, c.dataPath("userGroups", id, "userGroups"), ops); err != nil {
		return fmt.Errorf("guacamole: update parent groups of group %s: %w", id, err)
	}
	return nil
}
