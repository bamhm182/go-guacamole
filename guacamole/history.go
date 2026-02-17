package guacamole

import (
	"context"
	"fmt"
)

// HistoryEntry represents a single recorded connection session or login event.
type HistoryEntry struct {
	// Identifier is the unique identifier for this history entry.
	Identifier string `json:"identifier"`
	// UUID is the universally unique identifier for this history entry.
	UUID string `json:"uuid"`
	// Username is the name of the user who initiated the session.
	Username string `json:"username"`
	// RemoteHost is the IP address of the client that connected.
	RemoteHost string `json:"remoteHost"`
	// StartDate is the session start time in milliseconds since the Unix epoch.
	StartDate int64 `json:"startDate"`
	// EndDate is the session end time in milliseconds since the Unix epoch.
	// It will be zero if the session is still active.
	EndDate int64 `json:"endDate"`
	// Active indicates whether the session is still in progress.
	Active bool `json:"active"`
}

// ListConnectionHistory returns the global history of all connection sessions,
// optionally ordered by start date. Pass order as "-startDate" for descending
// or "startDate" for ascending; pass an empty string for the server default.
func (c *Client) ListConnectionHistory(ctx context.Context, order string) ([]HistoryEntry, error) {
	path := c.dataPath("history", "connections")
	if order != "" {
		path += "?order=" + order
	}
	var result []HistoryEntry
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("guacamole: list connection history: %w", err)
	}
	return result, nil
}

// GetConnectionHistory returns the session history for a specific connection.
func (c *Client) GetConnectionHistory(ctx context.Context, connectionID string) ([]HistoryEntry, error) {
	var result []HistoryEntry
	if err := c.get(ctx, c.dataPath("connections", connectionID, "history"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get connection history %s: %w", connectionID, err)
	}
	return result, nil
}

// GetUserHistory returns the login history for a specific user.
func (c *Client) GetUserHistory(ctx context.Context, username string) ([]HistoryEntry, error) {
	var result []HistoryEntry
	if err := c.get(ctx, c.dataPath("users", username, "history"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get user history %s: %w", username, err)
	}
	return result, nil
}
