package helpcenter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/helpcenter"
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

func setup() (svc *helpcenter.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = helpcenter.NewService(caller)
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

func TestService_ListCollections(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"total_count":2,
			"data":[
				{"type":"collection","id":"1","name":"Getting Started","order":1,"default_locale":"en"},
				{"type":"collection","id":"2","name":"FAQ","order":2,"default_locale":"en"}
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
	result, err := svc.ListCollections(ctx, nil)
	if err != nil {
		t.Fatalf("ListCollections returned error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %v, want 2", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(result.Data))
	}
	if result.Data[0].Name != "Getting Started" {
		t.Errorf("Data[0].Name = %v, want Getting Started", result.Data[0].Name)
	}
	if result.Data[1].Name != "FAQ" {
		t.Errorf("Data[1].Name = %v, want FAQ", result.Data[1].Name)
	}
}

func TestService_GetCollection(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"collection",
			"id":"123",
			"workspace_id":"ws1",
			"name":"Getting Started",
			"description":"A collection of getting started articles",
			"created_at":1672531200,
			"updated_at":1672617600,
			"url":"https://help.example.com/en/collections/123-getting-started",
			"icon":"book",
			"order":1,
			"default_locale":"en",
			"parent_id":"456",
			"help_center_id":99
		}`)
	})

	ctx := context.Background()
	collection, err := svc.GetCollection(ctx, "123")
	if err != nil {
		t.Fatalf("GetCollection returned error: %v", err)
	}
	if collection.ID != "123" {
		t.Errorf("Collection.ID = %v, want 123", collection.ID)
	}
	if collection.Name != "Getting Started" {
		t.Errorf("Collection.Name = %v, want Getting Started", collection.Name)
	}
	if collection.Description != "A collection of getting started articles" {
		t.Errorf("Collection.Description = %v, want description text", collection.Description)
	}
	if collection.Icon != "book" {
		t.Errorf("Collection.Icon = %v, want book", collection.Icon)
	}
	if collection.Order != 1 {
		t.Errorf("Collection.Order = %v, want 1", collection.Order)
	}
	if collection.ParentID != "456" {
		t.Errorf("Collection.ParentID = %v, want 456", collection.ParentID)
	}
	if collection.HelpCenterID != 99 {
		t.Errorf("Collection.HelpCenterID = %v, want 99", collection.HelpCenterID)
	}
}

func TestService_GetCollection_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"not_found","message":"Collection not found"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.GetCollection(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestService_CreateCollection(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body helpcenter.CreateCollectionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "New Collection" {
			t.Errorf("Create body name = %v, want New Collection", body.Name)
		}
		if body.Description != "A new collection" {
			t.Errorf("Create body description = %v, want A new collection", body.Description)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"type":"collection",
			"id":"789",
			"name":"New Collection",
			"description":"A new collection",
			"created_at":1672531200,
			"updated_at":1672531200
		}`)
	})

	ctx := context.Background()
	collection, err := svc.CreateCollection(ctx, &helpcenter.CreateCollectionRequest{
		Name:        "New Collection",
		Description: "A new collection",
	})
	if err != nil {
		t.Fatalf("CreateCollection returned error: %v", err)
	}
	if collection.ID != "789" {
		t.Errorf("Collection.ID = %v, want 789", collection.ID)
	}
	if collection.Name != "New Collection" {
		t.Errorf("Collection.Name = %v, want New Collection", collection.Name)
	}
}

func TestService_CreateCollection_WithParent(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["parent_id"] != "456" {
			t.Errorf("Create body parent_id = %v, want 456", body["parent_id"])
		}
		if body["help_center_id"] != float64(99) {
			t.Errorf("Create body help_center_id = %v, want 99", body["help_center_id"])
		}
		fmt.Fprint(w, `{
			"type":"collection",
			"id":"789",
			"name":"Sub Collection",
			"parent_id":"456",
			"help_center_id":99
		}`)
	})

	ctx := context.Background()
	hcID := 99
	collection, err := svc.CreateCollection(ctx, &helpcenter.CreateCollectionRequest{
		Name:         "Sub Collection",
		ParentID:     "456",
		HelpCenterID: &hcID,
	})
	if err != nil {
		t.Fatalf("CreateCollection returned error: %v", err)
	}
	if collection.ParentID != "456" {
		t.Errorf("Collection.ParentID = %v, want 456", collection.ParentID)
	}
	if collection.HelpCenterID != 99 {
		t.Errorf("Collection.HelpCenterID = %v, want 99", collection.HelpCenterID)
	}
}

