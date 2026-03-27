package intercom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestExportReportingService_Enqueue(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/enqueue", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body EnqueueExportReportingRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.DatasetID != "conversation" {
			t.Errorf("body.DatasetID = %v, want conversation", body.DatasetID)
		}
		if len(body.AttributeIDs) != 2 {
			t.Errorf("body.AttributeIDs length = %v, want 2", len(body.AttributeIDs))
		}
		if body.StartTime != 1717490000 {
			t.Errorf("body.StartTime = %v, want 1717490000", body.StartTime)
		}
		if body.EndTime != 1717510000 {
			t.Errorf("body.EndTime = %v, want 1717510000", body.EndTime)
		}
		fmt.Fprint(w, `{
			"job_identifier":"job1",
			"status":"pending",
			"download_url":"",
			"download_expires_at":""
		}`)
	})

	ctx := context.Background()
	job, err := client.ExportReporting.Enqueue(ctx, &EnqueueExportReportingRequest{
		DatasetID:    "conversation",
		AttributeIDs: []string{"conversation_id", "conversation_started_at"},
		StartTime:    1717490000,
		EndTime:      1717510000,
	})
	if err != nil {
		t.Fatalf("ExportReporting.Enqueue returned error: %v", err)
	}
	if job.JobIdentifier != "job1" {
		t.Errorf("JobIdentifier = %v, want job1", job.JobIdentifier)
	}
	if job.Status != "pending" {
		t.Errorf("Status = %v, want pending", job.Status)
	}
}

func TestExportReportingService_Enqueue_BadRequest(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/enqueue", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"bad_request","message":"'dataset_id' is a required parameter"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.ExportReporting.Enqueue(ctx, &EnqueueExportReportingRequest{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestExportReportingService_Enqueue_RateLimited(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/enqueue", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-456",
			"errors":[{"code":"rate_limit_exceeded","message":"Exceeded rate limit of 5 pending reporting dataset export jobs"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.ExportReporting.Enqueue(ctx, &EnqueueExportReportingRequest{
		DatasetID:    "conversation",
		AttributeIDs: []string{"conversation_id"},
		StartTime:    1717490000,
		EndTime:      1717510000,
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsRateLimited(err) {
		t.Errorf("Expected IsRateLimited, got: %v", err)
	}
}

func TestExportReportingService_GetStatus(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/job1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"job_identifier":"job1",
			"status":"complete",
			"download_url":"https://api.intercom.test/download/reporting_data/job1",
			"download_expires_at":"1674917488"
		}`)
	})

	ctx := context.Background()
	job, err := client.ExportReporting.GetStatus(ctx, "job1")
	if err != nil {
		t.Fatalf("ExportReporting.GetStatus returned error: %v", err)
	}
	if job.JobIdentifier != "job1" {
		t.Errorf("JobIdentifier = %v, want job1", job.JobIdentifier)
	}
	if job.Status != "complete" {
		t.Errorf("Status = %v, want complete", job.Status)
	}
	if job.DownloadURL != "https://api.intercom.test/download/reporting_data/job1" {
		t.Errorf("DownloadURL = %v, want https://api.intercom.test/download/reporting_data/job1", job.DownloadURL)
	}
}

func TestExportReportingService_GetStatus_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-789",
			"errors":[{"code":"not_found","message":"Export job not found for identifier: nonexistent"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.ExportReporting.GetStatus(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestExportReportingService_GetDatasets(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/get_datasets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"id":"conversation",
					"name":"Conversation",
					"attributes":[
						{"id":"conversation_id","name":"Conversation ID","type":"string"},
						{"id":"conversation_started_at","name":"Started At","type":"datetime"}
					]
				},
				{
					"id":"conversation_part",
					"name":"Conversation Part",
					"attributes":[
						{"id":"part_id","name":"Part ID","type":"string"}
					]
				}
			]
		}`)
	})

	ctx := context.Background()
	datasets, err := client.ExportReporting.GetDatasets(ctx)
	if err != nil {
		t.Fatalf("ExportReporting.GetDatasets returned error: %v", err)
	}
	if len(datasets) != 2 {
		t.Fatalf("len(datasets) = %v, want 2", len(datasets))
	}
	if datasets[0].ID != "conversation" {
		t.Errorf("datasets[0].ID = %v, want conversation", datasets[0].ID)
	}
	if datasets[0].Name != "Conversation" {
		t.Errorf("datasets[0].Name = %v, want Conversation", datasets[0].Name)
	}
	if len(datasets[0].Attributes) != 2 {
		t.Errorf("len(datasets[0].Attributes) = %v, want 2", len(datasets[0].Attributes))
	}
	if datasets[0].Attributes[0].ID != "conversation_id" {
		t.Errorf("datasets[0].Attributes[0].ID = %v, want conversation_id", datasets[0].Attributes[0].ID)
	}
}

