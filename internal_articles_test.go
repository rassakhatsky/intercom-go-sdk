package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestInternalArticlesService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles/45", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"internal_article",
			"id":"45",
			"title":"Internal Guide",
			"body":"Body of the Article",
			"owner_id":991266252,
			"author_id":991266252,
			"locale":"en",
			"created_at":1672928359,
			"updated_at":1672928610
		}`)
	})

	ctx := context.Background()
	article, err := client.InternalArticles.Get(ctx, "45")
	if err != nil {
		t.Fatalf("InternalArticles.Get returned error: %v", err)
	}
	if article.ID != "45" {
		t.Errorf("ID = %v, want 45", article.ID)
	}
	if article.Title != "Internal Guide" {
		t.Errorf("Title = %v, want Internal Guide", article.Title)
	}
	if article.Body != "Body of the Article" {
		t.Errorf("Body = %v, want Body of the Article", article.Body)
	}
	if article.OwnerID != 991266252 {
		t.Errorf("OwnerID = %v, want 991266252", article.OwnerID)
	}
	if article.AuthorID != 991266252 {
		t.Errorf("AuthorID = %v, want 991266252", article.AuthorID)
	}
	if article.Locale != "en" {
		t.Errorf("Locale = %v, want en", article.Locale)
	}
}

func TestInternalArticlesService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"test-req-id",
			"errors":[{"code":"not_found","message":"Resource Not Found"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.InternalArticles.Get(ctx, "999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestInternalArticlesService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateInternalArticleRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Title != "Thanks for everything" {
			t.Errorf("Title = %v, want Thanks for everything", body.Title)
		}
		if body.OwnerID != 991266252 {
			t.Errorf("OwnerID = %v, want 991266252", body.OwnerID)
		}
		if body.AuthorID != 991266252 {
			t.Errorf("AuthorID = %v, want 991266252", body.AuthorID)
		}
		fmt.Fprint(w, `{
			"type":"internal_article",
			"id":"42",
			"title":"Thanks for everything",
			"body":"Body of the Article",
			"owner_id":991266252,
			"author_id":991266252,
			"locale":"en"
		}`)
	})

	ctx := context.Background()
	article, err := client.InternalArticles.Create(ctx, &CreateInternalArticleRequest{
		Title:    "Thanks for everything",
		Body:     "Body of the Article",
		OwnerID:  991266252,
		AuthorID: 991266252,
	})
	if err != nil {
		t.Fatalf("InternalArticles.Create returned error: %v", err)
	}
	if article.ID != "42" {
		t.Errorf("ID = %v, want 42", article.ID)
	}
	if article.Title != "Thanks for everything" {
		t.Errorf("Title = %v, want Thanks for everything", article.Title)
	}
}

func TestInternalArticlesService_Update(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles/48", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateInternalArticleRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Title != "Christmas is here!" {
			t.Errorf("Title = %v, want Christmas is here!", body.Title)
		}
		fmt.Fprint(w, `{
			"type":"internal_article",
			"id":"48",
			"title":"Christmas is here!",
			"body":"<p>New gifts in store for the jolly season</p>",
			"owner_id":991266252,
			"author_id":991266252,
			"locale":"en"
		}`)
	})

	ctx := context.Background()
	article, err := client.InternalArticles.Update(ctx, "48", &UpdateInternalArticleRequest{
		Title: "Christmas is here!",
		Body:  "<p>New gifts in store for the jolly season</p>",
	})
	if err != nil {
		t.Fatalf("InternalArticles.Update returned error: %v", err)
	}
	if article.ID != "48" {
		t.Errorf("ID = %v, want 48", article.ID)
	}
	if article.Title != "Christmas is here!" {
		t.Errorf("Title = %v, want Christmas is here!", article.Title)
	}
}

func TestInternalArticlesService_Delete(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles/51", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{
			"id":"51",
			"object":"internal_article",
			"deleted":true
		}`)
	})

	ctx := context.Background()
	deleted, err := client.InternalArticles.Delete(ctx, "51")
	if err != nil {
		t.Fatalf("InternalArticles.Delete returned error: %v", err)
	}
	if deleted.ID != "51" {
		t.Errorf("ID = %v, want 51", deleted.ID)
	}
	if !deleted.Deleted {
		t.Error("Deleted = false, want true")
	}
	if deleted.Object != "internal_article" {
		t.Errorf("Object = %v, want internal_article", deleted.Object)
	}
}

