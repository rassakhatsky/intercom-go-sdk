package tags_test

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
	"github.com/rassakhatsky/intercom-go-sdk/tags"
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

func setup() (svc *tags.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = tags.NewService(caller)
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

func TestService_Get(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"tag",
			"id":"123",
			"name":"VIP"
		}`)
	})

	ctx := context.Background()
	tag, err := svc.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if tag.ID != "123" {
		t.Errorf("Tag.ID = %v, want 123", tag.ID)
	}
	if tag.Name != "VIP" {
		t.Errorf("Tag.Name = %v, want VIP", tag.Name)
	}
	if tag.Type != "tag" {
		t.Errorf("Tag.Type = %v, want tag", tag.Type)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"not_found","message":"Tag not found"}]
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

func TestService_List(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"tag","id":"1","name":"VIP"},
				{"type":"tag","id":"2","name":"Trial"},
				{"type":"tag","id":"3","name":"Enterprise"}
			]
		}`)
	})

	ctx := context.Background()
	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if list.Type != "list" {
		t.Errorf("List.Type = %v, want list", list.Type)
	}
	if len(list.Data) != 3 {
		t.Fatalf("List.Data length = %d, want 3", len(list.Data))
	}
	if list.Data[0].Name != "VIP" {
		t.Errorf("Data[0].Name = %v, want VIP", list.Data[0].Name)
	}
	if list.Data[2].Name != "Enterprise" {
		t.Errorf("Data[2].Name = %v, want Enterprise", list.Data[2].Name)
	}
}

func TestService_CreateOrUpdate_Create(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body tags.CreateOrUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "NewTag" {
			t.Errorf("body.Name = %v, want NewTag", body.Name)
		}
		if body.ID != "" {
			t.Errorf("body.ID = %v, want empty", body.ID)
		}
		fmt.Fprint(w, `{
			"type":"tag",
			"id":"456",
			"name":"NewTag"
		}`)
	})

	ctx := context.Background()
	tag, err := svc.CreateOrUpdate(ctx, &tags.CreateOrUpdateRequest{
		Name: "NewTag",
	})
	if err != nil {
		t.Fatalf("CreateOrUpdate returned error: %v", err)
	}
	if tag.ID != "456" {
		t.Errorf("Tag.ID = %v, want 456", tag.ID)
	}
	if tag.Name != "NewTag" {
		t.Errorf("Tag.Name = %v, want NewTag", tag.Name)
	}
}

func TestService_CreateOrUpdate_Update(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body tags.CreateOrUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "RenamedTag" {
			t.Errorf("body.Name = %v, want RenamedTag", body.Name)
		}
		if body.ID != "456" {
			t.Errorf("body.ID = %v, want 456", body.ID)
		}
		fmt.Fprint(w, `{
			"type":"tag",
			"id":"456",
			"name":"RenamedTag"
		}`)
	})

	ctx := context.Background()
	tag, err := svc.CreateOrUpdate(ctx, &tags.CreateOrUpdateRequest{
		Name: "RenamedTag",
		ID:   "456",
	})
	if err != nil {
		t.Fatalf("CreateOrUpdate returned error: %v", err)
	}
	if tag.Name != "RenamedTag" {
		t.Errorf("Tag.Name = %v, want RenamedTag", tag.Name)
	}
}

func TestService_Delete(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	err := svc.Delete(ctx, "123")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

func TestService_TagCompany(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["name"] != "VIP" {
			t.Errorf("body.name = %v, want VIP", body["name"])
		}
		companies, ok := body["companies"].([]any)
		if !ok {
			t.Fatal("body.companies is not an array")
		}
		if len(companies) != 1 {
			t.Fatalf("companies length = %d, want 1", len(companies))
		}
		comp := companies[0].(map[string]any)
		if comp["id"] != "comp_123" {
			t.Errorf("companies[0].id = %v, want comp_123", comp["id"])
		}
		fmt.Fprint(w, `{
			"type":"tag",
			"id":"789",
			"name":"VIP"
		}`)
	})

	ctx := context.Background()
	tag, err := svc.TagCompany(ctx, &tags.TagCompanyRequest{
		Name: "VIP",
		Companies: []tags.TagCompanyItem{
			{ID: "comp_123"},
		},
	})
	if err != nil {
		t.Fatalf("TagCompany returned error: %v", err)
	}
	if tag.Name != "VIP" {
		t.Errorf("Tag.Name = %v, want VIP", tag.Name)
	}
}

func TestService_TagCompany_WithCompanyID(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		companies := body["companies"].([]any)
		comp := companies[0].(map[string]any)
		if comp["company_id"] != "ext_456" {
			t.Errorf("companies[0].company_id = %v, want ext_456", comp["company_id"])
		}
		fmt.Fprint(w, `{
			"type":"tag",
			"id":"789",
			"name":"VIP"
		}`)
	})

	ctx := context.Background()
	tag, err := svc.TagCompany(ctx, &tags.TagCompanyRequest{
		Name: "VIP",
		Companies: []tags.TagCompanyItem{
			{CompanyID: "ext_456"},
		},
	})
	if err != nil {
		t.Fatalf("TagCompany returned error: %v", err)
	}
	if tag.ID != "789" {
		t.Errorf("Tag.ID = %v, want 789", tag.ID)
	}
}

