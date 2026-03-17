package cloudlayer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// requestOptions controls per-request behavior.
type requestOptions struct {
	params       map[string]string // query parameters
	retryable    bool              // eligible for auto-retry on 429/5xx
	absolutePath bool              // use path as-is, don't prepend apiVersion
}

// apiErrorBody is the shape of API error response JSON bodies.
type apiErrorBody struct {
	StatusCode *int    `json:"statusCode,omitempty"`
	ErrorField *string `json:"error,omitempty"`
	Message    *string `json:"message,omitempty"`
}

// doRequest performs a JSON API request with optional retry.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}, opts *requestOptions) (*ResponseHeaders, error) {
	if opts == nil {
		opts = &requestOptions{}
	}

	fullURL := c.buildURL(path, opts)

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	maxAttempts := 1
	if opts.retryable {
		maxAttempts = 1 + c.maxRetries
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := c.retryDelay(attempt, lastErr)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			// Re-create body reader for retry
			if body != nil {
				data, _ := json.Marshal(body)
				bodyReader = bytes.NewReader(data)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		c.setHeaders(req, "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			lastErr = &NetworkError{
				Message:       err.Error(),
				Err:           err,
				RequestPath:   path,
				RequestMethod: method,
			}
			if !opts.retryable {
				return nil, lastErr
			}
			continue
		}

		headers := parseResponseHeaders(resp)

		// Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if resp.StatusCode == 204 || result == nil {
				_ = resp.Body.Close()
				return headers, nil
			}
			defer func() { _ = resp.Body.Close() }()
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				return headers, &APIError{
					StatusCode:    resp.StatusCode,
					StatusText:    resp.Status,
					Message:       "failed to decode response: " + err.Error(),
					RequestPath:   path,
					RequestMethod: method,
				}
			}
			return headers, nil
		}

		// Error — read body for error details
		apiErr := c.buildError(resp, path, method)
		_ = resp.Body.Close()

		// Should we retry?
		if opts.retryable && isRetryableStatus(resp.StatusCode) && attempt < maxAttempts-1 {
			lastErr = apiErr
			continue
		}

		return headers, apiErr
	}

	return nil, lastErr
}

// doRawRequest performs a request and returns the raw response for binary handling.
// The caller is responsible for closing resp.Body.
func (c *Client) doRawRequest(ctx context.Context, method, path string, body interface{}, opts *requestOptions) (*http.Response, *ResponseHeaders, error) {
	if opts == nil {
		opts = &requestOptions{}
	}

	fullURL := c.buildURL(path, opts)

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, nil, fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req, "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		return nil, nil, &NetworkError{
			Message:       err.Error(),
			Err:           err,
			RequestPath:   path,
			RequestMethod: method,
		}
	}

	headers := parseResponseHeaders(resp)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, headers, nil
	}

	apiErr := c.buildError(resp, path, method)
	_ = resp.Body.Close()
	return nil, headers, apiErr
}

