package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestDataEventsService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body CreateDataEventRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.EventName != "invited-friend" {
			t.Errorf("body.EventName = %v, want invited-friend", body.EventName)
		}
		if body.CreatedAt != 1671028894 {
			t.Errorf("body.CreatedAt = %v, want 1671028894", body.CreatedAt)
		}
		if body.UserID != "314159" {
			t.Errorf("body.UserID = %v, want 314159", body.UserID)
		}
		w.WriteHeader(http.StatusAccepted)
	})

	ctx := context.Background()
	err := client.DataEvents.Create(ctx, &CreateDataEventRequest{
		EventName: "invited-friend",
		CreatedAt: 1671028894,
		UserID:    "314159",
	})
	if err != nil {
		t.Fatalf("DataEvents.Create returned error: %v", err)
	}
}

func TestDataEventsService_Create_WithMetadata(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["event_name"] != "placed-order" {
			t.Errorf("body.event_name = %v, want placed-order", body["event_name"])
		}
		metadata, ok := body["metadata"].(map[string]any)
		if !ok {
			t.Fatal("body.metadata is not a map")
		}
		if metadata["order_id"] != "1234" {
			t.Errorf("metadata.order_id = %v, want 1234", metadata["order_id"])
		}
		if metadata["source"] != "desktop" {
			t.Errorf("metadata.source = %v, want desktop", metadata["source"])
		}
		w.WriteHeader(http.StatusAccepted)
	})

	ctx := context.Background()
	err := client.DataEvents.Create(ctx, &CreateDataEventRequest{
		EventName: "placed-order",
		CreatedAt: 1671028894,
		Email:     "frodo@example.com",
		Metadata: map[string]any{
			"order_id": "1234",
			"source":   "desktop",
		},
	})
	if err != nil {
		t.Fatalf("DataEvents.Create returned error: %v", err)
	}
}

func TestDataEventsService_Create_WithContactID(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		var body CreateDataEventRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.ID != "8a88a590-e1c3-41e2-a502-e0649dbf721c" {
			t.Errorf("body.ID = %v, want 8a88a590-e1c3-41e2-a502-e0649dbf721c", body.ID)
		}
		w.WriteHeader(http.StatusAccepted)
	})

	ctx := context.Background()
	err := client.DataEvents.Create(ctx, &CreateDataEventRequest{
		EventName: "signed-up",
		CreatedAt: 1671028894,
		ID:        "8a88a590-e1c3-41e2-a502-e0649dbf721c",
	})
	if err != nil {
		t.Fatalf("DataEvents.Create returned error: %v", err)
	}
}

func TestDataEventsService_Create_Unauthorized(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	err := client.DataEvents.Create(ctx, &CreateDataEventRequest{
		EventName: "invited-friend",
		CreatedAt: 1671028894,
		UserID:    "314159",
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}

func TestDataEventsService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		if got := r.URL.Query().Get("type"); got != "user" {
			t.Errorf("query type = %v, want user", got)
		}
		if got := r.URL.Query().Get("user_id"); got != "314159" {
			t.Errorf("query user_id = %v, want 314159", got)
		}
		fmt.Fprint(w, `{
			"type":"event.summary",
			"events":[
				{
					"event_name":"invited-friend",
					"created_at":1671028894,
					"user_id":"314159",
					"metadata":{"invite_code":"ADDAFRIEND"}
				},
				{
					"event_name":"placed-order",
					"created_at":1671028900,
					"user_id":"314159"
				}
			],
			"pages":{
				"next":"https://api.intercom.io/events?next_page"
			},
			"email":"frodo@example.com",
			"intercom_user_id":"63a0979a5eeebeaf28dd56ba",
			"user_id":"314159"
		}`)
	})

	ctx := context.Background()
	result, err := client.DataEvents.List(ctx, &ListDataEventsOptions{
		Type:   "user",
		UserID: "314159",
	})
	if err != nil {
		t.Fatalf("DataEvents.List returned error: %v", err)
	}
	if result.Type != "event.summary" {
		t.Errorf("Type = %v, want event.summary", result.Type)
	}
	if len(result.Events) != 2 {
		t.Fatalf("Events length = %d, want 2", len(result.Events))
	}
	if result.Events[0].EventName != "invited-friend" {
		t.Errorf("Events[0].EventName = %v, want invited-friend", result.Events[0].EventName)
	}
	if result.Events[0].CreatedAt != 1671028894 {
		t.Errorf("Events[0].CreatedAt = %v, want 1671028894", result.Events[0].CreatedAt)
	}
	if result.UserID != "314159" {
		t.Errorf("UserID = %v, want 314159", result.UserID)
	}
	if result.Email != "frodo@example.com" {
		t.Errorf("Email = %v, want frodo@example.com", result.Email)
	}
}

func TestDataEventsService_List_ByEmail(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("type"); got != "user" {
			t.Errorf("query type = %v, want user", got)
		}
		if got := r.URL.Query().Get("email"); got != "frodo@example.com" {
			t.Errorf("query email = %v, want frodo@example.com", got)
		}
		fmt.Fprint(w, `{
			"type":"event.summary",
			"events":[],
			"email":"frodo@example.com"
		}`)
	})

	ctx := context.Background()
	result, err := client.DataEvents.List(ctx, &ListDataEventsOptions{
		Type:  "user",
		Email: "frodo@example.com",
	})
	if err != nil {
		t.Fatalf("DataEvents.List returned error: %v", err)
	}
	if len(result.Events) != 0 {
		t.Errorf("Events length = %d, want 0", len(result.Events))
	}
}

