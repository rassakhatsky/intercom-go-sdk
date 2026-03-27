package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCallsService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls/call_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"call",
			"id":"call_123",
			"conversation_id":"conv_1",
			"admin_id":"admin_1",
			"contact_id":"contact_1",
			"state":"ended",
			"call_type":"inbound",
			"direction":"inbound",
			"phone":"+15551234567",
			"created_at":1700000000,
			"updated_at":1700000100
		}`)
	})

	ctx := context.Background()
	call, err := client.Calls.Get(ctx, "call_123")
	if err != nil {
		t.Fatalf("Calls.Get returned error: %v", err)
	}
	if call.ID != "call_123" {
		t.Errorf("Call.ID = %v, want call_123", call.ID)
	}
	if call.ConversationID != "conv_1" {
		t.Errorf("Call.ConversationID = %v, want conv_1", call.ConversationID)
	}
	if call.State != "ended" {
		t.Errorf("Call.State = %v, want ended", call.State)
	}
	if call.Phone != "+15551234567" {
		t.Errorf("Call.Phone = %v, want +15551234567", call.Phone)
	}
}

func TestCallsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Call not found"}]}`)
	})

	ctx := context.Background()
	_, err := client.Calls.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestCallsService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"call","id":"call_1","state":"ended"},
				{"type":"call","id":"call_2","state":"in_progress"}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":25,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Calls.List(ctx, nil)
	if err != nil {
		t.Fatalf("Calls.List returned error: %v", err)
	}
	if result.Type != "list" {
		t.Errorf("Type = %v, want list", result.Type)
	}
	if len(result.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestCallsService_List_WithOptions(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("page = %v, want 2", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("per_page") != "10" {
			t.Errorf("per_page = %v, want 10", r.URL.Query().Get("per_page"))
		}
		fmt.Fprint(w, `{"type":"list","data":[],"total_count":0,"pages":{"type":"pages","page":2,"per_page":10,"total_pages":2}}`)
	})

	ctx := context.Background()
	_, err := client.Calls.List(ctx, &CallListOptions{Page: 2, PerPage: 10})
	if err != nil {
		t.Fatalf("Calls.List returned error: %v", err)
	}
}

func TestCallsService_Search(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		ids := body["conversation_ids"].([]any)
		if len(ids) != 2 {
			t.Errorf("conversation_ids length = %d, want 2", len(ids))
		}
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"call","id":"call_1","transcript_status":"completed","transcript":[{"speaker":"agent","text":"Hello"}]}
			]
		}`)
	})

	ctx := context.Background()
	result, err := client.Calls.Search(ctx, &CallSearchRequest{
		ConversationIDs: []string{"conv_1", "conv_2"},
	})
	if err != nil {
		t.Fatalf("Calls.Search returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(result.Data))
	}
	if result.Data[0].TranscriptStatus != "completed" {
		t.Errorf("TranscriptStatus = %v, want completed", result.Data[0].TranscriptStatus)
	}
}

func TestCallsService_GetRecordingURL(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls/call_123/recording", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Location", "https://example.com/recording.mp3")
		w.WriteHeader(http.StatusFound)
	})

	ctx := context.Background()
	url, err := client.Calls.GetRecordingURL(ctx, "call_123")
	if err != nil {
		t.Fatalf("Calls.GetRecordingURL returned error: %v", err)
	}
	if url != "https://example.com/recording.mp3" {
		t.Errorf("URL = %v, want https://example.com/recording.mp3", url)
	}
}

func TestCallsService_GetTranscript(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls/call_123/transcript", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "Agent: Hello\nCustomer: Hi there")
	})

	ctx := context.Background()
	transcript, err := client.Calls.GetTranscript(ctx, "call_123")
	if err != nil {
		t.Fatalf("Calls.GetTranscript returned error: %v", err)
	}
	if transcript != "Agent: Hello\nCustomer: Hi there" {
		t.Errorf("Transcript = %v, want Agent: Hello\\nCustomer: Hi there", transcript)
	}
}

// --- Raw companion method tests ---

func TestCallsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls/call_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-call-get")
		fmt.Fprint(w, `{"type":"call","id":"call_123","state":"ended","phone":"+15551234567"}`)
	})

	ctx := context.Background()
	result, err := client.Calls.GetRaw(ctx, "call_123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	call, err := ParseCallGetResult(result)
	if err != nil {
		t.Fatalf("ParseCallGetResult returned error: %v", err)
	}
	if call.ID != "call_123" {
		t.Errorf("Call.ID = %v, want call_123", call.ID)
	}
	if call.State != "ended" {
		t.Errorf("Call.State = %v, want ended", call.State)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-call-get" {
		t.Errorf("Header X-Request-Id = %q, want req-call-get", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestCallsService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Call not found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Calls.GetRaw(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetRaw returned Go error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Result.Error is nil, want non-nil")
	}
	if result.Error.Code != "not_found" {
		t.Errorf("Error.Code = %q, want not_found", result.Error.Code)
	}
	if result.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", result.StatusCode)
	}
}

func TestCallsService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-call-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"call","id":"call_1"},{"type":"call","id":"call_2"}],"total_count":2}`)
	})

	ctx := context.Background()
	result, err := client.Calls.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	list, err := ParseCallListResult(result)
	if err != nil {
		t.Fatalf("ParseCallListResult returned error: %v", err)
	}
	if len(list.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(list.Data))
	}
	if list.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", list.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-call-list" {
		t.Errorf("Header X-Request-Id = %q, want req-call-list", result.Header.Get("X-Request-Id"))
	}
}

