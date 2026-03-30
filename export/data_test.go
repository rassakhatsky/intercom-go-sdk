package export_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/export"
	"github.com/rassakhatsky/intercom-go-sdk/api"
)

func setupData() (svc *export.DataService, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = export.NewDataService(caller)
	return svc, mux, server.Close
}

func TestDataService_Create(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	mux.HandleFunc("/export/content/data", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body export.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.CreatedAtAfter != 1527811200 {
			t.Errorf("body.CreatedAtAfter = %v, want 1527811200", body.CreatedAtAfter)
		}
		if body.CreatedAtBefore != 1530403200 {
			t.Errorf("body.CreatedAtBefore = %v, want 1530403200", body.CreatedAtBefore)
		}
		fmt.Fprint(w, `{
			"job_identifier":"orzzsbd7hk67xyu",
			"status":"pending",
			"download_url":"",
			"download_expires_at":""
		}`)
	})

	ctx := context.Background()
	de, err := svc.Create(ctx, &export.CreateRequest{
		CreatedAtAfter:  1527811200,
		CreatedAtBefore: 1530403200,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if de.JobIdentifier != "orzzsbd7hk67xyu" {
		t.Errorf("JobIdentifier = %v, want orzzsbd7hk67xyu", de.JobIdentifier)
	}
	if de.Status != "pending" {
		t.Errorf("Status = %v, want pending", de.Status)
	}
}

func TestDataService_GetStatus(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	mux.HandleFunc("/export/content/data/orzzsbd7hk67xyu", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"job_identifier":"orzzsbd7hk67xyu",
			"status":"completed",
			"download_url":"https://api.intercom.test/download/content/data/orzzsbd7hk67xyu",
			"download_expires_at":"1674917488"
		}`)
	})

	ctx := context.Background()
	de, err := svc.GetStatus(ctx, "orzzsbd7hk67xyu")
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}
	if de.JobIdentifier != "orzzsbd7hk67xyu" {
		t.Errorf("JobIdentifier = %v, want orzzsbd7hk67xyu", de.JobIdentifier)
	}
	if de.Status != "completed" {
		t.Errorf("Status = %v, want completed", de.Status)
	}
	if de.DownloadURL != "https://api.intercom.test/download/content/data/orzzsbd7hk67xyu" {
		t.Errorf("DownloadURL = %v, want https://api.intercom.test/download/content/data/orzzsbd7hk67xyu", de.DownloadURL)
	}
	if de.DownloadExpiresAt != "1674917488" {
		t.Errorf("DownloadExpiresAt = %v, want 1674917488", de.DownloadExpiresAt)
	}
}

func TestDataService_GetStatus_NotFound(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	mux.HandleFunc("/export/content/data/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"not_found","message":"Resource not found"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.GetStatus(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestDataService_Cancel(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	mux.HandleFunc("/export/cancel/orzzsbd7hk67xyu", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"job_identifier":"orzzsbd7hk67xyu",
			"status":"canceled",
			"download_url":"",
			"download_expires_at":""
		}`)
	})

	ctx := context.Background()
	de, err := svc.Cancel(ctx, "orzzsbd7hk67xyu")
	if err != nil {
		t.Fatalf("Cancel returned error: %v", err)
	}
	if de.JobIdentifier != "orzzsbd7hk67xyu" {
		t.Errorf("JobIdentifier = %v, want orzzsbd7hk67xyu", de.JobIdentifier)
	}
	if de.Status != "canceled" {
		t.Errorf("Status = %v, want canceled", de.Status)
	}
}

func TestDataService_Download(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	expectedData := []byte("user_id,email,name\n123,test@example.com,Test User\n")

	mux.HandleFunc("/download/content/data/orzzsbd7hk67xyu", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(expectedData)
	})

	ctx := context.Background()
	var buf bytes.Buffer
	err := svc.Download(ctx, "orzzsbd7hk67xyu", &buf)
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), expectedData) {
		t.Errorf("Download data = %q, want %q", buf.String(), string(expectedData))
	}
}

