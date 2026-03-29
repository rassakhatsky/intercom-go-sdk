package news_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
	"github.com/rassakhatsky/intercom-go-sdk/news"
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

func setup() (svc *news.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = news.NewService(caller)
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

func TestService_ListItems(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"news-item",
					"id":"1",
					"workspace_id":"ws1",
					"title":"Product Update",
					"body":"<p>New features</p>",
					"sender_id":42,
					"state":"live",
					"labels":["Product","Update"],
					"cover_image_url":"https://example.com/cover.jpg",
					"reactions":["😆","😅"],
					"deliver_silently":false,
					"created_at":1672531200,
					"updated_at":1672617600,
					"newsfeed_assignments":[{"newsfeed_id":10,"published_at":1672531200}]
				},
				{
					"type":"news-item",
					"id":"2",
					"title":"Company News",
					"sender_id":43,
					"state":"draft"
				}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":10,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListItems(ctx, nil)
	if err != nil {
		t.Fatalf("ListItems returned error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %v, want 2", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("len(Data) = %v, want 2", len(result.Data))
	}
	item := result.Data[0]
	if item.ID != "1" {
		t.Errorf("ID = %v, want 1", item.ID)
	}
	if item.Title != "Product Update" {
		t.Errorf("Title = %v, want Product Update", item.Title)
	}
	if item.SenderID != 42 {
		t.Errorf("SenderID = %v, want 42", item.SenderID)
	}
	if item.State != "live" {
		t.Errorf("State = %v, want live", item.State)
	}
	if len(item.Labels) != 2 {
		t.Errorf("len(Labels) = %v, want 2", len(item.Labels))
	}
	if item.CoverImageURL != "https://example.com/cover.jpg" {
		t.Errorf("CoverImageURL = %v, want https://example.com/cover.jpg", item.CoverImageURL)
	}
	if len(item.Reactions) != 2 {
		t.Errorf("len(Reactions) = %v, want 2", len(item.Reactions))
	}
	if item.DeliverSilently != false {
		t.Errorf("DeliverSilently = %v, want false", item.DeliverSilently)
	}
	if len(item.NewsfeedAssignments) != 1 {
		t.Fatalf("len(NewsfeedAssignments) = %v, want 1", len(item.NewsfeedAssignments))
	}
	if item.NewsfeedAssignments[0].NewsfeedID != 10 {
		t.Errorf("NewsfeedAssignment.NewsfeedID = %v, want 10", item.NewsfeedAssignments[0].NewsfeedID)
	}
}

func TestService_GetItem(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"news-item",
			"id":"42",
			"title":"Big Announcement",
			"body":"<p>Hello world</p>",
			"sender_id":100,
			"state":"live",
			"labels":["Announcement"],
			"reactions":["👍"],
			"deliver_silently":true,
			"created_at":1672531200,
			"updated_at":1672617600,
			"newsfeed_assignments":[{"newsfeed_id":5,"published_at":1672531200}]
		}`)
	})

	ctx := context.Background()
	item, err := svc.GetItem(ctx, "42")
	if err != nil {
		t.Fatalf("GetItem returned error: %v", err)
	}
	if item.ID != "42" {
		t.Errorf("ID = %v, want 42", item.ID)
	}
	if item.Title != "Big Announcement" {
		t.Errorf("Title = %v, want Big Announcement", item.Title)
	}
	if item.DeliverSilently != true {
		t.Errorf("DeliverSilently = %v, want true", item.DeliverSilently)
	}
}

func TestService_GetItem_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-1",
			"errors":[{"code":"not_found","message":"Resource Not Found"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.GetItem(ctx, "999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestService_CreateItem(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body news.CreateItemRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Title != "Halloween is here!" {
			t.Errorf("Title = %v, want Halloween is here!", body.Title)
		}
		if body.SenderID != 123 {
			t.Errorf("SenderID = %v, want 123", body.SenderID)
		}
		if body.State != "live" {
			t.Errorf("State = %v, want live", body.State)
		}
		if body.DeliverSilently == nil || !*body.DeliverSilently {
			t.Errorf("DeliverSilently = %v, want true", body.DeliverSilently)
		}
		if len(body.Labels) != 2 {
			t.Errorf("len(Labels) = %v, want 2", len(body.Labels))
		}
		if len(body.NewsfeedAssignments) != 1 {
			t.Errorf("len(NewsfeedAssignments) = %v, want 1", len(body.NewsfeedAssignments))
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"type":"news-item",
			"id":"50",
			"title":"Halloween is here!",
			"body":"<p>New costumes</p>",
			"sender_id":123,
			"state":"live",
			"deliver_silently":true,
			"labels":["Product","Update"],
			"created_at":1672531200,
			"updated_at":1672531200
		}`)
	})

	ctx := context.Background()
	boolTrue := true
	item, err := svc.CreateItem(ctx, &news.CreateItemRequest{
		Title:           "Halloween is here!",
		Body:            "<p>New costumes</p>",
		SenderID:        123,
		State:           "live",
		DeliverSilently: &boolTrue,
		Labels:          []string{"Product", "Update"},
		Reactions:       []string{"😆", "😅"},
		NewsfeedAssignments: []news.NewsfeedAssignment{
			{NewsfeedID: 53, PublishedAt: 1664638214},
		},
	})
	if err != nil {
		t.Fatalf("CreateItem returned error: %v", err)
	}
	if item.ID != "50" {
		t.Errorf("ID = %v, want 50", item.ID)
	}
	if item.Title != "Halloween is here!" {
		t.Errorf("Title = %v, want Halloween is here!", item.Title)
	}
}

