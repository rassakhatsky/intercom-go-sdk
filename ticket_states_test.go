package intercom

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestTicketStatesService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_states", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"ticket_state",
					"id":"8269",
					"category":"submitted",
					"internal_label":"Submitted",
					"external_label":"Submitted",
					"archived":false
				},
				{
					"type":"ticket_state",
					"id":"8270",
					"category":"in_progress",
					"internal_label":"In progress",
					"external_label":"In progress",
					"archived":false
				},
				{
					"type":"ticket_state",
					"id":"8271",
					"category":"waiting_on_customer",
					"internal_label":"Waiting on customer",
					"external_label":"Waiting on customer",
					"archived":true
				}
			]
		}`)
	})

	ctx := context.Background()
	result, err := client.TicketStates.List(ctx)
	if err != nil {
		t.Fatalf("TicketStates.List returned error: %v", err)
	}
	if result.Type != "list" {
		t.Errorf("Type = %v, want list", result.Type)
	}
	if len(result.Data) != 3 {
		t.Fatalf("Data length = %d, want 3", len(result.Data))
	}

	ts := result.Data[0]
	if ts.Type != "ticket_state" {
		t.Errorf("Data[0].Type = %v, want ticket_state", ts.Type)
	}
	if ts.ID != "8269" {
		t.Errorf("Data[0].ID = %v, want 8269", ts.ID)
	}
	if ts.Category != "submitted" {
		t.Errorf("Data[0].Category = %v, want submitted", ts.Category)
	}
	if ts.InternalLabel != "Submitted" {
		t.Errorf("Data[0].InternalLabel = %v, want Submitted", ts.InternalLabel)
	}
	if ts.ExternalLabel != "Submitted" {
		t.Errorf("Data[0].ExternalLabel = %v, want Submitted", ts.ExternalLabel)
	}
	if ts.Archived != false {
		t.Errorf("Data[0].Archived = %v, want false", ts.Archived)
	}

	ts2 := result.Data[1]
	if ts2.Category != "in_progress" {
		t.Errorf("Data[1].Category = %v, want in_progress", ts2.Category)
	}

	ts3 := result.Data[2]
	if ts3.Archived != true {
		t.Errorf("Data[2].Archived = %v, want true", ts3.Archived)
	}
}

func TestTicketStatesService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_states", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ts-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"ticket_state","id":"8269","category":"submitted","internal_label":"Submitted","external_label":"Submitted","archived":false}]}`)
	})

	ctx := context.Background()
	result, err := client.TicketStates.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-ts-list" {
		t.Errorf("Header X-Request-Id = %q, want req-ts-list", result.Header.Get("X-Request-Id"))
	}
	tsl, err := ParseTicketStateListResult(result)
	if err != nil {
		t.Fatalf("ParseTicketStateListResult returned error: %v", err)
	}
	if len(tsl.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(tsl.Data))
	}
	if tsl.Data[0].ID != "8269" {
		t.Errorf("Data[0].ID = %v, want 8269", tsl.Data[0].ID)
	}
	if tsl.Data[0].Category != "submitted" {
		t.Errorf("Data[0].Category = %v, want submitted", tsl.Data[0].Category)
	}
}

func TestTicketStatesService_List_Unauthorized(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_states", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-456",
			"errors":[{"code":"unauthorized","message":"Unauthorized"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.TicketStates.List(ctx)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}
