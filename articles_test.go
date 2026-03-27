package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestArticlesService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"article",
			"id":"123",
			"workspace_id":"ws1",
			"title":"Getting Started",
			"description":"A guide to getting started",
			"body":"<p>Welcome</p>",
			"author_id":42,
			"state":"published",
			"created_at":1672531200,
			"updated_at":1672617600,
			"url":"https://help.example.com/en/articles/123-getting-started",
			"parent_id":5,
			"parent_ids":[1,5],
			"parent_type":"collection",
			"default_locale":"en",
			"statistics":{
				"type":"article_statistics",
				"views":100,
				"conversations":5,
				"reactions":10,
				"happy_reaction_percentage":80.0,
				"neutral_reaction_percentage":15.0,
				"sad_reaction_percentage":5.0
			}
		}`)
	})

	ctx := context.Background()
	article, err := client.Articles.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Articles.Get returned error: %v", err)
	}
	if article.ID != "123" {
		t.Errorf("Article.ID = %v, want 123", article.ID)
	}
	if article.Title != "Getting Started" {
		t.Errorf("Article.Title = %v, want Getting Started", article.Title)
	}
	if article.AuthorID != 42 {
		t.Errorf("Article.AuthorID = %v, want 42", article.AuthorID)
	}
	if article.State != "published" {
		t.Errorf("Article.State = %v, want published", article.State)
	}
	if article.ParentID == nil || *article.ParentID != 5 {
		t.Errorf("Article.ParentID = %v, want 5", article.ParentID)
	}
	if article.ParentType == nil || *article.ParentType != "collection" {
		t.Errorf("Article.ParentType = %v, want collection", article.ParentType)
	}
	if article.Statistics == nil {
		t.Fatal("Article.Statistics is nil")
	}
	if article.Statistics.Views != 100 {
		t.Errorf("Statistics.Views = %v, want 100", article.Statistics.Views)
	}
	if article.Statistics.HappyReactionPercentage != 80.0 {
		t.Errorf("Statistics.HappyReactionPercentage = %v, want 80.0", article.Statistics.HappyReactionPercentage)
	}
}

func TestArticlesService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"not_found","message":"Article not found"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.Articles.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestArticlesService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateArticleRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Title != "New Article" {
			t.Errorf("Create body title = %v, want New Article", body.Title)
		}
		if body.AuthorID != 42 {
			t.Errorf("Create body author_id = %v, want 42", body.AuthorID)
		}
		if body.State != "draft" {
			t.Errorf("Create body state = %v, want draft", body.State)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"type":"article",
			"id":"456",
			"title":"New Article",
			"author_id":42,
			"state":"draft",
			"created_at":1672531200,
			"updated_at":1672531200
		}`)
	})

	ctx := context.Background()
	article, err := client.Articles.Create(ctx, &CreateArticleRequest{
		Title:    "New Article",
		AuthorID: 42,
		State:    "draft",
	})
	if err != nil {
		t.Fatalf("Articles.Create returned error: %v", err)
	}
	if article.ID != "456" {
		t.Errorf("Article.ID = %v, want 456", article.ID)
	}
	if article.Title != "New Article" {
		t.Errorf("Article.Title = %v, want New Article", article.Title)
	}
}

func TestArticlesService_Create_WithTranslatedContent(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		tc, ok := body["translated_content"].(map[string]any)
		if !ok {
			t.Fatal("translated_content missing from body")
		}
		fr, ok := tc["fr"].(map[string]any)
		if !ok {
			t.Fatal("translated_content.fr missing from body")
		}
		if fr["title"] != "Nouvel Article" {
			t.Errorf("translated_content.fr.title = %v, want Nouvel Article", fr["title"])
		}
		fmt.Fprint(w, `{
			"type":"article",
			"id":"456",
			"title":"New Article",
			"author_id":42,
			"state":"draft",
			"translated_content":{
				"type":"article_translated_content",
				"fr":{
					"type":"article_content",
					"title":"Nouvel Article",
					"body":"<p>Bienvenue</p>",
					"state":"draft"
				}
			}
		}`)
	})

	ctx := context.Background()
	article, err := client.Articles.Create(ctx, &CreateArticleRequest{
		Title:    "New Article",
		AuthorID: 42,
		TranslatedContent: &ArticleTranslatedContent{
			Type: "article_translated_content",
			FR: &ArticleContent{
				Title: "Nouvel Article",
				Body:  "<p>Bienvenue</p>",
				State: "draft",
			},
		},
	})
	if err != nil {
		t.Fatalf("Articles.Create returned error: %v", err)
	}
	if article.TranslatedContent == nil {
		t.Fatal("Article.TranslatedContent is nil")
	}
	if article.TranslatedContent.FR == nil {
		t.Fatal("Article.TranslatedContent.FR is nil")
	}
	if article.TranslatedContent.FR.Title != "Nouvel Article" {
		t.Errorf("TranslatedContent.FR.Title = %v, want Nouvel Article", article.TranslatedContent.FR.Title)
	}
}

