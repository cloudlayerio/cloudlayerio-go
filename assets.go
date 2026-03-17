package cloudlayer

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// ListAssets returns up to 10 most recent assets (newest first).
//
// Note: the server limits results to 10 records. This method should not be
// polled frequently as each call reads documents from Firestore.
func (c *Client) ListAssets(ctx context.Context) ([]Asset, error) {
	var assets []Asset
	_, err := c.doRequest(ctx, "GET", "/assets", nil, &assets, &requestOptions{retryable: true})
	if err != nil {
		return nil, err
	}
	if assets == nil {
		assets = []Asset{}
	}
	return assets, nil
}

// GetAsset retrieves a single asset by ID.
func (c *Client) GetAsset(ctx context.Context, assetID string) (*Asset, error) {
	if assetID == "" {
		return nil, &ValidationError{Field: "assetID", Message: "assetID must not be empty"}
	}
	var asset Asset
	_, err := c.doRequest(ctx, "GET", "/assets/"+assetID, nil, &asset, &requestOptions{retryable: true})
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// DownloadJobResult fetches the binary file from a completed job's asset URL.
//
// This is essential for v2 workflows where conversion endpoints return Job
// objects instead of raw binary. Use [Client.WaitForJob] first to ensure the
// job is complete, then call this to download the result.
//
// The asset URL is a presigned S3 URL with a TTL — if it has expired, this
// method returns an error suggesting the URL may have expired.
func (c *Client) DownloadJobResult(ctx context.Context, job *Job) ([]byte, error) {
	if job == nil {
		return nil, &ValidationError{Field: "job", Message: "job must not be nil"}
	}
	if job.AssetURL == nil || *job.AssetURL == "" {
		return nil, &ValidationError{
			Field:   "job.AssetURL",
			Message: "job has no asset URL — the job may still be pending",
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", *job.AssetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating download request: %w", err)
	}
	// Do NOT send X-API-Key to S3 presigned URLs
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &NetworkError{
			Message:       err.Error(),
			Err:           err,
			RequestPath:   *job.AssetURL,
			RequestMethod: "GET",
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 403 {
		return nil, &APIError{
			StatusCode:    403,
			StatusText:    "Forbidden",
			Message:       "asset URL may have expired (presigned URLs have a TTL)",
			RequestPath:   *job.AssetURL,
			RequestMethod: "GET",
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode:    resp.StatusCode,
			StatusText:    http.StatusText(resp.StatusCode),
			Message:       fmt.Sprintf("download failed with status %d", resp.StatusCode),
			RequestPath:   *job.AssetURL,
			RequestMethod: "GET",
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading download response: %w", err)
	}

	return data, nil
}
