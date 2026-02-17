package guacamole

import "encoding/json"

// NullableStringMap is a map[string]string that correctly round-trips with the
// Guacamole API's attribute JSON:
//
//   - On unmarshal: null values are silently converted to empty strings so
//     callers always get a plain map[string]string regardless of which keys
//     the server chose to return as null vs. "".
//
//   - On marshal: a nil map serializes as {} (not null and not omitted).
//     Guacamole returns HTTP 500 if the attributes field is missing or null.
type NullableStringMap map[string]string

func (m *NullableStringMap) UnmarshalJSON(data []byte) error {
	var raw map[string]*string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*m = make(NullableStringMap, len(raw))
	for k, v := range raw {
		if v != nil {
			(*m)[k] = *v
		}
	}
	return nil
}

// MarshalJSON always encodes as a JSON object, even when nil, because
// Guacamole rejects requests where the attributes field is missing or null.
func (m NullableStringMap) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]string(m))
}

// PatchOperation represents a single RFC 6902 JSON Patch operation. Guacamole
// uses JSON Patch for permission and membership modifications.
type PatchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// AuthResponse is returned by POST /api/tokens.
type AuthResponse struct {
	AuthToken            string   `json:"authToken"`
	Username             string   `json:"username"`
	DataSource           string   `json:"dataSource"`
	AvailableDataSources []string `json:"availableDataSources"`
}

// Connection represents a Guacamole remote desktop connection.
//
// Parameters holds the protocol-specific settings (hostname, port, credentials,
// etc.) and is only populated when explicitly requested via the /parameters
// endpoint. On create/update, set Parameters to supply these values; on read,
// call GetConnectionParameters separately.
type Connection struct {
	Identifier        string            `json:"identifier,omitempty"`
	Name              string            `json:"name"`
	ParentIdentifier  string            `json:"parentIdentifier,omitempty"`
	Protocol          string            `json:"protocol"`
	Parameters        map[string]string `json:"parameters,omitempty"`
	Attributes        NullableStringMap `json:"attributes"`
	ActiveConnections int               `json:"activeConnections,omitempty"`
}

// ConnectionGroup represents an organizational or load-balancing group of
// connections.
//
// Type must be either "ORGANIZATIONAL" or "BALANCING".
type ConnectionGroup struct {
	Identifier            string            `json:"identifier,omitempty"`
	Name                  string            `json:"name"`
	ParentIdentifier      string            `json:"parentIdentifier,omitempty"`
	Type                  string            `json:"type"`
	Attributes            NullableStringMap `json:"attributes"`
	ActiveConnections     int               `json:"activeConnections,omitempty"`
	ChildConnections      []Connection      `json:"childConnections,omitempty"`
	ChildConnectionGroups []ConnectionGroup `json:"childConnectionGroups,omitempty"`
}

// User represents a Guacamole user account.
//
// Password is write-only: it is accepted on create/update but never returned
// by GET. Attributes contains optional profile and restriction fields; null
// values from the API are normalised to empty strings.
type User struct {
	Username   string            `json:"username"`
	Password   string            `json:"password,omitempty"`
	Disabled   bool              `json:"disabled,omitempty"`
	Attributes NullableStringMap `json:"attributes"`
	LastActive int64             `json:"lastActive,omitempty"`
}

// UserGroup represents a Guacamole user group.
type UserGroup struct {
	Identifier string            `json:"identifier"`
	Disabled   bool              `json:"disabled,omitempty"`
	Attributes NullableStringMap `json:"attributes"`
}

// SharingProfile represents a sharing profile attached to a connection. It
// defines a secondary set of connection parameters used when sharing a
// session, most commonly {"read-only": "true"}.
type SharingProfile struct {
	Identifier                  string            `json:"identifier,omitempty"`
	Name                        string            `json:"name"`
	PrimaryConnectionIdentifier string            `json:"primaryConnectionIdentifier"`
	Parameters                  map[string]string `json:"parameters,omitempty"`
	Attributes                  NullableStringMap `json:"attributes"`
}

// Permissions holds the full permission set for a user or user group.
//
// Map keys for *Permissions fields are resource identifiers (connection ID,
// username, etc.). Values are slices of permission strings: "READ", "UPDATE",
// "DELETE", "ADMINISTER".
//
// SystemPermissions contains zero or more of: "CREATE_USER",
// "CREATE_USER_GROUP", "CREATE_CONNECTION", "CREATE_CONNECTION_GROUP",
// "CREATE_SHARING_PROFILE", "ADMINISTER".
type Permissions struct {
	ConnectionPermissions       map[string][]string `json:"connectionPermissions"`
	ConnectionGroupPermissions  map[string][]string `json:"connectionGroupPermissions"`
	SharingProfilePermissions   map[string][]string `json:"sharingProfilePermissions"`
	ActiveConnectionPermissions map[string][]string `json:"activeConnectionPermissions"`
	UserPermissions             map[string][]string `json:"userPermissions"`
	UserGroupPermissions        map[string][]string `json:"userGroupPermissions"`
	SystemPermissions           []string            `json:"systemPermissions"`
}
