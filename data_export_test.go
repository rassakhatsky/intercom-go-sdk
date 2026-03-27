package intercom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestDataExportService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/content/data", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body CreateDataExportRequest
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
	export, err := client.DataExport.Create(ctx, &CreateDataExportRequest{
		CreatedAtAfter:  1527811200,
		CreatedAtBefore: 1530403200,
	})
	if err != nil {
		t.Fatalf("DataExport.Create returned error: %v", err)
	}
	if export.JobIdentifier != "orzzsbd7hk67xyu" {
		t.Errorf("JobIdentifier = %v, want orzzsbd7hk67xyu", export.JobIdentifier)
	}
	if export.Status != "pending" {
		t.Errorf("Status = %v, want pending", export.Status)
	}
}

func TestDataExportService_GetStatus(t *testing.T) {
	client, mux, teardown := setup()
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
	export, err := client.DataExport.GetStatus(ctx, "orzzsbd7hk67xyu")
	if err != nil {
		t.Fatalf("DataExport.GetStatus returned error: %v", err)
	}
	if export.JobIdentifier != "orzzsbd7hk67xyu" {
		t.Errorf("JobIdentifier = %v, want orzzsbd7hk67xyu", export.JobIdentifier)
	}
	if export.Status != "completed" {
		t.Errorf("Status = %v, want completed", export.Status)
	}
	if export.DownloadURL != "https://api.intercom.test/download/content/data/orzzsbd7hk67xyu" {
		t.Errorf("DownloadURL = %v, want https://api.intercom.test/download/content/data/orzzsbd7hk67xyu", export.DownloadURL)
	}
	if export.DownloadExpiresAt != "1674917488" {
		t.Errorf("DownloadExpiresAt = %v, want 1674917488", export.DownloadExpiresAt)
	}
}

func TestDataExportService_GetStatus_NotFound(t *testing.T) {
	client, mux, teardown := setup()
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
	_, err := client.DataExport.GetStatus(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestDataExportService_Cancel(t *testing.T) {
	client, mux, teardown := setup()
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
	export, err := client.DataExport.Cancel(ctx, "orzzsbd7hk67xyu")
	if err != nil {
		t.Fatalf("DataExport.Cancel returned error: %v", err)
	}
	if export.JobIdentifier != "orzzsbd7hk67xyu" {
		t.Errorf("JobIdentifier = %v, want orzzsbd7hk67xyu", export.JobIdentifier)
	}
	if export.Status != "canceled" {
		t.Errorf("Status = %v, want canceled", export.Status)
	}
}

func TestDataExportService_Download(t *testing.T) {
	client, mux, teardown := setup()
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
	err := client.DataExport.Download(ctx, "orzzsbd7hk67xyu", &buf)
	if err != nil {
		t.Fatalf("DataExport.Download returned error: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), expectedData) {
		t.Errorf("Download data = %q, want %q", buf.String(), string(expectedData))
	}
}

func TestDataExportService_Download_NotFound(t *testing.T) {
	client, mux, teardown := setup()
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
	err := client.DataExport.Download(ctx, "nonexistent", &buf)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestDataExportService_Create_RateLimited(t *testing.T) {
	client, mux, teardown := setup()
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
	_, err := client.DataExport.Create(ctx, &CreateDataExportRequest{
		CreatedAtAfter:  1527811200,
		CreatedAtBefore: 1530403200,
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsRateLimited(err) {
		t.Errorf("Expected IsRateLimited, got: %v", err)
	}
}

func TestDataExportService_CreateRaw(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.DataExport.CreateRaw(ctx, &CreateDataExportRequest{
		CreatedAtAfter:  1527811200,
		CreatedAtBefore: 1530403200,
	})
	if err != nil {
		t.Fatalf("DataExport.CreateRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	export, err := ParseDataExportCreateResult(result)
	if err != nil {
		t.Fatalf("ParseDataExportCreateResult returned error: %v", err)
	}
	if export.JobIdentifier != "orzzsbd7hk67xyu" {
		t.Errorf("JobIdentifier = %v, want orzzsbd7hk67xyu", export.JobIdentifier)
	}
	if export.Status != "pending" {
		t.Errorf("Status = %v, want pending", export.Status)
	}
}

func TestDataExportService_GetStatusRaw(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.DataExport.GetStatusRaw(ctx, "orzzsbd7hk67xyu")
	if err != nil {
		t.Fatalf("DataExport.GetStatusRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	export, err := ParseDataExportGetStatusResult(result)
	if err != nil {
		t.Fatalf("ParseDataExportGetStatusResult returned error: %v", err)
	}
	if export.JobIdentifier != "orzzsbd7hk67xyu" {
		t.Errorf("JobIdentifier = %v, want orzzsbd7hk67xyu", export.JobIdentifier)
	}
	if export.Status != "completed" {
		t.Errorf("Status = %v, want completed", export.Status)
	}
}

func TestDataExportService_GetStatusRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.DataExport.GetStatusRaw(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("DataExport.GetStatusRaw returned error: %v", err)
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

func TestDataExportService_CancelRaw(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.DataExport.CancelRaw(ctx, "orzzsbd7hk67xyu")
	if err != nil {
		t.Fatalf("DataExport.CancelRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	export, err := ParseDataExportCancelResult(result)
	if err != nil {
		t.Fatalf("ParseDataExportCancelResult returned error: %v", err)
	}
	if export.Status != "canceled" {
		t.Errorf("Status = %v, want canceled", export.Status)
	}
}