func TestArticlesService_Update(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateArticleRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Title != "Updated Title" {
			t.Errorf("Update body title = %v, want Updated Title", body.Title)
		}
		if body.State != "published" {
			t.Errorf("Update body state = %v, want published", body.State)
		}
		fmt.Fprint(w, `{
			"type":"article",
			"id":"123",
			"title":"Updated Title",
			"state":"published",
			"author_id":42,
			"updated_at":1672617600
		}`)
	})

	ctx := context.Background()
	article, err := client.Articles.Update(ctx, "123", &UpdateArticleRequest{
		Title: "Updated Title",
		State: "published",
	})
	if err != nil {
		t.Fatalf("Articles.Update returned error: %v", err)
	}
	if article.Title != "Updated Title" {
		t.Errorf("Article.Title = %v, want Updated Title", article.Title)
	}
	if article.State != "published" {
		t.Errorf("Article.State = %v, want published", article.State)
	}
}

func TestArticlesService_Delete(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{
			"id":"123",
			"object":"article",
			"deleted":true
		}`)
	})

	ctx := context.Background()
	deleted, err := client.Articles.Delete(ctx, "123")
	if err != nil {
		t.Fatalf("Articles.Delete returned error: %v", err)
	}
	if deleted.ID != "123" {
		t.Errorf("Deleted.ID = %v, want 123", deleted.ID)
	}
	if !deleted.Deleted {
		t.Error("Deleted.Deleted = false, want true")
	}
	if deleted.Object != "article" {
		t.Errorf("Deleted.Object = %v, want article", deleted.Object)
	}
}

func TestArticlesService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"total_count":2,
			"data":[
				{"type":"article","id":"1","title":"Article One","author_id":42,"state":"published"},
				{"type":"article","id":"2","title":"Article Two","author_id":43,"state":"draft"}
			],
			"pages":{
				"type":"pages",
				"page":1,
				"per_page":20,
				"total_pages":1
			}
		}`)
	})

	ctx := context.Background()
	result, err := client.Articles.List(ctx, nil)
	if err != nil {
		t.Fatalf("Articles.List returned error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %v, want 2", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(result.Data))
	}
	if result.Data[0].Title != "Article One" {
		t.Errorf("Data[0].Title = %v, want Article One", result.Data[0].Title)
	}
	if result.Data[1].State != "draft" {
		t.Errorf("Data[1].State = %v, want draft", result.Data[1].State)
	}
}

func TestArticlesService_ListAll(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	callCount := 0
	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		callCount++
		if callCount == 1 {
			fmt.Fprint(w, `{
				"type":"list",
				"total_count":3,
				"data":[
					{"type":"article","id":"1","title":"Article One","author_id":42},
					{"type":"article","id":"2","title":"Article Two","author_id":42}
				],
				"pages":{
					"type":"pages",
					"page":1,
					"per_page":2,
					"total_pages":2,
					"next":{"page":2,"starting_after":"art_2"}
				}
			}`)
		} else {
			fmt.Fprint(w, `{
				"type":"list",
				"total_count":3,
				"data":[
					{"type":"article","id":"3","title":"Article Three","author_id":42}
				],
				"pages":{
					"type":"pages",
					"page":2,
					"per_page":2,
					"total_pages":2
				}
			}`)
		}
	})

	ctx := context.Background()
	iter := client.Articles.ListAll(ctx, nil)
	var articles []Article
	for iter.Next() {
		articles = append(articles, iter.Current())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("ListAll iterator error: %v", err)
	}
	if len(articles) != 3 {
		t.Fatalf("ListAll returned %d articles, want 3", len(articles))
	}
	if articles[2].Title != "Article Three" {
		t.Errorf("articles[2].Title = %v, want Article Three", articles[2].Title)
	}
}