func TestDataService_Download_NotFound(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	mux.HandleFunc("/download/content/data/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-456",
			"errors":[{"code":"not_found","message":"Resource not found"}]
		}`)
	})

	ctx := context.Background()
	var buf bytes.Buffer
	err := svc.Download(ctx, "nonexistent", &buf)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestDataService_Create_RateLimited(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	mux.HandleFunc("/export/content/data", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-789",
			"errors":[{"code":"rate_limit_exceeded","message":"You have exceeded the rate limit"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.Create(ctx, &export.CreateRequest{
		CreatedAtAfter:  1527811200,
		CreatedAtBefore: 1530403200,
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsRateLimited(err) {
		t.Errorf("Expected IsRateLimited, got: %v", err)
	}
}

func TestDataService_CreateRaw(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	mux.HandleFunc("/export/content/data", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"job_identifier":"orzzsbd7hk67xyu",
			"status":"pending",
			"download_url":"",
			"download_expires_at":""
		}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &export.CreateRequest{
		CreatedAtAfter:  1527811200,
		CreatedAtBefore: 1530403200,
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	de, err := export.ParseCreateResult(result)
	if err != nil {
		t.Fatalf("ParseCreateResult returned error: %v", err)
	}
	if de.JobIdentifier != "orzzsbd7hk67xyu" {
		t.Errorf("JobIdentifier = %v, want orzzsbd7hk67xyu", de.JobIdentifier)
	}
	if de.Status != "pending" {
		t.Errorf("Status = %v, want pending", de.Status)
	}
}

func TestDataService_GetStatusRaw(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	mux.HandleFunc("/export/content/data/orzzsbd7hk67xyu", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"job_identifier":"orzzsbd7hk67xyu",
			"status":"completed",
			"download_url":"https://api.intercom.test/download/content/data/orzzsbd7hk67xyu",
			"download_expires_at":"1674917488"
		}`)
	})

	ctx := context.Background()
	result, err := svc.GetStatusRaw(ctx, "orzzsbd7hk67xyu")
	if err != nil {
		t.Fatalf("GetStatusRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	de, err := export.ParseDataGetStatusResult(result)
	if err != nil {
		t.Fatalf("ParseDataGetStatusResult returned error: %v", err)
	}
	if de.JobIdentifier != "orzzsbd7hk67xyu" {
		t.Errorf("JobIdentifier = %v, want orzzsbd7hk67xyu", de.JobIdentifier)
	}
	if de.Status != "completed" {
		t.Errorf("Status = %v, want completed", de.Status)
	}
}

func TestDataService_GetStatusRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	mux.HandleFunc("/export/content/data/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"not_found","message":"Resource not found"}]
		}`)
	})

	ctx := context.Background()
	result, err := svc.GetStatusRaw(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetStatusRaw returned error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Expected Error to be populated")
	}
	if result.Error.Code != "not_found" {
		t.Errorf("Error.Code = %v, want not_found", result.Error.Code)
	}
	if result.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusNotFound)
	}
}

func TestDataService_CancelRaw(t *testing.T) {
	svc, mux, teardown := setupData()
	defer teardown()

	mux.HandleFunc("/export/cancel/orzzsbd7hk67xyu", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"job_identifier":"orzzsbd7hk67xyu",
			"status":"canceled",
			"download_url":"",
			"download_expires_at":""
		}`)
	})

	ctx := context.Background()
	result, err := svc.CancelRaw(ctx, "orzzsbd7hk67xyu")
	if err != nil {
		t.Fatalf("CancelRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	de, err := export.ParseCancelResult(result)
	if err != nil {
		t.Fatalf("ParseCancelResult returned error: %v", err)
	}
	if de.Status != "canceled" {
		t.Errorf("Status = %v, want canceled", de.Status)
	}
}
