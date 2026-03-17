package cloudlayer

import (
	"context"
	"encoding/json"
)

// ListStorage returns all storage configurations for the account.
func (c *Client) ListStorage(ctx context.Context) ([]StorageListItem, error) {
	var items []StorageListItem
	_, err := c.doRequest(ctx, "GET", "/storage", nil, &items, &requestOptions{retryable: true})
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []StorageListItem{}
	}
	return items, nil
}

// GetStorage retrieves a single storage configuration by ID.
func (c *Client) GetStorage(ctx context.Context, storageID string) (*StorageDetail, error) {
	if storageID == "" {
		return nil, &ValidationError{Field: "storageID", Message: "storageID must not be empty"}
	}
	var detail StorageDetail
	_, err := c.doRequest(ctx, "GET", "/storage/"+storageID, nil, &detail, &requestOptions{retryable: true})
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// AddStorage creates a new storage configuration.
//
// If the user's plan does not support custom storage, the server returns HTTP 200
// with {allowed: false, reason: "...", statusCode: N}. In this case, AddStorage
// returns an [APIError] with the reason and status code from the response.
func (c *Client) AddStorage(ctx context.Context, params *StorageParams) (*StorageCreateResponse, error) {
	if params == nil {
		return nil, &ValidationError{Field: "params", Message: "storage params must not be nil"}
	}
	if params.Title == "" {
		return nil, &ValidationError{Field: "title", Message: "title must not be empty"}
	}
	if params.Bucket == "" {
		return nil, &ValidationError{Field: "bucket", Message: "bucket must not be empty"}
	}
	if params.Region == "" {
		return nil, &ValidationError{Field: "region", Message: "region must not be empty"}
	}
	if params.AccessKeyID == "" {
		return nil, &ValidationError{Field: "accessKeyId", Message: "accessKeyId must not be empty"}
	}
	if params.SecretAccessKey == "" {
		return nil, &ValidationError{Field: "secretAccessKey", Message: "secretAccessKey must not be empty"}
	}

	// We need to inspect the raw response to detect the "not allowed" case
	var raw json.RawMessage
	_, err := c.doRequest(ctx, "POST", "/storage", params, &raw, &requestOptions{retryable: false})
	if err != nil {
		return nil, err
	}

	// Check if it's a "not allowed" response
	var notAllowed StorageNotAllowedResponse
	if json.Unmarshal(raw, &notAllowed) == nil && !notAllowed.Allowed && notAllowed.Reason != "" {
		return nil, &APIError{
			StatusCode:    notAllowed.StatusCode,
			Message:       notAllowed.Reason,
			RequestPath:   "/storage",
			RequestMethod: "POST",
		}
	}

	var result StorageCreateResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, &APIError{
			StatusCode:    200,
			StatusText:    "OK",
			Message:       "failed to decode storage response: " + err.Error(),
			RequestPath:   "/storage",
			RequestMethod: "POST",
		}
	}

	return &result, nil
}

// DeleteStorage deletes a storage configuration by ID.
func (c *Client) DeleteStorage(ctx context.Context, storageID string) error {
	if storageID == "" {
		return &ValidationError{Field: "storageID", Message: "storageID must not be empty"}
	}
	_, err := c.doRequest(ctx, "DELETE", "/storage/"+storageID, nil, nil, &requestOptions{retryable: false})
	return err
}
