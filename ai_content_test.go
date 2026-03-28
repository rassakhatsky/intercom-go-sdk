package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// --- Content Import Sources ---

func TestAIContentService_ListContentImportSources(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"content_import_source",
					"id":1,
					"url":"https://support.example.com/us/1",
					"sync_behavior":"automatic",
					"status":"active",
					"last_synced_at":1734537259,
					"created_at":1734537259,
					"updated_at":1734537259
				},
				{
					"type":"content_import_source",
					"id":2,
					"url":"https://support.example.com/us/2",
					"sync_behavior":"api",
					"status":"active",
					"last_synced_at":1734537260,
					"created_at":1734537260,
					"updated_at":1734537260
				}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":10,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.ListContentImportSources(ctx)
	if err != nil {
		t.Fatalf("AIContent.ListContentImportSources returned error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %v, want 2", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("len(Data) = %v, want 2", len(result.Data))
	}
	src := result.Data[0]
	if src.ID != 1 {
		t.Errorf("ID = %v, want 1", src.ID)
	}
	if src.Type != "content_import_source" {
		t.Errorf("Type = %v, want content_import_source", src.Type)
	}
	if src.URL != "https://support.example.com/us/1" {
		t.Errorf("URL = %v, want https://support.example.com/us/1", src.URL)
	}
	if src.SyncBehavior != "automatic" {
		t.Errorf("SyncBehavior = %v, want automatic", src.SyncBehavior)
	}
	if src.Status != "active" {
		t.Errorf("Status = %v, want active", src.Status)
	}
}

