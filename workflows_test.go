package intercom

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestWorkflowsService_Export(t *testing.T) {
	client, mux, teardown := setup()
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
	export, err := client.Workflows.Export(ctx, "67890")
	if err != nil {
		t.Fatalf("Workflows.Export returned error: %v", err)
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

func TestWorkflowsService_Export_NotFound(t *testing.T) {
	client, mux, teardown := setup()
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
	_, err := client.Workflows.Export(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

func TestWorkflowsService_ExportRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/export/workflows/67890", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"export_version": "1.0"}`)
	})

	ctx := context.Background()
	result, err := client.Workflows.ExportRaw(ctx, "67890")
	if err != nil {
		t.Fatalf("Workflows.ExportRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %v, want %v", result.StatusCode, http.StatusOK)
	}
}