func TestCallsService_SearchRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-call-search")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		ids := body["conversation_ids"].([]any)
		if len(ids) != 1 {
			t.Errorf("conversation_ids length = %d, want 1", len(ids))
		}
		fmt.Fprint(w, `{"type":"list","data":[{"type":"call","id":"call_1","transcript_status":"completed"}]}`)
	})

	ctx := context.Background()
	result, err := client.Calls.SearchRaw(ctx, &CallSearchRequest{
		ConversationIDs: []string{"conv_1"},
	})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	list, err := ParseCallSearchResult(result)
	if err != nil {
		t.Fatalf("ParseCallSearchResult returned error: %v", err)
	}
	if len(list.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(list.Data))
	}
	if list.Data[0].TranscriptStatus != "completed" {
		t.Errorf("TranscriptStatus = %v, want completed", list.Data[0].TranscriptStatus)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCallsService_GetRecordingURLRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls/call_123/recording", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Location", "https://example.com/recording.mp3")
		w.WriteHeader(http.StatusFound)
	})

	ctx := context.Background()
	result, err := client.Calls.GetRecordingURLRaw(ctx, "call_123")
	if err != nil {
		t.Fatalf("GetRecordingURLRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusFound {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusFound)
	}
	loc, err := ParseCallGetRecordingURLResult(result)
	if err != nil {
		t.Fatalf("ParseCallGetRecordingURLResult returned error: %v", err)
	}
	if loc != "https://example.com/recording.mp3" {
		t.Errorf("Location = %q, want https://example.com/recording.mp3", loc)
	}
}

func TestCallsService_GetTranscriptRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/calls/call_123/transcript", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Request-Id", "req-transcript")
		fmt.Fprint(w, "Agent: Hello\nCustomer: Hi there")
	})

	ctx := context.Background()
	result, err := client.Calls.GetTranscriptRaw(ctx, "call_123")
	if err != nil {
		t.Fatalf("GetTranscriptRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	transcript, err := ParseCallGetTranscriptResult(result)
	if err != nil {
		t.Fatalf("ParseCallGetTranscriptResult returned error: %v", err)
	}
	if transcript != "Agent: Hello\nCustomer: Hi there" {
		t.Errorf("Transcript = %q, want Agent: Hello\\nCustomer: Hi there", transcript)
	}
	if result.Header.Get("X-Request-Id") != "req-transcript" {
		t.Errorf("Header X-Request-Id = %q, want req-transcript", result.Header.Get("X-Request-Id"))
	}
}