func TestAIContentService_GetContentImportSource(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources/5", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"content_import_source",
			"id":5,
			"url":"https://support.example.com/us/5",
			"sync_behavior":"api",
			"status":"active",
			"last_synced_at":1734537265,
			"created_at":1734537265,
			"updated_at":1734537265
		}`)
	})

	ctx := context.Background()
	src, err := client.AIContent.GetContentImportSource(ctx, "5")
	if err != nil {
		t.Fatalf("AIContent.GetContentImportSource returned error: %v", err)
	}
	if src.ID != 5 {
		t.Errorf("ID = %v, want 5", src.ID)
	}
	if src.URL != "https://support.example.com/us/5" {
		t.Errorf("URL = %v, want https://support.example.com/us/5", src.URL)
	}
	if src.SyncBehavior != "api" {
		t.Errorf("SyncBehavior = %v, want api", src.SyncBehavior)
	}
}

func TestAIContentService_CreateContentImportSource(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateContentImportSourceRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.SyncBehavior != "api" {
			t.Errorf("SyncBehavior = %v, want api", body.SyncBehavior)
		}
		if body.URL != "https://www.example.com" {
			t.Errorf("URL = %v, want https://www.example.com", body.URL)
		}
		fmt.Fprint(w, `{
			"type":"content_import_source",
			"id":10,
			"url":"https://www.example.com",
			"sync_behavior":"api",
			"status":"active",
			"last_synced_at":1734537261,
			"created_at":1734537261,
			"updated_at":1734537261
		}`)
	})

	ctx := context.Background()
	src, err := client.AIContent.CreateContentImportSource(ctx, &CreateContentImportSourceRequest{
		SyncBehavior: "api",
		URL:          "https://www.example.com",
	})
	if err != nil {
		t.Fatalf("AIContent.CreateContentImportSource returned error: %v", err)
	}
	if src.ID != 10 {
		t.Errorf("ID = %v, want 10", src.ID)
	}
	if src.URL != "https://www.example.com" {
		t.Errorf("URL = %v, want https://www.example.com", src.URL)
	}
}

func TestAIContentService_UpdateContentImportSource(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources/10", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateContentImportSourceRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.SyncBehavior != "api" {
			t.Errorf("SyncBehavior = %v, want api", body.SyncBehavior)
		}
		if body.URL != "https://www.example.com" {
			t.Errorf("URL = %v, want https://www.example.com", body.URL)
		}
		fmt.Fprint(w, `{
			"type":"content_import_source",
			"id":10,
			"url":"https://www.example.com",
			"sync_behavior":"api",
			"status":"active",
			"last_synced_at":1734537267,
			"created_at":1734537261,
			"updated_at":1734537267
		}`)
	})

	ctx := context.Background()
	src, err := client.AIContent.UpdateContentImportSource(ctx, "10", &UpdateContentImportSourceRequest{
		SyncBehavior: "api",
		URL:          "https://www.example.com",
	})
	if err != nil {
		t.Fatalf("AIContent.UpdateContentImportSource returned error: %v", err)
	}
	if src.ID != 10 {
		t.Errorf("ID = %v, want 10", src.ID)
	}
	if src.UpdatedAt != 1734537267 {
		t.Errorf("UpdatedAt = %v, want 1734537267", src.UpdatedAt)
	}
}

func TestAIContentService_DeleteContentImportSource(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources/10", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	err := client.AIContent.DeleteContentImportSource(ctx, "10")
	if err != nil {
		t.Fatalf("AIContent.DeleteContentImportSource returned error: %v", err)
	}
}

func TestAIContentService_GetContentImportSource_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-1",
			"errors":[{"code":"not_found","message":"Resource Not Found"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.AIContent.GetContentImportSource(ctx, "999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

// --- External Pages ---

func TestAIContentService_ListExternalPages(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"external_page",
					"id":"1",
					"title":"My External Content",
					"html":"<h1>Hello world</h1>",
					"url":"https://support.example.com/us/1",
					"ai_agent_availability":true,
					"ai_copilot_availability":true,
					"locale":"en",
					"source_id":100,
					"external_id":"ext-1",
					"created_at":1672928359,
					"updated_at":1672928610,
					"last_ingested_at":1672928610
				},
				{
					"type":"external_page",
					"id":"2",
					"title":"Another Page",
					"html":"<p>Content</p>",
					"url":"https://support.example.com/us/2",
					"ai_agent_availability":false,
					"ai_copilot_availability":true,
					"locale":"en",
					"source_id":100,
					"external_id":"ext-2",
					"created_at":1672928400,
					"updated_at":1672928700,
					"last_ingested_at":1672928700
				}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":10,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.ListExternalPages(ctx)
	if err != nil {
		t.Fatalf("AIContent.ListExternalPages returned error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %v, want 2", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("len(Data) = %v, want 2", len(result.Data))
	}
	page := result.Data[0]
	if page.ID != "1" {
		t.Errorf("ID = %v, want 1", page.ID)
	}
	if page.Title != "My External Content" {
		t.Errorf("Title = %v, want My External Content", page.Title)
	}
	if page.HTML != "<h1>Hello world</h1>" {
		t.Errorf("HTML = %v, want <h1>Hello world</h1>", page.HTML)
	}
	if page.URL != "https://support.example.com/us/1" {
		t.Errorf("URL = %v, want https://support.example.com/us/1", page.URL)
	}
	if !page.AIAgentAvailability {
		t.Errorf("AIAgentAvailability = %v, want true", page.AIAgentAvailability)
	}
	if !page.AICopilotAvailability {
		t.Errorf("AICopilotAvailability = %v, want true", page.AICopilotAvailability)
	}
	if page.SourceID != 100 {
		t.Errorf("SourceID = %v, want 100", page.SourceID)
	}
	if page.ExternalID != "ext-1" {
		t.Errorf("ExternalID = %v, want ext-1", page.ExternalID)
	}
}

func TestAIContentService_GetExternalPage(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages/22", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"external_page",
			"id":"22",
			"title":"My External Content",
			"html":"<h1>Hello world</h1>",
			"url":"https://support.example.com/us/5",
			"ai_agent_availability":true,
			"ai_copilot_availability":true,
			"locale":"en",
			"source_id":100,
			"external_id":"ext-22",
			"created_at":1672928359,
			"updated_at":1672928610,
			"last_ingested_at":1672928610
		}`)
	})

	ctx := context.Background()
	page, err := client.AIContent.GetExternalPage(ctx, "22")
	if err != nil {
		t.Fatalf("AIContent.GetExternalPage returned error: %v", err)
	}
	if page.ID != "22" {
		t.Errorf("ID = %v, want 22", page.ID)
	}
	if page.Title != "My External Content" {
		t.Errorf("Title = %v, want My External Content", page.Title)
	}
	if page.ExternalID != "ext-22" {
		t.Errorf("ExternalID = %v, want ext-22", page.ExternalID)
	}
}

