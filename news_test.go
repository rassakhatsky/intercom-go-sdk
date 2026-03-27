package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestNewsService_ListNewsItems(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.News.ListNewsItems(ctx, nil)
	if err != nil {
		t.Fatalf("News.ListNewsItems returned error: %v", err)
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

func TestNewsService_GetNewsItem(t *testing.T) {
	client, mux, teardown := setup()
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
	item, err := client.News.GetNewsItem(ctx, "42")
	if err != nil {
		t.Fatalf("News.GetNewsItem returned error: %v", err)
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

func TestNewsService_GetNewsItem_NotFound(t *testing.T) {
	client, mux, teardown := setup()
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
	_, err := client.News.GetNewsItem(ctx, "999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestNewsService_CreateNewsItem(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateNewsItemRequest
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
		if !body.DeliverSilently {
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
	item, err := client.News.CreateNewsItem(ctx, &CreateNewsItemRequest{
		Title:           "Halloween is here!",
		Body:            "<p>New costumes</p>",
		SenderID:        123,
		State:           "live",
		DeliverSilently: true,
		Labels:          []string{"Product", "Update"},
		Reactions:       []string{"😆", "😅"},
		NewsfeedAssignments: []NewsfeedAssignment{
			{NewsfeedID: 53, PublishedAt: 1664638214},
		},
	})
	if err != nil {
		t.Fatalf("News.CreateNewsItem returned error: %v", err)
	}
	if item.ID != "50" {
		t.Errorf("ID = %v, want 50", item.ID)
	}
	if item.Title != "Halloween is here!" {
		t.Errorf("Title = %v, want Halloween is here!", item.Title)
	}
}

func TestNewsService_UpdateNewsItem(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/50", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateNewsItemRequest
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
	item, err := client.News.UpdateNewsItem(ctx, "50", &UpdateNewsItemRequest{
		Title: "Updated Title",
		Body:  "<p>Updated body</p>",
	})
	if err != nil {
		t.Fatalf("News.UpdateNewsItem returned error: %v", err)
	}
	if item.ID != "50" {
		t.Errorf("ID = %v, want 50", item.ID)
	}
	if item.Title != "Updated Title" {
		t.Errorf("Title = %v, want Updated Title", item.Title)
	}
}

func TestNewsService_DeleteNewsItem(t *testing.T) {
	client, mux, teardown := setup()
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
	deleted, err := client.News.DeleteNewsItem(ctx, "50")
	if err != nil {
		t.Fatalf("News.DeleteNewsItem returned error: %v", err)
	}
	if deleted.ID != "50" {
		t.Errorf("ID = %v, want 50", deleted.ID)
	}
	if !deleted.Deleted {
		t.Errorf("Deleted = %v, want true", deleted.Deleted)
	}
}

func TestNewsService_ListNewsfeeds(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.News.ListNewsfeeds(ctx, nil)
	if err != nil {
		t.Fatalf("News.ListNewsfeeds returned error: %v", err)
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

func TestNewsService_GetNewsfeed(t *testing.T) {
	client, mux, teardown := setup()
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
	feed, err := client.News.GetNewsfeed(ctx, "10")
	if err != nil {
		t.Fatalf("News.GetNewsfeed returned error: %v", err)
	}
	if feed.ID != "10" {
		t.Errorf("ID = %v, want 10", feed.ID)
	}
	if feed.Name != "Visitor Feed" {
		t.Errorf("Name = %v, want Visitor Feed", feed.Name)
	}
}

func TestNewsService_ListNewsfeedItems(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.News.ListNewsfeedItems(ctx, "10", nil)
	if err != nil {
		t.Fatalf("News.ListNewsfeedItems returned error: %v", err)
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

func TestNewsService_ListNewsItemsRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ni-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"news-item","id":"1","title":"Product Update"}],"total_count":1,"pages":{"type":"pages","page":1,"per_page":10,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := client.News.ListNewsItemsRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListNewsItemsRaw returned error: %v", err)
	}
	data, err := ParseNewsListNewsItemsResult(result)
	if err != nil {
		t.Fatalf("ParseNewsListNewsItemsResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestNewsService_GetNewsItemRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ni-get")
		fmt.Fprint(w, `{"type":"news-item","id":"42","title":"Big Announcement","state":"live"}`)
	})

	ctx := context.Background()
	result, err := client.News.GetNewsItemRaw(ctx, "42")
	if err != nil {
		t.Fatalf("GetNewsItemRaw returned error: %v", err)
	}
	data, err := ParseNewsGetNewsItemResult(result)
	if err != nil {
		t.Fatalf("ParseNewsGetNewsItemResult returned error: %v", err)
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

func TestNewsService_GetNewsItemRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := client.News.GetNewsItemRaw(ctx, "999")
	if err != nil {
		t.Fatalf("GetNewsItemRaw returned Go error: %v", err)
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

func TestNewsService_CreateNewsItemRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-ni-create")
		fmt.Fprint(w, `{"type":"news-item","id":"50","title":"Halloween is here!"}`)
	})

	ctx := context.Background()
	result, err := client.News.CreateNewsItemRaw(ctx, &CreateNewsItemRequest{
		Title:    "Halloween is here!",
		SenderID: 123,
	})
	if err != nil {
		t.Fatalf("CreateNewsItemRaw returned error: %v", err)
	}
	data, err := ParseNewsCreateNewsItemResult(result)
	if err != nil {
		t.Fatalf("ParseNewsCreateNewsItemResult returned error: %v", err)
	}
	if data.ID != "50" {
		t.Errorf("Data.ID = %v, want 50", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestNewsService_UpdateNewsItemRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/50", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-ni-update")
		fmt.Fprint(w, `{"type":"news-item","id":"50","title":"Updated Title"}`)
	})

	ctx := context.Background()
	result, err := client.News.UpdateNewsItemRaw(ctx, "50", &UpdateNewsItemRequest{Title: "Updated Title"})
	if err != nil {
		t.Fatalf("UpdateNewsItemRaw returned error: %v", err)
	}
	data, err := ParseNewsUpdateNewsItemResult(result)
	if err != nil {
		t.Fatalf("ParseNewsUpdateNewsItemResult returned error: %v", err)
	}
	if data.Title != "Updated Title" {
		t.Errorf("Data.Title = %v, want Updated Title", data.Title)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestNewsService_DeleteNewsItemRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/news_items/50", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-ni-del")
		fmt.Fprint(w, `{"id":"50","object":"news-item","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.News.DeleteNewsItemRaw(ctx, "50")
	if err != nil {
		t.Fatalf("DeleteNewsItemRaw returned error: %v", err)
	}
	data, err := ParseNewsDeleteNewsItemResult(result)
	if err != nil {
		t.Fatalf("ParseNewsDeleteNewsItemResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestNewsService_ListNewsfeedsRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/newsfeeds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-nf-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"newsfeed","id":"10","name":"Visitor Feed"}],"total_count":1,"pages":{"type":"pages","page":1,"per_page":10,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := client.News.ListNewsfeedsRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListNewsfeedsRaw returned error: %v", err)
	}
	data, err := ParseNewsListNewsfeedsResult(result)
	if err != nil {
		t.Fatalf("ParseNewsListNewsfeedsResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestNewsService_GetNewsfeedRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/newsfeeds/10", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-nf-get")
		fmt.Fprint(w, `{"type":"newsfeed","id":"10","name":"Visitor Feed"}`)
	})

	ctx := context.Background()
	result, err := client.News.GetNewsfeedRaw(ctx, "10")
	if err != nil {
		t.Fatalf("GetNewsfeedRaw returned error: %v", err)
	}
	data, err := ParseNewsGetNewsfeedResult(result)
	if err != nil {
		t.Fatalf("ParseNewsGetNewsfeedResult returned error: %v", err)
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

func TestNewsService_ListNewsfeedItemsRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/news/newsfeeds/10/items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-nfi-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"news-item","id":"1","title":"Feed Item"}],"total_count":1,"pages":{"type":"pages","page":1,"per_page":20,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := client.News.ListNewsfeedItemsRaw(ctx, "10", nil)
	if err != nil {
		t.Fatalf("ListNewsfeedItemsRaw returned error: %v", err)
	}
	data, err := ParseNewsListNewsfeedItemsResult(result)
	if err != nil {
		t.Fatalf("ParseNewsListNewsfeedItemsResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}
