package guacamole

import (
	"context"
	"fmt"
)

// ActiveConnection represents a currently-active remote desktop session.
type ActiveConnection struct {
	Identifier        string `json:"identifier"`
	ConnectionIdentifier string `json:"connectionIdentifier"`
	StartDate         int64  `json:"startDate"`
	RemoteHost        string `json:"remoteHost"`
	Username          string `json:"username"`
	Active            bool   `json:"active"`
}

// ListActiveConnections returns all currently-active sessions, keyed by
// active-connection identifier. The map is empty when no sessions are open.
func (c *Client) ListActiveConnections(ctx context.Context) (map[string]ActiveConnection, error) {
	var result map[string]ActiveConnection
	if err := c.get(ctx, c.dataPath("activeConnections"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: list active connections: %w", err)
	}
	return result, nil
}

// KillActiveConnection forcibly terminates the active session with the given
// identifier.
func (c *Client) KillActiveConnection(ctx context.Context, id string) error {
	if err := c.delete(ctx, c.dataPath("activeConnections", id)); err != nil {
		return fmt.Errorf("guacamole: kill active connection %s: %w", id, err)
	}
	return nil
}
