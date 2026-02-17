package guacamole

import (
	"errors"
	"fmt"
)

// Guacamole API error type constants.
const (
	ErrTypeNotFound        = "NOT_FOUND"
	ErrTypePermissionDenied = "PERMISSION_DENIED"
)

// APIError represents an error response from the Guacamole REST API.
type APIError struct {
	// Message is the human-readable error description.
	Message string `json:"message"`
	// Type is the machine-readable error category (e.g. "NOT_FOUND").
	Type string `json:"type"`
	// HTTPStatus is the HTTP status code of the response.
	HTTPStatus int `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("guacamole API error (HTTP %d, type %s): %s", e.HTTPStatus, e.Type, e.Message)
}

// IsNotFound reports whether the error indicates the requested resource does
// not exist (HTTP 404 / type "NOT_FOUND").
func (e *APIError) IsNotFound() bool {
	return e.Type == ErrTypeNotFound
}

// IsPermissionDenied reports whether the error indicates the caller lacks
// permission to perform the requested operation (HTTP 403 / type
// "PERMISSION_DENIED").
func (e *APIError) IsPermissionDenied() bool {
	return e.Type == ErrTypePermissionDenied
}

// IsNotFound is a convenience function that returns true when err (or any
// error in its chain) is an *APIError with type "NOT_FOUND". It returns false
// for any other error type, including nil.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsNotFound()
	}
	return false
}

// IsPermissionDenied is a convenience function that returns true when err (or
// any error in its chain) is an *APIError with type "PERMISSION_DENIED".
func IsPermissionDenied(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsPermissionDenied()
	}
	return false
}
