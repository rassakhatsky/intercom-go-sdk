package api

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

type testItem struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// mockFetcher returns a PageFetcher that serves pages from a predefined map.
// Pages are keyed by the StartingAfter cursor ("" for the first page).
func mockFetcher(pages map[string]*PagedResult[testItem]) PageFetcher[testItem] {
	return func(_ context.Context, opts *ListOptions) (*PagedResult[testItem], error) {
		page, ok := pages[opts.StartingAfter]
		if !ok {
			return nil, fmt.Errorf("unexpected cursor: %q", opts.StartingAfter)
		}
		return page, nil
	}
}

func TestIter_MultiplePages(t *testing.T) {
	pages := map[string]*PagedResult[testItem]{
		"": {
			Type:       "list",
			Data:       []testItem{{ID: "1", Name: "a"}, {ID: "2", Name: "b"}},
			TotalCount: 4,
			Pages: CursorPages{
				Type: "pages", Page: 1, PerPage: 2, TotalPages: 2,
				Next: &StartingAfterPage{PerPage: 2, StartingAfter: "cursor-page2"},
			},
		},
		"cursor-page2": {
			Type:       "list",
			Data:       []testItem{{ID: "3", Name: "c"}, {ID: "4", Name: "d"}},
			TotalCount: 4,
			Pages:      CursorPages{Type: "pages", Page: 2, PerPage: 2, TotalPages: 2},
		},
	}

	iter := NewIter[testItem](context.Background(), nil, mockFetcher(pages))

	var got []testItem
	for iter.Next() {
		got = append(got, iter.Current())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("Iter error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d items, want 4", len(got))
	}
	for i, item := range got {
		wantID := fmt.Sprintf("%d", i+1)
		if item.ID != wantID {
			t.Errorf("item[%d].ID = %q, want %q", i, item.ID, wantID)
		}
	}
}

func TestIter_EmptyFirstPage(t *testing.T) {
	pages := map[string]*PagedResult[testItem]{
		"": {
			Type:       "list",
			Data:       []testItem{},
			TotalCount: 0,
			Pages:      CursorPages{Type: "pages", Page: 1, PerPage: 20, TotalPages: 0},
		},
	}

	iter := NewIter[testItem](context.Background(), nil, mockFetcher(pages))

	count := 0
	for iter.Next() {
		count++
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("Iter error: %v", err)
	}
	if count != 0 {
		t.Errorf("got %d items, want 0", count)
	}
}

func TestIter_ErrorMidPagination(t *testing.T) {
	fetchErr := fmt.Errorf("server error")
	fetcher := func(_ context.Context, opts *ListOptions) (*PagedResult[testItem], error) {
		if opts.StartingAfter == "" {
			return &PagedResult[testItem]{
				Type:       "list",
				Data:       []testItem{{ID: "1"}},
				TotalCount: 2,
				Pages: CursorPages{
					Type: "pages", Page: 1, PerPage: 1, TotalPages: 2,
					Next: &StartingAfterPage{PerPage: 1, StartingAfter: "cursor-page2"},
				},
			}, nil
		}
		return nil, fetchErr
	}

	iter := NewIter[testItem](context.Background(), nil, fetcher)

	if !iter.Next() {
		t.Fatal("expected first Next() to return true")
	}
	if iter.Current().ID != "1" {
		t.Errorf("first item ID = %q, want %q", iter.Current().ID, "1")
	}

	if iter.Next() {
		t.Fatal("expected second Next() to return false after error")
	}
	if iter.Err() == nil {
		t.Fatal("expected non-nil error after failed page fetch")
	}
}

func TestIter_Current(t *testing.T) {
	pages := map[string]*PagedResult[testItem]{
		"": {
			Type:       "list",
			Data:       []testItem{{ID: "1", Name: "x"}, {ID: "2", Name: "y"}},
			TotalCount: 2,
			Pages:      CursorPages{Type: "pages", Page: 1, PerPage: 2, TotalPages: 1},
		},
	}

	iter := NewIter[testItem](context.Background(), nil, mockFetcher(pages))

	iter.Next()
	if c := iter.Current(); c.ID != "1" || c.Name != "x" {
		t.Errorf("Current() = %+v, want {ID:1 Name:x}", c)
	}
	iter.Next()
	if c := iter.Current(); c.ID != "2" || c.Name != "y" {
		t.Errorf("Current() = %+v, want {ID:2 Name:y}", c)
	}
}

