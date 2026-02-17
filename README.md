# go-guacamole

A Go client library for the [Apache Guacamole](https://guacamole.apache.org/) REST API, designed for use in Terraform providers and infrastructure automation tools.

## Features

- Full CRUD for all Guacamole resources: connections, connection groups, users, user groups, and sharing profiles
- Permission management via JSON Patch helpers
- Group membership and nesting (user→group, group→group)
- Session and history inspection (active connections, login history, session history)
- Correct handling of Guacamole's nullable attribute JSON (`null` → `""`)
- URL-encoding of identifiers with spaces or special characters
- Zero external dependencies — standard library only
- Context support on every method

## Installation

```sh
go get github.com/bamhm182/go-guacamole
```

## Quick start

```go
import "github.com/bamhm182/go-guacamole/guacamole"

client := guacamole.NewClient("http://localhost:8080/guacamole")
if err := client.Authenticate(ctx, "guacadmin", "guacadmin"); err != nil {
    log.Fatal(err)
}

// Create an SSH connection
conn, err := client.CreateConnection(ctx, guacamole.Connection{
    Name:             "My SSH Server",
    Protocol:         "ssh",
    ParentIdentifier: guacamole.RootConnectionGroupIdentifier,
    Parameters: map[string]string{
        "hostname": "192.168.1.100",
        "port":     "22",
        "username": "admin",
    },
})
```

## Resources

### Connections

```go
// List all connections (keyed by identifier)
conns, err := client.ListConnections(ctx)

// Create
conn, err := client.CreateConnection(ctx, guacamole.Connection{
    Name:             "Production DB",
    Protocol:         "rdp",
    ParentIdentifier: "ROOT",
    Parameters: map[string]string{
        "hostname":    "10.0.0.5",
        "port":        "3389",
        "username":    "Administrator",
        "ignore-cert": "true",
    },
    Attributes: guacamole.NullableStringMap{
        "max-connections":          "5",
        "max-connections-per-user": "1",
    },
})

// Read — note: parameters are fetched separately
conn, err := client.GetConnection(ctx, "42")
params, err := client.GetConnectionParameters(ctx, "42")

// Update
err = client.UpdateConnection(ctx, "42", updatedConn)

// Delete
err = client.DeleteConnection(ctx, "42")
```

`Parameters` holds the protocol-specific connection settings (hostname, port, credentials, etc.). The Guacamole API returns them from a separate `/parameters` endpoint, so `GetConnection` does not populate `Parameters`; call `GetConnectionParameters` explicitly and merge as needed.

### Connection groups

Connection groups are either `"ORGANIZATIONAL"` (folder-like) or `"BALANCING"` (load-balanced pool).

```go
// Full hierarchy from ROOT — most efficient way to discover the topology
tree, err := client.GetConnectionGroupTree(ctx, guacamole.RootConnectionGroupIdentifier)

// Subtree rooted at a specific group
subtree, err := client.GetConnectionGroupTree(ctx, "7")

// CRUD
cg, err := client.CreateConnectionGroup(ctx, guacamole.ConnectionGroup{
    Name:             "Data Centers",
    Type:             guacamole.ConnectionGroupTypeOrganizational,
    ParentIdentifier: guacamole.RootConnectionGroupIdentifier,
})
cg, err  = client.GetConnectionGroup(ctx, cg.Identifier)
err      = client.UpdateConnectionGroup(ctx, cg.Identifier, updatedCG)
err      = client.DeleteConnectionGroup(ctx, cg.Identifier)
```

### Users

```go
// CRUD
user, err := client.CreateUser(ctx, guacamole.User{
    Username: "alice",
    Password: "s3cr3t",
    Attributes: guacamole.NullableStringMap{
        "guac-full-name":    "Alice Smith",
        "guac-email-address": "alice@example.com",
    },
})
user, err  = client.GetUser(ctx, "alice")
err        = client.UpdateUser(ctx, "alice", updatedUser)
err        = client.DeleteUser(ctx, "alice")

// Permissions
perms, err    := client.GetUserPermissions(ctx, "alice")
effPerms, err := client.GetUserEffectivePermissions(ctx, "alice") // includes group inheritance

err = client.UpdateUserPermissions(ctx, "alice", []guacamole.PatchOperation{
    guacamole.AddConnectionPermission("42", guacamole.PermissionRead),
    guacamole.AddSystemPermission(guacamole.SystemPermissionCreateConnection),
})

// Group membership
groups, err := client.GetUserGroups(ctx, "alice")
err = client.UpdateUserGroups(ctx, "alice", []guacamole.PatchOperation{
    guacamole.AddGroupMembership("admins"),
    guacamole.RemoveGroupMembership("temps"),
})
```

### User groups

```go
// CRUD
ug, err := client.CreateUserGroup(ctx, guacamole.UserGroup{Identifier: "admins"})
ug, err  = client.GetUserGroup(ctx, "admins")
err      = client.UpdateUserGroup(ctx, "admins", updatedUG)
err      = client.DeleteUserGroup(ctx, "admins")

// Permissions
err = client.UpdateUserGroupPermissions(ctx, "admins", []guacamole.PatchOperation{
    guacamole.AddConnectionPermission("42", guacamole.PermissionRead),
})

// Member users (users inside this group)
users, err := client.GetUserGroupMemberUsers(ctx, "admins")
err = client.UpdateUserGroupMemberUsers(ctx, "admins", []guacamole.PatchOperation{
    guacamole.AddGroupMembership("alice"),
})

// Member groups (child groups nested inside this group)
children, err := client.GetUserGroupMemberGroups(ctx, "all-staff")
err = client.UpdateUserGroupMemberGroups(ctx, "all-staff", []guacamole.PatchOperation{
    guacamole.AddGroupMembership("devs"),
})

// Parent groups (groups this group belongs to — the inverse of the above)
parents, err := client.GetUserGroupParentGroups(ctx, "devs")
err = client.UpdateUserGroupParentGroups(ctx, "devs", []guacamole.PatchOperation{
    guacamole.AddGroupMembership("all-staff"),
})
```

### Sharing profiles

```go
sp, err := client.CreateSharingProfile(ctx, guacamole.SharingProfile{
    Name:                        "Read-only Share",
    PrimaryConnectionIdentifier: "42",
    Parameters:                  map[string]string{"read-only": "true"},
})
sp, err    = client.GetSharingProfile(ctx, sp.Identifier)
params, err = client.GetSharingProfileParameters(ctx, sp.Identifier)
err        = client.UpdateSharingProfile(ctx, sp.Identifier, updatedSP)
err        = client.DeleteSharingProfile(ctx, sp.Identifier)
```

### Active connections and history

```go
// Currently active sessions
active, err := client.ListActiveConnections(ctx)
err = client.KillActiveConnection(ctx, "session-id")

// Historical records
connHistory, err := client.ListConnectionHistory(ctx, "-startDate") // "" for default order
perConnHistory, err := client.GetConnectionHistory(ctx, "42")
userHistory, err := client.GetUserHistory(ctx, "alice")
```

### Current user (self)

```go
self, err         := client.GetSelf(ctx)
perms, err        := client.GetSelfPermissions(ctx)
effPerms, err     := client.GetSelfEffectivePermissions(ctx)
```

## Permission management

Permissions are modified using JSON Patch–style operations. Helper functions produce the correct `PatchOperation` values:

| Helper | Description |
|---|---|
| `AddConnectionPermission(id, perm)` | Grant connection permission |
| `RemoveConnectionPermission(id, perm)` | Revoke connection permission |
| `AddConnectionGroupPermission(id, perm)` | Grant connection group permission |
| `RemoveConnectionGroupPermission(id, perm)` | Revoke connection group permission |
| `AddSharingProfilePermission(id, perm)` | Grant sharing profile permission |
| `RemoveSharingProfilePermission(id, perm)` | Revoke sharing profile permission |
| `AddUserPermission(username, perm)` | Grant permission on a user account |
| `RemoveUserPermission(username, perm)` | Revoke permission on a user account |
| `AddUserGroupPermission(id, perm)` | Grant permission on a user group |
| `RemoveUserGroupPermission(id, perm)` | Revoke permission on a user group |
| `AddSystemPermission(perm)` | Grant a system-level permission |
| `RemoveSystemPermission(perm)` | Revoke a system-level permission |
| `AddGroupMembership(id)` | Add to a membership list |
| `RemoveGroupMembership(id)` | Remove from a membership list |

**Object permission constants:** `PermissionRead`, `PermissionUpdate`, `PermissionDelete`, `PermissionAdminister`

**System permission constants:** `SystemPermissionCreateUser`, `SystemPermissionCreateUserGroup`, `SystemPermissionCreateConnection`, `SystemPermissionCreateConnectionGroup`, `SystemPermissionCreateSharingProfile`, `SystemPermissionAdminister`

## Error handling

All methods return `nil` on success and a wrapped `*APIError` on failure. Use the package-level helpers to inspect specific error types:

```go
conn, err := client.GetConnection(ctx, id)
if guacamole.IsNotFound(err) {
    // Resource doesn't exist — safe to create or treat as already deleted
} else if guacamole.IsPermissionDenied(err) {
    // Caller lacks permission
} else if err != nil {
    // Network error, server error, etc.
}
```

Both helpers use `errors.As` internally, so they work correctly when the `*APIError` has been wrapped by `fmt.Errorf("... %w", err)`.

`*APIError` fields:

```go
type APIError struct {
    Message    string // human-readable description
    Type       string // "NOT_FOUND", "PERMISSION_DENIED", etc.
    HTTPStatus int    // raw HTTP status code
}
```

## Custom HTTP client

Supply your own `*http.Client` to configure TLS, proxies, or transport-level logging:

```go
httpClient := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    },
    Timeout: 60 * time.Second,
}
client := guacamole.NewClientWithHTTPClient("https://guacamole.example.com/guacamole", httpClient)
```

## Notes for Terraform provider authors

- **`attributes` is always serialised.** `NullableStringMap` marshals as `{}` when nil. Guacamole returns HTTP 500 if the field is missing or `null`, so never use `omitempty` on attributes fields.
- **`parameters` is fetched separately.** `GetConnection` and `GetSharingProfile` do not return protocol parameters. Call `GetConnectionParameters` / `GetSharingProfileParameters` and merge into your Terraform state.
- **Identifiers are numeric strings for connections and groups** (e.g. `"42"`), but free-form strings for users and user groups. URL-encoding is handled automatically by the client.
- **`IsNotFound`** is the right check for Terraform's `resource.RetryContext` and for detecting resources deleted outside Terraform.
- **`dataSource`** is set automatically from the `Authenticate` response. It reflects the active database backend (e.g. `"postgresql"`). There is no need to set it manually.