func TestService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-tag-get")
		fmt.Fprint(w, `{"type":"tag","id":"123","name":"VIP"}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	data, err := tags.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
	}
	if data.ID != "123" {
		t.Errorf("Data.ID = %v, want 123", data.ID)
	}
	if data.Name != "VIP" {
		t.Errorf("Data.Name = %v, want VIP", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-tag-get" {
		t.Errorf("Header X-Request-Id = %q, want req-tag-get", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Tag not found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "nonexistent")
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

func TestService_ListRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-tag-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"tag","id":"1","name":"VIP"},{"type":"tag","id":"2","name":"Trial"}]}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	data, err := tags.ParseListResult(result)
	if err != nil {
		t.Fatalf("ParseListResult returned error: %v", err)
	}
	if len(data.Data) != 2 {
		t.Fatalf("Data.Data length = %d, want 2", len(data.Data))
	}
	if data.Data[0].Name != "VIP" {
		t.Errorf("Data.Data[0].Name = %v, want VIP", data.Data[0].Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-tag-list" {
		t.Errorf("Header X-Request-Id = %q, want req-tag-list", result.Header.Get("X-Request-Id"))
	}
}

func TestService_CreateOrUpdateRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-tag-create")
		fmt.Fprint(w, `{"type":"tag","id":"456","name":"NewTag"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateOrUpdateRaw(ctx, &tags.CreateOrUpdateRequest{Name: "NewTag"})
	if err != nil {
		t.Fatalf("CreateOrUpdateRaw returned error: %v", err)
	}
	data, err := tags.ParseCreateOrUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseCreateOrUpdateResult returned error: %v", err)
	}
	if data.ID != "456" {
		t.Errorf("Data.ID = %v, want 456", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_DeleteRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-tag-del")
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	result, err := svc.DeleteRaw(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	if result.StatusCode != 204 {
		t.Errorf("StatusCode = %d, want 204", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-tag-del" {
		t.Errorf("Header X-Request-Id = %q, want req-tag-del", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestService_TagCompanyRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-tag-company")
		fmt.Fprint(w, `{"type":"tag","id":"789","name":"VIP"}`)
	})

	ctx := context.Background()
	result, err := svc.TagCompanyRaw(ctx, &tags.TagCompanyRequest{
		Name:      "VIP",
		Companies: []tags.TagCompanyItem{{ID: "comp_123"}},
	})
	if err != nil {
		t.Fatalf("TagCompanyRaw returned error: %v", err)
	}
	data, err := tags.ParseTagCompanyResult(result)
	if err != nil {
		t.Fatalf("ParseTagCompanyResult returned error: %v", err)
	}
	if data.Name != "VIP" {
		t.Errorf("Data.Name = %v, want VIP", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_UntagCompanyRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-untag-company")
		fmt.Fprint(w, `{"type":"tag","id":"789","name":"VIP"}`)
	})

	ctx := context.Background()
	result, err := svc.UntagCompanyRaw(ctx, &tags.UntagCompanyRequest{
		Name:      "VIP",
		Companies: []tags.UntagCompanyItem{{ID: "comp_123", Untag: true}},
	})
	if err != nil {
		t.Fatalf("UntagCompanyRaw returned error: %v", err)
	}
	data, err := tags.ParseUntagCompanyResult(result)
	if err != nil {
		t.Fatalf("ParseUntagCompanyResult returned error: %v", err)
	}
	if data.Name != "VIP" {
		t.Errorf("Data.Name = %v, want VIP", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_UntagCompany(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["name"] != "VIP" {
			t.Errorf("body.name = %v, want VIP", body["name"])
		}
		companies := body["companies"].([]any)
		if len(companies) != 1 {
			t.Fatalf("companies length = %d, want 1", len(companies))
		}
		comp := companies[0].(map[string]any)
		if comp["id"] != "comp_123" {
			t.Errorf("companies[0].id = %v, want comp_123", comp["id"])
		}
		if comp["untag"] != true {
			t.Errorf("companies[0].untag = %v, want true", comp["untag"])
		}
		fmt.Fprint(w, `{
			"type":"tag",
			"id":"789",
			"name":"VIP"
		}`)
	})

	ctx := context.Background()
	tag, err := svc.UntagCompany(ctx, &tags.UntagCompanyRequest{
		Name: "VIP",
		Companies: []tags.UntagCompanyItem{
			{ID: "comp_123", Untag: true},
		},
	})
	if err != nil {
		t.Fatalf("UntagCompany returned error: %v", err)
	}
	if tag.Name != "VIP" {
		t.Errorf("Tag.Name = %v, want VIP", tag.Name)
	}
}
