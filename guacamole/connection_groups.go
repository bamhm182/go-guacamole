package guacamole

import (
	"context"
	"fmt"
)

// ConnectionGroupTypeOrganizational is the type value for an organizational
// (folder-like) connection group.
const ConnectionGroupTypeOrganizational = "ORGANIZATIONAL"

// ConnectionGroupTypeBalancing is the type value for a load-balancing
// connection group.
const ConnectionGroupTypeBalancing = "BALANCING"

// RootConnectionGroupIdentifier is the identifier of the root connection
// group, which is the parent of all top-level connections and groups.
const RootConnectionGroupIdentifier = "ROOT"

// ListConnectionGroups returns all connection groups visible to the
// authenticated user, keyed by identifier.
func (c *Client) ListConnectionGroups(ctx context.Context) (map[string]ConnectionGroup, error) {
	var result map[string]ConnectionGroup
	if err := c.get(ctx, c.dataPath("connectionGroups"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: list connection groups: %w", err)
	}
	return result, nil
}

// GetConnectionGroupTree returns the connection group hierarchy rooted at the
// given group identifier, including all nested groups and their child
// connections. Pass RootConnectionGroupIdentifier ("ROOT") to retrieve the
// complete topology, or pass a specific group identifier to retrieve a subtree.
func (c *Client) GetConnectionGroupTree(ctx context.Context, rootID string) (*ConnectionGroup, error) {
	var result ConnectionGroup
	if err := c.get(ctx, c.dataPath("connectionGroups", rootID, "tree"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get connection group tree %s: %w", rootID, err)
	}
	return &result, nil
}

// CreateConnectionGroup creates a new connection group and returns the created
// resource with its server-assigned identifier.
func (c *Client) CreateConnectionGroup(ctx context.Context, group ConnectionGroup) (*ConnectionGroup, error) {
	var result ConnectionGroup
	if err := c.post(ctx, c.dataPath("connectionGroups"), group, &result); err != nil {
		return nil, fmt.Errorf("guacamole: create connection group: %w", err)
	}
	return &result, nil
}

// GetConnectionGroup retrieves the connection group with the given identifier.
func (c *Client) GetConnectionGroup(ctx context.Context, id string) (*ConnectionGroup, error) {
	var result ConnectionGroup
	if err := c.get(ctx, c.dataPath("connectionGroups", id), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get connection group %s: %w", id, err)
	}
	return &result, nil
}

// UpdateConnectionGroup replaces the connection group identified by id with
// the supplied ConnectionGroup. The identifier field within group is ignored;
// id is used.
func (c *Client) UpdateConnectionGroup(ctx context.Context, id string, group ConnectionGroup) error {
	if err := c.put(ctx, c.dataPath("connectionGroups", id), group); err != nil {
		return fmt.Errorf("guacamole: update connection group %s: %w", id, err)
	}
	return nil
}

// DeleteConnectionGroup permanently removes the connection group with the
// given identifier.
func (c *Client) DeleteConnectionGroup(ctx context.Context, id string) error {
	if err := c.delete(ctx, c.dataPath("connectionGroups", id)); err != nil {
		return fmt.Errorf("guacamole: delete connection group %s: %w", id, err)
	}
	return nil
}
