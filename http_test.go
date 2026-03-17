package cloudlayer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient creates a Client pointed at the given test server.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient("test-key", V2, WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestDoRequest_Headers(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("X-API-Key = %q", r.Header.Get("X-API-Key"))
		}
		if !strings.HasPrefix(r.Header.Get("User-Agent"), "cloudlayerio-go/") {
			t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{}`)
	})

	var result map[string]interface{}
	_, err := c.doRequest(context.Background(), "GET", "/test", nil, &result, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoRequest_CustomHeaders(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "hello" {
			t.Errorf("X-Custom = %q", r.Header.Get("X-Custom"))
		}
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{}`)
	})
	c.customHeaders = map[string]string{"X-Custom": "hello"}

	var result map[string]interface{}
	_, err := c.doRequest(context.Background(), "GET", "/test", nil, &result, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoRequest_URLConstruction(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/jobs" {
			t.Errorf("path = %q, want /v2/jobs", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `[]`)
	})

	var result []interface{}
	_, err := c.doRequest(context.Background(), "GET", "/jobs", nil, &result, &requestOptions{retryable: true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoRequest_AbsolutePath(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/templates" {
			t.Errorf("path = %q, want /v2/templates", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `[]`)
	})

	var result []interface{}
	_, err := c.doRequest(context.Background(), "GET", "/v2/templates", nil, &result, &requestOptions{absolutePath: true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoRequest_QueryParams(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("category") != "business" {
			t.Errorf("category = %q", r.URL.Query().Get("category"))
		}
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `[]`)
	})

	var result []interface{}
	_, err := c.doRequest(context.Background(), "GET", "/v2/templates", nil, &result, &requestOptions{
		absolutePath: true,
		params:       map[string]string{"category": "business"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoRequest_JSONBody(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if body["url"] != "https://example.com" {
			t.Errorf("body.url = %v", body["url"])
		}
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{"id":"job-1","status":"pending"}`)
	})

	var result Job
	_, err := c.doRequest(context.Background(), "POST", "/url/pdf", map[string]interface{}{
		"url": "https://example.com",
	}, &result, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "job-1" {
		t.Errorf("result.ID = %q", result.ID)
	}
}

func TestDoRequest_ResponseHeaders(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cl-worker-job-id", "wj-123")
		w.Header().Set("cl-calls-remaining", "950")
		w.Header().Set("cl-process-time-cost", "0.5")
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{}`)
	})

	var result map[string]interface{}
	headers, err := c.doRequest(context.Background(), "GET", "/test", nil, &result, nil)
	if err != nil {
		t.Fatal(err)
	}
	if headers.WorkerJobID == nil || *headers.WorkerJobID != "wj-123" {
		t.Error("WorkerJobID not parsed")
	}
	if headers.CallsRemaining == nil || *headers.CallsRemaining != 950 {
		t.Error("CallsRemaining not parsed")
	}
	if headers.ProcessTimeCost == nil || *headers.ProcessTimeCost != 0.5 {
		t.Error("ProcessTimeCost not parsed")
	}
}

func TestDoRequest_204NoContent(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})

	_, err := c.doRequest(context.Background(), "DELETE", "/storage/123", nil, nil, nil)
	if err != nil {
		t.Fatalf("204 should not return error: %v", err)
	}
}

func TestDoRequest_401AuthError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = fmt.Fprint(w, `{"message":"Invalid API key"}`)
	})

	var result interface{}
	_, err := c.doRequest(context.Background(), "GET", "/account", nil, &result, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected *AuthError, got %T: %v", err, err)
	}
	if authErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d", authErr.StatusCode)
	}
}

func TestDoRequest_403AuthError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = fmt.Fprint(w, `{"message":"Forbidden"}`)
	})

	var result interface{}
	_, err := c.doRequest(context.Background(), "GET", "/test", nil, &result, nil)
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected *AuthError for 403, got %T", err)
	}
}

func TestDoRequest_429RateLimitError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(429)
		_, _ = fmt.Fprint(w, `{"message":"Rate limit exceeded"}`)
	})

	var result interface{}
	_, err := c.doRequest(context.Background(), "GET", "/test", nil, &result, nil)
	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}
	if rlErr.RetryAfter == nil || *rlErr.RetryAfter != 30 {
		t.Errorf("RetryAfter = %v", rlErr.RetryAfter)
	}
}

func TestDoRequest_400APIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = fmt.Fprint(w, `{"message":"Bad request"}`)
	})

	var result interface{}
	_, err := c.doRequest(context.Background(), "POST", "/test", nil, &result, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
}

