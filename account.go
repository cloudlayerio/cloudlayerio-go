package cloudlayer

import "context"

// GetAccount retrieves account information including usage statistics.
func (c *Client) GetAccount(ctx context.Context) (*AccountInfo, error) {
	var info AccountInfo
	_, err := c.doRequest(ctx, "GET", "/account", nil, &info, &requestOptions{retryable: true})
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetStatus checks the API health status.
//
// Note: the legacy API returns {"status": "ok "} with a trailing space.
func (c *Client) GetStatus(ctx context.Context) (*StatusResponse, error) {
	var status StatusResponse
	_, err := c.doRequest(ctx, "GET", "/getStatus", nil, &status, &requestOptions{retryable: true})
	if err != nil {
		return nil, err
	}
	return &status, nil
}
