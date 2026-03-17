package cloudlayer

import (
	"net/http"
	"net/url"
	"time"
)

// Version is the current version of the cloudlayer-go SDK.
const Version = "0.1.0"

const defaultBaseURL = "https://api.cloudlayer.io"

// Client is the CloudLayer.io API client. Create one with [NewClient].
type Client struct {
	apiKey        string
	apiVersion    APIVersion
	baseURL       string
	httpClient    *http.Client
	maxRetries    int
	userAgent     string
	customHeaders map[string]string
}

// NewClient creates a new CloudLayer.io API client.
//
// The apiKey and apiVersion parameters are required. Use [V1] or [V2] for the
// API version. Functional options can be used to customize the client.
//
//	client, err := cloudlayer.NewClient("your-api-key", cloudlayer.V2)
func NewClient(apiKey string, apiVersion APIVersion, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, &ConfigError{Message: "apiKey must not be empty"}
	}
	if apiVersion != V1 && apiVersion != V2 {
		return nil, &ConfigError{Message: "apiVersion must be V1 or V2"}
	}

	c := &Client{
		apiKey:     apiKey,
		apiVersion: apiVersion,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 2,
		userAgent:  "cloudlayerio-go/" + Version,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Validate baseURL after options are applied
	if _, err := url.ParseRequestURI(c.baseURL); err != nil {
		return nil, &ConfigError{Message: "baseURL is not a valid URL: " + c.baseURL}
	}

	// Validate timeout
	if c.httpClient.Timeout <= 0 {
		return nil, &ConfigError{Message: "timeout must be greater than 0"}
	}

	return c, nil
}