func TestIter_PageResponse(t *testing.T) {
	pages := map[string]*PagedResult[testItem]{
		"": {
			Type:       "list",
			Data:       []testItem{{ID: "1"}},
			TotalCount: 42,
			Pages:      CursorPages{Type: "pages", Page: 1, PerPage: 10, TotalPages: 5},
		},
	}

	iter := NewIter[testItem](context.Background(), nil, mockFetcher(pages))
	iter.Next()

	pr := iter.PageResponse()
	if pr == nil {
		t.Fatal("PageResponse() returned nil")
	}
	if pr.TotalCount != 42 {
		t.Errorf("TotalCount = %d, want 42", pr.TotalCount)
	}
	if pr.Pages.Page != 1 {
		t.Errorf("Pages.Page = %d, want 1", pr.Pages.Page)
	}
	if pr.Pages.TotalPages != 5 {
		t.Errorf("Pages.TotalPages = %d, want 5", pr.Pages.TotalPages)
	}
}

func TestPagedResult_JSON(t *testing.T) {
	raw := `{
		"type": "list",
		"data": [{"id":"1"},{"id":"2"}],
		"total_count": 10,
		"pages": {
			"type": "pages",
			"page": 1,
			"per_page": 2,
			"total_pages": 5,
			"next": {"per_page": 2, "starting_after": "abc123"}
		}
	}`

	var result PagedResult[testItem]
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if result.Type != "list" {
		t.Errorf("Type = %q, want %q", result.Type, "list")
	}
	if len(result.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(result.Data))
	}
	if result.Data[0].ID != "1" {
		t.Errorf("Data[0].ID = %q, want %q", result.Data[0].ID, "1")
	}
	if result.TotalCount != 10 {
		t.Errorf("TotalCount = %d, want 10", result.TotalCount)
	}
	if result.Pages.Next == nil {
		t.Fatal("Pages.Next is nil")
	}
	if result.Pages.Next.StartingAfter != "abc123" {
		t.Errorf("Pages.Next.StartingAfter = %q, want %q", result.Pages.Next.StartingAfter, "abc123")
	}
}

func TestIter_Collect_MultiplePages(t *testing.T) {
	pages := map[string]*PagedResult[testItem]{
		"": {
			Type:       "list",
			Data:       []testItem{{ID: "1", Name: "a"}, {ID: "2", Name: "b"}},
			TotalCount: 4,
			Pages: CursorPages{
				Type: "pages", Page: 1, PerPage: 2, TotalPages: 2,
				Next: &StartingAfterPage{PerPage: 2, StartingAfter: "cursor-page2"},
			},
		},
		"cursor-page2": {
			Type:       "list",
			Data:       []testItem{{ID: "3", Name: "c"}, {ID: "4", Name: "d"}},
			TotalCount: 4,
			Pages:      CursorPages{Type: "pages", Page: 2, PerPage: 2, TotalPages: 2},
		},
	}

	iter := NewIter[testItem](context.Background(), nil, mockFetcher(pages))
	got, err := iter.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("Collect() returned %d items, want 4", len(got))
	}
	for i, item := range got {
		wantID := fmt.Sprintf("%d", i+1)
		if item.ID != wantID {
			t.Errorf("item[%d].ID = %q, want %q", i, item.ID, wantID)
		}
	}
}

