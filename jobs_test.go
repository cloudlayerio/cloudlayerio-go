package cloudlayer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestListJobs(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/jobs" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"id":"j1","status":"success"},{"id":"j2","status":"pending"}]`)
	})

	jobs, err := c.ListJobs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("len(jobs) = %d, want 2", len(jobs))
	}
	if jobs[0].ID != "j1" {
		t.Errorf("jobs[0].ID = %q", jobs[0].ID)
	}
}

func TestListJobs_EmptyArray(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[]`)
	})

	jobs, err := c.ListJobs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if jobs == nil {
		t.Error("empty response should return empty slice, not nil")
	}
	if len(jobs) != 0 {
		t.Errorf("len(jobs) = %d, want 0", len(jobs))
	}
}

func TestGetJob_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/jobs/abc123" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"abc123","status":"success","assetUrl":"https://s3.example.com/file.pdf"}`)
	})

	job, err := c.GetJob(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != "abc123" {
		t.Errorf("ID = %q", job.ID)
	}
	if job.AssetURL == nil || *job.AssetURL != "https://s3.example.com/file.pdf" {
		t.Error("AssetURL not parsed")
	}
}

func TestGetJob_EmptyID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	})

	_, err := c.GetJob(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestGetJob_404(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = fmt.Fprint(w, `{"message":"Job not found"}`)
	})

	_, err := c.GetJob(context.Background(), "nonexistent")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
}

func TestWaitForJob_ImmediateSuccess(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"j1","status":"success","assetUrl":"https://example.com/file.pdf"}`)
	})

	job, err := c.WaitForJob(context.Background(), "j1")
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != JobSuccess {
		t.Errorf("status = %q", job.Status)
	}
}

func TestWaitForJob_PollsThenSuccess(t *testing.T) {
	var calls int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if n < 3 {
			_, _ = fmt.Fprint(w, `{"id":"j1","status":"pending"}`)
		} else {
			_, _ = fmt.Fprint(w, `{"id":"j1","status":"success"}`)
		}
	})

	job, err := c.WaitForJob(context.Background(), "j1", WithPollInterval(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != JobSuccess {
		t.Errorf("status = %q", job.Status)
	}
	if atomic.LoadInt32(&calls) < 3 {
		t.Errorf("expected at least 3 polls, got %d", atomic.LoadInt32(&calls))
	}
}

func TestWaitForJob_ErrorStatus(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"j1","status":"error","error":"generation failed"}`)
	})

	_, err := c.WaitForJob(context.Background(), "j1")
	if err == nil {
		t.Fatal("expected error for failed job")
	}
	if !errors.Is(err, err) { // not nil
		t.Log("error:", err)
	}
}

func TestWaitForJob_ContextCancelled(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"j1","status":"pending"}`)
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := c.WaitForJob(ctx, "j1", WithPollInterval(2*time.Second))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %T: %v", err, err)
	}
}

func TestWaitForJob_PollIntervalTooShort(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	})

	_, err := c.WaitForJob(context.Background(), "j1", WithPollInterval(500*time.Millisecond))
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestWaitForJob_EmptyID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	})

	_, err := c.WaitForJob(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}