func TestDoRequest_RetryOn429(t *testing.T) {
	var attempts int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			_, _ = fmt.Fprint(w, `{"message":"rate limited"}`)
			return
		}
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})
	c.maxRetries = 2

	var result map[string]interface{}
	_, err := c.doRequest(context.Background(), "GET", "/test", nil, &result, &requestOptions{retryable: true})
	if err != nil {
		t.Fatalf("expected success after retry: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("attempts = %d, want 2", atomic.LoadInt32(&attempts))
	}
}

func TestDoRequest_NoRetryWhenNotRetryable(t *testing.T) {
	var attempts int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(500)
		_, _ = fmt.Fprint(w, `{"message":"server error"}`)
	})
	c.maxRetries = 2

	var result interface{}
	_, err := c.doRequest(context.Background(), "POST", "/url/pdf", nil, &result, &requestOptions{retryable: false})
	if err == nil {
		t.Fatal("expected error")
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1 (no retry for non-retryable)", atomic.LoadInt32(&attempts))
	}
}

func TestDoRequest_NoRetryOn400(t *testing.T) {
	var attempts int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(400)
		_, _ = fmt.Fprint(w, `{"message":"bad request"}`)
	})
	c.maxRetries = 2

	var result interface{}
	_, err := c.doRequest(context.Background(), "GET", "/test", nil, &result, &requestOptions{retryable: true})
	if err == nil {
		t.Fatal("expected error")
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1 (no retry for 400)", atomic.LoadInt32(&attempts))
	}
}

func TestDoRequest_ContextCancelled(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Will be cancelled
		w.WriteHeader(200)
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var result interface{}
	_, err := c.doRequest(ctx, "GET", "/test", nil, &result, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %T: %v", err, err)
	}
}

func TestDoRequest_ContextDeadlineExceeded(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(200)
	})
	c.httpClient.Timeout = 0 // Disable client timeout, use context only

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var result interface{}
	_, err := c.doRequest(ctx, "GET", "/test", nil, &result, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %T: %v", err, err)
	}
}

func TestDoRequest_RetryExhausted(t *testing.T) {
	var attempts int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(500)
		_, _ = fmt.Fprint(w, `{"message":"server error"}`)
	})
	c.maxRetries = 1

	var result interface{}
	_, err := c.doRequest(context.Background(), "GET", "/test", nil, &result, &requestOptions{retryable: true})
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("attempts = %d, want 2 (1 initial + 1 retry)", atomic.LoadInt32(&attempts))
	}
}

func TestDoRequest_RetryOn500(t *testing.T) {
	var attempts int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 2 {
			w.WriteHeader(500)
			_, _ = fmt.Fprint(w, `{"message":"error"}`)
			return
		}
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{}`)
	})
	c.maxRetries = 3

	var result map[string]interface{}
	_, err := c.doRequest(context.Background(), "GET", "/test", nil, &result, &requestOptions{retryable: true})
	if err != nil {
		t.Fatalf("expected success after retries: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", atomic.LoadInt32(&attempts))
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  *int
	}{
		{"seconds", "30", intPtr(30)},
		{"empty", "", nil},
		{"invalid", "abc", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfter(tt.value)
			if tt.want == nil && got != nil {
				t.Errorf("got %v, want nil", *got)
			}
			if tt.want != nil && (got == nil || *got != *tt.want) {
				t.Errorf("got %v, want %d", got, *tt.want)
			}
		})
	}
}

func TestDoRequest_NetworkError(t *testing.T) {
	c, err := NewClient("key", V2, WithBaseURL("http://127.0.0.1:1"))
	if err != nil {
		t.Fatal(err)
	}
	c.httpClient.Timeout = 1 * time.Second

	var result interface{}
	_, reqErr := c.doRequest(context.Background(), "GET", "/test", nil, &result, nil)
	if reqErr == nil {
		t.Fatal("expected network error")
	}
	var netErr *NetworkError
	if !errors.As(reqErr, &netErr) {
		t.Errorf("expected *NetworkError, got %T: %v", reqErr, reqErr)
	}
}

func TestDoMultipartRequest(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("X-API-Key = %q", r.Header.Get("X-API-Key"))
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			t.Fatal(err)
		}
		f, fh, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("form file: %v", err)
		}
		defer func() { _ = f.Close() }()
		if fh.Filename != "test.docx" {
			t.Errorf("filename = %q", fh.Filename)
		}
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{"id":"job-2","status":"pending"}`)
	})

	var result Job
	_, err := c.doMultipartRequest(context.Background(), "/docx/pdf", &FileInput{
		Reader:   strings.NewReader("fake docx content"),
		Filename: "test.docx",
	}, nil, &result, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "job-2" {
		t.Errorf("result.ID = %q", result.ID)
	}
}
