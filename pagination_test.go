package intercom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

func TestIter_MultiplePages(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	// Page 1: items 1,2 with cursor to page 2
	// Page 2: items 3,4 with no next cursor
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")

		after := r.URL.Query().Get("starting_after")
		switch after {
		case "":
			fmt.Fprint(w, `{
				"type": "list",
				"data": [{"id":"1","name":"a"},{"id":"2","name":"b"}],
				"total_count": 4,
				"pages": {
					"type": "pages",
					"page": 1,
					"per_page": 2,
					"total_pages": 2,
					"next": {"per_page": 2, "starting_after": "cursor-page2"}
				}
			}`)
		case "cursor-page2":
			fmt.Fprint(w, `{
				"type": "list",
				"data": [{"id":"3","name":"c"},{"id":"4","name":"d"}],
				"total_count": 4,
				"pages": {
					"type": "pages",
					"page": 2,
					"per_page": 2,
					"total_pages": 2,
					"next": null
				}
			}`)
		default:
			t.Errorf("unexpected starting_after = %q", after)
		}
	})

	type item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	iter := NewIter[item](context.Background(), nil, fetcher)

	var got []item
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
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"type": "list",
			"data": [],
			"total_count": 0,
			"pages": {
				"type": "pages",
				"page": 1,
				"per_page": 20,
				"total_pages": 0,
				"next": null
			}
		}`)
	})

	type item struct {
		ID string `json:"id"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	iter := NewIter[item](context.Background(), nil, fetcher)

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
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		after := r.URL.Query().Get("starting_after")
		if after == "" {
			fmt.Fprint(w, `{
				"type": "list",
				"data": [{"id":"1"}],
				"total_count": 2,
				"pages": {
					"type": "pages",
					"page": 1,
					"per_page": 1,
					"total_pages": 2,
					"next": {"per_page": 1, "starting_after": "cursor-page2"}
				}
			}`)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"type":"error.list","errors":[{"code":"server_error","message":"boom"}]}`)
		}
	})

	type item struct {
		ID string `json:"id"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	iter := NewIter[item](context.Background(), nil, fetcher)

	// First item should succeed
	if !iter.Next() {
		t.Fatal("expected first Next() to return true")
	}
	if iter.Current().ID != "1" {
		t.Errorf("first item ID = %q, want %q", iter.Current().ID, "1")
	}

	// Second call should fail (page 2 returns 500)
	if iter.Next() {
		t.Fatal("expected second Next() to return false after error")
	}
	if iter.Err() == nil {
		t.Fatal("expected non-nil error after failed page fetch")
	}
}

func TestIter_Current(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"type": "list",
			"data": [{"id":"1","val":"x"},{"id":"2","val":"y"}],
			"total_count": 2,
			"pages": {"type":"pages","page":1,"per_page":2,"total_pages":1}
		}`)
	})

	type item struct {
		ID  string `json:"id"`
		Val string `json:"val"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	iter := NewIter[item](context.Background(), nil, fetcher)

	iter.Next()
	if c := iter.Current(); c.ID != "1" || c.Val != "x" {
		t.Errorf("Current() = %+v, want {ID:1 Val:x}", c)
	}
	iter.Next()
	if c := iter.Current(); c.ID != "2" || c.Val != "y" {
		t.Errorf("Current() = %+v, want {ID:2 Val:y}", c)
	}
}

func TestIter_PageResponse(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"type": "list",
			"data": [{"id":"1"}],
			"total_count": 42,
			"pages": {"type":"pages","page":1,"per_page":10,"total_pages":5}
		}`)
	})

	type item struct {
		ID string `json:"id"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	iter := NewIter[item](context.Background(), nil, fetcher)

	// Must call Next at least once to have a page loaded
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
	type item struct {
		ID string `json:"id"`
	}

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

	var result PagedResult[item]
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
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")

		after := r.URL.Query().Get("starting_after")
		switch after {
		case "":
			fmt.Fprint(w, `{
				"type": "list",
				"data": [{"id":"1","name":"a"},{"id":"2","name":"b"}],
				"total_count": 4,
				"pages": {
					"type": "pages",
					"page": 1,
					"per_page": 2,
					"total_pages": 2,
					"next": {"per_page": 2, "starting_after": "cursor-page2"}
				}
			}`)
		case "cursor-page2":
			fmt.Fprint(w, `{
				"type": "list",
				"data": [{"id":"3","name":"c"},{"id":"4","name":"d"}],
				"total_count": 4,
				"pages": {
					"type": "pages",
					"page": 2,
					"per_page": 2,
					"total_pages": 2,
					"next": null
				}
			}`)
		default:
			t.Errorf("unexpected starting_after = %q", after)
		}
	})

	type item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	iter := NewIter[item](context.Background(), nil, fetcher)
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
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"type": "list",
			"data": [],
			"total_count": 0,
			"pages": {
				"type": "pages",
				"page": 1,
				"per_page": 20,
				"total_pages": 0,
				"next": null
			}
		}`)
	})

	type item struct {
		ID string `json:"id"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	iter := NewIter[item](context.Background(), nil, fetcher)
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
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		after := r.URL.Query().Get("starting_after")
		if after == "" {
			fmt.Fprint(w, `{
				"type": "list",
				"data": [{"id":"1"},{"id":"2"}],
				"total_count": 4,
				"pages": {
					"type": "pages",
					"page": 1,
					"per_page": 2,
					"total_pages": 2,
					"next": {"per_page": 2, "starting_after": "cursor-page2"}
				}
			}`)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"type":"error.list","errors":[{"code":"server_error","message":"boom"}]}`)
		}
	})

	type item struct {
		ID string `json:"id"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	iter := NewIter[item](context.Background(), nil, fetcher)
	got, err := iter.Collect()
	if err == nil {
		t.Fatal("Collect() expected error, got nil")
	}
	// Verify the error is an ErrorResponse with the expected code
	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("Collect() error type = %T, want *ErrorResponse", err)
	}
	if len(errResp.Errors) == 0 || errResp.Errors[0].Code != "server_error" {
		t.Errorf("error code = %v, want server_error", errResp.Errors)
	}
	// Should return the items collected before the error
	if len(got) != 2 {
		t.Errorf("Collect() returned %d partial items, want 2", len(got))
	}
	if got[0].ID != "1" || got[1].ID != "2" {
		t.Errorf("partial items = %+v, want [{ID:1} {ID:2}]", got)
	}
}

