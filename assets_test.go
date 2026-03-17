package cloudlayer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestListAssets(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"uid":"u1","fileId":"f1"},{"uid":"u2","fileId":"f2"}]`)
	})

	assets, err := c.ListAssets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 2 {
		t.Errorf("len = %d", len(assets))
	}
}

func TestListAssets_Empty(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[]`)
	})

	assets, err := c.ListAssets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if assets == nil {
		t.Error("should return empty slice, not nil")
	}
}

func TestGetAsset_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/assets/a123" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"uid":"u1","fileId":"f1","url":"https://s3.example.com/file.pdf"}`)
	})

	asset, err := c.GetAsset(context.Background(), "a123")
	if err != nil {
		t.Fatal(err)
	}
	if asset.FileID != "f1" {
		t.Errorf("FileID = %q", asset.FileID)
	}
}

func TestGetAsset_EmptyID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	})

	_, err := c.GetAsset(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestDownloadJobResult_Valid(t *testing.T) {
	srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	// Create a separate server for the download URL
	downloadSrv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Should NOT have X-API-Key header
		if r.Header.Get("X-API-Key") != "" {
			t.Error("X-API-Key should not be sent to S3 URL")
		}
		_, _ = w.Write([]byte("%PDF-1.4 content"))
	})
	_ = srv

	job := &Job{
		AssetURL: stringPtr(downloadSrv.baseURL + "/file.pdf"),
	}

	data, err := downloadSrv.DownloadJobResult(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

func TestDownloadJobResult_NilJob(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})

	_, err := c.DownloadJobResult(context.Background(), nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestDownloadJobResult_EmptyAssetURL(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})

	_, err := c.DownloadJobResult(context.Background(), &Job{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestDownloadJobResult_403Expired(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})

	job := &Job{AssetURL: stringPtr(c.baseURL + "/expired.pdf")}
	_, err := c.DownloadJobResult(context.Background(), job)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
}
