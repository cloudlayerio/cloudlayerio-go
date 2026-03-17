package cloudlayer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestListTemplates_NoOptions(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/templates" {
			t.Errorf("path = %q, want /v2/templates", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"id":"t1","name":"Invoice"}]`)
	})

	templates, err := c.ListTemplates(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 1 {
		t.Fatalf("len = %d", len(templates))
	}
	if templates[0].ID != "t1" {
		t.Errorf("ID = %q", templates[0].ID)
	}
}

func TestListTemplates_WithFilters(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("category") != "business" {
			t.Errorf("category = %q", r.URL.Query().Get("category"))
		}
		if r.URL.Query().Get("type") != "pdf" {
			t.Errorf("type = %q", r.URL.Query().Get("type"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[]`)
	})

	_, err := c.ListTemplates(context.Background(), &ListTemplatesOptions{
		Category: stringPtr("business"),
		Type:     stringPtr("pdf"),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestListTemplates_Empty(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[]`)
	})

	templates, err := c.ListTemplates(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if templates == nil {
		t.Error("should return empty slice, not nil")
	}
}

func TestGetTemplate_Valid(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/template/t1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"t1","name":"Invoice","type":"pdf"}`)
	})

	tmpl, err := c.GetTemplate(context.Background(), "t1")
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.ID != "t1" {
		t.Errorf("ID = %q", tmpl.ID)
	}
}

func TestGetTemplate_EmptyID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})

	_, err := c.GetTemplate(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}