func TestInternalArticlesService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"internal_article","id":"39","title":"Article 1","owner_id":1,"author_id":1,"locale":"en"},
				{"type":"internal_article","id":"40","title":"Article 2","owner_id":1,"author_id":1,"locale":"en"}
			],
			"total_count":2,
			"pages":{
				"type":"pages",
				"page":1,
				"per_page":25,
				"total_pages":1
			}
		}`)
	})

	ctx := context.Background()
	result, err := client.InternalArticles.List(ctx, nil)
	if err != nil {
		t.Fatalf("InternalArticles.List returned error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %v, want 2", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("len(Data) = %v, want 2", len(result.Data))
	}
	if result.Data[0].ID != "39" {
		t.Errorf("Data[0].ID = %v, want 39", result.Data[0].ID)
	}
	if result.Data[1].Title != "Article 2" {
		t.Errorf("Data[1].Title = %v, want Article 2", result.Data[1].Title)
	}
}

func TestInternalArticlesService_ListAll(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	callCount := 0
	mux.HandleFunc("/internal_articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		callCount++
		if callCount == 1 {
			fmt.Fprint(w, `{
				"type":"list",
				"data":[
					{"type":"internal_article","id":"1","title":"First","owner_id":1,"author_id":1,"locale":"en"}
				],
				"total_count":2,
				"pages":{
					"type":"pages",
					"page":1,
					"per_page":1,
					"total_pages":2,
					"next":{"page":2,"starting_after":"cursor1"}
				}
			}`)
		} else {
			fmt.Fprint(w, `{
				"type":"list",
				"data":[
					{"type":"internal_article","id":"2","title":"Second","owner_id":1,"author_id":1,"locale":"en"}
				],
				"total_count":2,
				"pages":{
					"type":"pages",
					"page":2,
					"per_page":1,
					"total_pages":2
				}
			}`)
		}
	})

	ctx := context.Background()
	iter := client.InternalArticles.ListAll(ctx, nil)
	var articles []InternalArticle
	for iter.Next() {
		articles = append(articles, iter.Current())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("Iter returned error: %v", err)
	}
	if len(articles) != 2 {
		t.Fatalf("got %d articles, want 2", len(articles))
	}
	if articles[0].ID != "1" {
		t.Errorf("articles[0].ID = %v, want 1", articles[0].ID)
	}
	if articles[1].ID != "2" {
		t.Errorf("articles[1].ID = %v, want 2", articles[1].ID)
	}
}

func TestInternalArticlesService_Search(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("folder_id"); got != "123" {
			t.Errorf("folder_id = %v, want 123", got)
		}
		fmt.Fprint(w, `{
			"type":"list",
			"total_count":1,
			"data":{
				"internal_articles":[
					{"type":"internal_article","id":"55","body":"Body of the Article","owner_id":991266252,"author_id":991266252,"locale":"en"}
				]
			},
			"pages":{
				"type":"pages",
				"page":1,
				"total_pages":1,
				"per_page":10
			}
		}`)
	})

	ctx := context.Background()
	result, err := client.InternalArticles.Search(ctx, &InternalArticleSearchOptions{
		FolderID: "123",
	})
	if err != nil {
		t.Fatalf("InternalArticles.Search returned error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", result.TotalCount)
	}
	if len(result.Data.InternalArticles) != 1 {
		t.Fatalf("len(InternalArticles) = %v, want 1", len(result.Data.InternalArticles))
	}
	if result.Data.InternalArticles[0].ID != "55" {
		t.Errorf("ID = %v, want 55", result.Data.InternalArticles[0].ID)
	}
}

func TestInternalArticlesService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles/45", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ia-get")
		fmt.Fprint(w, `{"type":"internal_article","id":"45","title":"Internal Guide","owner_id":991266252,"author_id":991266252}`)
	})

	ctx := context.Background()
	result, err := client.InternalArticles.GetRaw(ctx, "45")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	data, err := ParseInternalArticleGetResult(result)
	if err != nil {
		t.Fatalf("ParseInternalArticleGetResult returned error: %v", err)
	}
	if data.ID != "45" {
		t.Errorf("Data.ID = %v, want 45", data.ID)
	}
	if data.Title != "Internal Guide" {
		t.Errorf("Data.Title = %v, want Internal Guide", data.Title)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestInternalArticlesService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := client.InternalArticles.GetRaw(ctx, "999")
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

func TestInternalArticlesService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ia-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"internal_article","id":"39","title":"Article 1"}],"total_count":1,"pages":{"type":"pages","page":1,"per_page":25,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := client.InternalArticles.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	data, err := ParseInternalArticleListResult(result)
	if err != nil {
		t.Fatalf("ParseInternalArticleListResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestInternalArticlesService_CreateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-ia-create")
		fmt.Fprint(w, `{"type":"internal_article","id":"42","title":"Thanks for everything"}`)
	})

	ctx := context.Background()
	result, err := client.InternalArticles.CreateRaw(ctx, &CreateInternalArticleRequest{
		Title:    "Thanks for everything",
		OwnerID:  991266252,
		AuthorID: 991266252,
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	data, err := ParseInternalArticleCreateResult(result)
	if err != nil {
		t.Fatalf("ParseInternalArticleCreateResult returned error: %v", err)
	}
	if data.ID != "42" {
		t.Errorf("Data.ID = %v, want 42", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestInternalArticlesService_UpdateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles/48", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-ia-update")
		fmt.Fprint(w, `{"type":"internal_article","id":"48","title":"Christmas is here!"}`)
	})

	ctx := context.Background()
	result, err := client.InternalArticles.UpdateRaw(ctx, "48", &UpdateInternalArticleRequest{Title: "Christmas is here!"})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	data, err := ParseInternalArticleUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseInternalArticleUpdateResult returned error: %v", err)
	}
	if data.Title != "Christmas is here!" {
		t.Errorf("Data.Title = %v, want Christmas is here!", data.Title)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestInternalArticlesService_DeleteRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles/51", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-ia-del")
		fmt.Fprint(w, `{"id":"51","object":"internal_article","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.InternalArticles.DeleteRaw(ctx, "51")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	data, err := ParseInternalArticleDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseInternalArticleDeleteResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestInternalArticlesService_SearchRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/internal_articles/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ia-search")
		fmt.Fprint(w, `{"type":"list","total_count":1,"data":{"internal_articles":[{"type":"internal_article","id":"55"}]},"pages":{"type":"pages","page":1,"total_pages":1,"per_page":10}}`)
	})

	ctx := context.Background()
	result, err := client.InternalArticles.SearchRaw(ctx, &InternalArticleSearchOptions{FolderID: "123"})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	data, err := ParseInternalArticleSearchResult(result)
	if err != nil {
		t.Fatalf("ParseInternalArticleSearchResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}
