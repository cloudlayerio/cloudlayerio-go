package cloudlayer

import (
	"context"
	"fmt"
)

// ListTemplates returns public templates from the CloudLayer.io gallery.
//
// This endpoint always uses /v2/ regardless of the client's API version,
// and does not require API key authentication.
func (c *Client) ListTemplates(ctx context.Context, opts *ListTemplatesOptions) ([]PublicTemplate, error) {
	params := make(map[string]string)
	if opts != nil {
		if opts.Type != nil {
			params["type"] = *opts.Type
		}
		if opts.Category != nil {
			params["category"] = *opts.Category
		}
		if opts.Tags != nil {
			params["tags"] = *opts.Tags
		}
		if opts.Expand != nil {
			params["expand"] = fmt.Sprintf("%t", *opts.Expand)
		}
	}

	var templates []PublicTemplate
	_, err := c.doRequest(ctx, "GET", "/v2/templates", nil, &templates, &requestOptions{
		absolutePath: true,
		retryable:    true,
		params:       params,
	})
	if err != nil {
		return nil, err
	}
	if templates == nil {
		templates = []PublicTemplate{}
	}
	return templates, nil
}

// GetTemplate retrieves a single public template by ID.
//
// This endpoint always uses /v2/ regardless of the client's API version.
func (c *Client) GetTemplate(ctx context.Context, templateID string) (*PublicTemplate, error) {
	if templateID == "" {
		return nil, &ValidationError{Field: "templateID", Message: "templateID must not be empty"}
	}
	var tmpl PublicTemplate
	_, err := c.doRequest(ctx, "GET", "/v2/template/"+templateID, nil, &tmpl, &requestOptions{
		absolutePath: true,
		retryable:    true,
	})
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}
