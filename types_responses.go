package cloudlayer

import (
	"encoding/json"
	"fmt"
	"math"
)

// Job represents a CloudLayer.io conversion job.
type Job struct {
	ID              string          `json:"id"`
	UID             string          `json:"uid"`
	Name            *string         `json:"name,omitempty"`
	Type            *JobType        `json:"type,omitempty"`
	Status          JobStatus       `json:"status"`
	Timestamp       json.RawMessage `json:"timestamp,omitempty"`
	WorkerName      *string         `json:"workerName,omitempty"`
	ProcessTime     *int            `json:"processTime,omitempty"`
	APIKeyUsed      *string         `json:"apiKeyUsed,omitempty"`
	ProcessTimeCost *float64        `json:"processTimeCost,omitempty"`
	APICreditCost   *float64        `json:"apiCreditCost,omitempty"`
	BandwidthCost   *float64        `json:"bandwidthCost,omitempty"`
	TotalCost       *float64        `json:"totalCost,omitempty"`
	Size            *int            `json:"size,omitempty"`
	Params          json.RawMessage `json:"params,omitempty"`
	AssetURL        *string         `json:"assetUrl,omitempty"`
	PreviewURL      *string         `json:"previewUrl,omitempty"`
	Self            *string         `json:"self,omitempty"`
	AssetID         *string         `json:"assetId,omitempty"`
	ProjectID       *string         `json:"projectId,omitempty"`
	Error           *string         `json:"error,omitempty"`
}

// TimestampUnix returns the job timestamp as Unix milliseconds.
// It handles both numeric timestamps and Firestore timestamp objects
// ({_seconds, _nanoseconds}).
func (j *Job) TimestampUnix() int64 {
	if j == nil || len(j.Timestamp) == 0 {
		return 0
	}
	return parseTimestamp(j.Timestamp)
}

// Asset represents a generated file asset.
type Asset struct {
	UID           string          `json:"uid"`
	ID            *string         `json:"id,omitempty"`
	FileID        string          `json:"fileId"`
	PreviewFileID *string         `json:"previewFileId,omitempty"`
	Type          *string         `json:"type,omitempty"`
	Ext           *string         `json:"ext,omitempty"`
	PreviewExt    *string         `json:"previewExt,omitempty"`
	URL           *string         `json:"url,omitempty"`
	PreviewURL    *string         `json:"previewUrl,omitempty"`
	Size          *int            `json:"size,omitempty"`
	Timestamp     json.RawMessage `json:"timestamp,omitempty"`
	ProjectID     *string         `json:"projectId,omitempty"`
	JobID         *string         `json:"jobId,omitempty"`
	Name          *string         `json:"name,omitempty"`
}

// TimestampUnix returns the asset timestamp as Unix milliseconds.
// It handles both numeric timestamps and Firestore timestamp objects.
func (a *Asset) TimestampUnix() int64 {
	if a == nil || len(a.Timestamp) == 0 {
		return 0
	}
	return parseTimestamp(a.Timestamp)
}

// parseTimestamp extracts Unix milliseconds from either a number or a
// Firestore timestamp object ({_seconds, _nanoseconds}).
func parseTimestamp(data json.RawMessage) int64 {
	// Try numeric first (most common)
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		return int64(n)
	}

	// Try Firestore timestamp object
	var ts struct {
		Seconds     int64 `json:"_seconds"`
		Nanoseconds int64 `json:"_nanoseconds"`
	}
	if err := json.Unmarshal(data, &ts); err == nil && ts.Seconds > 0 {
		return ts.Seconds*1000 + ts.Nanoseconds/1_000_000
	}

	return 0
}

// StorageListItem is returned by [Client.ListStorage].
// It contains only the ID and title — use [Client.GetStorage] for full details.
type StorageListItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// StorageDetail is returned by [Client.GetStorage].
type StorageDetail struct {
	Data  string `json:"data"`
	ID    string `json:"id"`
	Title string `json:"title"`
	UID   string `json:"uid"`
}

// StorageParams are the parameters for [Client.AddStorage].
type StorageParams struct {
	Title           string  `json:"title"`
	Region          string  `json:"region"`
	AccessKeyID     string  `json:"accessKeyId"`
	SecretAccessKey string  `json:"secretAccessKey"`
	Bucket          string  `json:"bucket"`
	Endpoint        *string `json:"endpoint,omitempty"`
}

// StorageCreateResponse is returned by [Client.AddStorage] on success.
type StorageCreateResponse struct {
	Title string `json:"title"`
	ID    string `json:"id"`
}

// StorageNotAllowedResponse is returned when the user's plan does not
// support custom storage. Note: this comes as HTTP 200, not an error status.
type StorageNotAllowedResponse struct {
	Allowed    bool   `json:"allowed"`
	Reason     string `json:"reason"`
	StatusCode int    `json:"statusCode"`
}

// AccountInfo contains account usage and subscription information.
type AccountInfo struct {
	Email            string                 `json:"-"`
	CallsLimit       int                    `json:"-"`
	Calls            int                    `json:"-"`
	StorageUsed      int                    `json:"-"`
	StorageLimit     int                    `json:"-"`
	Subscription     string                 `json:"-"`
	BytesTotal       int                    `json:"-"`
	BytesLimit       int                    `json:"-"`
	ComputeTimeTotal int                    `json:"-"`
	ComputeTimeLimit int                    `json:"-"`
	SubType          string                 `json:"-"`
	UID              string                 `json:"-"`
	Credit           *float64               `json:"-"`
	SubActive        bool                   `json:"-"`
	Extra            map[string]interface{} `json:"-"`
}