func TestDataEventsService_List_ByIntercomUserID(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("intercom_user_id"); got != "63a0979a5eeebeaf28dd56ba" {
			t.Errorf("query intercom_user_id = %v, want 63a0979a5eeebeaf28dd56ba", got)
		}
		fmt.Fprint(w, `{
			"type":"event.summary",
			"events":[],
			"intercom_user_id":"63a0979a5eeebeaf28dd56ba"
		}`)
	})

	ctx := context.Background()
	result, err := client.DataEvents.List(ctx, &ListDataEventsOptions{
		Type:           "user",
		IntercomUserID: "63a0979a5eeebeaf28dd56ba",
	})
	if err != nil {
		t.Fatalf("DataEvents.List returned error: %v", err)
	}
	if result.IntercomUserID != "63a0979a5eeebeaf28dd56ba" {
		t.Errorf("IntercomUserID = %v, want 63a0979a5eeebeaf28dd56ba", result.IntercomUserID)
	}
}

func TestDataEventsService_List_WithSummary(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("summary"); got != "true" {
			t.Errorf("query summary = %v, want true", got)
		}
		fmt.Fprint(w, `{
			"type":"event.summary",
			"events":[
				{
					"event_name":"placed-order",
					"created_at":1671028894,
					"user_id":"314159"
				}
			],
			"user_id":"314159"
		}`)
	})

	ctx := context.Background()
	summary := true
	result, err := client.DataEvents.List(ctx, &ListDataEventsOptions{
		Type:    "user",
		UserID:  "314159",
		Summary: &summary,
	})
	if err != nil {
		t.Fatalf("DataEvents.List returned error: %v", err)
	}
	if len(result.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(result.Events))
	}
}

func TestDataEventsService_CreateSummaries(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events/summaries", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body CreateEventSummariesRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.UserID != "314159" {
			t.Errorf("body.UserID = %v, want 314159", body.UserID)
		}
		if len(body.EventSummaries) != 1 {
			t.Fatalf("body.EventSummaries length = %d, want 1", len(body.EventSummaries))
		}
		if body.EventSummaries[0].EventName != "placed-order" {
			t.Errorf("body.EventSummaries[0].EventName = %v, want placed-order", body.EventSummaries[0].EventName)
		}
		if body.EventSummaries[0].Count != 5 {
			t.Errorf("body.EventSummaries[0].Count = %v, want 5", body.EventSummaries[0].Count)
		}
		if body.EventSummaries[0].First != 1671028894 {
			t.Errorf("body.EventSummaries[0].First = %v, want 1671028894", body.EventSummaries[0].First)
		}
		if body.EventSummaries[0].Last != 1671128894 {
			t.Errorf("body.EventSummaries[0].Last = %v, want 1671128894", body.EventSummaries[0].Last)
		}
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	err := client.DataEvents.CreateSummaries(ctx, &CreateEventSummariesRequest{
		UserID: "314159",
		EventSummaries: []EventSummary{{
			EventName: "placed-order",
			Count:     5,
			First:     1671028894,
			Last:      1671128894,
		}},
	})
	if err != nil {
		t.Fatalf("DataEvents.CreateSummaries returned error: %v", err)
	}
}

func TestDataEventsService_CreateSummaries_Unauthorized(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events/summaries", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-789",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	err := client.DataEvents.CreateSummaries(ctx, &CreateEventSummariesRequest{
		UserID: "314159",
		EventSummaries: []EventSummary{{
			EventName: "placed-order",
			Count:     1,
			First:     1671028894,
			Last:      1671028894,
		}},
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}

func TestDataEventsService_CreateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusAccepted)
	})

	ctx := context.Background()
	result, err := client.DataEvents.CreateRaw(ctx, &CreateDataEventRequest{
		EventName: "invited-friend",
		CreatedAt: 1671028894,
		UserID:    "314159",
	})
	if err != nil {
		t.Fatalf("DataEvents.CreateRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusAccepted {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusAccepted)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
}

func TestDataEventsService_CreateRaw_Unauthorized(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	result, err := client.DataEvents.CreateRaw(ctx, &CreateDataEventRequest{
		EventName: "invited-friend",
		CreatedAt: 1671028894,
		UserID:    "314159",
	})
	if err != nil {
		t.Fatalf("DataEvents.CreateRaw returned error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Expected Error to be populated")
	}
	if result.Error.Code != "unauthorized" {
		t.Errorf("Error.Code = %v, want unauthorized", result.Error.Code)
	}
	if result.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnauthorized)
	}
}

func TestDataEventsService_ListRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"event.summary",
			"events":[
				{"event_name":"invited-friend","created_at":1671028894,"user_id":"314159"}
			],
			"user_id":"314159"
		}`)
	})

	ctx := context.Background()
	result, err := client.DataEvents.ListRaw(ctx, &ListDataEventsOptions{
		Type:   "user",
		UserID: "314159",
	})
	if err != nil {
		t.Fatalf("DataEvents.ListRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	data, err := ParseDataEventListResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if data.Type != "event.summary" {
		t.Errorf("Type = %v, want event.summary", data.Type)
	}
	if len(data.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(data.Events))
	}
	if data.Events[0].EventName != "invited-friend" {
		t.Errorf("Events[0].EventName = %v, want invited-friend", data.Events[0].EventName)
	}
}

func TestDataEventsService_CreateSummariesRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/events/summaries", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	result, err := client.DataEvents.CreateSummariesRaw(ctx, &CreateEventSummariesRequest{
		UserID: "314159",
		EventSummaries: []EventSummary{{
			EventName: "placed-order",
			Count:     5,
			First:     1671028894,
			Last:      1671128894,
		}},
	})
	if err != nil {
		t.Fatalf("DataEvents.CreateSummariesRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
}
