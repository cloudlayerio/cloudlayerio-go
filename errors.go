package cloudlayer

import (
	"encoding/json"
	"fmt"
)

// APIError represents an error response from the CloudLayer.io API.
type APIError struct {
	StatusCode    int             `json:"statusCode"`
	StatusText    string          `json:"statusText"`
	Message       string          `json:"message"`
	Body          json.RawMessage `json:"body,omitempty"`
	RequestPath   string          `json:"requestPath"`
	RequestMethod string          `json:"requestMethod"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("cloudlayer: API error %d %s: %s (path: %s %s)",
		e.StatusCode, e.StatusText, e.Message, e.RequestMethod, e.RequestPath)
}

// AuthError is returned when the API rejects authentication (HTTP 401 or 403).
type AuthError struct {
	APIError
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("cloudlayer: authentication error %d: %s (path: %s %s)",
		e.StatusCode, e.Message, e.RequestMethod, e.RequestPath)
}

// RateLimitError is returned when the API rate limit is exceeded (HTTP 429).
type RateLimitError struct {
	APIError
	// RetryAfter is the number of seconds to wait before retrying, if provided
	// by the server. Nil if the header was not present.
	RetryAfter *int
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter != nil {
		return fmt.Sprintf("cloudlayer: rate limit exceeded, retry after %ds (path: %s %s)",
			*e.RetryAfter, e.RequestMethod, e.RequestPath)
	}
	return fmt.Sprintf("cloudlayer: rate limit exceeded (path: %s %s)",
		e.RequestMethod, e.RequestPath)
}

// TimeoutError is returned when a request times out due to the SDK's own
// timeout logic (not context cancellation — use errors.Is(err, context.Canceled)
// or errors.Is(err, context.DeadlineExceeded) for context errors).
type TimeoutError struct {
	// Timeout is the timeout in milliseconds.
	Timeout       int
	RequestPath   string
	RequestMethod string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("cloudlayer: request timed out after %dms (path: %s %s)",
		e.Timeout, e.RequestMethod, e.RequestPath)
}

// NetworkError is returned when a request fails due to a network-level problem
// (DNS resolution, connection refused, TLS errors, etc.).
type NetworkError struct {
	Message       string
	Err           error
	RequestPath   string
	RequestMethod string
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("cloudlayer: network error: %s (path: %s %s)",
		e.Message, e.RequestMethod, e.RequestPath)
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// ValidationError is returned when client-side input validation fails.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("cloudlayer: validation error on %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("cloudlayer: validation error: %s", e.Message)
}

// ConfigError is returned when client configuration is invalid.
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("cloudlayer: config error: %s", e.Message)
}
