package cloudlayer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestListStorage(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"id":"s1","title":"Production S3"}]`)
	})

	items, err := c.ListStorage(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "s1" {
		t.Errorf("unexpected items: %v", items)
	}
}

func TestGetStorage_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/storage/s1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"s1","title":"Production","data":"encrypted","uid":"u1"}`)
	})

	detail, err := c.GetStorage(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Title != "Production" {
		t.Errorf("Title = %q", detail.Title)
	}
}

func TestGetStorage_EmptyID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})

	_, err := c.GetStorage(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestAddStorage_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["title"] != "My S3" {
			t.Errorf("title = %v", body["title"])
		}
		if body["accessKeyId"] != "AKIA123" {
			t.Errorf("accessKeyId = %v", body["accessKeyId"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"title":"My S3","id":"s-new"}`)
	})

	result, err := c.AddStorage(context.Background(), &StorageParams{
		Title:           "My S3",
		Bucket:          "my-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "AKIA123",
		SecretAccessKey: "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "s-new" {
		t.Errorf("ID = %q", result.ID)
	}
}

func TestAddStorage_NotAllowed(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"allowed":false,"reason":"Plan does not support custom storage","statusCode":401}`)
	})

	_, err := c.AddStorage(context.Background(), &StorageParams{
		Title:           "Test",
		Bucket:          "b",
		Region:          "r",
		AccessKeyID:     "k",
		SecretAccessKey: "s",
	})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
}

func TestAddStorage_MissingFields(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})

	tests := []struct {
		name   string
		params *StorageParams
	}{
		{"nil params", nil},
		{"empty title", &StorageParams{Bucket: "b", Region: "r", AccessKeyID: "k", SecretAccessKey: "s"}},
		{"empty bucket", &StorageParams{Title: "t", Region: "r", AccessKeyID: "k", SecretAccessKey: "s"}},
		{"empty region", &StorageParams{Title: "t", Bucket: "b", AccessKeyID: "k", SecretAccessKey: "s"}},
		{"empty accessKeyId", &StorageParams{Title: "t", Bucket: "b", Region: "r", SecretAccessKey: "s"}},
		{"empty secretAccessKey", &StorageParams{Title: "t", Bucket: "b", Region: "r", AccessKeyID: "k"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.AddStorage(context.Background(), tt.params)
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Errorf("expected *ValidationError, got %T", err)
			}
		})
	}
}

func TestDeleteStorage_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %q", r.Method)
		}
		if r.URL.Path != "/v2/storage/s1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"status":"success"}`)
	})

	err := c.DeleteStorage(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteStorage_EmptyID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})

	err := c.DeleteStorage(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}
