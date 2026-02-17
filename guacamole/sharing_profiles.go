package guacamole

import (
	"context"
	"fmt"
)

// ListSharingProfiles returns all sharing profiles visible to the authenticated
// user, keyed by identifier.
func (c *Client) ListSharingProfiles(ctx context.Context) (map[string]SharingProfile, error) {
	var result map[string]SharingProfile
	if err := c.get(ctx, c.dataPath("sharingProfiles"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: list sharing profiles: %w", err)
	}
	return result, nil
}

// CreateSharingProfile creates a new sharing profile and returns the created
// resource with its server-assigned identifier.
func (c *Client) CreateSharingProfile(ctx context.Context, profile SharingProfile) (*SharingProfile, error) {
	var result SharingProfile
	if err := c.post(ctx, c.dataPath("sharingProfiles"), profile, &result); err != nil {
		return nil, fmt.Errorf("guacamole: create sharing profile: %w", err)
	}
	return &result, nil
}

// GetSharingProfile retrieves the sharing profile with the given identifier.
// Note: the returned SharingProfile does not include parameters; call
// GetSharingProfileParameters separately to obtain those.
func (c *Client) GetSharingProfile(ctx context.Context, id string) (*SharingProfile, error) {
	var result SharingProfile
	if err := c.get(ctx, c.dataPath("sharingProfiles", id), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get sharing profile %s: %w", id, err)
	}
	return &result, nil
}

// GetSharingProfileParameters returns the parameters for the sharing profile
// with the given identifier (e.g. {"read-only": "true"}).
func (c *Client) GetSharingProfileParameters(ctx context.Context, id string) (map[string]string, error) {
	var result map[string]string
	if err := c.get(ctx, c.dataPath("sharingProfiles", id, "parameters"), &result); err != nil {
		return nil, fmt.Errorf("guacamole: get sharing profile parameters %s: %w", id, err)
	}
	return result, nil
}

// UpdateSharingProfile replaces the sharing profile identified by id with the
// supplied SharingProfile. The identifier field within profile is ignored; id
// is used.
func (c *Client) UpdateSharingProfile(ctx context.Context, id string, profile SharingProfile) error {
	if err := c.put(ctx, c.dataPath("sharingProfiles", id), profile); err != nil {
		return fmt.Errorf("guacamole: update sharing profile %s: %w", id, err)
	}
	return nil
}

// DeleteSharingProfile permanently removes the sharing profile with the given
// identifier.
func (c *Client) DeleteSharingProfile(ctx context.Context, id string) error {
	if err := c.delete(ctx, c.dataPath("sharingProfiles", id)); err != nil {
		return fmt.Errorf("guacamole: delete sharing profile %s: %w", id, err)
	}
	return nil
}