func TestService_UpdateCollection(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body helpcenter.UpdateCollectionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Updated Collection" {
			t.Errorf("Update body name = %v, want Updated Collection", body.Name)
		}
		fmt.Fprint(w, `{
			"type":"collection",
			"id":"123",
			"name":"Updated Collection",
			"description":"Updated description",
			"updated_at":1672617600
		}`)
	})

	ctx := context.Background()
	collection, err := svc.UpdateCollection(ctx, "123", &helpcenter.UpdateCollectionRequest{
		Name: "Updated Collection",
	})
	if err != nil {
		t.Fatalf("UpdateCollection returned error: %v", err)
	}
	if collection.Name != "Updated Collection" {
		t.Errorf("Collection.Name = %v, want Updated Collection", collection.Name)
	}
}

func TestService_DeleteCollection(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{
			"id":"123",
			"object":"collection",
			"deleted":true
		}`)
	})

	ctx := context.Background()
	deleted, err := svc.DeleteCollection(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteCollection returned error: %v", err)
	}
	if deleted.ID != "123" {
		t.Errorf("Deleted.ID = %v, want 123", deleted.ID)
	}
	if !deleted.Deleted {
		t.Error("Deleted.Deleted = false, want true")
	}
	if deleted.Object != "collection" {
		t.Errorf("Deleted.Object = %v, want collection", deleted.Object)
	}
}

func TestService_ListHelpCenters(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/help_centers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"help_center",
					"id":"1",
					"workspace_id":"ws1",
					"identifier":"help",
					"website_turned_on":true,
					"display_name":"Help Center",
					"created_at":1672531200,
					"updated_at":1672617600
				},
				{
					"type":"help_center",
					"id":"2",
					"workspace_id":"ws1",
					"identifier":"docs",
					"website_turned_on":false,
					"display_name":"Documentation",
					"created_at":1672531200,
					"updated_at":1672617600
				}
			]
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListHelpCenters(ctx)
	if err != nil {
		t.Fatalf("ListHelpCenters returned error: %v", err)
	}
	if result.Type != "list" {
		t.Errorf("Type = %v, want list", result.Type)
	}
	if len(result.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(result.Data))
	}
	if result.Data[0].DisplayName != "Help Center" {
		t.Errorf("Data[0].DisplayName = %v, want Help Center", result.Data[0].DisplayName)
	}
	if !result.Data[0].WebsiteTurnedOn {
		t.Error("Data[0].WebsiteTurnedOn = false, want true")
	}
	if result.Data[1].Identifier != "docs" {
		t.Errorf("Data[1].Identifier = %v, want docs", result.Data[1].Identifier)
	}
}

func TestService_GetHelpCenter(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/help_centers/42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"help_center",
			"id":"42",
			"workspace_id":"ws1",
			"identifier":"help",
			"website_turned_on":true,
			"display_name":"Help Center",
			"url":"https://help.example.com",
			"custom_domain":"help.example.com",
			"created_at":1672531200,
			"updated_at":1672617600
		}`)
	})

	ctx := context.Background()
	hc, err := svc.GetHelpCenter(ctx, "42")
	if err != nil {
		t.Fatalf("GetHelpCenter returned error: %v", err)
	}
	if hc.ID != "42" {
		t.Errorf("HelpCenter.ID = %v, want 42", hc.ID)
	}
	if hc.DisplayName != "Help Center" {
		t.Errorf("HelpCenter.DisplayName = %v, want Help Center", hc.DisplayName)
	}
	if hc.URL != "https://help.example.com" {
		t.Errorf("HelpCenter.URL = %v, want https://help.example.com", hc.URL)
	}
	if hc.CustomDomain != "help.example.com" {
		t.Errorf("HelpCenter.CustomDomain = %v, want help.example.com", hc.CustomDomain)
	}
	if !hc.WebsiteTurnedOn {
		t.Error("HelpCenter.WebsiteTurnedOn = false, want true")
	}
}