// UnmarshalJSON implements custom unmarshaling for AccountInfo.
// Known fields are extracted into typed struct fields; remaining fields
// are collected into the Extra map.
func (a *AccountInfo) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshaling AccountInfo: %w", err)
	}

	known := map[string]bool{
		"email": true, "callsLimit": true, "calls": true,
		"storageUsed": true, "storageLimit": true, "subscription": true,
		"bytesTotal": true, "bytesLimit": true, "computeTimeTotal": true,
		"computeTimeLimit": true, "subType": true, "uid": true,
		"credit": true, "subActive": true,
	}

	unmarshalString := func(key string) string {
		if v, ok := raw[key]; ok {
			var s string
			if json.Unmarshal(v, &s) == nil {
				return s
			}
		}
		return ""
	}

	unmarshalInt := func(key string) int {
		if v, ok := raw[key]; ok {
			var n float64
			if json.Unmarshal(v, &n) == nil {
				return int(math.Round(n))
			}
		}
		return 0
	}

	unmarshalBool := func(key string) bool {
		if v, ok := raw[key]; ok {
			var b bool
			if json.Unmarshal(v, &b) == nil {
				return b
			}
		}
		return false
	}

	a.Email = unmarshalString("email")
	a.CallsLimit = unmarshalInt("callsLimit")
	a.Calls = unmarshalInt("calls")
	a.StorageUsed = unmarshalInt("storageUsed")
	a.StorageLimit = unmarshalInt("storageLimit")
	a.Subscription = unmarshalString("subscription")
	a.BytesTotal = unmarshalInt("bytesTotal")
	a.BytesLimit = unmarshalInt("bytesLimit")
	a.ComputeTimeTotal = unmarshalInt("computeTimeTotal")
	a.ComputeTimeLimit = unmarshalInt("computeTimeLimit")
	a.SubType = unmarshalString("subType")
	a.UID = unmarshalString("uid")
	a.SubActive = unmarshalBool("subActive")

	if v, ok := raw["credit"]; ok {
		var c float64
		if json.Unmarshal(v, &c) == nil {
			a.Credit = &c
		}
	}

	a.Extra = make(map[string]interface{})
	for key, val := range raw {
		if !known[key] {
			var v interface{}
			if json.Unmarshal(val, &v) == nil {
				a.Extra[key] = v
			}
		}
	}

	return nil
}

// MarshalJSON implements custom marshaling for AccountInfo.
func (a AccountInfo) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"email":            a.Email,
		"callsLimit":       a.CallsLimit,
		"calls":            a.Calls,
		"storageUsed":      a.StorageUsed,
		"storageLimit":     a.StorageLimit,
		"subscription":     a.Subscription,
		"bytesTotal":       a.BytesTotal,
		"bytesLimit":       a.BytesLimit,
		"computeTimeTotal": a.ComputeTimeTotal,
		"computeTimeLimit": a.ComputeTimeLimit,
		"subType":          a.SubType,
		"uid":              a.UID,
		"subActive":        a.SubActive,
	}
	if a.Credit != nil {
		m["credit"] = *a.Credit
	}
	for k, v := range a.Extra {
		m[k] = v
	}
	return json.Marshal(m)
}

// StatusResponse is returned by [Client.GetStatus].
type StatusResponse struct {
	Status string `json:"status"`
}

// PublicTemplate represents a public template from the CloudLayer.io gallery.
type PublicTemplate struct {
	ID       string          `json:"id"`
	Name     *string         `json:"name,omitempty"`
	Type     *string         `json:"type,omitempty"`
	Category *string         `json:"category,omitempty"`
	Tags     *string         `json:"tags,omitempty"`
	Raw      json.RawMessage `json:"-"`
}

// UnmarshalJSON implements custom unmarshaling for PublicTemplate.
// Known fields are extracted; the full raw JSON is preserved in Raw.
func (t *PublicTemplate) UnmarshalJSON(data []byte) error {
	// Preserve the raw JSON
	t.Raw = make(json.RawMessage, len(data))
	copy(t.Raw, data)

	// Extract known fields using an alias to avoid infinite recursion
	type Alias PublicTemplate
	var alias struct {
		Alias
	}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	t.ID = alias.ID
	t.Name = alias.Name
	t.Type = alias.Type
	t.Category = alias.Category
	t.Tags = alias.Tags
	return nil
}

// ConversionResult wraps the response from a conversion endpoint.
type ConversionResult struct {
	// Job is populated for v2 responses and v1 async responses.
	Job *Job
	// Data is populated for v1 synchronous responses (raw binary).
	Data []byte
	// Headers contains CloudLayer-specific response headers.
	Headers *ResponseHeaders
	// Status is the HTTP status code.
	Status int
	// Filename is from the Content-Disposition header (v1 only).
	Filename string
}

// ResponseHeaders contains parsed cl-* response headers from the CloudLayer.io API.
type ResponseHeaders struct {
	WorkerJobID     *string  `json:"cl-worker-job-id,omitempty"`
	ClusterID       *string  `json:"cl-cluster-id,omitempty"`
	Worker          *string  `json:"cl-worker,omitempty"`
	Bandwidth       *int     `json:"cl-bandwidth,omitempty"`
	ProcessTime     *int     `json:"cl-process-time,omitempty"`
	CallsRemaining  *int     `json:"cl-calls-remaining,omitempty"`
	ChargedTime     *int     `json:"cl-charged-time,omitempty"`
	BandwidthCost   *float64 `json:"cl-bandwidth-cost,omitempty"`
	ProcessTimeCost *float64 `json:"cl-process-time-cost,omitempty"`
	APICreditCost   *float64 `json:"cl-api-credit-cost,omitempty"`
}