func TestExportReportingService_Download(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	expectedData := []byte("conversation_id,started_at\n123,2024-01-01\n")

	mux.HandleFunc("/download/reporting_data/job1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		if accept := r.Header.Get("Accept"); accept != "application/octet-stream" {
			t.Errorf("Accept header = %v, want application/octet-stream", accept)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(expectedData)
	})

	ctx := context.Background()
	var buf bytes.Buffer
	err := client.ExportReporting.Download(ctx, "job1", &buf)
	if err != nil {
		t.Fatalf("ExportReporting.Download returned error: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), expectedData) {
		t.Errorf("Download data = %q, want %q", buf.String(), string(expectedData))
	}
}

func TestExportReportingService_Download_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/download/reporting_data/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-abc",
			"errors":[{"code":"not_found","message":"Export job not found for identifier: nonexistent"}]
		}`)
	})

	ctx := context.Background()
	var buf bytes.Buffer
	err := client.ExportReporting.Download(ctx, "nonexistent", &buf)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestExportReportingService_EnqueueRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/enqueue", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"job_identifier":"job1",
			"status":"pending",
			"download_url":"",
			"download_expires_at":""
		}`)
	})

	ctx := context.Background()
	result, err := client.ExportReporting.EnqueueRaw(ctx, &EnqueueExportReportingRequest{
		DatasetID:    "conversation",
		AttributeIDs: []string{"conversation_id", "conversation_started_at"},
		StartTime:    1717490000,
		EndTime:      1717510000,
	})
	if err != nil {
		t.Fatalf("ExportReporting.EnqueueRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	job, err := ParseExportReportingEnqueueResult(result)
	if err != nil {
		t.Fatalf("ParseExportReportingEnqueueResult returned error: %v", err)
	}
	if job.JobIdentifier != "job1" {
		t.Errorf("JobIdentifier = %v, want job1", job.JobIdentifier)
	}
}

func TestExportReportingService_GetStatusRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/job1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"job_identifier":"job1",
			"status":"complete",
			"download_url":"https://api.intercom.test/download/reporting_data/job1",
			"download_expires_at":"1674917488"
		}`)
	})

	ctx := context.Background()
	result, err := client.ExportReporting.GetStatusRaw(ctx, "job1")
	if err != nil {
		t.Fatalf("ExportReporting.GetStatusRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	job, err := ParseExportReportingGetStatusResult(result)
	if err != nil {
		t.Fatalf("ParseExportReportingGetStatusResult returned error: %v", err)
	}
	if job.Status != "complete" {
		t.Errorf("Status = %v, want complete", job.Status)
	}
}

func TestExportReportingService_GetStatusRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-789",
			"errors":[{"code":"not_found","message":"Export job not found"}]
		}`)
	})

	ctx := context.Background()
	result, err := client.ExportReporting.GetStatusRaw(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("ExportReporting.GetStatusRaw returned error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Expected Error to be populated")
	}
	if result.Error.Code != "not_found" {
		t.Errorf("Error.Code = %v, want not_found", result.Error.Code)
	}
}

func TestExportReportingService_GetDatasetsRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/get_datasets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"id":"conversation",
					"name":"Conversation",
					"attributes":[
						{"id":"conversation_id","name":"Conversation ID","type":"string"}
					]
				}
			]
		}`)
	})

	ctx := context.Background()
	result, err := client.ExportReporting.GetDatasetsRaw(ctx)
	if err != nil {
		t.Fatalf("ExportReporting.GetDatasetsRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	resp, err := ParseExportReportingGetDatasetsResult(result)
	if err != nil {
		t.Fatalf("ParseExportReportingGetDatasetsResult returned error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
	}
	if resp.Data[0].ID != "conversation" {
		t.Errorf("Data[0].ID = %v, want conversation", resp.Data[0].ID)
	}
}
