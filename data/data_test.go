package data_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/data"
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

func setupCaller() (*testCaller, *http.ServeMux, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	return caller, mux, server.Close
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

// --- DataEvents tests ---

func TestEventsService_Create(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body data.CreateEventRequest
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
	err := svc.Create(ctx, &data.CreateEventRequest{
		EventName: "invited-friend",
		CreatedAt: 1671028894,
		UserID:    "314159",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

func TestEventsService_Create_WithMetadata(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

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
	err := svc.Create(ctx, &data.CreateEventRequest{
		EventName: "placed-order",
		CreatedAt: 1671028894,
		Email:     "frodo@example.com",
		Metadata: map[string]any{
			"order_id": "1234",
			"source":   "desktop",
		},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

func TestEventsService_Create_WithContactID(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		var body data.CreateEventRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.ID != "8a88a590-e1c3-41e2-a502-e0649dbf721c" {
			t.Errorf("body.ID = %v, want 8a88a590-e1c3-41e2-a502-e0649dbf721c", body.ID)
		}
		w.WriteHeader(http.StatusAccepted)
	})

	ctx := context.Background()
	err := svc.Create(ctx, &data.CreateEventRequest{
		EventName: "signed-up",
		CreatedAt: 1671028894,
		ID:        "8a88a590-e1c3-41e2-a502-e0649dbf721c",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

func TestEventsService_Create_Unauthorized(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	err := svc.Create(ctx, &data.CreateEventRequest{
		EventName: "invited-friend",
		CreatedAt: 1671028894,
		UserID:    "314159",
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}

func TestEventsService_List(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

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
	result, err := svc.List(ctx, &data.ListOptions{
		Type:   "user",
		UserID: "314159",
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
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

func TestEventsService_List_ByEmail(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

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
	result, err := svc.List(ctx, &data.ListOptions{
		Type:  "user",
		Email: "frodo@example.com",
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(result.Events) != 0 {
		t.Errorf("Events length = %d, want 0", len(result.Events))
	}
}

func TestEventsService_List_ByIntercomUserID(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

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
	result, err := svc.List(ctx, &data.ListOptions{
		Type:           "user",
		IntercomUserID: "63a0979a5eeebeaf28dd56ba",
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if result.IntercomUserID != "63a0979a5eeebeaf28dd56ba" {
		t.Errorf("IntercomUserID = %v, want 63a0979a5eeebeaf28dd56ba", result.IntercomUserID)
	}
}

func TestEventsService_List_WithSummary(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

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
	result, err := svc.List(ctx, &data.ListOptions{
		Type:    "user",
		UserID:  "314159",
		Summary: &summary,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(result.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(result.Events))
	}
}

func TestEventsService_CreateSummaries(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

	mux.HandleFunc("/events/summaries", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body data.CreateSummariesRequest
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
	err := svc.CreateSummaries(ctx, &data.CreateSummariesRequest{
		UserID: "314159",
		EventSummaries: []data.EventSummary{{
			EventName: "placed-order",
			Count:     5,
			First:     1671028894,
			Last:      1671128894,
		}},
	})
	if err != nil {
		t.Fatalf("CreateSummaries returned error: %v", err)
	}
}

func TestEventsService_CreateSummaries_Unauthorized(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

	mux.HandleFunc("/events/summaries", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-789",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	err := svc.CreateSummaries(ctx, &data.CreateSummariesRequest{
		UserID: "314159",
		EventSummaries: []data.EventSummary{{
			EventName: "placed-order",
			Count:     1,
			First:     1671028894,
			Last:      1671028894,
		}},
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}

func TestEventsService_CreateRaw(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusAccepted)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &data.CreateEventRequest{
		EventName: "invited-friend",
		CreatedAt: 1671028894,
		UserID:    "314159",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusAccepted {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusAccepted)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
}

func TestEventsService_CreateRaw_Unauthorized(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &data.CreateEventRequest{
		EventName: "invited-friend",
		CreatedAt: 1671028894,
		UserID:    "314159",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
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

func TestEventsService_ListRaw(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

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
	result, err := svc.ListRaw(ctx, &data.ListOptions{
		Type:   "user",
		UserID: "314159",
	})
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	resp, err := data.ParseListResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if resp.Type != "event.summary" {
		t.Errorf("Type = %v, want event.summary", resp.Type)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(resp.Events))
	}
	if resp.Events[0].EventName != "invited-friend" {
		t.Errorf("Events[0].EventName = %v, want invited-friend", resp.Events[0].EventName)
	}
}

func TestEventsService_CreateSummariesRaw(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewEventsService(caller)

	mux.HandleFunc("/events/summaries", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	result, err := svc.CreateSummariesRaw(ctx, &data.CreateSummariesRequest{
		UserID: "314159",
		EventSummaries: []data.EventSummary{{
			EventName: "placed-order",
			Count:     5,
			First:     1671028894,
			Last:      1671128894,
		}},
	})
	if err != nil {
		t.Fatalf("CreateSummariesRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
}

// --- DataAttributes tests ---

func TestAttributesService_List(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"data_attribute",
					"id":1,
					"name":"name",
					"full_name":"name",
					"label":"Company name",
					"description":"The name of a company",
					"data_type":"string",
					"api_writable":true,
					"ui_writable":false,
					"messenger_writable":true,
					"custom":false,
					"archived":false,
					"model":"company"
				},
				{
					"type":"data_attribute",
					"id":2,
					"name":"paid_subscriber",
					"full_name":"custom_attributes.paid_subscriber",
					"label":"Paid Subscriber",
					"description":"Whether the user is a paid subscriber",
					"data_type":"boolean",
					"api_writable":true,
					"ui_writable":false,
					"messenger_writable":false,
					"custom":true,
					"archived":false,
					"model":"contact",
					"admin_id":"12345",
					"created_at":1671028894,
					"updated_at":1671028894
				}
			]
		}`)
	})

	ctx := context.Background()
	list, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if list.Type != "list" {
		t.Errorf("Type = %v, want list", list.Type)
	}
	if len(list.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(list.Data))
	}
	if list.Data[0].Name != "name" {
		t.Errorf("Data[0].Name = %v, want name", list.Data[0].Name)
	}
	if list.Data[0].Model != "company" {
		t.Errorf("Data[0].Model = %v, want company", list.Data[0].Model)
	}
	if list.Data[1].Custom != true {
		t.Errorf("Data[1].Custom = %v, want true", list.Data[1].Custom)
	}
	if list.Data[1].AdminID != "12345" {
		t.Errorf("Data[1].AdminID = %v, want 12345", list.Data[1].AdminID)
	}
}

func TestAttributesService_List_WithOptions(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("model"); got != "company" {
			t.Errorf("query model = %v, want company", got)
		}
		if got := r.URL.Query().Get("include_archived"); got != "true" {
			t.Errorf("query include_archived = %v, want true", got)
		}
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"data_attribute",
					"id":1,
					"name":"name",
					"model":"company"
				}
			]
		}`)
	})

	ctx := context.Background()
	includeArchived := true
	list, err := svc.List(ctx, &data.ListAttributesOptions{
		Model:           "company",
		IncludeArchived: &includeArchived,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(list.Data))
	}
}

func TestAttributesService_Create(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body data.CreateAttributeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Mithril Shirt" {
			t.Errorf("body.Name = %v, want Mithril Shirt", body.Name)
		}
		if body.Model != "company" {
			t.Errorf("body.Model = %v, want company", body.Model)
		}
		if body.DataType != "string" {
			t.Errorf("body.DataType = %v, want string", body.DataType)
		}
		fmt.Fprint(w, `{
			"id":37,
			"type":"data_attribute",
			"name":"Mithril Shirt",
			"full_name":"custom_attributes.Mithril Shirt",
			"label":"Mithril Shirt",
			"data_type":"string",
			"api_writable":true,
			"ui_writable":false,
			"messenger_writable":false,
			"custom":true,
			"archived":false,
			"admin_id":"12345",
			"created_at":1734537756,
			"updated_at":1734537756,
			"model":"company"
		}`)
	})

	ctx := context.Background()
	attr, err := svc.Create(ctx, &data.CreateAttributeRequest{
		Name:     "Mithril Shirt",
		Model:    "company",
		DataType: "string",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if attr.ID != 37 {
		t.Errorf("ID = %v, want 37", attr.ID)
	}
	if attr.Name != "Mithril Shirt" {
		t.Errorf("Name = %v, want Mithril Shirt", attr.Name)
	}
	if attr.Custom != true {
		t.Errorf("Custom = %v, want true", attr.Custom)
	}
	if attr.Model != "company" {
		t.Errorf("Model = %v, want company", attr.Model)
	}
}

func TestAttributesService_Create_WithOptions(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["data_type"] != "options" {
			t.Errorf("body.data_type = %v, want options", body["data_type"])
		}
		opts, ok := body["options"].([]any)
		if !ok {
			t.Fatal("body.options is not an array")
		}
		if len(opts) != 2 {
			t.Fatalf("options length = %d, want 2", len(opts))
		}
		fmt.Fprint(w, `{
			"id":38,
			"type":"data_attribute",
			"name":"Size Range",
			"data_type":"string",
			"options":["1-10","11-50"],
			"custom":true,
			"model":"company"
		}`)
	})

	ctx := context.Background()
	attr, err := svc.Create(ctx, &data.CreateAttributeRequest{
		Name:     "Size Range",
		Model:    "company",
		DataType: "options",
		Options:  []data.AttributeOption{{Value: "1-10"}, {Value: "11-50"}},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if attr.ID != 38 {
		t.Errorf("ID = %v, want 38", attr.ID)
	}
	if len(attr.Options) != 2 {
		t.Fatalf("Options length = %d, want 2", len(attr.Options))
	}
}

func TestAttributesService_Update(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes/44", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body data.UpdateAttributeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Description != "Just a plain old ring" {
			t.Errorf("body.Description = %v, want Just a plain old ring", body.Description)
		}
		if *body.MessengerWritable != true {
			t.Errorf("body.MessengerWritable = %v, want true", *body.MessengerWritable)
		}
		fmt.Fprint(w, `{
			"id":44,
			"type":"data_attribute",
			"name":"The One Ring",
			"full_name":"custom_attributes.The One Ring",
			"label":"The One Ring",
			"description":"Just a plain old ring",
			"data_type":"string",
			"options":["1-10","11-20"],
			"api_writable":true,
			"ui_writable":false,
			"messenger_writable":true,
			"custom":true,
			"archived":false,
			"admin_id":"12345",
			"created_at":1734537762,
			"updated_at":1734537763,
			"model":"company"
		}`)
	})

	ctx := context.Background()
	messengerWritable := true
	attr, err := svc.Update(ctx, 44, &data.UpdateAttributeRequest{
		Description:       "Just a plain old ring",
		MessengerWritable: &messengerWritable,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if attr.ID != 44 {
		t.Errorf("ID = %v, want 44", attr.ID)
	}
	if attr.Description != "Just a plain old ring" {
		t.Errorf("Description = %v, want Just a plain old ring", attr.Description)
	}
	if attr.MessengerWritable != true {
		t.Errorf("MessengerWritable = %v, want true", attr.MessengerWritable)
	}
}

func TestAttributesService_Update_Archive(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes/44", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["archived"] != true {
			t.Errorf("body.archived = %v, want true", body["archived"])
		}
		fmt.Fprint(w, `{
			"id":44,
			"type":"data_attribute",
			"name":"The One Ring",
			"archived":true,
			"model":"company"
		}`)
	})

	ctx := context.Background()
	archived := true
	attr, err := svc.Update(ctx, 44, &data.UpdateAttributeRequest{
		Archived: &archived,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if attr.Archived != true {
		t.Errorf("Archived = %v, want true", attr.Archived)
	}
}

func TestAttributesService_Update_NotFound(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"field_not_found","message":"We couldn't find that data attribute to update"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.Update(ctx, 999, &data.UpdateAttributeRequest{
		Description: "test",
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestAttributesService_ListRaw_Success(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-da-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"data_attribute","id":1,"name":"name","model":"company"}]}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	d, err := data.ParseAttributeListResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(d.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(d.Data))
	}
	if d.Data[0].Name != "name" {
		t.Errorf("Data[0].Name = %v, want name", d.Data[0].Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-da-list" {
		t.Errorf("Header X-Request-Id = %q, want req-da-list", result.Header.Get("X-Request-Id"))
	}
}

func TestAttributesService_CreateRaw_Success(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-da-create")
		fmt.Fprint(w, `{"id":37,"type":"data_attribute","name":"Mithril Shirt","data_type":"string","custom":true,"model":"company"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &data.CreateAttributeRequest{
		Name:     "Mithril Shirt",
		Model:    "company",
		DataType: "string",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	d, err := data.ParseAttributeCreateResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if d.ID != 37 {
		t.Errorf("ID = %v, want 37", d.ID)
	}
	if d.Name != "Mithril Shirt" {
		t.Errorf("Name = %v, want Mithril Shirt", d.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestAttributesService_UpdateRaw_Success(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes/44", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-da-update")
		fmt.Fprint(w, `{"id":44,"type":"data_attribute","name":"The One Ring","description":"Just a plain old ring","model":"company"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateRaw(ctx, 44, &data.UpdateAttributeRequest{
		Description: "Just a plain old ring",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	d, err := data.ParseAttributeUpdateResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if d.ID != 44 {
		t.Errorf("ID = %v, want 44", d.ID)
	}
	if d.Description != "Just a plain old ring" {
		t.Errorf("Description = %v, want Just a plain old ring", d.Description)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-da-update" {
		t.Errorf("Header X-Request-Id = %q, want req-da-update", result.Header.Get("X-Request-Id"))
	}
}

func TestAttributesService_UpdateRaw_NotFound(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-123","errors":[{"code":"field_not_found","message":"We couldn't find that data attribute to update"}]}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateRaw(ctx, 999, &data.UpdateAttributeRequest{
		Description: "test",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned Go error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Result.Error is nil, want non-nil")
	}
	if result.Error.Code != "field_not_found" {
		t.Errorf("Error.Code = %q, want field_not_found", result.Error.Code)
	}
	if result.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", result.StatusCode)
	}
}

func TestAttributesService_List_Unauthorized(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewAttributesService(caller)

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-456",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.List(ctx, nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}

// --- CustomObjects tests ---

func TestObjectsService_Get(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order/obj_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"id":"obj_123",
			"external_id":"ext_123",
			"type":"Order",
			"created_at":1700000000,
			"updated_at":1700000100,
			"custom_attributes":{"status":"shipped"}
		}`)
	})

	ctx := context.Background()
	obj, err := svc.Get(ctx, "Order", "obj_123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if obj.ID != "obj_123" {
		t.Errorf("ID = %v, want obj_123", obj.ID)
	}
	if obj.ExternalID != "ext_123" {
		t.Errorf("ExternalID = %v, want ext_123", obj.ExternalID)
	}
	if obj.Type != "Order" {
		t.Errorf("Type = %v, want Order", obj.Type)
	}
	if obj.CustomAttributes["status"] != "shipped" {
		t.Errorf("CustomAttributes[status] = %v, want shipped", obj.CustomAttributes["status"])
	}
}

func TestObjectsService_Get_NotFound(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Object not found"}]}`)
	})

	ctx := context.Background()
	_, err := svc.Get(ctx, "Order", "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestObjectsService_GetByExternalID(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if r.URL.Query().Get("external_id") != "ext_123" {
			t.Errorf("external_id = %v, want ext_123", r.URL.Query().Get("external_id"))
		}
		fmt.Fprint(w, `{
			"id":"obj_123",
			"external_id":"ext_123",
			"type":"Order",
			"custom_attributes":{"status":"pending"}
		}`)
	})

	ctx := context.Background()
	obj, err := svc.GetByExternalID(ctx, "Order", "ext_123")
	if err != nil {
		t.Fatalf("GetByExternalID returned error: %v", err)
	}
	if obj.ExternalID != "ext_123" {
		t.Errorf("ExternalID = %v, want ext_123", obj.ExternalID)
	}
}

func TestObjectsService_CreateOrUpdate(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["external_id"] != "ext_456" {
			t.Errorf("external_id = %v, want ext_456", body["external_id"])
		}
		fmt.Fprint(w, `{
			"id":"obj_456",
			"external_id":"ext_456",
			"type":"Order",
			"created_at":1700000000,
			"updated_at":1700000000,
			"custom_attributes":{"status":"new"}
		}`)
	})

	ctx := context.Background()
	obj, err := svc.CreateOrUpdate(ctx, "Order", &data.CreateOrUpdateObjectRequest{
		ExternalID:       "ext_456",
		CustomAttributes: map[string]any{"status": "new"},
	})
	if err != nil {
		t.Fatalf("CreateOrUpdate returned error: %v", err)
	}
	if obj.ID != "obj_456" {
		t.Errorf("ID = %v, want obj_456", obj.ID)
	}
}

func TestObjectsService_Delete(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order/obj_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"obj_123","object":"Order","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.Delete(ctx, "Order", "obj_123")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if !result.Deleted {
		t.Error("Deleted = false, want true")
	}
	if result.ID != "obj_123" {
		t.Errorf("ID = %v, want obj_123", result.ID)
	}
}

func TestObjectsService_DeleteByExternalID(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		if r.URL.Query().Get("external_id") != "ext_123" {
			t.Errorf("external_id = %v, want ext_123", r.URL.Query().Get("external_id"))
		}
		fmt.Fprint(w, `{"id":"obj_123","object":"Order","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteByExternalID(ctx, "Order", "ext_123")
	if err != nil {
		t.Fatalf("DeleteByExternalID returned error: %v", err)
	}
	if !result.Deleted {
		t.Error("Deleted = false, want true")
	}
}

func TestObjectsService_GetRaw_Success(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order/obj_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-co-get")
		fmt.Fprint(w, `{"id":"obj_123","external_id":"ext_123","type":"Order","custom_attributes":{"status":"shipped"}}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "Order", "obj_123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	d, err := data.ParseObjectGetResult(result)
	if err != nil {
		t.Fatalf("ParseObjectGetResult returned error: %v", err)
	}
	if d.ID != "obj_123" {
		t.Errorf("Data.ID = %v, want obj_123", d.ID)
	}
	if d.CustomAttributes["status"] != "shipped" {
		t.Errorf("Data.CustomAttributes[status] = %v, want shipped", d.CustomAttributes["status"])
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-co-get" {
		t.Errorf("Header X-Request-Id = %q, want req-co-get", result.Header.Get("X-Request-Id"))
	}
}

func TestObjectsService_GetByExternalIDRaw_Success(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-co-extid")
		fmt.Fprint(w, `{"id":"obj_123","external_id":"ext_123","type":"Order"}`)
	})

	ctx := context.Background()
	result, err := svc.GetByExternalIDRaw(ctx, "Order", "ext_123")
	if err != nil {
		t.Fatalf("GetByExternalIDRaw returned error: %v", err)
	}
	d, err := data.ParseObjectGetByExternalIDResult(result)
	if err != nil {
		t.Fatalf("ParseObjectGetByExternalIDResult returned error: %v", err)
	}
	if d.ExternalID != "ext_123" {
		t.Errorf("Data.ExternalID = %v, want ext_123", d.ExternalID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestObjectsService_CreateOrUpdateRaw_Success(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-co-create")
		fmt.Fprint(w, `{"id":"obj_456","external_id":"ext_456","type":"Order","custom_attributes":{"status":"new"}}`)
	})

	ctx := context.Background()
	result, err := svc.CreateOrUpdateRaw(ctx, "Order", &data.CreateOrUpdateObjectRequest{
		ExternalID:       "ext_456",
		CustomAttributes: map[string]any{"status": "new"},
	})
	if err != nil {
		t.Fatalf("CreateOrUpdateRaw returned error: %v", err)
	}
	d, err := data.ParseObjectCreateOrUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseObjectCreateOrUpdateResult returned error: %v", err)
	}
	if d.ID != "obj_456" {
		t.Errorf("Data.ID = %v, want obj_456", d.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestObjectsService_DeleteRaw_Success(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order/obj_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-co-del")
		fmt.Fprint(w, `{"id":"obj_123","object":"Order","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteRaw(ctx, "Order", "obj_123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	d, err := data.ParseObjectDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseObjectDeleteResult returned error: %v", err)
	}
	if !d.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestObjectsService_DeleteByExternalIDRaw_Success(t *testing.T) {
	caller, mux, teardown := setupCaller()
	defer teardown()
	svc := data.NewObjectsService(caller)

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-co-del-ext")
		fmt.Fprint(w, `{"id":"obj_123","object":"Order","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteByExternalIDRaw(ctx, "Order", "ext_123")
	if err != nil {
		t.Fatalf("DeleteByExternalIDRaw returned error: %v", err)
	}
	d, err := data.ParseObjectDeleteByExternalIDResult(result)
	if err != nil {
		t.Fatalf("ParseObjectDeleteByExternalIDResult returned error: %v", err)
	}
	if !d.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-co-del-ext" {
		t.Errorf("Header X-Request-Id = %q, want req-co-del-ext", result.Header.Get("X-Request-Id"))
	}
}
