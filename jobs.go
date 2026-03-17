package cloudlayer

import (
	"context"
	"fmt"
	"time"
)

// ListJobs returns up to 10 most recent jobs (newest first).
//
// Note: the server limits results to 10 records. This method should not be
// polled frequently as each call reads documents from Firestore.
func (c *Client) ListJobs(ctx context.Context) ([]Job, error) {
	var jobs []Job
	_, err := c.doRequest(ctx, "GET", "/jobs", nil, &jobs, &requestOptions{retryable: true})
	if err != nil {
		return nil, err
	}
	if jobs == nil {
		jobs = []Job{}
	}
	return jobs, nil
}

// GetJob retrieves a single job by ID.
func (c *Client) GetJob(ctx context.Context, jobID string) (*Job, error) {
	if jobID == "" {
		return nil, &ValidationError{Field: "jobID", Message: "jobID must not be empty"}
	}
	var job Job
	_, err := c.doRequest(ctx, "GET", "/jobs/"+jobID, nil, &job, &requestOptions{retryable: true})
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// WaitOption configures [Client.WaitForJob] behavior.
type WaitOption func(*waitConfig)

type waitConfig struct {
	pollInterval time.Duration
	maxWait      time.Duration
}

// WithPollInterval sets the polling interval for WaitForJob.
// Default is 5 seconds. Minimum is 2 seconds — values below 2s return
// a [ValidationError].
func WithPollInterval(d time.Duration) WaitOption {
	return func(cfg *waitConfig) {
		cfg.pollInterval = d
	}
}

// WithMaxWait sets the maximum time to wait for a job to complete.
// Default is 5 minutes.
func WithMaxWait(d time.Duration) WaitOption {
	return func(cfg *waitConfig) {
		cfg.maxWait = d
	}
}

// WaitForJob polls [Client.GetJob] until the job reaches a terminal status
// ("success" or "error").
//
// On success: returns the completed Job.
// On error status: returns an error containing the job ID and error message.
// On context cancellation: returns the context error.
// On max wait exceeded: returns a [TimeoutError].
func (c *Client) WaitForJob(ctx context.Context, jobID string, opts ...WaitOption) (*Job, error) {
	if jobID == "" {
		return nil, &ValidationError{Field: "jobID", Message: "jobID must not be empty"}
	}

	cfg := &waitConfig{
		pollInterval: 5 * time.Second,
		maxWait:      5 * time.Minute,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.pollInterval < 2*time.Second {
		return nil, &ValidationError{
			Field:   "pollInterval",
			Message: "poll interval must be at least 2 seconds",
		}
	}

	deadline := time.Now().Add(cfg.maxWait)

	for {
		job, err := c.GetJob(ctx, jobID)
		if err != nil {
			return nil, err
		}

		switch job.Status {
		case JobSuccess:
			return job, nil
		case JobError:
			msg := "job failed"
			if job.Error != nil {
				msg = *job.Error
			}
			return nil, fmt.Errorf("cloudlayer: job %s failed: %s", jobID, msg)
		}

		// Check deadline
		if time.Now().After(deadline) {
			return nil, &TimeoutError{
				Timeout:     int(cfg.maxWait.Milliseconds()),
				RequestPath: "/jobs/" + jobID,
			}
		}

		// Wait for next poll
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(cfg.pollInterval):
		}
	}
}
