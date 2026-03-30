package segments_test

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
	"github.com/rassakhatsky/intercom-go-sdk/segments"
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

func setup() (svc *segments.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = segments.NewService(caller)
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

func TestService_Get(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments/seg_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"segment",
			"id":"seg_123",
			"name":"Active Users",
			"created_at":1394621988,
			"updated_at":1394622004,
			"person_type":"user",
			"count":54
		}`)
	})

	ctx := context.Background()
	seg, err := svc.Get(ctx, "seg_123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if seg.ID != "seg_123" {
		t.Errorf("Segment.ID = %v, want seg_123", seg.ID)
	}
	if seg.Name != "Active Users" {
		t.Errorf("Segment.Name = %v, want Active Users", seg.Name)
	}
	if seg.Type != "segment" {
		t.Errorf("Segment.Type = %v, want segment", seg.Type)
	}
	if seg.PersonType != "user" {
		t.Errorf("Segment.PersonType = %v, want user", seg.PersonType)
	}
	if seg.CreatedAt != 1394621988 {
		t.Errorf("Segment.CreatedAt = %v, want 1394621988", seg.CreatedAt)
	}
	if seg.UpdatedAt != 1394622004 {
		t.Errorf("Segment.UpdatedAt = %v, want 1394622004", seg.UpdatedAt)
	}
	if seg.Count != 54 {
		t.Errorf("Segment.Count = %v, want 54", seg.Count)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"not_found","message":"Segment not found"}]
		}`)
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
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if r.URL.Query().Get("include_count") != "" {
			t.Errorf("unexpected include_count param: %v", r.URL.Query().Get("include_count"))
		}
		fmt.Fprint(w, `{
			"type":"segment.list",
			"segments":[
				{"type":"segment","id":"seg_1","name":"Active","person_type":"user"},
				{"type":"segment","id":"seg_2","name":"Slipping Away","person_type":"user"},
				{"type":"segment","id":"seg_3","name":"New","person_type":"contact"}
			]
		}`)
	})

	ctx := context.Background()
	result, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if result.Type != "segment.list" {
		t.Errorf("Type = %v, want segment.list", result.Type)
	}
	if len(result.Segments) != 3 {
		t.Fatalf("Segments length = %d, want 3", len(result.Segments))
	}
	if result.Segments[0].Name != "Active" {
		t.Errorf("Segments[0].Name = %v, want Active", result.Segments[0].Name)
	}
	if result.Segments[2].PersonType != "contact" {
		t.Errorf("Segments[2].PersonType = %v, want contact", result.Segments[2].PersonType)
	}
}

func TestService_List_WithIncludeCount(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if r.URL.Query().Get("include_count") != "true" {
			t.Errorf("include_count = %v, want true", r.URL.Query().Get("include_count"))
		}
		fmt.Fprint(w, `{
			"type":"segment.list",
			"segments":[
				{"type":"segment","id":"seg_1","name":"Active","person_type":"user","count":120}
			]
		}`)
	})

	ctx := context.Background()
	includeCount := true
	result, err := svc.List(ctx, &segments.ListOptions{IncludeCount: &includeCount})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(result.Segments) != 1 {
		t.Fatalf("Segments length = %d, want 1", len(result.Segments))
	}
	if result.Segments[0].Count != 120 {
		t.Errorf("Segments[0].Count = %v, want 120", result.Segments[0].Count)
	}
}

func TestService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments/seg_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-seg-get")
		fmt.Fprint(w, `{"type":"segment","id":"seg_123","name":"Active Users","person_type":"user","count":54}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "seg_123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-seg-get" {
		t.Errorf("Header X-Request-Id = %q, want req-seg-get", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	seg, err := segments.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
	}
	if seg.ID != "seg_123" {
		t.Errorf("ID = %v, want seg_123", seg.ID)
	}
	if seg.Name != "Active Users" {
		t.Errorf("Name = %v, want Active Users", seg.Name)
	}
}

func TestService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Segment not found"}]}`)
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
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-seg-list")
		fmt.Fprint(w, `{"type":"segment.list","segments":[{"type":"segment","id":"seg_1","name":"Active"},{"type":"segment","id":"seg_2","name":"New"}]}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-seg-list" {
		t.Errorf("Header X-Request-Id = %q, want req-seg-list", result.Header.Get("X-Request-Id"))
	}
	list, err := segments.ParseListResult(result)
	if err != nil {
		t.Fatalf("ParseListResult returned error: %v", err)
	}
	if len(list.Segments) != 2 {
		t.Fatalf("Segments length = %d, want 2", len(list.Segments))
	}
	if list.Segments[0].Name != "Active" {
		t.Errorf("Segments[0].Name = %v, want Active", list.Segments[0].Name)
	}
}