func TestService_GetHelpCenter_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/help_centers/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-456",
			"errors":[{"code":"not_found","message":"Help Center not found"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.GetHelpCenter(ctx, "999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestService_ListCollectionsRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-lc")
		fmt.Fprint(w, `{"type":"list","total_count":1,"data":[{"type":"collection","id":"1","name":"Getting Started"}],"pages":{"type":"pages","page":1,"per_page":20,"total_pages":1}}`)
	})

	ctx := context.Background()
	result, err := svc.ListCollectionsRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListCollectionsRaw returned error: %v", err)
	}
	data, err := helpcenter.ParseListCollectionsResult(result)
	if err != nil {
		t.Fatalf("ParseListCollectionsResult returned error: %v", err)
	}
	if data.TotalCount != 1 {
		t.Errorf("TotalCount = %v, want 1", data.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_GetCollectionRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-gc")
		fmt.Fprint(w, `{"type":"collection","id":"123","name":"Getting Started"}`)
	})

	ctx := context.Background()
	result, err := svc.GetCollectionRaw(ctx, "123")
	if err != nil {
		t.Fatalf("GetCollectionRaw returned error: %v", err)
	}
	data, err := helpcenter.ParseGetCollectionResult(result)
	if err != nil {
		t.Fatalf("ParseGetCollectionResult returned error: %v", err)
	}
	if data.ID != "123" {
		t.Errorf("Data.ID = %v, want 123", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestService_GetCollectionRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Collection not found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetCollectionRaw(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetCollectionRaw returned Go error: %v", err)
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

func TestService_CreateCollectionRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-cc")
		fmt.Fprint(w, `{"type":"collection","id":"789","name":"New Collection"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateCollectionRaw(ctx, &helpcenter.CreateCollectionRequest{Name: "New Collection"})
	if err != nil {
		t.Fatalf("CreateCollectionRaw returned error: %v", err)
	}
	data, err := helpcenter.ParseCreateCollectionResult(result)
	if err != nil {
		t.Fatalf("ParseCreateCollectionResult returned error: %v", err)
	}
	if data.ID != "789" {
		t.Errorf("Data.ID = %v, want 789", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_UpdateCollectionRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-uc")
		fmt.Fprint(w, `{"type":"collection","id":"123","name":"Updated Collection"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateCollectionRaw(ctx, "123", &helpcenter.UpdateCollectionRequest{Name: "Updated Collection"})
	if err != nil {
		t.Fatalf("UpdateCollectionRaw returned error: %v", err)
	}
	data, err := helpcenter.ParseUpdateCollectionResult(result)
	if err != nil {
		t.Fatalf("ParseUpdateCollectionResult returned error: %v", err)
	}
	if data.Name != "Updated Collection" {
		t.Errorf("Data.Name = %v, want Updated Collection", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_DeleteCollectionRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/collections/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-dc")
		fmt.Fprint(w, `{"id":"123","object":"collection","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteCollectionRaw(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteCollectionRaw returned error: %v", err)
	}
	data, err := helpcenter.ParseDeleteCollectionResult(result)
	if err != nil {
		t.Fatalf("ParseDeleteCollectionResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_ListHelpCentersRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/help_centers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-lhc")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"help_center","id":"1","display_name":"Help Center"}]}`)
	})

	ctx := context.Background()
	result, err := svc.ListHelpCentersRaw(ctx)
	if err != nil {
		t.Fatalf("ListHelpCentersRaw returned error: %v", err)
	}
	data, err := helpcenter.ParseListHelpCentersResult(result)
	if err != nil {
		t.Fatalf("ParseListHelpCentersResult returned error: %v", err)
	}
	if len(data.Data) != 1 {
		t.Fatalf("Data.Data length = %d, want 1", len(data.Data))
	}
	if data.Data[0].DisplayName != "Help Center" {
		t.Errorf("Data.Data[0].DisplayName = %v, want Help Center", data.Data[0].DisplayName)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_GetHelpCenterRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/help_center/help_centers/42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ghc")
		fmt.Fprint(w, `{"type":"help_center","id":"42","display_name":"Help Center","website_turned_on":true}`)
	})

	ctx := context.Background()
	result, err := svc.GetHelpCenterRaw(ctx, "42")
	if err != nil {
		t.Fatalf("GetHelpCenterRaw returned error: %v", err)
	}
	data, err := helpcenter.ParseGetHelpCenterResult(result)
	if err != nil {
		t.Fatalf("ParseGetHelpCenterResult returned error: %v", err)
	}
	if data.ID != "42" {
		t.Errorf("Data.ID = %v, want 42", data.ID)
	}
	if !data.WebsiteTurnedOn {
		t.Error("Data.WebsiteTurnedOn = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}
