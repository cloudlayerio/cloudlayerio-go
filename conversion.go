package cloudlayer

import (
	"context"
	"encoding/json"
	"io"
	"mime"
)

// URLToPDF converts a URL to PDF.
//
// v2 returns a [Job] object (use [Client.WaitForJob] + [Client.DownloadJobResult]).
// v1 returns raw binary data in [ConversionResult.Data].
func (c *Client) URLToPDF(ctx context.Context, opts *URLToPDFOptions) (*ConversionResult, error) {
	if err := validateURLOptions(&opts.URLOptions); err != nil {
		return nil, err
	}
	return c.doConversion(ctx, "/url/pdf", opts)
}

// URLToImage converts a URL to an image.
func (c *Client) URLToImage(ctx context.Context, opts *URLToImageOptions) (*ConversionResult, error) {
	if err := validateURLOptions(&opts.URLOptions); err != nil {
		return nil, err
	}
	return c.doConversion(ctx, "/url/image", opts)
}

// HTMLToPDF converts HTML content to PDF.
//
// The HTML field must be base64-encoded. Use [EncodeHTML] to encode raw HTML.
func (c *Client) HTMLToPDF(ctx context.Context, opts *HTMLToPDFOptions) (*ConversionResult, error) {
	if opts.HTML == "" {
		return nil, &ValidationError{Field: "html", Message: "html must not be empty"}
	}
	return c.doConversion(ctx, "/html/pdf", opts)
}

// HTMLToImage converts HTML content to an image.
func (c *Client) HTMLToImage(ctx context.Context, opts *HTMLToImageOptions) (*ConversionResult, error) {
	if opts.HTML == "" {
		return nil, &ValidationError{Field: "html", Message: "html must not be empty"}
	}
	return c.doConversion(ctx, "/html/image", opts)
}

// TemplateToPDF generates a PDF from a template.
//
// Either TemplateID or Template (base64-encoded) must be provided, but not both.
func (c *Client) TemplateToPDF(ctx context.Context, opts *TemplateToPDFOptions) (*ConversionResult, error) {
	if err := validateTemplateOptions(&opts.TemplateOptions); err != nil {
		return nil, err
	}
	return c.doConversion(ctx, "/template/pdf", opts)
}

// TemplateToImage generates an image from a template.
func (c *Client) TemplateToImage(ctx context.Context, opts *TemplateToImageOptions) (*ConversionResult, error) {
	if err := validateTemplateOptions(&opts.TemplateOptions); err != nil {
		return nil, err
	}
	return c.doConversion(ctx, "/template/image", opts)
}

// MergePDFs merges multiple PDFs into one.
// Uses URL-based input (not file uploads).
func (c *Client) MergePDFs(ctx context.Context, opts *MergePDFsOptions) (*ConversionResult, error) {
	if err := validateURLOptions(&opts.URLOptions); err != nil {
		return nil, err
	}
	return c.doConversion(ctx, "/pdf/merge", opts)
}

// doConversion is the internal method for all JSON-body conversion endpoints.
func (c *Client) doConversion(ctx context.Context, path string, body interface{}) (*ConversionResult, error) {
	resp, headers, err := c.doRawRequest(ctx, "POST", path, body, &requestOptions{retryable: false})
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	result := &ConversionResult{
		Headers: headers,
		Status:  resp.StatusCode,
	}

	// Detect response type from Content-Type header
	ct := resp.Header.Get("Content-Type")
	if isJSONContentType(ct) {
		// v2 response (or v1 async) — parse Job
		var job Job
		if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
			return nil, &APIError{
				StatusCode:    resp.StatusCode,
				StatusText:    resp.Status,
				Message:       "failed to decode conversion response: " + err.Error(),
				RequestPath:   path,
				RequestMethod: "POST",
			}
		}
		result.Job = &job
	} else {
		// v1 sync response — raw binary
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, &APIError{
				StatusCode:    resp.StatusCode,
				StatusText:    resp.Status,
				Message:       "failed to read binary response: " + err.Error(),
				RequestPath:   path,
				RequestMethod: "POST",
			}
		}
		result.Data = data
		result.Filename = parseContentDisposition(resp.Header.Get("Content-Disposition"))
	}

	return result, nil
}

func validateURLOptions(opts *URLOptions) error {
	hasURL := opts.URL != nil && *opts.URL != ""
	hasBatch := opts.Batch != nil && len(opts.Batch.URLs) > 0
	if !hasURL && !hasBatch {
		return &ValidationError{Field: "url", Message: "either url or batch must be provided"}
	}
	return nil
}

func validateTemplateOptions(opts *TemplateOptions) error {
	hasID := opts.TemplateID != nil && *opts.TemplateID != ""
	hasTemplate := opts.Template != nil && *opts.Template != ""
	if !hasID && !hasTemplate {
		return &ValidationError{
			Field:   "templateId",
			Message: "either templateId or template must be provided",
		}
	}
	return nil
}

func isJSONContentType(ct string) bool {
	mediaType, _, _ := mime.ParseMediaType(ct)
	return mediaType == "application/json"
}

func parseContentDisposition(header string) string {
	if header == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(header)
	if err != nil {
		return ""
	}
	if fn, ok := params["filename"]; ok {
		return fn
	}
	return ""
}

// DOCXToPDF converts a DOCX document to PDF via multipart upload.
func (c *Client) DOCXToPDF(ctx context.Context, opts *DOCXToPDFOptions) (*ConversionResult, error) {
	if opts.File == nil {
		return nil, &ValidationError{Field: "file", Message: "file must not be nil"}
	}
	return c.doDocumentConversion(ctx, "/docx/pdf", opts.File, &opts.BaseOptions)
}

// DOCXToHTML converts a DOCX document to HTML via multipart upload.
func (c *Client) DOCXToHTML(ctx context.Context, opts *DOCXToHTMLOptions) (*ConversionResult, error) {
	if opts.File == nil {
		return nil, &ValidationError{Field: "file", Message: "file must not be nil"}
	}
	return c.doDocumentConversion(ctx, "/docx/html", opts.File, &opts.BaseOptions)
}

// PDFToDOCX converts a PDF to DOCX via multipart upload.
func (c *Client) PDFToDOCX(ctx context.Context, opts *PDFToDOCXOptions) (*ConversionResult, error) {
	if opts.File == nil {
		return nil, &ValidationError{Field: "file", Message: "file must not be nil"}
	}
	return c.doDocumentConversion(ctx, "/pdf/docx", opts.File, &opts.BaseOptions)
}

// doDocumentConversion handles multipart document conversion endpoints.
func (c *Client) doDocumentConversion(ctx context.Context, path string, file *FileInput, baseOpts *BaseOptions) (*ConversionResult, error) {
	// Marshal base options to fields
	fields := make(map[string]string)
	if baseOpts != nil {
		optJSON, err := json.Marshal(baseOpts)
		if err == nil && string(optJSON) != "{}" {
			fields["options"] = string(optJSON)
		}
	}

	var job Job
	headers, err := c.doMultipartRequest(ctx, path, file, fields, &job, &requestOptions{retryable: false})
	if err != nil {
		return nil, err
	}

	return &ConversionResult{
		Job:     &job,
		Headers: headers,
		Status:  200,
	}, nil
}
