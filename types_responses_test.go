package cloudlayer

import (
	"encoding/json"
	"testing"
)

func TestJob_TimestampUnix_Numeric(t *testing.T) {
	j := &Job{
		Timestamp: json.RawMessage(`1705312200000`),
	}
	got := j.TimestampUnix()
	if got != 1705312200000 {
		t.Errorf("TimestampUnix() = %d, want 1705312200000", got)
	}
}

func TestJob_TimestampUnix_FirestoreObject(t *testing.T) {
	j := &Job{
		Timestamp: json.RawMessage(`{"_seconds":1705312200,"_nanoseconds":500000000}`),
	}
	got := j.TimestampUnix()
	want := int64(1705312200*1000 + 500)
	if got != want {
		t.Errorf("TimestampUnix() = %d, want %d", got, want)
	}
}

func TestJob_TimestampUnix_Empty(t *testing.T) {
	j := &Job{}
	if got := j.TimestampUnix(); got != 0 {
		t.Errorf("TimestampUnix() on empty = %d, want 0", got)
	}
}

func TestJob_TimestampUnix_Nil(t *testing.T) {
	var j *Job
	if got := j.TimestampUnix(); got != 0 {
		t.Errorf("TimestampUnix() on nil = %d, want 0", got)
	}
}

func TestAsset_TimestampUnix(t *testing.T) {
	a := &Asset{
		Timestamp: json.RawMessage(`1705312200000`),
	}
	got := a.TimestampUnix()
	if got != 1705312200000 {
		t.Errorf("Asset.TimestampUnix() = %d, want 1705312200000", got)
	}
}

func TestAccountInfo_UnmarshalJSON(t *testing.T) {
	input := `{
		"email": "user@example.com",
		"uid": "abc123",
		"subscription": "price-starter-1k",
		"subType": "limit",
		"subActive": true,
		"calls": 842,
		"callsLimit": 1000,
		"bytesTotal": 1073741824,
		"bytesLimit": 5368709120,
		"computeTimeTotal": 384200,
		"computeTimeLimit": 600000,
		"storageUsed": 52428800,
		"storageLimit": 1073741824,
		"html-pdf": 500,
		"url-pdf": 342
	}`

	var info AccountInfo
	if err := json.Unmarshal([]byte(input), &info); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if info.Email != "user@example.com" {
		t.Errorf("Email = %q, want user@example.com", info.Email)
	}
	if info.Calls != 842 {
		t.Errorf("Calls = %d, want 842", info.Calls)
	}
	if info.SubActive != true {
		t.Error("SubActive should be true")
	}
	if info.Credit != nil {
		t.Error("Credit should be nil when not present")
	}

	// Extra fields
	if v, ok := info.Extra["html-pdf"]; !ok {
		t.Error("Extra should contain html-pdf")
	} else if v.(float64) != 500 {
		t.Errorf("Extra[html-pdf] = %v, want 500", v)
	}
	if v, ok := info.Extra["url-pdf"]; !ok {
		t.Error("Extra should contain url-pdf")
	} else if v.(float64) != 342 {
		t.Errorf("Extra[url-pdf] = %v, want 342", v)
	}
}

func TestAccountInfo_UnmarshalJSON_WithCredit(t *testing.T) {
	input := `{
		"email": "user@example.com",
		"uid": "abc123",
		"subscription": "price-growth-usage",
		"subType": "usage",
		"subActive": true,
		"calls": 15234,
		"callsLimit": -1,
		"credit": 4500.5,
		"bytesTotal": 0,
		"bytesLimit": -1,
		"computeTimeTotal": 0,
		"computeTimeLimit": -1,
		"storageUsed": 0,
		"storageLimit": -1
	}`

	var info AccountInfo
	if err := json.Unmarshal([]byte(input), &info); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if info.Credit == nil {
		t.Fatal("Credit should not be nil")
	}
	if *info.Credit != 4500.5 {
		t.Errorf("Credit = %f, want 4500.5", *info.Credit)
	}
	if info.CallsLimit != -1 {
		t.Errorf("CallsLimit = %d, want -1", info.CallsLimit)
	}
}

func TestAccountInfo_MarshalJSON_RoundTrip(t *testing.T) {
	credit := 100.0
	original := AccountInfo{
		Email:        "test@test.com",
		UID:          "uid1",
		Subscription: "sub1",
		SubType:      "limit",
		SubActive:    true,
		Calls:        10,
		CallsLimit:   100,
		Credit:       &credit,
		Extra:        map[string]interface{}{"html-pdf": float64(5)},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded AccountInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Email != original.Email {
		t.Errorf("Email = %q, want %q", decoded.Email, original.Email)
	}
	if decoded.Credit == nil || *decoded.Credit != *original.Credit {
		t.Error("Credit round-trip mismatch")
	}
	if v, ok := decoded.Extra["html-pdf"]; !ok || v.(float64) != 5 {
		t.Error("Extra round-trip mismatch")
	}
}

func TestPublicTemplate_UnmarshalJSON(t *testing.T) {
	input := `{"id":"tmpl-1","name":"Invoice","type":"pdf","category":"business","tags":"invoice,billing","extra_field":"value"}`

	var tmpl PublicTemplate
	if err := json.Unmarshal([]byte(input), &tmpl); err != nil {
		t.Fatal(err)
	}

	if tmpl.ID != "tmpl-1" {
		t.Errorf("ID = %q, want tmpl-1", tmpl.ID)
	}
	if tmpl.Name == nil || *tmpl.Name != "Invoice" {
		t.Error("Name should be Invoice")
	}
	if tmpl.Raw == nil {
		t.Error("Raw should be preserved")
	}
	// Verify raw contains the extra field
	var raw map[string]interface{}
	if err := json.Unmarshal(tmpl.Raw, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["extra_field"] != "value" {
		t.Error("Raw should contain extra_field")
	}
}

func TestStatusResponse_Unmarshal(t *testing.T) {
	input := `{"status":"ok "}`
	var s StatusResponse
	if err := json.Unmarshal([]byte(input), &s); err != nil {
		t.Fatal(err)
	}
	if s.Status != "ok " {
		t.Errorf("Status = %q, want %q", s.Status, "ok ")
	}
}
