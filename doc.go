// Package cloudlayer provides a Go client for the CloudLayer.io document
// generation API. It supports URL/HTML/template to PDF/image conversion,
// document format conversion, job management, and storage configuration.
//
// Basic usage:
//
//	client, err := cloudlayer.NewClient("your-api-key", cloudlayer.V2)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	result, err := client.HTMLToPDF(ctx, &cloudlayer.HTMLToPDFOptions{
//	    HTMLOptions: cloudlayer.HTMLOptions{
//	        HTML: cloudlayer.EncodeHTML("<h1>Hello World</h1>"),
//	    },
//	})
//
// # API Versions
//
// CloudLayer.io supports two API versions with different response behaviors:
//
//   - V1: Synchronous by default. Conversion endpoints return raw binary data.
//   - V2: Asynchronous by default. Conversion endpoints return a [Job] object.
//     Use [Client.WaitForJob] to poll until completion, then
//     [Client.DownloadJobResult] to fetch the binary.
//
// The API version must be specified when creating the client — there is no default.
//
// # Error Handling
//
// All errors returned by the SDK are concrete types that support [errors.As]:
//
//	var authErr *cloudlayer.AuthError
//	if errors.As(err, &authErr) {
//	    // Handle authentication failure (401 or 403)
//	}
//
//	var rateLimitErr *cloudlayer.RateLimitError
//	if errors.As(err, &rateLimitErr) {
//	    // Handle rate limiting (429), check rateLimitErr.RetryAfter
//	}
package cloudlayer