// doMultipartRequest performs a multipart form request.
func (c *Client) doMultipartRequest(ctx context.Context, path string, file *FileInput, fields map[string]string, result interface{}, opts *requestOptions) (*ResponseHeaders, error) {
	if opts == nil {
		opts = &requestOptions{}
	}

	fullURL := c.buildURL(path, opts)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}
	if _, err := io.Copy(part, file.Reader); err != nil {
		return nil, fmt.Errorf("writing file content: %w", err)
	}

	// Add other fields
	for key, val := range fields {
		if err := writer.WriteField(key, val); err != nil {
			return nil, fmt.Errorf("writing form field %q: %w", key, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req, writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &NetworkError{
			Message:       err.Error(),
			Err:           err,
			RequestPath:   path,
			RequestMethod: http.MethodPost,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	headers := parseResponseHeaders(resp)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if resp.StatusCode == 204 || result == nil {
			return headers, nil
		}
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return headers, &APIError{
				StatusCode:    resp.StatusCode,
				StatusText:    resp.Status,
				Message:       "failed to decode response: " + err.Error(),
				RequestPath:   path,
				RequestMethod: http.MethodPost,
			}
		}
		return headers, nil
	}

	return headers, c.buildError(resp, path, http.MethodPost)
}

// buildURL constructs the full request URL.
func (c *Client) buildURL(path string, opts *requestOptions) string {
	var base string
	if opts != nil && opts.absolutePath {
		base = c.baseURL + path
	} else {
		base = c.baseURL + "/" + string(c.apiVersion) + path
	}

	if opts != nil && len(opts.params) > 0 {
		parts := make([]string, 0, len(opts.params))
		for k, v := range opts.params {
			parts = append(parts, k+"="+v)
		}
		base += "?" + strings.Join(parts, "&")
	}

	return base
}

// setHeaders sets common headers on a request.
func (c *Client) setHeaders(req *http.Request, contentType string) {
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range c.customHeaders {
		req.Header.Set(k, v)
	}
}

// buildError constructs an appropriate error type from an HTTP response.
func (c *Client) buildError(resp *http.Response, path, method string) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	message := http.StatusText(resp.StatusCode)

	// Try to parse error body for a message
	var errBody apiErrorBody
	if json.Unmarshal(body, &errBody) == nil {
		if errBody.Message != nil && *errBody.Message != "" {
			message = *errBody.Message
		} else if errBody.ErrorField != nil && *errBody.ErrorField != "" {
			message = *errBody.ErrorField
		}
	}

	baseErr := APIError{
		StatusCode:    resp.StatusCode,
		StatusText:    http.StatusText(resp.StatusCode),
		Message:       message,
		Body:          body,
		RequestPath:   path,
		RequestMethod: method,
	}

	switch resp.StatusCode {
	case 401, 403:
		return &AuthError{APIError: baseErr}
	case 429:
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return &RateLimitError{APIError: baseErr, RetryAfter: retryAfter}
	default:
		return &baseErr
	}
}

// retryDelay calculates the delay before the next retry attempt.
func (c *Client) retryDelay(attempt int, lastErr error) time.Duration {
	// Respect Retry-After header from rate limit errors
	if rl, ok := lastErr.(*RateLimitError); ok && rl.RetryAfter != nil {
		return time.Duration(*rl.RetryAfter) * time.Second
	}

	// Exponential backoff with jitter
	baseDelay := time.Second
	delay := baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
	jitter := time.Duration(rand.Int63n(int64(baseDelay)))
	return delay + jitter
}

// isRetryableStatus returns true for status codes that warrant a retry.
func isRetryableStatus(code int) bool {
	switch code {
	case 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

// parseRetryAfter parses the Retry-After header value as seconds.
func parseRetryAfter(value string) *int {
	if value == "" {
		return nil
	}
	// Try seconds
	if n, err := strconv.Atoi(value); err == nil {
		return &n
	}
	// Try HTTP-date
	if t, err := http.ParseTime(value); err == nil {
		seconds := int(time.Until(t).Seconds())
		if seconds < 0 {
			seconds = 0
		}
		return &seconds
	}
	return nil
}

// parseResponseHeaders extracts cl-* headers from the response.
func parseResponseHeaders(resp *http.Response) *ResponseHeaders {
	h := &ResponseHeaders{}

	if v := resp.Header.Get("cl-worker-job-id"); v != "" {
		h.WorkerJobID = &v
	}
	if v := resp.Header.Get("cl-cluster-id"); v != "" {
		h.ClusterID = &v
	}
	if v := resp.Header.Get("cl-worker"); v != "" {
		h.Worker = &v
	}
	if v := resp.Header.Get("cl-bandwidth"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			h.Bandwidth = &n
		}
	}
	if v := resp.Header.Get("cl-process-time"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			h.ProcessTime = &n
		}
	}
	if v := resp.Header.Get("cl-calls-remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			h.CallsRemaining = &n
		}
	}
	if v := resp.Header.Get("cl-charged-time"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			h.ChargedTime = &n
		}
	}
	if v := resp.Header.Get("cl-bandwidth-cost"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			h.BandwidthCost = &f
		}
	}
	if v := resp.Header.Get("cl-process-time-cost"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			h.ProcessTimeCost = &f
		}
	}
	if v := resp.Header.Get("cl-api-credit-cost"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			h.APICreditCost = &f
		}
	}

	return h
}
