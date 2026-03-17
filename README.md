# cloudlayer.io Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/cloudlayerio/cloudlayerio-go.svg)](https://pkg.go.dev/github.com/cloudlayerio/cloudlayerio-go)
[![CI](https://github.com/cloudlayerio/cloudlayerio-go/actions/workflows/ci.yml/badge.svg)](https://github.com/cloudlayerio/cloudlayerio-go/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Official Go SDK for the [cloudlayer.io](https://cloudlayer.io) document generation API. Convert URLs, HTML, and templates to PDF and images.

## Installation

```bash
go get github.com/cloudlayerio/cloudlayerio-go
```

Requires Go 1.21+. Zero external dependencies.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    cloudlayer "github.com/cloudlayerio/cloudlayerio-go"
)

func main() {
    client, err := cloudlayer.NewClient("your-api-key", cloudlayer.V2)
    if err != nil {
        log.Fatal(err)
    }

    // Generate a PDF from HTML
    result, err := client.HTMLToPDF(context.Background(), &cloudlayer.HTMLToPDFOptions{
        HTMLOptions: cloudlayer.HTMLOptions{
            HTML: cloudlayer.EncodeHTML("<h1>Hello World</h1>"),
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // v2 returns a Job — wait for completion, then download
    job, err := client.WaitForJob(context.Background(), result.Job.ID)
    if err != nil {
        log.Fatal(err)
    }

    data, err := client.DownloadJobResult(context.Background(), job)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Downloaded %d bytes\n", len(data))
}
```

## API Versions

| Feature | v1 | v2 |
|---------|----|----|
| Response | Raw binary | JSON Job object |
| Default mode | Synchronous | Asynchronous |
| Storage | Optional | Enabled by default |
| Binary access | Direct from response | Via `DownloadJobResult()` |

**v2 never returns binary directly.** Always use `WaitForJob()` + `DownloadJobResult()`.

```go
// v2 (recommended)
client, _ := cloudlayer.NewClient("key", cloudlayer.V2)

// v1 (legacy — returns raw binary)
client, _ := cloudlayer.NewClient("key", cloudlayer.V1)
```

## Document Generation

### URL to PDF/Image

```go
result, err := client.URLToPDF(ctx, &cloudlayer.URLToPDFOptions{
    URLOptions: cloudlayer.URLOptions{
        URL: cloudlayer.StringPtr("https://example.com"),
    },
    PDFOptions: cloudlayer.PDFOptions{
        Format: &cloudlayer.FormatA4,
    },
})
```

### HTML to PDF/Image

```go
result, err := client.HTMLToPDF(ctx, &cloudlayer.HTMLToPDFOptions{
    HTMLOptions: cloudlayer.HTMLOptions{
        HTML: cloudlayer.EncodeHTML("<h1>Invoice</h1><p>Amount: $100</p>"),
    },
    PDFOptions: cloudlayer.PDFOptions{
        PrintBackground: cloudlayer.BoolPtr(true),
    },
})
```

### Template to PDF/Image

```go
result, err := client.TemplateToPDF(ctx, &cloudlayer.TemplateToPDFOptions{
    TemplateOptions: cloudlayer.TemplateOptions{
        TemplateID: cloudlayer.StringPtr("professional-invoice"),
        Data: map[string]interface{}{
            "invoiceNumber": "INV-001",
            "total":         450.00,
        },
    },
})
```

## Document Conversion

```go
// DOCX to PDF
file, _ := cloudlayer.FileInputFromPath("document.docx")
result, err := client.DOCXToPDF(ctx, &cloudlayer.DOCXToPDFOptions{File: file})

// Merge PDFs
result, err := client.MergePDFs(ctx, &cloudlayer.MergePDFsOptions{
    URLOptions: cloudlayer.URLOptions{
        Batch: &cloudlayer.Batch{
            URLs: []string{"https://example.com/1.pdf", "https://example.com/2.pdf"},
        },
    },
})
```

## Data Management

```go
// List recent jobs
jobs, err := client.ListJobs(ctx)

// Get job details
job, err := client.GetJob(ctx, "job-id")

// List assets
assets, err := client.ListAssets(ctx)

// Account info
account, err := client.GetAccount(ctx)

// Storage configuration
configs, err := client.ListStorage(ctx)
```

## Working with v2 Jobs

```go
// 1. Start conversion
result, err := client.HTMLToPDF(ctx, &cloudlayer.HTMLToPDFOptions{
    HTMLOptions: cloudlayer.HTMLOptions{HTML: cloudlayer.EncodeHTML(html)},
})

// 2. Wait for completion (polls every 5s, max 5min)
job, err := client.WaitForJob(ctx, result.Job.ID)

// 3. Download the result
data, err := client.DownloadJobResult(ctx, job)
os.WriteFile("output.pdf", data, 0644)
```

Custom polling options:

```go
job, err := client.WaitForJob(ctx, jobID,
    cloudlayer.WithPollInterval(3*time.Second),
    cloudlayer.WithMaxWait(10*time.Minute),
)
```

## Error Handling

All errors are concrete types supporting `errors.As`:

```go
result, err := client.HTMLToPDF(ctx, opts)
if err != nil {
    var authErr *cloudlayer.AuthError
    if errors.As(err, &authErr) {
        log.Fatal("Invalid API key")
    }

    var rateLimitErr *cloudlayer.RateLimitError
    if errors.As(err, &rateLimitErr) {
        log.Printf("Rate limited, retry after %ds", *rateLimitErr.RetryAfter)
    }

    var validationErr *cloudlayer.ValidationError
    if errors.As(err, &validationErr) {
        log.Printf("Invalid input: %s", validationErr.Message)
    }

    log.Fatal(err)
}
```

| Error Type | When |
|------------|------|
| `*AuthError` | 401 or 403 response |
| `*RateLimitError` | 429 response |
| `*APIError` | Other API errors (400, 404, 500, etc.) |
| `*ValidationError` | Client-side input validation failure |
| `*ConfigError` | Invalid client configuration |
| `*NetworkError` | Connection failures, DNS errors |
| `*TimeoutError` | SDK internal timeout (e.g., WaitForJob max wait) |

## Client Configuration

```go
client, err := cloudlayer.NewClient("key", cloudlayer.V2,
    cloudlayer.WithTimeout(60*time.Second),
    cloudlayer.WithMaxRetries(3),
    cloudlayer.WithBaseURL("https://custom-endpoint.example.com"),
    cloudlayer.WithHeaders(map[string]string{"X-Custom": "value"}),
)
```

## Other SDKs

- **JavaScript/TypeScript:** [@cloudlayerio/sdk](https://www.npmjs.com/package/@cloudlayerio/sdk) ([GitHub](https://github.com/cloudlayerio/cloudlayerio-js))
- **Python:** [cloudlayerio](https://pypi.org/project/cloudlayerio/) ([GitHub](https://github.com/cloudlayerio/cloudlayerio-python))
- **PHP:** [cloudlayerio/sdk](https://packagist.org/packages/cloudlayerio/sdk) ([GitHub](https://github.com/cloudlayerio/cloudlayerio-php))
- **.NET C#:** [cloudlayerio-dotnet](https://www.nuget.org/packages/cloudlayerio-dotnet/) ([GitHub](https://github.com/cloudlayerio/cloudlayerio-dotnet))

## License

MIT - see [LICENSE](LICENSE)
