package cloudlayer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestURLToPDF_ValidURL(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/v2/url/pdf" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["url"] != "https://example.com" {
			t.Errorf("body.url = %v", body["url"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"j1","status":"pending"}`))
	})

	result, err := c.URLToPDF(context.Background(), &URLToPDFOptions{
		URLOptions: URLOptions{URL: stringPtr("https://example.com")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Job == nil || result.Job.ID != "j1" {
		t.Error("expected Job with ID j1")
	}
}

func TestURLToPDF_EmptyURL(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	})

	_, err := c.URLToPDF(context.Background(), &URLToPDFOptions{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestURLToPDF_WithBatch(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		batch, ok := body["batch"].(map[string]interface{})
		if !ok {
			t.Fatal("expected batch in body")
		}
		urls, ok := batch["urls"].([]interface{})
		if !ok || len(urls) != 2 {
			t.Errorf("expected 2 batch URLs, got %v", batch["urls"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"j2","status":"pending"}`))
	})

	result, err := c.URLToPDF(context.Background(), &URLToPDFOptions{
		URLOptions: URLOptions{
			Batch: &Batch{URLs: []string{"https://a.com", "https://b.com"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Job.ID != "j2" {
		t.Errorf("Job.ID = %q", result.Job.ID)
	}
}

func TestURLToImage_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/url/image" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"j3","status":"pending"}`))
	})

	result, err := c.URLToImage(context.Background(), &URLToImageOptions{
		URLOptions: URLOptions{URL: stringPtr("https://example.com")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Job.ID != "j3" {
		t.Errorf("Job.ID = %q", result.Job.ID)
	}
}

func TestHTMLToPDF_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/html/pdf" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["html"] != "PGgxPkhlbGxvPC9oMT4=" {
			t.Errorf("body.html = %v", body["html"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"j4","status":"pending"}`))
	})

	result, err := c.HTMLToPDF(context.Background(), &HTMLToPDFOptions{
		HTMLOptions: HTMLOptions{HTML: EncodeHTML("<h1>Hello</h1>")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Job.ID != "j4" {
		t.Errorf("Job.ID = %q", result.Job.ID)
	}
}

func TestHTMLToPDF_EmptyHTML(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	})

	_, err := c.HTMLToPDF(context.Background(), &HTMLToPDFOptions{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestHTMLToImage_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/html/image" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"j5","status":"pending"}`))
	})

	result, err := c.HTMLToImage(context.Background(), &HTMLToImageOptions{
		HTMLOptions: HTMLOptions{HTML: EncodeHTML("<h1>Hi</h1>")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Job.ID != "j5" {
		t.Errorf("Job.ID = %q", result.Job.ID)
	}
}

func TestTemplateToPDF_WithTemplateID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["templateId"] != "invoice-01" {
			t.Errorf("templateId = %v", body["templateId"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"j6","status":"pending"}`))
	})

	result, err := c.TemplateToPDF(context.Background(), &TemplateToPDFOptions{
		TemplateOptions: TemplateOptions{
			TemplateID: stringPtr("invoice-01"),
			Data:       map[string]interface{}{"name": "Test"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Job.ID != "j6" {
		t.Errorf("Job.ID = %q", result.Job.ID)
	}
}

func TestTemplateToPDF_MissingBoth(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	})

	_, err := c.TemplateToPDF(context.Background(), &TemplateToPDFOptions{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestMergePDFs_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/pdf/merge" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"j7","status":"pending"}`))
	})

	result, err := c.MergePDFs(context.Background(), &MergePDFsOptions{
		URLOptions: URLOptions{
			Batch: &Batch{URLs: []string{"https://a.com/1.pdf", "https://a.com/2.pdf"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Job.ID != "j7" {
		t.Errorf("Job.ID = %q", result.Job.ID)
	}
}

func TestDOCXToPDF_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/docx/pdf" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"j8","status":"pending"}`))
	})

	result, err := c.DOCXToPDF(context.Background(), &DOCXToPDFOptions{
		File: &FileInput{
			Reader:   strings.NewReader("fake docx"),
			Filename: "test.docx",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Job.ID != "j8" {
		t.Errorf("Job.ID = %q", result.Job.ID)
	}
}

func TestDOCXToPDF_NilFile(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	})

	_, err := c.DOCXToPDF(context.Background(), &DOCXToPDFOptions{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestConversion_V1BinaryResponse(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `attachment; filename="output.pdf"`)
		_, _ = w.Write([]byte("%PDF-1.4 fake content"))
	})
	c.apiVersion = V1

	result, err := c.HTMLToPDF(context.Background(), &HTMLToPDFOptions{
		HTMLOptions: HTMLOptions{HTML: EncodeHTML("<h1>Hi</h1>")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Job != nil {
		t.Error("v1 binary should have nil Job")
	}
	if len(result.Data) == 0 {
		t.Error("v1 binary should have Data")
	}
	if result.Filename != "output.pdf" {
		t.Errorf("Filename = %q, want output.pdf", result.Filename)
	}
}

func TestConversion_V2JSONResponse(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"j10","status":"pending"}`))
	})

	result, err := c.HTMLToPDF(context.Background(), &HTMLToPDFOptions{
		HTMLOptions: HTMLOptions{HTML: EncodeHTML("<h1>Hi</h1>")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Data != nil {
		t.Error("v2 should have nil Data")
	}
	if result.Job == nil || result.Job.ID != "j10" {
		t.Error("v2 should have Job")
	}
}

func TestConversion_NoRetry(t *testing.T) {
	calls := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"message":"error"}`))
	})
	c.maxRetries = 3

	_, err := c.HTMLToPDF(context.Background(), &HTMLToPDFOptions{
		HTMLOptions: HTMLOptions{HTML: EncodeHTML("<h1>Hi</h1>")},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (conversion should not retry)", calls)
	}
}
