package intercom

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestSegmentsService_Get(t *testing.T) {
	client, mux, teardown := setup()
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
	seg, err := client.Segments.Get(ctx, "seg_123")
	if err != nil {
		t.Fatalf("Segments.Get returned error: %v", err)
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

func TestSegmentsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
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
	_, err := client.Segments.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestSegmentsService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		// Should not have include_count when not requested
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
	result, err := client.Segments.List(ctx, nil)
	if err != nil {
		t.Fatalf("Segments.List returned error: %v", err)
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

func TestSegmentsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments/seg_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-seg-get")
		fmt.Fprint(w, `{"type":"segment","id":"seg_123","name":"Active Users","person_type":"user","count":54}`)
	})

	ctx := context.Background()
	result, err := client.Segments.GetRaw(ctx, "seg_123")
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
	seg, err := ParseSegmentGetResult(result)
	if err != nil {
		t.Fatalf("ParseSegmentGetResult returned error: %v", err)
	}
	if seg.ID != "seg_123" {
		t.Errorf("ID = %v, want seg_123", seg.ID)
	}
	if seg.Name != "Active Users" {
		t.Errorf("Name = %v, want Active Users", seg.Name)
	}
}

func TestSegmentsService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Segment not found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Segments.GetRaw(ctx, "nonexistent")
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

func TestSegmentsService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-seg-list")
		fmt.Fprint(w, `{"type":"segment.list","segments":[{"type":"segment","id":"seg_1","name":"Active"},{"type":"segment","id":"seg_2","name":"New"}]}`)
	})

	ctx := context.Background()
	result, err := client.Segments.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-seg-list" {
		t.Errorf("Header X-Request-Id = %q, want req-seg-list", result.Header.Get("X-Request-Id"))
	}
	segments, err := ParseSegmentListResult(result)
	if err != nil {
		t.Fatalf("ParseSegmentListResult returned error: %v", err)
	}
	if len(segments.Segments) != 2 {
		t.Fatalf("Segments length = %d, want 2", len(segments.Segments))
	}
	if segments.Segments[0].Name != "Active" {
		t.Errorf("Segments[0].Name = %v, want Active", segments.Segments[0].Name)
	}
}

func TestSegmentsService_List_WithIncludeCount(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.Segments.List(ctx, &SegmentListOptions{IncludeCount: &includeCount})
	if err != nil {
		t.Fatalf("Segments.List returned error: %v", err)
	}
	if len(result.Segments) != 1 {
		t.Fatalf("Segments length = %d, want 1", len(result.Segments))
	}
	if result.Segments[0].Count != 120 {
		t.Errorf("Segments[0].Count = %v, want 120", result.Segments[0].Count)
	}
}