func TestIter_Collect_EmptyResult(t *testing.T) {
	pages := map[string]*PagedResult[testItem]{
		"": {
			Type:       "list",
			Data:       []testItem{},
			TotalCount: 0,
			Pages:      CursorPages{Type: "pages", Page: 1, PerPage: 20, TotalPages: 0},
		},
	}

	iter := NewIter[testItem](context.Background(), nil, mockFetcher(pages))
	got, err := iter.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if got == nil {
		t.Fatal("Collect() returned nil slice, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Errorf("Collect() returned %d items, want 0", len(got))
	}
}

func TestIter_Collect_Error(t *testing.T) {
	fetchErr := fmt.Errorf("server error")
	fetcher := func(_ context.Context, opts *ListOptions) (*PagedResult[testItem], error) {
		if opts.StartingAfter == "" {
			return &PagedResult[testItem]{
				Type:       "list",
				Data:       []testItem{{ID: "1"}, {ID: "2"}},
				TotalCount: 4,
				Pages: CursorPages{
					Type: "pages", Page: 1, PerPage: 2, TotalPages: 2,
					Next: &StartingAfterPage{PerPage: 2, StartingAfter: "cursor-page2"},
				},
			}, nil
		}
		return nil, fetchErr
	}

	iter := NewIter[testItem](context.Background(), nil, fetcher)
	got, err := iter.Collect()
	if err == nil {
		t.Fatal("Collect() expected error, got nil")
	}
	if len(got) != 2 {
		t.Errorf("Collect() returned %d partial items, want 2", len(got))
	}
	if got[0].ID != "1" || got[1].ID != "2" {
		t.Errorf("partial items = %+v, want [{ID:1} {ID:2}]", got)
	}
}

func TestIter_ForEach_AllItems(t *testing.T) {
	pages := map[string]*PagedResult[testItem]{
		"": {
			Type:       "list",
			Data:       []testItem{{ID: "1", Name: "a"}, {ID: "2", Name: "b"}},
			TotalCount: 4,
			Pages: CursorPages{
				Type: "pages", Page: 1, PerPage: 2, TotalPages: 2,
				Next: &StartingAfterPage{PerPage: 2, StartingAfter: "cursor-page2"},
			},
		},
		"cursor-page2": {
			Type:       "list",
			Data:       []testItem{{ID: "3", Name: "c"}, {ID: "4", Name: "d"}},
			TotalCount: 4,
			Pages:      CursorPages{Type: "pages", Page: 2, PerPage: 2, TotalPages: 2},
		},
	}

	iter := NewIter[testItem](context.Background(), nil, mockFetcher(pages))
	var got []testItem
	err := iter.ForEach(func(it testItem) error {
		got = append(got, it)
		return nil
	})
	if err != nil {
		t.Fatalf("ForEach() error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("ForEach() visited %d items, want 4", len(got))
	}
	for i, it := range got {
		wantID := fmt.Sprintf("%d", i+1)
		if it.ID != wantID {
			t.Errorf("item[%d].ID = %q, want %q", i, it.ID, wantID)
		}
	}
}

func TestIter_ForEach_EarlyExit(t *testing.T) {
	pages := map[string]*PagedResult[testItem]{
		"": {
			Type:       "list",
			Data:       []testItem{{ID: "1"}, {ID: "2"}, {ID: "3"}},
			TotalCount: 3,
			Pages:      CursorPages{Type: "pages", Page: 1, PerPage: 3, TotalPages: 1},
		},
	}

	stopErr := fmt.Errorf("stop after 2")
	iter := NewIter[testItem](context.Background(), nil, mockFetcher(pages))
	var count int
	err := iter.ForEach(func(it testItem) error {
		count++
		if count == 2 {
			return stopErr
		}
		return nil
	})
	if err != stopErr {
		t.Fatalf("ForEach() error = %v, want %v", err, stopErr)
	}
	if count != 2 {
		t.Errorf("callback called %d times, want 2", count)
	}
	if iter.Err() != nil {
		t.Errorf("iter.Err() = %v, want nil (callback error, not fetch error)", iter.Err())
	}
}

func TestIter_ForEach_FetchError(t *testing.T) {
	fetchErr := fmt.Errorf("server error")
	fetcher := func(_ context.Context, opts *ListOptions) (*PagedResult[testItem], error) {
		if opts.StartingAfter == "" {
			return &PagedResult[testItem]{
				Type:       "list",
				Data:       []testItem{{ID: "1"}},
				TotalCount: 2,
				Pages: CursorPages{
					Type: "pages", Page: 1, PerPage: 1, TotalPages: 2,
					Next: &StartingAfterPage{PerPage: 1, StartingAfter: "cursor-page2"},
				},
			}, nil
		}
		return nil, fetchErr
	}

	iter := NewIter[testItem](context.Background(), nil, fetcher)
	var visited int
	err := iter.ForEach(func(it testItem) error {
		visited++
		return nil
	})
	if err == nil {
		t.Fatal("ForEach() expected error, got nil")
	}
	if visited != 1 {
		t.Errorf("callback called %d times, want 1 (before fetch error)", visited)
	}
	if iter.Err() == nil {
		t.Error("iter.Err() should be non-nil for fetch errors")
	}
}

func TestNewIter_NilOpts(t *testing.T) {
	pages := map[string]*PagedResult[testItem]{
		"": {
			Type:  "list",
			Data:  []testItem{{ID: "1"}},
			Pages: CursorPages{Type: "pages", Page: 1, PerPage: 1, TotalPages: 1},
		},
	}

	iter := NewIter[testItem](context.Background(), nil, mockFetcher(pages))
	if !iter.Next() {
		t.Fatal("expected Next() to return true")
	}
	if iter.Current().ID != "1" {
		t.Errorf("Current().ID = %q, want %q", iter.Current().ID, "1")
	}
}

func TestNewIter_CopiesOpts(t *testing.T) {
	opts := &ListOptions{PerPage: 10, StartingAfter: "abc"}
	iter := NewIter[testItem](context.Background(), opts, func(_ context.Context, o *ListOptions) (*PagedResult[testItem], error) {
		return &PagedResult[testItem]{Data: []testItem{}}, nil
	})
	// Mutating the original opts should not affect the iterator
	opts.StartingAfter = "changed"
	iter.Next()
	// The iterator should have used the original value
	if iter.opts.StartingAfter != "abc" {
		t.Errorf("iter.opts.StartingAfter = %q, want %q (original should be copied)", iter.opts.StartingAfter, "abc")
	}
}
