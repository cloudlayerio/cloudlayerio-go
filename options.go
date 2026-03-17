package cloudlayer

import (
	"net/http"
	"time"
)

// ClientOption configures a [Client]. Use With* functions to create options.
type ClientOption func(*Client)

// WithBaseURL sets the API base URL. Defaults to "https://api.cloudlayer.io".
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom [http.Client] for requests.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithTimeout sets the HTTP client timeout. Defaults to 30 seconds.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithMaxRetries sets the maximum number of retries for retryable requests.
// Clamped to [0, 5]. Defaults to 2.
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) {
		if n < 0 {
			n = 0
		}
		if n > 5 {
			n = 5
		}
		c.maxRetries = n
	}
}

// WithUserAgent sets the User-Agent header. Defaults to "cloudlayerio-go/{version}".
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// WithHeaders sets additional headers to include on every request.
func WithHeaders(h map[string]string) ClientOption {
	return func(c *Client) {
		c.customHeaders = h
	}
}