func TestArticlesService_Search(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("phrase"); got != "getting started" {
			t.Errorf("phrase = %v, want getting started", got)
		}
		if got := r.URL.Query().Get("state"); got != "published" {
			t.Errorf("state = %v, want published", got)
		}
		if got := r.URL.Query().Get("highlight"); got != "true" {
			t.Errorf("highlight = %v, want true", got)
		}
		fmt.Fprint(w, `{
			"type":"list",
			"total_count":1,
			"data":{
				"articles":[
					{
						"type":"article",
						"id":"123",
						"title":"Getting Started",
						"state":"published",
						"author_id":42
					}
				],
				"highlights":[
					{
						"article_id":"123",
						"highlighted_title":[
							{"type":"highlight","text":"Getting Started"}
						]
					}
				]
			},
			"pages":{
				"type":"pages",
				"page":1,
				"total_pages":1,
				"per_page":25
			}
		}`)
	})

	ctx := context.Background()
	highlight := true
	result, err := client.Articles.Search(ctx, &ArticleSearchOptions{
		Phrase:    "getting started",
		State:     "published",
		Highlight: &highlight,
	})
	if err != nil {
		t.Fatalf("Articles.Search returned error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", result.TotalCount)
	}
	if len(result.Data.Articles) != 1 {
		t.Fatalf("Articles length = %d, want 1", len(result.Data.Articles))
	}
	if result.Data.Articles[0].Title != "Getting Started" {
		t.Errorf("Articles[0].Title = %v, want Getting Started", result.Data.Articles[0].Title)
	}
	if len(result.Data.Highlights) != 1 {
		t.Fatalf("Highlights length = %d, want 1", len(result.Data.Highlights))
	}
	if result.Data.Highlights[0].ArticleID != "123" {
		t.Errorf("Highlights[0].ArticleID = %v, want 123", result.Data.Highlights[0].ArticleID)
	}
}

func TestArticlesService_Search_WithHelpCenterID(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("help_center_id"); got != "99" {
			t.Errorf("help_center_id = %v, want 99", got)
		}
		fmt.Fprint(w, `{
			"type":"list",
			"total_count":0,
			"data":{"articles":[]},
			"pages":{"type":"pages","page":1,"total_pages":0,"per_page":25}
		}`)
	})

	ctx := context.Background()
	result, err := client.Articles.Search(ctx, &ArticleSearchOptions{
		HelpCenterID: 99,
	})
	if err != nil {
		t.Fatalf("Articles.Search returned error: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("TotalCount = %v, want 0", result.TotalCount)
	}
}

// --- Raw companion method tests ---

func TestArticlesService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-abc")
		fmt.Fprint(w, `{"type":"article","id":"123","title":"Getting Started","author_id":42}`)
	})

	ctx := context.Background()
	result, err := client.Articles.GetRaw(ctx, "123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	data, err := ParseArticleGetResult(result)
	if err != nil {
		t.Fatalf("ParseArticleGetResult returned error: %v", err)
	}
	if data.ID != "123" {
		t.Errorf("Data.ID = %v, want 123", data.ID)
	}
	if data.Title != "Getting Started" {
		t.Errorf("Data.Title = %v, want Getting Started", data.Title)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-abc" {
		t.Errorf("Header X-Request-Id = %q, want req-abc", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestArticlesService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Article not found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Articles.GetRaw(ctx, "999")
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

func TestArticlesService_ListRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"article","id":"1","title":"One"},{"type":"article","id":"2","title":"Two"}],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":20,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Articles.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	data, err := ParseArticleListResult(result)
	if err != nil {
		t.Fatalf("ParseArticleListResult returned error: %v", err)
	}
	if len(data.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(data.Data))
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestArticlesService_CreateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"type":"article","id":"456","title":"New Article","author_id":42}`)
	})

	ctx := context.Background()
	result, err := client.Articles.CreateRaw(ctx, &CreateArticleRequest{
		Title:    "New Article",
		AuthorID: 42,
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	data, err := ParseArticleCreateResult(result)
	if err != nil {
		t.Fatalf("ParseArticleCreateResult returned error: %v", err)
	}
	if data.ID != "456" {
		t.Errorf("Data.ID = %v, want 456", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestArticlesService_UpdateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"type":"article","id":"123","title":"Updated","state":"published"}`)
	})

	ctx := context.Background()
	result, err := client.Articles.UpdateRaw(ctx, "123", &UpdateArticleRequest{
		Title: "Updated",
		State: "published",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	data, err := ParseArticleUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseArticleUpdateResult returned error: %v", err)
	}
	if data.Title != "Updated" {
		t.Errorf("Data.Title = %v, want Updated", data.Title)
	}
}

func TestArticlesService_DeleteRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"123","object":"article","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.Articles.DeleteRaw(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	data, err := ParseArticleDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseArticleDeleteResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestArticlesService_SearchRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-RateLimit-Remaining", "99")
		fmt.Fprint(w, `{
			"type":"list",
			"total_count":1,
			"data":{"articles":[{"type":"article","id":"123","title":"Getting Started"}]},
			"pages":{"type":"pages","page":1,"total_pages":1,"per_page":25}
		}`)
	})

	ctx := context.Background()
	result, err := client.Articles.SearchRaw(ctx, &ArticleSearchOptions{
		Phrase: "getting started",
	})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	data, err := ParseArticleSearchResult(result)
	if err != nil {
		t.Fatalf("ParseArticleSearchResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("Data.TotalCount = %d, want 1", data.TotalCount)
	}
	if result.Header.Get("X-RateLimit-Remaining") != "99" {
		t.Errorf("Header X-RateLimit-Remaining = %q, want 99", result.Header.Get("X-RateLimit-Remaining"))
	}
}
