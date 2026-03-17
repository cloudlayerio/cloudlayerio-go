package cloudlayer

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewClient_Valid(t *testing.T) {
	c, err := NewClient("test-key", V2)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c.apiKey != "test-key" {
		t.Errorf("apiKey = %q, want test-key", c.apiKey)
	}
	if c.apiVersion != V2 {
		t.Errorf("apiVersion = %q, want v2", c.apiVersion)
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
	if c.maxRetries != 2 {
		t.Errorf("maxRetries = %d, want 2", c.maxRetries)
	}
}

func TestNewClient_EmptyAPIKey(t *testing.T) {
	_, err := NewClient("", V2)
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Errorf("expected *ConfigError, got %T", err)
	}
}

func TestNewClient_InvalidAPIVersion(t *testing.T) {
	_, err := NewClient("key", APIVersion("v3"))
	if err == nil {
		t.Fatal("expected error for invalid API version")
	}
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Errorf("expected *ConfigError, got %T", err)
	}
}

func TestNewClient_V1(t *testing.T) {
	c, err := NewClient("key", V1)
	if err != nil {
		t.Fatal(err)
	}
	if c.apiVersion != V1 {
		t.Errorf("apiVersion = %q, want v1", c.apiVersion)
	}
}

func TestNewClient_V2(t *testing.T) {
	c, err := NewClient("key", V2)
	if err != nil {
		t.Fatal(err)
	}
	if c.apiVersion != V2 {
		t.Errorf("apiVersion = %q, want v2", c.apiVersion)
	}
}

func TestNewClient_WithBaseURL(t *testing.T) {
	c, err := NewClient("key", V2, WithBaseURL("https://custom.api.com"))
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL != "https://custom.api.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestNewClient_WithTimeout(t *testing.T) {
	c, err := NewClient("key", V2, WithTimeout(10*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient.Timeout != 10*time.Second {
		t.Errorf("timeout = %v", c.httpClient.Timeout)
	}
}

func TestNewClient_WithMaxRetries_Zero(t *testing.T) {
	c, err := NewClient("key", V2, WithMaxRetries(0))
	if err != nil {
		t.Fatal(err)
	}
	if c.maxRetries != 0 {
		t.Errorf("maxRetries = %d, want 0", c.maxRetries)
	}
}

func TestNewClient_WithMaxRetries_ClampedToFive(t *testing.T) {
	c, err := NewClient("key", V2, WithMaxRetries(10))
	if err != nil {
		t.Fatal(err)
	}
	if c.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want 5", c.maxRetries)
	}
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 60 * time.Second}
	c, err := NewClient("key", V2, WithHTTPClient(custom))
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient != custom {
		t.Error("httpClient not set to custom client")
	}
}

func TestNewClient_WithHeaders(t *testing.T) {
	headers := map[string]string{"X-Custom": "test"}
	c, err := NewClient("key", V2, WithHeaders(headers))
	if err != nil {
		t.Fatal(err)
	}
	if c.customHeaders["X-Custom"] != "test" {
		t.Error("custom headers not set")
	}
}

func TestNewClient_WithUserAgent(t *testing.T) {
	c, err := NewClient("key", V2, WithUserAgent("my-app/1.0"))
	if err != nil {
		t.Fatal(err)
	}
	if c.userAgent != "my-app/1.0" {
		t.Errorf("userAgent = %q", c.userAgent)
	}
}