func TestService_UpdateItem(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/50", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body news.UpdateItemRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Title != "Updated Title" {
			t.Errorf("Title = %v, want Updated Title", body.Title)
		}

		fmt.Fprint(w, `{
			"type":"news-item",
			"id":"50",
			"title":"Updated Title",
			"body":"<p>Updated body</p>",
			"sender_id":123,
			"state":"live",
			"created_at":1672531200,
			"updated_at":1672617600
		}`)
	})

	ctx := context.Background()
	item, err := svc.UpdateItem(ctx, "50", &news.UpdateItemRequest{
		Title: "Updated Title",
		Body:  "<p>Updated body</p>",
	})
	if err != nil {
		t.Fatalf("UpdateItem returned error: %v", err)
	}
	if item.ID != "50" {
		t.Errorf("ID = %v, want 50", item.ID)
	}
	if item.Title != "Updated Title" {
		t.Errorf("Title = %v, want Updated Title", item.Title)
	}
}

func TestService_DeleteItem(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/50", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{
			"id":"50",
			"object":"news-item",
			"deleted":true
		}`)
	})

	ctx := context.Background()
	deleted, err := svc.DeleteItem(ctx, "50")
	if err != nil {
		t.Fatalf("DeleteItem returned error: %v", err)
	}
	if deleted.ID != "50" {
		t.Errorf("ID = %v, want 50", deleted.ID)
	}
	if !deleted.Deleted {
		t.Errorf("Deleted = %v, want true", deleted.Deleted)
	}
}

func TestService_ListNewsfeeds(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/newsfeeds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"newsfeed","id":"10","name":"Visitor Feed","created_at":1672531200,"updated_at":1672617600},
				{"type":"newsfeed","id":"11","name":"User Feed","created_at":1672531200,"updated_at":1672617600}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":10,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListNewsfeeds(ctx, nil)
	if err != nil {
		t.Fatalf("ListNewsfeeds returned error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %v, want 2", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("len(Data) = %v, want 2", len(result.Data))
	}
	if result.Data[0].Name != "Visitor Feed" {
		t.Errorf("Name = %v, want Visitor Feed", result.Data[0].Name)
	}
}

func TestService_GetNewsfeed(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/newsfeeds/10", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"newsfeed",
			"id":"10",
			"name":"Visitor Feed",
			"created_at":1672531200,
			"updated_at":1672617600
		}`)
	})

	ctx := context.Background()
	feed, err := svc.GetNewsfeed(ctx, "10")
	if err != nil {
		t.Fatalf("GetNewsfeed returned error: %v", err)
	}
	if feed.ID != "10" {
		t.Errorf("ID = %v, want 10", feed.ID)
	}
	if feed.Name != "Visitor Feed" {
		t.Errorf("Name = %v, want Visitor Feed", feed.Name)
	}
}

func TestService_ListNewsfeedItems(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/newsfeeds/10/items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"news-item",
					"id":"1",
					"title":"Feed Item",
					"sender_id":42,
					"state":"live"
				}
			],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":20,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListNewsfeedItems(ctx, "10", nil)
	if err != nil {
		t.Fatalf("ListNewsfeedItems returned error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", result.TotalCount)
	}
	if len(result.Data) != 1 {
		t.Fatalf("len(Data) = %v, want 1", len(result.Data))
	}
	if result.Data[0].Title != "Feed Item" {
		t.Errorf("Title = %v, want Feed Item", result.Data[0].Title)
	}
}

