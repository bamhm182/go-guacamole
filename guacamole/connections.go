package guacamole

import (
	"context"
	"fmt"
)

// ListConnections returns all connections visible to the authenticated user,
// keyed by connection identifier.
func (c *Client) ListConnections(ctx context.Context) (map[string]Connection, error) {
	var result map[string]Connection
	if err := c.get(ctx, c.dataPath("connections"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: list connections: %w", err)
	}
	return result, nil
}

// CreateConnection creates a new connection and returns the created resource
// with its server-assigned identifier.
func (c *Client) CreateConnection(ctx context.Context, conn Connection) (*Connection, error) {
	var result Connection
	if err := c.post(ctx, c.dataPath("connections"), conn, &result); err != nil {
		return nil, fmt.Errorf("guacamole: create connection: %w", err)
	}
	return &result, nil
}

// GetConnection retrieves the connection with the given identifier.
// Note: the returned Connection does not include protocol parameters; call
// GetConnectionParameters separately to obtain those.
func (c *Client) GetConnection(ctx context.Context, id string) (*Connection, error) {
	var result Connection
	if err := c.get(ctx, c.dataPath("connections", id), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get connection %s: %w", id, err)
	}
	return &result, nil
}

// GetConnectionParameters returns the protocol-specific parameters for the
// connection with the given identifier (e.g. hostname, port, username).
func (c *Client) GetConnectionParameters(ctx context.Context, id string) (map[string]string, error) {
	var result map[string]string
	if err := c.get(ctx, c.dataPath("connections", id, "parameters"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get connection parameters %s: %w", id, err)
	}
	return result, nil
}

// UpdateConnection replaces the connection identified by id with the supplied
// Connection. The identifier field within conn is ignored; id is used.
func (c *Client) UpdateConnection(ctx context.Context, id string, conn Connection) error {
	if err := c.put(ctx, c.dataPath("connections", id), conn); err != nil {
		return fmt.Errorf("guacamole: update connection %s: %w", id, err)
	}
	return nil
}

// DeleteConnection permanently removes the connection with the given
// identifier.
func (c *Client) DeleteConnection(ctx context.Context, id string) error {
	if err := c.delete(ctx, c.dataPath("connections", id)); err != nil {
		return fmt.Errorf("guacamole: delete connection %s: %w", id, err)
	}
	return nil
}
