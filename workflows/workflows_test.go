package workflows_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/api"
	"github.com/rassakhatsky/intercom-go-sdk/workflows"
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

func (tc *testCaller) DoRawNoRedirect(ctx context.Context, req *http.Request) (*api.Result, error) {
	noRedirectClient := &http.Client{
		Transport: tc.client.Transport,
		Timeout:   tc.client.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req = req.WithContext(ctx)
	resp, err := noRedirectClient.Do(req)
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

func setup() (svc *workflows.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = workflows.NewService(caller)
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

func TestService_Export(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/workflows/67890", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"export_version": "1.0",
			"exported_at": "2026-01-26T12:00:00Z",
			"app_id": 12345,
			"workflow": {
				"id": "67890",
				"title": "My Workflow",
				"description": "A workflow that handles customer inquiries",
				"trigger_type": "inbound_conversation",
				"state": "live",
				"target_channels": ["chat"],
				"preferred_devices": ["desktop", "mobile"],
				"created_at": "2025-06-15T10:30:00Z",
				"updated_at": "2026-01-20T14:45:00Z",
				"targeting": {},
				"snapshot": {},
				"attributes": [],
				"embedded_rules": []
			}
		}`)
	})

	ctx := context.Background()
	export, err := svc.Export(ctx, "67890")
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	if export.ExportVersion != "1.0" {
		t.Errorf("ExportVersion = %v, want 1.0", export.ExportVersion)
	}
	if export.ExportedAt != "2026-01-26T12:00:00Z" {
		t.Errorf("ExportedAt = %v, want 2026-01-26T12:00:00Z", export.ExportedAt)
	}
	if export.AppID != 12345 {
		t.Errorf("AppID = %v, want 12345", export.AppID)
	}
	if export.Workflow == nil {
		t.Fatal("Workflow is nil")
	}
	if export.Workflow.ID != "67890" {
		t.Errorf("Workflow.ID = %v, want 67890", export.Workflow.ID)
	}
	if export.Workflow.Title != "My Workflow" {
		t.Errorf("Workflow.Title = %v, want My Workflow", export.Workflow.Title)
	}
	if export.Workflow.Description == nil || *export.Workflow.Description != "A workflow that handles customer inquiries" {
		t.Errorf("Workflow.Description = %v, want A workflow that handles customer inquiries", export.Workflow.Description)
	}
	if export.Workflow.TriggerType != "inbound_conversation" {
		t.Errorf("Workflow.TriggerType = %v, want inbound_conversation", export.Workflow.TriggerType)
	}
	if export.Workflow.State != "live" {
		t.Errorf("Workflow.State = %v, want live", export.Workflow.State)
	}
	if len(export.Workflow.TargetChannels) != 1 || export.Workflow.TargetChannels[0] != "chat" {
		t.Errorf("Workflow.TargetChannels = %v, want [chat]", export.Workflow.TargetChannels)
	}
	if len(export.Workflow.PreferredDevices) != 2 {
		t.Errorf("Workflow.PreferredDevices length = %v, want 2", len(export.Workflow.PreferredDevices))
	}
}

func TestService_Export_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/workflows/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type": "error.list",
			"request_id": "b3c8c472-8478-4f10-a29e-a23dbf921c46",
			"errors": [{"code": "not_found", "message": "Workflow not found"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.Export(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

func TestService_ExportRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/workflows/67890", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"export_version": "1.0"}`)
	})

	ctx := context.Background()
	result, err := svc.ExportRaw(ctx, "67890")
	if err != nil {
		t.Fatalf("ExportRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %v, want %v", result.StatusCode, http.StatusOK)
	}
}
