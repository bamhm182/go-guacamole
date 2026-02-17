// Package guacamole provides a Go client for the Apache Guacamole REST API,
// designed for use in Terraform providers and other infrastructure tooling.
//
// Usage:
//
//	client := guacamole.NewClient("http://localhost:8080/guacamole")
//	if err := client.Authenticate(ctx, "guacadmin", "guacadmin"); err != nil {
//	    log.Fatal(err)
//	}
//	conn, err := client.CreateConnection(ctx, guacamole.Connection{...})
package guacamole

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Client is a Guacamole REST API client. Create one with NewClient and call
// Authenticate before making resource requests.
type Client struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
	dataSource string
}

// NewClient creates a new Client targeting the given Guacamole base URL (e.g.
// "http://localhost:8080/guacamole"). The client uses a 30-second timeout by
// default.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithHTTPClient creates a new Client with a caller-supplied
// *http.Client. This is useful for supplying custom TLS configuration or
// transport-level logging.
func NewClientWithHTTPClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// Authenticate performs the Guacamole token exchange (POST /api/tokens) and
// stores the resulting token and data source for use in subsequent calls.
// It must be called before any resource method.
func (c *Client) Authenticate(ctx context.Context, username, password string) error {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/tokens",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("guacamole: build auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("guacamole: auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	var auth AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return fmt.Errorf("guacamole: decode auth response: %w", err)
	}

	c.authToken = auth.AuthToken
	c.dataSource = auth.DataSource
	return nil
}

// Logout invalidates the current session token (DELETE /api/session).
func (c *Client) Logout(ctx context.Context) error {
	return c.delete(ctx, "/api/session")
}

// DataSource returns the data source string that was received during
// authentication (e.g. "postgresql"). This is used in all API paths.
func (c *Client) DataSource() string {
	return c.dataSource
}

// AuthToken returns the current authentication token.
func (c *Client) AuthToken() string {
	return c.authToken
}

// dataPath builds a URL path prefixed with the session data source segment,
// percent-encoding each segment so that identifiers containing spaces, @, or
// other reserved characters are handled correctly.
//
// Example: dataPath("users", "bob@example.com") →
//
//	"/api/session/data/postgresql/users/bob%40example.com"
func (c *Client) dataPath(segments ...string) string {
	parts := make([]string, 0, len(segments)+2)
	parts = append(parts, url.PathEscape(c.dataSource))
	for _, s := range segments {
		parts = append(parts, url.PathEscape(s))
	}
	return "/api/session/data/" + path.Join(parts...)
}

// ── HTTP helpers ─────────────────────────────────────────────────────────────

// get makes a GET request and decodes the JSON response body into out.
func (c *Client) get(ctx context.Context, path string, out interface{}) error {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

// post makes a POST request with a JSON body and decodes the JSON response
// into out (may be nil if no response body is expected).
func (c *Client) post(ctx context.Context, path string, body, out interface{}) error {
	resp, err := c.do(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// put makes a PUT request with a JSON body. Guacamole returns 204 No Content
// for successful updates.
func (c *Client) put(ctx context.Context, path string, body interface{}) error {
	resp, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// delete makes a DELETE request.
func (c *Client) delete(ctx context.Context, path string) error {
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// patch makes a PATCH request with a JSON Patch body. Guacamole uses JSON
// Patch (RFC 6902) for permission and group-membership modifications.
func (c *Client) patch(ctx context.Context, path string, ops []PatchOperation) error {
	resp, err := c.do(ctx, http.MethodPatch, path, ops)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// do is the low-level HTTP request method. It serialises body to JSON (if
// non-nil), attaches the auth token header, executes the request, and returns
// an error for any non-2xx response.
func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("guacamole: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("guacamole: build request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.authToken != "" {
		req.Header.Set("Guacamole-Token", c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("guacamole: %s %s: %w", method, path, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, c.parseError(resp)
	}

	return resp, nil
}

// parseError reads an API error response body and returns an *APIError.
func (c *Client) parseError(resp *http.Response) error {
	apiErr := &APIError{HTTPStatus: resp.StatusCode}
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		apiErr.Message = http.StatusText(resp.StatusCode)
		return apiErr
	}
	if err := json.Unmarshal(body, apiErr); err != nil {
		apiErr.Message = string(body)
	}
	return apiErr
}