func TestIter_ForEach_AllItems(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")

		after := r.URL.Query().Get("starting_after")
		switch after {
		case "":
			fmt.Fprint(w, `{
				"type": "list",
				"data": [{"id":"1","name":"a"},{"id":"2","name":"b"}],
				"total_count": 4,
				"pages": {
					"type": "pages",
					"page": 1,
					"per_page": 2,
					"total_pages": 2,
					"next": {"per_page": 2, "starting_after": "cursor-page2"}
				}
			}`)
		case "cursor-page2":
			fmt.Fprint(w, `{
				"type": "list",
				"data": [{"id":"3","name":"c"},{"id":"4","name":"d"}],
				"total_count": 4,
				"pages": {
					"type": "pages",
					"page": 2,
					"per_page": 2,
					"total_pages": 2,
					"next": null
				}
			}`)
		default:
			t.Errorf("unexpected starting_after = %q", after)
		}
	})

	type item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	iter := NewIter[item](context.Background(), nil, fetcher)
	var got []item
	err := iter.ForEach(func(it item) error {
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
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"type": "list",
			"data": [{"id":"1"},{"id":"2"},{"id":"3"}],
			"total_count": 3,
			"pages": {"type":"pages","page":1,"per_page":3,"total_pages":1}
		}`)
	})

	type item struct {
		ID string `json:"id"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	stopErr := fmt.Errorf("stop after 2")
	iter := NewIter[item](context.Background(), nil, fetcher)
	var count int
	err := iter.ForEach(func(it item) error {
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
	// Fetch error should be nil — iteration stopped by callback, not by a fetch failure
	if iter.Err() != nil {
		t.Errorf("iter.Err() = %v, want nil (callback error, not fetch error)", iter.Err())
	}
}

func TestIter_ForEach_FetchError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		after := r.URL.Query().Get("starting_after")
		if after == "" {
			fmt.Fprint(w, `{
				"type": "list",
				"data": [{"id":"1"}],
				"total_count": 2,
				"pages": {
					"type": "pages",
					"page": 1,
					"per_page": 1,
					"total_pages": 2,
					"next": {"per_page": 1, "starting_after": "cursor-page2"}
				}
			}`)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"type":"error.list","errors":[{"code":"server_error","message":"boom"}]}`)
		}
	})

	type item struct {
		ID string `json:"id"`
	}

	fetcher := func(ctx context.Context, opts *ListOptions) (*PagedResult[item], error) {
		path, _ := api.AddQueryOptions("items", opts)
		req, _ := client.NewRequest(http.MethodGet, path, nil)
		result := new(PagedResult[item])
		_, err := client.Do(ctx, req, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	iter := NewIter[item](context.Background(), nil, fetcher)
	var visited int
	err := iter.ForEach(func(it item) error {
		visited++
		return nil
	})
	if err == nil {
		t.Fatal("ForEach() expected error, got nil")
	}
	// Verify the error is an ErrorResponse with the expected code
	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("ForEach() error type = %T, want *ErrorResponse", err)
	}
	if len(errResp.Errors) == 0 || errResp.Errors[0].Code != "server_error" {
		t.Errorf("error code = %v, want server_error", errResp.Errors)
	}
	if visited != 1 {
		t.Errorf("callback called %d times, want 1 (before fetch error)", visited)
	}
	// iter.Err() should be set for fetch errors
	if iter.Err() == nil {
		t.Error("iter.Err() should be non-nil for fetch errors")
	}
}

func TestListOptions_QueryParams(t *testing.T) {
	opts := &ListOptions{PerPage: 25, StartingAfter: "cursor-abc"}
	path, err := api.AddQueryOptions("contacts", opts)
	if err != nil {
		t.Fatalf("addQueryOptions error: %v", err)
	}
	if path != "contacts?per_page=25&starting_after=cursor-abc" {
		t.Errorf("path = %q, want query params for per_page and starting_after", path)
	}
}
