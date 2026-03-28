package export_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/export"
	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// testCaller implements api.Caller for testing, backed by an httptest.Server.
type testCaller struct {
	baseURL string
	client  *http.Client
}

func (tc *testCaller) NewRequest(method, urlStr string, body any) (*http.Request, error) {
	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(jsonBody)
	}
	req, err := http.NewRequest(method, tc.baseURL+"/"+urlStr, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (tc *testCaller) DoRaw(ctx context.Context, req *http.Request) (*api.Result, error) {
	req = req.WithContext(ctx)
	resp, err := tc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return api.BuildResult(resp, b), nil
}

func (tc *testCaller) Do(ctx context.Context, req *http.Request, v any) (*api.Response, error) {
	result, err := tc.DoRaw(ctx, req)
	if err != nil {
		return nil, err
	}
	response := &api.Response{Result: result}
	if result.Error != nil {
		return response, api.ResultError(result)
	}
	if v != nil && result.StatusCode != http.StatusNoContent && len(result.Body) > 0 {
		if err := json.Unmarshal(result.Body, v); err != nil {
			return response, err
		}
	}
	return response, nil
}

func (tc *testCaller) DoDownload(ctx context.Context, req *http.Request, w io.Writer) error {
	req = req.WithContext(ctx)
	resp, err := tc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("HTTP %d: failed to read error body: %w", resp.StatusCode, readErr)
		}
		result := api.BuildResult(resp, b)
		return api.ResultError(result)
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

func setupReporting() (svc *export.ReportingService, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = export.NewReportingService(caller)
	return svc, mux, server.Close
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method = %v, want %v", got, want)
	}
}

func testHeader(t *testing.T, r *http.Request, header, want string) {
	t.Helper()
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header %v = %q, want %q", header, got, want)
	}
}

func TestReportingService_Enqueue(t *testing.T) {
	svc, mux, teardown := setupReporting()
	defer teardown()

	mux.HandleFunc("/export/reporting_data/enqueue", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body export.EnqueueRequest
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
	job, err := svc.Enqueue(ctx, &export.EnqueueRequest{
		DatasetID:    "conversation",
		AttributeIDs: []string{"conversation_id", "conversation_started_at"},
		StartTime:    1717490000,
		EndTime:      1717510000,
	})
	if err != nil {
		t.Fatalf("Enqueue returned error: %v", err)
	}
	if job.JobIdentifier != "job1" {
		t.Errorf("JobIdentifier = %v, want job1", job.JobIdentifier)
	}
	if job.Status != "pending" {
		t.Errorf("Status = %v, want pending", job.Status)
	}
}

func TestReportingService_Enqueue_BadRequest(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
	_, err := svc.Enqueue(ctx, &export.EnqueueRequest{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestReportingService_Enqueue_RateLimited(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
	_, err := svc.Enqueue(ctx, &export.EnqueueRequest{
		DatasetID:    "conversation",
		AttributeIDs: []string{"conversation_id"},
		StartTime:    1717490000,
		EndTime:      1717510000,
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsRateLimited(err) {
		t.Errorf("Expected IsRateLimited, got: %v", err)
	}
}

func TestReportingService_GetStatus(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
	job, err := svc.GetStatus(ctx, "job1")
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
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

func TestReportingService_GetStatus_NotFound(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
	_, err := svc.GetStatus(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestReportingService_GetDatasets(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
	datasets, err := svc.GetDatasets(ctx)
	if err != nil {
		t.Fatalf("GetDatasets returned error: %v", err)
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

func TestReportingService_Download(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
	err := svc.Download(ctx, "job1", &buf)
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), expectedData) {
		t.Errorf("Download data = %q, want %q", buf.String(), string(expectedData))
	}
}

func TestReportingService_Download_NotFound(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
	err := svc.Download(ctx, "nonexistent", &buf)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestReportingService_EnqueueRaw(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
	result, err := svc.EnqueueRaw(ctx, &export.EnqueueRequest{
		DatasetID:    "conversation",
		AttributeIDs: []string{"conversation_id", "conversation_started_at"},
		StartTime:    1717490000,
		EndTime:      1717510000,
	})
	if err != nil {
		t.Fatalf("EnqueueRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	job, err := export.ParseEnqueueResult(result)
	if err != nil {
		t.Fatalf("ParseEnqueueResult returned error: %v", err)
	}
	if job.JobIdentifier != "job1" {
		t.Errorf("JobIdentifier = %v, want job1", job.JobIdentifier)
	}
}

func TestReportingService_GetStatusRaw(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
	result, err := svc.GetStatusRaw(ctx, "job1")
	if err != nil {
		t.Fatalf("GetStatusRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	job, err := export.ParseGetStatusResult(result)
	if err != nil {
		t.Fatalf("ParseGetStatusResult returned error: %v", err)
	}
	if job.Status != "complete" {
		t.Errorf("Status = %v, want complete", job.Status)
	}
}

func TestReportingService_GetStatusRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
}

func TestReportingService_GetDatasetsRaw(t *testing.T) {
	svc, mux, teardown := setupReporting()
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
	result, err := svc.GetDatasetsRaw(ctx)
	if err != nil {
		t.Fatalf("GetDatasetsRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	resp, err := export.ParseGetDatasetsResult(result)
	if err != nil {
		t.Fatalf("ParseGetDatasetsResult returned error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
	}
	if resp.Data[0].ID != "conversation" {
		t.Errorf("Data[0].ID = %v, want conversation", resp.Data[0].ID)
	}
}
