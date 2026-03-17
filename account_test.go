package cloudlayer

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestGetAccount(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/account" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{
			"email":"user@test.com",
			"uid":"u1",
			"subscription":"price-starter-1k",
			"subType":"limit",
			"subActive":true,
			"calls":100,
			"callsLimit":1000,
			"bytesTotal":0,
			"bytesLimit":5368709120,
			"computeTimeTotal":0,
			"computeTimeLimit":600000,
			"storageUsed":0,
			"storageLimit":1073741824
		}`)
	})

	info, err := c.GetAccount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Email != "user@test.com" {
		t.Errorf("Email = %q", info.Email)
	}
	if info.Calls != 100 {
		t.Errorf("Calls = %d", info.Calls)
	}
	if !info.SubActive {
		t.Error("SubActive should be true")
	}
}

func TestGetAccount_WithCredits(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{
			"email":"user@test.com",
			"uid":"u1",
			"subscription":"price-growth-usage",
			"subType":"usage",
			"subActive":true,
			"calls":15234,
			"callsLimit":-1,
			"credit":4500.5,
			"bytesTotal":0,
			"bytesLimit":-1,
			"computeTimeTotal":0,
			"computeTimeLimit":-1,
			"storageUsed":0,
			"storageLimit":-1
		}`)
	})

	info, err := c.GetAccount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Credit == nil || *info.Credit != 4500.5 {
		t.Errorf("Credit = %v", info.Credit)
	}
	if info.CallsLimit != -1 {
		t.Errorf("CallsLimit = %d", info.CallsLimit)
	}
}

func TestGetStatus(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/getStatus" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"status":"ok "}`)
	})

	status, err := c.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "ok " {
		t.Errorf("Status = %q, want %q", status.Status, "ok ")
	}
}
