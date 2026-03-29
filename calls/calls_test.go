package calls_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/calls"
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

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method = %v, want %v", got, want)
	}
}

func testHeader(t *testing.T, r *http.Request, header, want string) {
	t.Helper()
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header %v = %v, want %v", header, got, want)
	}
}

func setupCalls() (svc *calls.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = calls.NewService(caller)
	return svc, mux, server.Close
}

func setupRedirects() (svc *calls.RedirectsService, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = calls.NewRedirectsService(caller)
	return svc, mux, server.Close
}

// --- Calls Service Tests ---

func TestService_Get(t *testing.T) {
	svc, mux, teardown := setupCalls()
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
	call, err := svc.Get(ctx, "call_123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
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

func TestService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setupCalls()
	defer teardown()

	mux.HandleFunc("/calls/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Call not found"}]}`)
	})

	ctx := context.Background()
	_, err := svc.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestService_List(t *testing.T) {
	svc, mux, teardown := setupCalls()
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
	result, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
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

func TestService_List_WithOptions(t *testing.T) {
	svc, mux, teardown := setupCalls()
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
	_, err := svc.List(ctx, &calls.ListOptions{Page: 2, PerPage: 10})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
}

func TestService_Search(t *testing.T) {
	svc, mux, teardown := setupCalls()
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
	result, err := svc.Search(ctx, &calls.SearchRequest{
		ConversationIDs: []string{"conv_1", "conv_2"},
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(result.Data))
	}
	if result.Data[0].TranscriptStatus != "completed" {
		t.Errorf("TranscriptStatus = %v, want completed", result.Data[0].TranscriptStatus)
	}
}

func TestService_GetRecordingURL(t *testing.T) {
	svc, mux, teardown := setupCalls()
	defer teardown()

	mux.HandleFunc("/calls/call_123/recording", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Location", "https://example.com/recording.mp3")
		w.WriteHeader(http.StatusFound)
	})

	ctx := context.Background()
	url, err := svc.GetRecordingURL(ctx, "call_123")
	if err != nil {
		t.Fatalf("GetRecordingURL returned error: %v", err)
	}
	if url != "https://example.com/recording.mp3" {
		t.Errorf("URL = %v, want https://example.com/recording.mp3", url)
	}
}

func TestService_GetTranscript(t *testing.T) {
	svc, mux, teardown := setupCalls()
	defer teardown()

	mux.HandleFunc("/calls/call_123/transcript", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "Agent: Hello\nCustomer: Hi there")
	})

	ctx := context.Background()
	transcript, err := svc.GetTranscript(ctx, "call_123")
	if err != nil {
		t.Fatalf("GetTranscript returned error: %v", err)
	}
	if transcript != "Agent: Hello\nCustomer: Hi there" {
		t.Errorf("Transcript = %v, want Agent: Hello\\nCustomer: Hi there", transcript)
	}
}

// --- Raw companion method tests ---

func TestService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setupCalls()
	defer teardown()

	mux.HandleFunc("/calls/call_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-call-get")
		fmt.Fprint(w, `{"type":"call","id":"call_123","state":"ended","phone":"+15551234567"}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "call_123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	call, err := calls.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
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