func TestAIContentService_CreateExternalPage(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateExternalPageRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Title != "Test" {
			t.Errorf("Title = %v, want Test", body.Title)
		}
		if body.HTML != "<html><body><h1>Test</h1></body></html>" {
			t.Errorf("HTML = %v, want <html><body><h1>Test</h1></body></html>", body.HTML)
		}
		if body.SourceID != 1234 {
			t.Errorf("SourceID = %v, want 1234", body.SourceID)
		}
		if body.ExternalID != "abc1234" {
			t.Errorf("ExternalID = %v, want abc1234", body.ExternalID)
		}
		if body.Locale != "en" {
			t.Errorf("Locale = %v, want en", body.Locale)
		}

		fmt.Fprint(w, `{
			"type":"external_page",
			"id":"30",
			"title":"Test",
			"html":"<html><body><h1>Test</h1></body></html>",
			"url":"https://www.example.com",
			"ai_agent_availability":true,
			"ai_copilot_availability":true,
			"locale":"en",
			"source_id":1234,
			"external_id":"abc1234",
			"created_at":1672928359,
			"updated_at":1672928359,
			"last_ingested_at":1672928359
		}`)
	})

	ctx := context.Background()
	page, err := client.AIContent.CreateExternalPage(ctx, &CreateExternalPageRequest{
		Title:                 "Test",
		HTML:                  "<html><body><h1>Test</h1></body></html>",
		URL:                   "https://www.example.com",
		AIAgentAvailability:   boolPtr(true),
		AICopilotAvailability: boolPtr(true),
		Locale:                "en",
		SourceID:              1234,
		ExternalID:            "abc1234",
	})
	if err != nil {
		t.Fatalf("AIContent.CreateExternalPage returned error: %v", err)
	}
	if page.ID != "30" {
		t.Errorf("ID = %v, want 30", page.ID)
	}
	if page.Title != "Test" {
		t.Errorf("Title = %v, want Test", page.Title)
	}
	if page.SourceID != 1234 {
		t.Errorf("SourceID = %v, want 1234", page.SourceID)
	}
}

func TestAIContentService_UpdateExternalPage(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages/30", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateExternalPageRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Title != "Updated Title" {
			t.Errorf("Title = %v, want Updated Title", body.Title)
		}

		fmt.Fprint(w, `{
			"type":"external_page",
			"id":"30",
			"title":"Updated Title",
			"html":"<html><body><h1>Test</h1></body></html>",
			"url":"https://www.example.com",
			"ai_agent_availability":true,
			"ai_copilot_availability":true,
			"locale":"en",
			"source_id":1234,
			"external_id":"abc1234",
			"created_at":1672928359,
			"updated_at":1672928700,
			"last_ingested_at":1672928610
		}`)
	})

	ctx := context.Background()
	page, err := client.AIContent.UpdateExternalPage(ctx, "30", &UpdateExternalPageRequest{
		Title:    "Updated Title",
		SourceID: 1234,
	})
	if err != nil {
		t.Fatalf("AIContent.UpdateExternalPage returned error: %v", err)
	}
	if page.ID != "30" {
		t.Errorf("ID = %v, want 30", page.ID)
	}
	if page.Title != "Updated Title" {
		t.Errorf("Title = %v, want Updated Title", page.Title)
	}
	if page.UpdatedAt != 1672928700 {
		t.Errorf("UpdatedAt = %v, want 1672928700", page.UpdatedAt)
	}
}

