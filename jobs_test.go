package intercom

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestJobsService_GetStatus(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/jobs/status/job_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"job",
			"id":"job_123",
			"url":"https://api.intercom.io/jobs/status/job_123",
			"status":"success",
			"resource_type":"ticket",
			"resource_id":"ticket_456",
			"resource_url":"https://api.intercom.io/tickets/ticket_456"
		}`)
	})

	ctx := context.Background()
	job, err := client.Jobs.GetStatus(ctx, "job_123")
	if err != nil {
		t.Fatalf("Jobs.GetStatus returned error: %v", err)
	}
	if job.ID != "job_123" {
		t.Errorf("ID = %v, want job_123", job.ID)
	}
	if job.Status != "success" {
		t.Errorf("Status = %v, want success", job.Status)
	}
	if job.ResourceType != "ticket" {
		t.Errorf("ResourceType = %v, want ticket", job.ResourceType)
	}
	if job.ResourceID != "ticket_456" {
		t.Errorf("ResourceID = %v, want ticket_456", job.ResourceID)
	}
}

func TestJobsService_GetStatus_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/jobs/status/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Job not found"}]}`)
	})

	ctx := context.Background()
	_, err := client.Jobs.GetStatus(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestJobsService_GetStatus_Pending(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/jobs/status/job_pending", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"type":"job",
			"id":"job_pending",
			"url":"https://api.intercom.io/jobs/status/job_pending",
			"status":"pending",
			"resource_type":"ticket",
			"resource_id":null,
			"resource_url":null
		}`)
	})

	ctx := context.Background()
	job, err := client.Jobs.GetStatus(ctx, "job_pending")
	if err != nil {
		t.Fatalf("Jobs.GetStatus returned error: %v", err)
	}
	if job.Status != "pending" {
		t.Errorf("Status = %v, want pending", job.Status)
	}
	if job.ResourceID != "" {
		t.Errorf("ResourceID = %v, want empty", job.ResourceID)
	}
}

func TestJobsService_GetStatusRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/jobs/status/job_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"job",
			"id":"job_123",
			"status":"success",
			"resource_type":"ticket",
			"resource_id":"ticket_456"
		}`)
	})

	ctx := context.Background()
	result, err := client.Jobs.GetStatusRaw(ctx, "job_123")
	if err != nil {
		t.Fatalf("Jobs.GetStatusRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	job, err := ParseJobGetStatusResult(result)
	if err != nil {
		t.Fatalf("ParseJobGetStatusResult returned error: %v", err)
	}
	if job.ID != "job_123" {
		t.Errorf("ID = %v, want job_123", job.ID)
	}
	if job.Status != "success" {
		t.Errorf("Status = %v, want success", job.Status)
	}
}

func TestJobsService_GetStatusRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/jobs/status/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Job not found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Jobs.GetStatusRaw(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Jobs.GetStatusRaw returned error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Expected Error to be populated")
	}
	if result.Error.Code != "not_found" {
		t.Errorf("Error.Code = %v, want not_found", result.Error.Code)
	}
}