func TestService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setupCalls()
	defer teardown()

	mux.HandleFunc("/calls/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Call not found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "nonexistent")
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

func TestService_ListRaw_Success(t *testing.T) {
	svc, mux, teardown := setupCalls()
	defer teardown()

	mux.HandleFunc("/calls", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-call-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"call","id":"call_1"},{"type":"call","id":"call_2"}],"total_count":2}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	list, err := calls.ParseListResult(result)
	if err != nil {
		t.Fatalf("ParseListResult returned error: %v", err)
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

func TestService_SearchRaw_Success(t *testing.T) {
	svc, mux, teardown := setupCalls()
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
	result, err := svc.SearchRaw(ctx, &calls.SearchRequest{
		ConversationIDs: []string{"conv_1"},
	})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	list, err := calls.ParseSearchResult(result)
	if err != nil {
		t.Fatalf("ParseSearchResult returned error: %v", err)
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

func TestService_GetRecordingURLRaw_Success(t *testing.T) {
	svc, mux, teardown := setupCalls()
	defer teardown()

	mux.HandleFunc("/calls/call_123/recording", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Location", "https://example.com/recording.mp3")
		w.WriteHeader(http.StatusFound)
	})

	ctx := context.Background()
	result, err := svc.GetRecordingURLRaw(ctx, "call_123")
	if err != nil {
		t.Fatalf("GetRecordingURLRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusFound {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusFound)
	}
	loc, err := calls.ParseGetRecordingURLResult(result)
	if err != nil {
		t.Fatalf("ParseGetRecordingURLResult returned error: %v", err)
	}
	if loc != "https://example.com/recording.mp3" {
		t.Errorf("Location = %q, want https://example.com/recording.mp3", loc)
	}
}

func TestService_GetTranscriptRaw_Success(t *testing.T) {
	svc, mux, teardown := setupCalls()
	defer teardown()

	mux.HandleFunc("/calls/call_123/transcript", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Request-Id", "req-transcript")
		fmt.Fprint(w, "Agent: Hello\nCustomer: Hi there")
	})

	ctx := context.Background()
	result, err := svc.GetTranscriptRaw(ctx, "call_123")
	if err != nil {
		t.Fatalf("GetTranscriptRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	transcript, err := calls.ParseGetTranscriptResult(result)
	if err != nil {
		t.Fatalf("ParseGetTranscriptResult returned error: %v", err)
	}
	if transcript != "Agent: Hello\nCustomer: Hi there" {
		t.Errorf("Transcript = %q, want Agent: Hello\\nCustomer: Hi there", transcript)
	}
	if result.Header.Get("X-Request-Id") != "req-transcript" {
		t.Errorf("Header X-Request-Id = %q, want req-transcript", result.Header.Get("X-Request-Id"))
	}
}

// --- ParseGetTranscriptResult Tests ---

func TestParseGetTranscriptResult_Nil(t *testing.T) {
	_, err := calls.ParseGetTranscriptResult(nil)
	if err == nil {
		t.Fatal("Expected error for nil result, got nil")
	}
	if got := err.Error(); got != "intercom: nil result" {
		t.Errorf("Error = %q, want %q", got, "intercom: nil result")
	}
}

func TestParseGetTranscriptResult_EmptyBody(t *testing.T) {
	r := &api.Result{StatusCode: 200, Body: []byte{}}
	_, err := calls.ParseGetTranscriptResult(r)
	if err == nil {
		t.Fatal("Expected error for empty body, got nil")
	}
	if got := err.Error(); got != "intercom: empty transcript body" {
		t.Errorf("Error = %q, want %q", got, "intercom: empty transcript body")
	}
}

func TestParseGetTranscriptResult_ValidBody(t *testing.T) {
	r := &api.Result{StatusCode: 200, Body: []byte("Agent: Hello\nCustomer: Hi")}
	transcript, err := calls.ParseGetTranscriptResult(r)
	if err != nil {
		t.Fatalf("ParseGetTranscriptResult returned error: %v", err)
	}
	if transcript != "Agent: Hello\nCustomer: Hi" {
		t.Errorf("Transcript = %q, want %q", transcript, "Agent: Hello\nCustomer: Hi")
	}
}

// --- ParseGetRecordingURLResult Tests ---

func TestParseGetRecordingURLResult_NilResult(t *testing.T) {
	_, err := calls.ParseGetRecordingURLResult(nil)
	if err == nil {
		t.Fatal("Expected error for nil result, got nil")
	}
	if !strings.Contains(err.Error(), "nil result") {
		t.Errorf("Error should mention nil result, got: %s", err.Error())
	}
}

func TestParseGetRecordingURLResult_RedirectMissingLocation(t *testing.T) {
	r := &api.Result{
		StatusCode: http.StatusFound,
		Header:     http.Header{},
	}
	_, err := calls.ParseGetRecordingURLResult(r)
	if err == nil {
		t.Fatal("Expected error for redirect without Location header, got nil")
	}
	if !strings.Contains(err.Error(), "missing Location header") {
		t.Errorf("Error should mention missing Location header, got: %s", err.Error())
	}
}

func TestParseGetRecordingURLResult_UnexpectedStatus(t *testing.T) {
	r := &api.Result{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       []byte(`{"type":"error","message":"recording not available"}`),
	}
	_, err := calls.ParseGetRecordingURLResult(r)
	if err == nil {
		t.Fatal("Expected error for 200 status, got nil")
	}
	errMsg := err.Error()
	if got := errMsg; got == "" {
		t.Fatal("Error message is empty")
	}
	// Verify error includes status code
	if !strings.Contains(errMsg, "200") {
		t.Errorf("Error should contain status code 200, got: %s", errMsg)
	}
	// Verify error includes body content
	if !strings.Contains(errMsg, "recording not available") {
		t.Errorf("Error should contain body content, got: %s", errMsg)
	}
}

// --- Phone Call Redirects Tests ---

func TestRedirectsService_Create(t *testing.T) {
	svc, mux, teardown := setupRedirects()
	defer teardown()

	mux.HandleFunc("/phone_call_redirects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["phone"] != "+15551234567" {
			t.Errorf("phone = %v, want +15551234567", body["phone"])
		}
		fmt.Fprint(w, `{
			"type":"phone_call_redirect",
			"phone":"+15551234567"
		}`)
	})

	ctx := context.Background()
	result, err := svc.Create(ctx, &calls.CreateRedirectRequest{
		Phone: "+15551234567",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if result.Type != "phone_call_redirect" {
		t.Errorf("Type = %v, want phone_call_redirect", result.Type)
	}
	if result.Phone != "+15551234567" {
		t.Errorf("Phone = %v, want +15551234567", result.Phone)
	}
}

func TestRedirectsService_Create_WithCustomAttributes(t *testing.T) {
	svc, mux, teardown := setupRedirects()
	defer teardown()

	mux.HandleFunc("/phone_call_redirects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		attrs := body["custom_attributes"].(map[string]any)
		if attrs["plan"] != "premium" {
			t.Errorf("custom_attributes.plan = %v, want premium", attrs["plan"])
		}
		fmt.Fprint(w, `{
			"type":"phone_call_redirect",
			"phone":"+15551234567"
		}`)
	})

	ctx := context.Background()
	result, err := svc.Create(ctx, &calls.CreateRedirectRequest{
		Phone:            "+15551234567",
		CustomAttributes: map[string]any{"plan": "premium"},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if result.Phone != "+15551234567" {
		t.Errorf("Phone = %v, want +15551234567", result.Phone)
	}
}

func TestRedirectsService_CreateRaw_Success(t *testing.T) {
	svc, mux, teardown := setupRedirects()
	defer teardown()

	mux.HandleFunc("/phone_call_redirects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-pcr")
		fmt.Fprint(w, `{"type":"phone_call_redirect","phone":"+15551234567"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &calls.CreateRedirectRequest{
		Phone: "+15551234567",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-pcr" {
		t.Errorf("Header X-Request-Id = %q, want req-pcr", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	pcr, err := calls.ParseRedirectCreateResult(result)
	if err != nil {
		t.Fatalf("ParseRedirectCreateResult returned error: %v", err)
	}
	if pcr.Type != "phone_call_redirect" {
		t.Errorf("Type = %v, want phone_call_redirect", pcr.Type)
	}
	if pcr.Phone != "+15551234567" {
		t.Errorf("Phone = %v, want +15551234567", pcr.Phone)
	}
}