func TestService_ListItemsRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ni-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"news-item","id":"1","title":"Product Update"}],"total_count":1,"pages":{"type":"pages","page":1,"per_page":10,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := svc.ListItemsRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListItemsRaw returned error: %v", err)
	}
	data, err := news.ParseListItemsResult(result)
	if err != nil {
		t.Fatalf("ParseListItemsResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_GetItemRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ni-get")
		fmt.Fprint(w, `{"type":"news-item","id":"42","title":"Big Announcement","state":"live"}`)
	})

	ctx := context.Background()
	result, err := svc.GetItemRaw(ctx, "42")
	if err != nil {
		t.Fatalf("GetItemRaw returned error: %v", err)
	}
	data, err := news.ParseGetItemResult(result)
	if err != nil {
		t.Fatalf("ParseGetItemResult returned error: %v", err)
	}
	if data.ID != "42" {
		t.Errorf("Data.ID = %v, want 42", data.ID)
	}
	if data.Title != "Big Announcement" {
		t.Errorf("Data.Title = %v, want Big Announcement", data.Title)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestService_GetItemRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetItemRaw(ctx, "999")
	if err != nil {
		t.Fatalf("GetItemRaw returned Go error: %v", err)
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

func TestService_CreateItemRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-ni-create")
		fmt.Fprint(w, `{"type":"news-item","id":"50","title":"Halloween is here!"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateItemRaw(ctx, &news.CreateItemRequest{
		Title:    "Halloween is here!",
		SenderID: 123,
	})
	if err != nil {
		t.Fatalf("CreateItemRaw returned error: %v", err)
	}
	data, err := news.ParseCreateItemResult(result)
	if err != nil {
		t.Fatalf("ParseCreateItemResult returned error: %v", err)
	}
	if data.ID != "50" {
		t.Errorf("Data.ID = %v, want 50", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_UpdateItemRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/50", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-ni-update")
		fmt.Fprint(w, `{"type":"news-item","id":"50","title":"Updated Title"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateItemRaw(ctx, "50", &news.UpdateItemRequest{Title: "Updated Title"})
	if err != nil {
		t.Fatalf("UpdateItemRaw returned error: %v", err)
	}
	data, err := news.ParseUpdateItemResult(result)
	if err != nil {
		t.Fatalf("ParseUpdateItemResult returned error: %v", err)
	}
	if data.Title != "Updated Title" {
		t.Errorf("Data.Title = %v, want Updated Title", data.Title)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_DeleteItemRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/50", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-ni-del")
		fmt.Fprint(w, `{"id":"50","object":"news-item","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteItemRaw(ctx, "50")
	if err != nil {
		t.Fatalf("DeleteItemRaw returned error: %v", err)
	}
	data, err := news.ParseDeleteItemResult(result)
	if err != nil {
		t.Fatalf("ParseDeleteItemResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_ListNewsfeedsRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/newsfeeds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-nf-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"newsfeed","id":"10","name":"Visitor Feed"}],"total_count":1,"pages":{"type":"pages","page":1,"per_page":10,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := svc.ListNewsfeedsRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListNewsfeedsRaw returned error: %v", err)
	}
	data, err := news.ParseListNewsfeedsResult(result)
	if err != nil {
		t.Fatalf("ParseListNewsfeedsResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_GetNewsfeedRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/newsfeeds/10", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-nf-get")
		fmt.Fprint(w, `{"type":"newsfeed","id":"10","name":"Visitor Feed"}`)
	})

	ctx := context.Background()
	result, err := svc.GetNewsfeedRaw(ctx, "10")
	if err != nil {
		t.Fatalf("GetNewsfeedRaw returned error: %v", err)
	}
	data, err := news.ParseGetNewsfeedResult(result)
	if err != nil {
		t.Fatalf("ParseGetNewsfeedResult returned error: %v", err)
	}
	if data.ID != "10" {
		t.Errorf("Data.ID = %v, want 10", data.ID)
	}
	if data.Name != "Visitor Feed" {
		t.Errorf("Data.Name = %v, want Visitor Feed", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_ListNewsfeedItemsRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/newsfeeds/10/items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-nfi-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"news-item","id":"1","title":"Feed Item"}],"total_count":1,"pages":{"type":"pages","page":1,"per_page":20,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := svc.ListNewsfeedItemsRaw(ctx, "10", nil)
	if err != nil {
		t.Fatalf("ListNewsfeedItemsRaw returned error: %v", err)
	}
	data, err := news.ParseListNewsfeedItemsResult(result)
	if err != nil {
		t.Fatalf("ParseListNewsfeedItemsResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}
