package guacamole

import (
	"context"
	"fmt"
)

// Self represents the currently-authenticated user's profile as returned by
// the /self endpoint. It is similar to User but also includes a LastActive
// timestamp.
type Self struct {
	Username   string            `json:"username"`
	Disabled   bool              `json:"disabled"`
	LastActive int64             `json:"lastActive"`
	Attributes NullableStringMap `json:"attributes,omitempty"`
}

// GetSelf returns the profile of the currently-authenticated user. This is
// useful for validating credentials and retrieving the authenticated username
// without knowing it in advance.
func (c *Client) GetSelf(ctx context.Context) (*Self, error) {
	var result Self
	if err := c.get(ctx, c.dataPath("self"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get self: %w", err)
	}
	return &result, nil
}

// GetSelfPermissions returns the explicit permissions held by the
// currently-authenticated user. This does not include permissions inherited
// via group membership; use GetSelfEffectivePermissions for the full set.
func (c *Client) GetSelfPermissions(ctx context.Context) (*Permissions, error) {
	var result Permissions
	if err := c.get(ctx, c.dataPath("self", "permissions"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get self permissions: %w", err)
	}
	return &result, nil
}

// GetSelfEffectivePermissions returns the full resolved permission set for the
// currently-authenticated user, including permissions inherited from group
// memberships.
func (c *Client) GetSelfEffectivePermissions(ctx context.Context) (*Permissions, error) {
	var result Permissions
	if err := c.get(ctx, c.dataPath("self", "effectivePermissions"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get self effective permissions: %w", err)
	}
	return &result, nil
}
