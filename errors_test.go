package cloudlayer

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode:    400,
		StatusText:    "Bad Request",
		Message:       "invalid html field",
		RequestPath:   "/html/pdf",
		RequestMethod: "POST",
	}
	s := err.Error()
	if s == "" {
		t.Error("Error() returned empty string")
	}
}

func TestAuthError_ErrorsAs(t *testing.T) {
	err := &AuthError{APIError: APIError{StatusCode: 401, Message: "unauthorized"}}

	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Error("errors.As should match *AuthError")
	}

	// AuthError embeds APIError — access the embedded fields directly
	if authErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d", authErr.StatusCode)
	}
	if authErr.Message != "unauthorized" {
		t.Errorf("Message = %q", authErr.Message)
	}
}

func TestRateLimitError_RetryAfter(t *testing.T) {
	retryAfter := 30
	err := &RateLimitError{
		APIError:   APIError{StatusCode: 429, Message: "rate limited"},
		RetryAfter: &retryAfter,
	}
	s := err.Error()
	if s == "" {
		t.Error("Error() returned empty string")
	}
	if err.RetryAfter == nil || *err.RetryAfter != 30 {
		t.Errorf("RetryAfter = %v", err.RetryAfter)
	}
}

func TestNetworkError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("connection refused")
	err := &NetworkError{
		Message: "network error",
		Err:     inner,
	}
	if !errors.Is(err, inner) {
		t.Error("errors.Is should find inner error")
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{Field: "url", Message: "url must not be empty"}
	s := err.Error()
	if s == "" {
		t.Error("Error() returned empty string")
	}
}

func TestValidationError_NoField(t *testing.T) {
	err := &ValidationError{Message: "something is wrong"}
	s := err.Error()
	if s == "" {
		t.Error("Error() returned empty string")
	}
}

func TestConfigError_Error(t *testing.T) {
	err := &ConfigError{Message: "apiKey must not be empty"}
	s := err.Error()
	if s == "" {
		t.Error("Error() returned empty string")
	}
}

func TestTimeoutError_Error(t *testing.T) {
	err := &TimeoutError{
		Timeout:       30000,
		RequestPath:   "/jobs/123",
		RequestMethod: "GET",
	}
	s := err.Error()
	if s == "" {
		t.Error("Error() returned empty string")
	}
}