func TestAIContentService_DeleteExternalPage(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages/22", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{
			"type":"external_page",
			"id":"22",
			"title":"My External Content",
			"html":"",
			"url":"https://support.example.com/us/5",
			"ai_agent_availability":true,
			"ai_copilot_availability":true,
			"locale":"en",
			"source_id":100,
			"external_id":"ext-22",
			"created_at":1672928359,
			"updated_at":1672928610,
			"last_ingested_at":1672928610
		}`)
	})

	ctx := context.Background()
	page, err := client.AIContent.DeleteExternalPage(ctx, "22")
	if err != nil {
		t.Fatalf("AIContent.DeleteExternalPage returned error: %v", err)
	}
	if page.ID != "22" {
		t.Errorf("ID = %v, want 22", page.ID)
	}
}

func TestAIContentService_GetExternalPage_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-1",
			"errors":[{"code":"not_found","message":"Resource Not Found"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.AIContent.GetExternalPage(ctx, "999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

// --- Raw Content Import Sources ---

func TestAIContentService_ListContentImportSourcesRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-list-cis")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"content_import_source","id":1,"url":"https://example.com","sync_behavior":"api","status":"active","last_synced_at":1734537259,"created_at":1734537259,"updated_at":1734537259}],"total_count":1,"pages":{"type":"pages","page":1,"per_page":10,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.ListContentImportSourcesRaw(ctx)
	if err != nil {
		t.Fatalf("ListContentImportSourcesRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-list-cis" {
		t.Errorf("Header X-Request-Id = %q, want req-list-cis", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	data, err := ParseAIContentListContentImportSourcesResult(result)
	if err != nil {
		t.Fatalf("ParseAIContentListContentImportSourcesResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if len(data.Data) != 1 {
		t.Fatalf("len(Data) = %v, want 1", len(data.Data))
	}
	if data.Data[0].ID != 1 {
		t.Errorf("Data[0].ID = %v, want 1", data.Data[0].ID)
	}
}

func TestAIContentService_GetContentImportSourceRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources/5", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-get-cis")
		fmt.Fprint(w, `{"type":"content_import_source","id":5,"url":"https://example.com/5","sync_behavior":"api","status":"active","last_synced_at":1734537265,"created_at":1734537265,"updated_at":1734537265}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.GetContentImportSourceRaw(ctx, "5")
	if err != nil {
		t.Fatalf("GetContentImportSourceRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-get-cis" {
		t.Errorf("Header X-Request-Id = %q, want req-get-cis", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	data, err := ParseAIContentGetContentImportSourceResult(result)
	if err != nil {
		t.Fatalf("ParseAIContentGetContentImportSourceResult returned error: %v", err)
	}
	if data.ID != 5 {
		t.Errorf("ID = %v, want 5", data.ID)
	}
}

func TestAIContentService_GetContentImportSourceRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.GetContentImportSourceRaw(ctx, "999")
	if err != nil {
		t.Fatalf("GetContentImportSourceRaw returned Go error: %v", err)
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

func TestAIContentService_CreateContentImportSourceRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-create-cis")
		fmt.Fprint(w, `{"type":"content_import_source","id":10,"url":"https://www.example.com","sync_behavior":"api","status":"active","last_synced_at":1734537261,"created_at":1734537261,"updated_at":1734537261}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.CreateContentImportSourceRaw(ctx, &CreateContentImportSourceRequest{
		SyncBehavior: "api",
		URL:          "https://www.example.com",
	})
	if err != nil {
		t.Fatalf("CreateContentImportSourceRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-create-cis" {
		t.Errorf("Header X-Request-Id = %q, want req-create-cis", result.Header.Get("X-Request-Id"))
	}
	data, err := ParseAIContentCreateContentImportSourceResult(result)
	if err != nil {
		t.Fatalf("ParseAIContentCreateContentImportSourceResult returned error: %v", err)
	}
	if data.ID != 10 {
		t.Errorf("ID = %v, want 10", data.ID)
	}
}

func TestAIContentService_UpdateContentImportSourceRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources/10", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-update-cis")
		fmt.Fprint(w, `{"type":"content_import_source","id":10,"url":"https://www.example.com","sync_behavior":"api","status":"active","last_synced_at":1734537267,"created_at":1734537261,"updated_at":1734537267}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.UpdateContentImportSourceRaw(ctx, "10", &UpdateContentImportSourceRequest{
		SyncBehavior: "api",
		URL:          "https://www.example.com",
	})
	if err != nil {
		t.Fatalf("UpdateContentImportSourceRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-update-cis" {
		t.Errorf("Header X-Request-Id = %q, want req-update-cis", result.Header.Get("X-Request-Id"))
	}
	data, err := ParseAIContentUpdateContentImportSourceResult(result)
	if err != nil {
		t.Fatalf("ParseAIContentUpdateContentImportSourceResult returned error: %v", err)
	}
	if data.ID != 10 {
		t.Errorf("ID = %v, want 10", data.ID)
	}
}

func TestAIContentService_DeleteContentImportSourceRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/content_import_sources/10", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	result, err := client.AIContent.DeleteContentImportSourceRaw(ctx, "10")
	if err != nil {
		t.Fatalf("DeleteContentImportSourceRaw returned error: %v", err)
	}
	if result.StatusCode != 204 {
		t.Errorf("StatusCode = %d, want 204", result.StatusCode)
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

// --- Raw External Pages ---

func TestAIContentService_ListExternalPagesRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-list-ep")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"external_page","id":"1","title":"Page One","html":"<h1>Hello</h1>","url":"https://example.com/1","ai_agent_availability":true,"ai_copilot_availability":true,"locale":"en","source_id":100,"external_id":"ext-1","created_at":1672928359,"updated_at":1672928610,"last_ingested_at":1672928610}],"total_count":1,"pages":{"type":"pages","page":1,"per_page":10,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.ListExternalPagesRaw(ctx)
	if err != nil {
		t.Fatalf("ListExternalPagesRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-list-ep" {
		t.Errorf("Header X-Request-Id = %q, want req-list-ep", result.Header.Get("X-Request-Id"))
	}
	data, err := ParseAIContentListExternalPagesResult(result)
	if err != nil {
		t.Fatalf("ParseAIContentListExternalPagesResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if len(data.Data) != 1 {
		t.Fatalf("len(Data) = %v, want 1", len(data.Data))
	}
	if data.Data[0].ID != "1" {
		t.Errorf("Data[0].ID = %v, want 1", data.Data[0].ID)
	}
}

func TestAIContentService_GetExternalPageRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages/22", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-get-ep")
		fmt.Fprint(w, `{"type":"external_page","id":"22","title":"My Page","html":"<h1>Hello</h1>","url":"https://example.com/22","ai_agent_availability":true,"ai_copilot_availability":true,"locale":"en","source_id":100,"external_id":"ext-22","created_at":1672928359,"updated_at":1672928610,"last_ingested_at":1672928610}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.GetExternalPageRaw(ctx, "22")
	if err != nil {
		t.Fatalf("GetExternalPageRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-get-ep" {
		t.Errorf("Header X-Request-Id = %q, want req-get-ep", result.Header.Get("X-Request-Id"))
	}
	data, err := ParseAIContentGetExternalPageResult(result)
	if err != nil {
		t.Fatalf("ParseAIContentGetExternalPageResult returned error: %v", err)
	}
	if data.ID != "22" {
		t.Errorf("ID = %v, want 22", data.ID)
	}
}

func TestAIContentService_GetExternalPageRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.GetExternalPageRaw(ctx, "999")
	if err != nil {
		t.Fatalf("GetExternalPageRaw returned Go error: %v", err)
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

func TestAIContentService_CreateExternalPageRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-create-ep")
		fmt.Fprint(w, `{"type":"external_page","id":"30","title":"Test","html":"<h1>Test</h1>","url":"https://www.example.com","ai_agent_availability":true,"ai_copilot_availability":true,"locale":"en","source_id":1234,"external_id":"abc1234","created_at":1672928359,"updated_at":1672928359,"last_ingested_at":1672928359}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.CreateExternalPageRaw(ctx, &CreateExternalPageRequest{
		Title:      "Test",
		HTML:       "<h1>Test</h1>",
		URL:        "https://www.example.com",
		Locale:     "en",
		SourceID:   1234,
		ExternalID: "abc1234",
	})
	if err != nil {
		t.Fatalf("CreateExternalPageRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-create-ep" {
		t.Errorf("Header X-Request-Id = %q, want req-create-ep", result.Header.Get("X-Request-Id"))
	}
	data, err := ParseAIContentCreateExternalPageResult(result)
	if err != nil {
		t.Fatalf("ParseAIContentCreateExternalPageResult returned error: %v", err)
	}
	if data.ID != "30" {
		t.Errorf("ID = %v, want 30", data.ID)
	}
}

func TestAIContentService_UpdateExternalPageRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages/30", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-update-ep")
		fmt.Fprint(w, `{"type":"external_page","id":"30","title":"Updated","html":"<h1>Test</h1>","url":"https://www.example.com","ai_agent_availability":true,"ai_copilot_availability":true,"locale":"en","source_id":1234,"external_id":"abc1234","created_at":1672928359,"updated_at":1672928700,"last_ingested_at":1672928610}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.UpdateExternalPageRaw(ctx, "30", &UpdateExternalPageRequest{
		Title: "Updated",
	})
	if err != nil {
		t.Fatalf("UpdateExternalPageRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-update-ep" {
		t.Errorf("Header X-Request-Id = %q, want req-update-ep", result.Header.Get("X-Request-Id"))
	}
	data, err := ParseAIContentUpdateExternalPageResult(result)
	if err != nil {
		t.Fatalf("ParseAIContentUpdateExternalPageResult returned error: %v", err)
	}
	if data.Title != "Updated" {
		t.Errorf("Title = %v, want Updated", data.Title)
	}
}

func TestAIContentService_DeleteExternalPageRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages/22", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-delete-ep")
		fmt.Fprint(w, `{"type":"external_page","id":"22","title":"My Page","html":"","url":"https://example.com/22","ai_agent_availability":true,"ai_copilot_availability":true,"locale":"en","source_id":100,"external_id":"ext-22","created_at":1672928359,"updated_at":1672928610,"last_ingested_at":1672928610}`)
	})

	ctx := context.Background()
	result, err := client.AIContent.DeleteExternalPageRaw(ctx, "22")
	if err != nil {
		t.Fatalf("DeleteExternalPageRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-delete-ep" {
		t.Errorf("Header X-Request-Id = %q, want req-delete-ep", result.Header.Get("X-Request-Id"))
	}
	data, err := ParseAIContentDeleteExternalPageResult(result)
	if err != nil {
		t.Fatalf("ParseAIContentDeleteExternalPageResult returned error: %v", err)
	}
	if data.ID != "22" {
		t.Errorf("ID = %v, want 22", data.ID)
	}
}

func TestAIContentService_UpdateExternalPage_WithExternalID(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ai/external_pages/30", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["external_id"] != "new-ext-id" {
			t.Errorf("external_id = %v, want new-ext-id", body["external_id"])
		}
		fmt.Fprint(w, `{
			"type":"external_page",
			"id":"30",
			"title":"Title",
			"html":"<h1>Test</h1>",
			"url":"https://www.example.com",
			"ai_agent_availability":true,
			"ai_copilot_availability":true,
			"locale":"en",
			"source_id":1234,
			"external_id":"new-ext-id",
			"created_at":1672928359,
			"updated_at":1672928700,
			"last_ingested_at":1672928610
		}`)
	})

	ctx := context.Background()
	page, err := client.AIContent.UpdateExternalPage(ctx, "30", &UpdateExternalPageRequest{
		Title:      "Title",
		ExternalID: "new-ext-id",
	})
	if err != nil {
		t.Fatalf("AIContent.UpdateExternalPage returned error: %v", err)
	}
	if page.ExternalID != "new-ext-id" {
		t.Errorf("ExternalID = %v, want new-ext-id", page.ExternalID)
	}
}

// boolPtr returns a pointer to the given bool value.
func boolPtr(b bool) *bool {
	return &b
}
