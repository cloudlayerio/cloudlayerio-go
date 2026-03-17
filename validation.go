package cloudlayer

import (
	"encoding/base64"
)

// EncodeHTML base64-encodes an HTML string for use with [HTMLOptions.HTML]
// and [TemplateOptions.Template].
func EncodeHTML(html string) string {
	return base64.StdEncoding.EncodeToString([]byte(html))
}

// stringPtr returns a pointer to the given string.
func stringPtr(s string) *string {
	return &s
}

// boolPtr returns a pointer to the given bool.
func boolPtr(b bool) *bool {
	return &b
}

// intPtr returns a pointer to the given int.
func intPtr(n int) *int {
	return &n
}
