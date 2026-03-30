package articles_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/articles"
	"github.com/rassakhatsky/intercom-go-sdk/api"
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

func setup() (svc *articles.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = articles.NewService(caller)
	return svc, mux, server.Close
}

func setupInternal() (svc *articles.InternalService, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = articles.NewInternalService(caller)
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

// --- Articles Service Tests ---

func TestService_Get(t *testing.T) {
	svc, mux, teardown := setup()
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
	article, err := svc.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
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

func TestService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
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
	_, err := svc.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestService_Create(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body articles.CreateRequest
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
	article, err := svc.Create(ctx, &articles.CreateRequest{
		Title:    "New Article",
		AuthorID: 42,
		State:    "draft",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if article.ID != "456" {
		t.Errorf("Article.ID = %v, want 456", article.ID)
	}
	if article.Title != "New Article" {
		t.Errorf("Article.Title = %v, want New Article", article.Title)
	}
}

func TestService_Create_WithTranslatedContent(t *testing.T) {
	svc, mux, teardown := setup()
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
	article, err := svc.Create(ctx, &articles.CreateRequest{
		Title:    "New Article",
		AuthorID: 42,
		TranslatedContent: &api.ArticleTranslatedContent{
			Type: "article_translated_content",
			FR: &api.ArticleContent{
				Title: "Nouvel Article",
				Body:  "<p>Bienvenue</p>",
				State: "draft",
			},
		},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
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

func TestService_Update(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body articles.UpdateRequest
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
	article, err := svc.Update(ctx, "123", &articles.UpdateRequest{
		Title: "Updated Title",
		State: "published",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if article.Title != "Updated Title" {
		t.Errorf("Article.Title = %v, want Updated Title", article.Title)
	}
	if article.State != "published" {
		t.Errorf("Article.State = %v, want published", article.State)
	}
}

func TestService_Delete(t *testing.T) {
	svc, mux, teardown := setup()
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
	deleted, err := svc.Delete(ctx, "123")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
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

func TestService_List(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
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

func TestService_ListAll(t *testing.T) {
	svc, mux, teardown := setup()
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
	iter := svc.ListAll(ctx, nil)
	var allArticles []articles.Article
	for iter.Next() {
		allArticles = append(allArticles, iter.Current())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("ListAll iterator error: %v", err)
	}
	if len(allArticles) != 3 {
		t.Fatalf("ListAll returned %d articles, want 3", len(allArticles))
	}
	if allArticles[2].Title != "Article Three" {
		t.Errorf("articles[2].Title = %v, want Article Three", allArticles[2].Title)
	}
}

func TestService_Search(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.Search(ctx, &articles.SearchOptions{
		Phrase:    "getting started",
		State:     "published",
		Highlight: &highlight,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
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

func TestService_Search_WithHelpCenterID(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.Search(ctx, &articles.SearchOptions{
		HelpCenterID: 99,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("TotalCount = %v, want 0", result.TotalCount)
	}
}

// --- Raw companion method tests ---

func TestService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-abc")
		fmt.Fprint(w, `{"type":"article","id":"123","title":"Getting Started","author_id":42}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	data, err := articles.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
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

func TestService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Article not found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "999")
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

func TestService_ListRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	data, err := articles.ParseListResult(result)
	if err != nil {
		t.Fatalf("ParseListResult returned error: %v", err)
	}
	if len(data.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(data.Data))
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_CreateRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"type":"article","id":"456","title":"New Article","author_id":42}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &articles.CreateRequest{
		Title:    "New Article",
		AuthorID: 42,
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	data, err := articles.ParseCreateResult(result)
	if err != nil {
		t.Fatalf("ParseCreateResult returned error: %v", err)
	}
	if data.ID != "456" {
		t.Errorf("Data.ID = %v, want 456", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_UpdateRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"type":"article","id":"123","title":"Updated","state":"published"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateRaw(ctx, "123", &articles.UpdateRequest{
		Title: "Updated",
		State: "published",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	data, err := articles.ParseUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseUpdateResult returned error: %v", err)
	}
	if data.Title != "Updated" {
		t.Errorf("Data.Title = %v, want Updated", data.Title)
	}
}

func TestService_DeleteRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/articles/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"123","object":"article","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteRaw(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	data, err := articles.ParseDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseDeleteResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_SearchRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.SearchRaw(ctx, &articles.SearchOptions{
		Phrase: "getting started",
	})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	data, err := articles.ParseSearchResult(result)
	if err != nil {
		t.Fatalf("ParseSearchResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("Data.TotalCount = %d, want 1", data.TotalCount)
	}
	if result.Header.Get("X-RateLimit-Remaining") != "99" {
		t.Errorf("Header X-RateLimit-Remaining = %q, want 99", result.Header.Get("X-RateLimit-Remaining"))
	}
}

// --- Internal Articles Service Tests ---

func TestInternalService_Get(t *testing.T) {
	svc, mux, teardown := setupInternal()
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
	article, err := svc.Get(ctx, "45")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
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

func TestInternalService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setupInternal()
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
	_, err := svc.Get(ctx, "999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestInternalService_Create(t *testing.T) {
	svc, mux, teardown := setupInternal()
	defer teardown()

	mux.HandleFunc("/internal_articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body articles.CreateInternalRequest
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
	article, err := svc.Create(ctx, &articles.CreateInternalRequest{
		Title:    "Thanks for everything",
		Body:     "Body of the Article",
		OwnerID:  991266252,
		AuthorID: 991266252,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if article.ID != "42" {
		t.Errorf("ID = %v, want 42", article.ID)
	}
	if article.Title != "Thanks for everything" {
		t.Errorf("Title = %v, want Thanks for everything", article.Title)
	}
}

func TestInternalService_Update(t *testing.T) {
	svc, mux, teardown := setupInternal()
	defer teardown()

	mux.HandleFunc("/internal_articles/48", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body articles.UpdateInternalRequest
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
	article, err := svc.Update(ctx, "48", &articles.UpdateInternalRequest{
		Title: "Christmas is here!",
		Body:  "<p>New gifts in store for the jolly season</p>",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if article.ID != "48" {
		t.Errorf("ID = %v, want 48", article.ID)
	}
	if article.Title != "Christmas is here!" {
		t.Errorf("Title = %v, want Christmas is here!", article.Title)
	}
}

func TestInternalService_Delete(t *testing.T) {
	svc, mux, teardown := setupInternal()
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
	deleted, err := svc.Delete(ctx, "51")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
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

func TestInternalService_List(t *testing.T) {
	svc, mux, teardown := setupInternal()
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
	result, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
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

func TestInternalService_ListAll(t *testing.T) {
	svc, mux, teardown := setupInternal()
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
	iter := svc.ListAll(ctx, nil)
	var allArticles []articles.InternalArticle
	for iter.Next() {
		allArticles = append(allArticles, iter.Current())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("Iter returned error: %v", err)
	}
	if len(allArticles) != 2 {
		t.Fatalf("got %d articles, want 2", len(allArticles))
	}
	if allArticles[0].ID != "1" {
		t.Errorf("articles[0].ID = %v, want 1", allArticles[0].ID)
	}
	if allArticles[1].ID != "2" {
		t.Errorf("articles[1].ID = %v, want 2", allArticles[1].ID)
	}
}

func TestInternalService_Search(t *testing.T) {
	svc, mux, teardown := setupInternal()
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
	result, err := svc.Search(ctx, &articles.InternalSearchOptions{
		FolderID: "123",
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
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

func TestInternalService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setupInternal()
	defer teardown()

	mux.HandleFunc("/internal_articles/45", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ia-get")
		fmt.Fprint(w, `{"type":"internal_article","id":"45","title":"Internal Guide","owner_id":991266252,"author_id":991266252}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "45")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	data, err := articles.ParseInternalGetResult(result)
	if err != nil {
		t.Fatalf("ParseInternalGetResult returned error: %v", err)
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

func TestInternalService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setupInternal()
	defer teardown()

	mux.HandleFunc("/internal_articles/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "999")
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

func TestInternalService_ListRaw_Success(t *testing.T) {
	svc, mux, teardown := setupInternal()
	defer teardown()

	mux.HandleFunc("/internal_articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ia-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"internal_article","id":"39","title":"Article 1"}],"total_count":1,"pages":{"type":"pages","page":1,"per_page":25,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	data, err := articles.ParseInternalListResult(result)
	if err != nil {
		t.Fatalf("ParseInternalListResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestInternalService_CreateRaw_Success(t *testing.T) {
	svc, mux, teardown := setupInternal()
	defer teardown()

	mux.HandleFunc("/internal_articles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-ia-create")
		fmt.Fprint(w, `{"type":"internal_article","id":"42","title":"Thanks for everything"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &articles.CreateInternalRequest{
		Title:    "Thanks for everything",
		OwnerID:  991266252,
		AuthorID: 991266252,
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	data, err := articles.ParseInternalCreateResult(result)
	if err != nil {
		t.Fatalf("ParseInternalCreateResult returned error: %v", err)
	}
	if data.ID != "42" {
		t.Errorf("Data.ID = %v, want 42", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestInternalService_UpdateRaw_Success(t *testing.T) {
	svc, mux, teardown := setupInternal()
	defer teardown()

	mux.HandleFunc("/internal_articles/48", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-ia-update")
		fmt.Fprint(w, `{"type":"internal_article","id":"48","title":"Christmas is here!"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateRaw(ctx, "48", &articles.UpdateInternalRequest{Title: "Christmas is here!"})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	data, err := articles.ParseInternalUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseInternalUpdateResult returned error: %v", err)
	}
	if data.Title != "Christmas is here!" {
		t.Errorf("Data.Title = %v, want Christmas is here!", data.Title)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestInternalService_DeleteRaw_Success(t *testing.T) {
	svc, mux, teardown := setupInternal()
	defer teardown()

	mux.HandleFunc("/internal_articles/51", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-ia-del")
		fmt.Fprint(w, `{"id":"51","object":"internal_article","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteRaw(ctx, "51")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	data, err := articles.ParseInternalDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseInternalDeleteResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestInternalService_SearchRaw_Success(t *testing.T) {
	svc, mux, teardown := setupInternal()
	defer teardown()

	mux.HandleFunc("/internal_articles/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ia-search")
		fmt.Fprint(w, `{"type":"list","total_count":1,"data":{"internal_articles":[{"type":"internal_article","id":"55"}]},"pages":{"type":"pages","page":1,"total_pages":1,"per_page":10}}`)
	})

	ctx := context.Background()
	result, err := svc.SearchRaw(ctx, &articles.InternalSearchOptions{FolderID: "123"})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	data, err := articles.ParseInternalSearchResult(result)
	if err != nil {
		t.Fatalf("ParseInternalSearchResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}
