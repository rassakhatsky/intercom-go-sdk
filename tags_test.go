package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestTagsService_Get(t *testing.T) {
	client, mux, teardown := setup()
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
	tag, err := client.Tags.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Tags.Get returned error: %v", err)
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

func TestTagsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
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
	_, err := client.Tags.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestTagsService_List(t *testing.T) {
	client, mux, teardown := setup()
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
	tags, err := client.Tags.List(ctx)
	if err != nil {
		t.Fatalf("Tags.List returned error: %v", err)
	}
	if tags.Type != "list" {
		t.Errorf("TagList.Type = %v, want list", tags.Type)
	}
	if len(tags.Data) != 3 {
		t.Fatalf("TagList.Data length = %d, want 3", len(tags.Data))
	}
	if tags.Data[0].Name != "VIP" {
		t.Errorf("Data[0].Name = %v, want VIP", tags.Data[0].Name)
	}
	if tags.Data[2].Name != "Enterprise" {
		t.Errorf("Data[2].Name = %v, want Enterprise", tags.Data[2].Name)
	}
}

func TestTagsService_CreateOrUpdate_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateOrUpdateTagRequest
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
	tag, err := client.Tags.CreateOrUpdate(ctx, &CreateOrUpdateTagRequest{
		Name: "NewTag",
	})
	if err != nil {
		t.Fatalf("Tags.CreateOrUpdate returned error: %v", err)
	}
	if tag.ID != "456" {
		t.Errorf("Tag.ID = %v, want 456", tag.ID)
	}
	if tag.Name != "NewTag" {
		t.Errorf("Tag.Name = %v, want NewTag", tag.Name)
	}
}

func TestTagsService_CreateOrUpdate_Update(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateOrUpdateTagRequest
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
	tag, err := client.Tags.CreateOrUpdate(ctx, &CreateOrUpdateTagRequest{
		Name: "RenamedTag",
		ID:   "456",
	})
	if err != nil {
		t.Fatalf("Tags.CreateOrUpdate returned error: %v", err)
	}
	if tag.Name != "RenamedTag" {
		t.Errorf("Tag.Name = %v, want RenamedTag", tag.Name)
	}
}

func TestTagsService_Delete(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	err := client.Tags.Delete(ctx, "123")
	if err != nil {
		t.Fatalf("Tags.Delete returned error: %v", err)
	}
}

func TestTagsService_TagCompany(t *testing.T) {
	client, mux, teardown := setup()
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
	tag, err := client.Tags.TagCompany(ctx, &TagCompanyRequest{
		Name: "VIP",
		Companies: []TagCompanyItem{
			{ID: "comp_123"},
		},
	})
	if err != nil {
		t.Fatalf("Tags.TagCompany returned error: %v", err)
	}
	if tag.Name != "VIP" {
		t.Errorf("Tag.Name = %v, want VIP", tag.Name)
	}
}

func TestTagsService_TagCompany_WithCompanyID(t *testing.T) {
	client, mux, teardown := setup()
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
	tag, err := client.Tags.TagCompany(ctx, &TagCompanyRequest{
		Name: "VIP",
		Companies: []TagCompanyItem{
			{CompanyID: "ext_456"},
		},
	})
	if err != nil {
		t.Fatalf("Tags.TagCompany returned error: %v", err)
	}
	if tag.ID != "789" {
		t.Errorf("Tag.ID = %v, want 789", tag.ID)
	}
}

func TestTagsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-tag-get")
		fmt.Fprint(w, `{"type":"tag","id":"123","name":"VIP"}`)
	})

	ctx := context.Background()
	result, err := client.Tags.GetRaw(ctx, "123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	data, err := ParseTagGetResult(result)
	if err != nil {
		t.Fatalf("ParseTagGetResult returned error: %v", err)
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

func TestTagsService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Tag not found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Tags.GetRaw(ctx, "nonexistent")
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

func TestTagsService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-tag-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"tag","id":"1","name":"VIP"},{"type":"tag","id":"2","name":"Trial"}]}`)
	})

	ctx := context.Background()
	result, err := client.Tags.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	data, err := ParseTagListResult(result)
	if err != nil {
		t.Fatalf("ParseTagListResult returned error: %v", err)
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

func TestTagsService_CreateOrUpdateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-tag-create")
		fmt.Fprint(w, `{"type":"tag","id":"456","name":"NewTag"}`)
	})

	ctx := context.Background()
	result, err := client.Tags.CreateOrUpdateRaw(ctx, &CreateOrUpdateTagRequest{Name: "NewTag"})
	if err != nil {
		t.Fatalf("CreateOrUpdateRaw returned error: %v", err)
	}
	data, err := ParseTagCreateOrUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseTagCreateOrUpdateResult returned error: %v", err)
	}
	if data.ID != "456" {
		t.Errorf("Data.ID = %v, want 456", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTagsService_DeleteRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-tag-del")
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	result, err := client.Tags.DeleteRaw(ctx, "123")
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

func TestTagsService_TagCompanyRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-tag-company")
		fmt.Fprint(w, `{"type":"tag","id":"789","name":"VIP"}`)
	})

	ctx := context.Background()
	result, err := client.Tags.TagCompanyRaw(ctx, &TagCompanyRequest{
		Name:      "VIP",
		Companies: []TagCompanyItem{{ID: "comp_123"}},
	})
	if err != nil {
		t.Fatalf("TagCompanyRaw returned error: %v", err)
	}
	data, err := ParseTagTagCompanyResult(result)
	if err != nil {
		t.Fatalf("ParseTagTagCompanyResult returned error: %v", err)
	}
	if data.Name != "VIP" {
		t.Errorf("Data.Name = %v, want VIP", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTagsService_UntagCompanyRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-untag-company")
		fmt.Fprint(w, `{"type":"tag","id":"789","name":"VIP"}`)
	})

	ctx := context.Background()
	result, err := client.Tags.UntagCompanyRaw(ctx, &UntagCompanyRequest{
		Name:      "VIP",
		Companies: []UntagCompanyItem{{ID: "comp_123", Untag: true}},
	})
	if err != nil {
		t.Fatalf("UntagCompanyRaw returned error: %v", err)
	}
	data, err := ParseTagUntagCompanyResult(result)
	if err != nil {
		t.Fatalf("ParseTagUntagCompanyResult returned error: %v", err)
	}
	if data.Name != "VIP" {
		t.Errorf("Data.Name = %v, want VIP", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTagsService_UntagCompany(t *testing.T) {
	client, mux, teardown := setup()
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
	tag, err := client.Tags.UntagCompany(ctx, &UntagCompanyRequest{
		Name: "VIP",
		Companies: []UntagCompanyItem{
			{ID: "comp_123", Untag: true},
		},
	})
	if err != nil {
		t.Fatalf("Tags.UntagCompany returned error: %v", err)
	}
	if tag.Name != "VIP" {
		t.Errorf("Tag.Name = %v, want VIP", tag.Name)
	}
}
