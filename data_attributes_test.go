package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestDataAttributesService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"data_attribute",
					"id":1,
					"name":"name",
					"full_name":"name",
					"label":"Company name",
					"description":"The name of a company",
					"data_type":"string",
					"api_writable":true,
					"ui_writable":false,
					"messenger_writable":true,
					"custom":false,
					"archived":false,
					"model":"company"
				},
				{
					"type":"data_attribute",
					"id":2,
					"name":"paid_subscriber",
					"full_name":"custom_attributes.paid_subscriber",
					"label":"Paid Subscriber",
					"description":"Whether the user is a paid subscriber",
					"data_type":"boolean",
					"api_writable":true,
					"ui_writable":false,
					"messenger_writable":false,
					"custom":true,
					"archived":false,
					"model":"contact",
					"admin_id":"12345",
					"created_at":1671028894,
					"updated_at":1671028894
				}
			]
		}`)
	})

	ctx := context.Background()
	list, err := client.DataAttributes.List(ctx, nil)
	if err != nil {
		t.Fatalf("DataAttributes.List returned error: %v", err)
	}
	if list.Type != "list" {
		t.Errorf("Type = %v, want list", list.Type)
	}
	if len(list.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(list.Data))
	}
	if list.Data[0].Name != "name" {
		t.Errorf("Data[0].Name = %v, want name", list.Data[0].Name)
	}
	if list.Data[0].Model != "company" {
		t.Errorf("Data[0].Model = %v, want company", list.Data[0].Model)
	}
	if list.Data[1].Custom != true {
		t.Errorf("Data[1].Custom = %v, want true", list.Data[1].Custom)
	}
	if list.Data[1].AdminID != "12345" {
		t.Errorf("Data[1].AdminID = %v, want 12345", list.Data[1].AdminID)
	}
}

func TestDataAttributesService_List_WithOptions(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("model"); got != "company" {
			t.Errorf("query model = %v, want company", got)
		}
		if got := r.URL.Query().Get("include_archived"); got != "true" {
			t.Errorf("query include_archived = %v, want true", got)
		}
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"data_attribute",
					"id":1,
					"name":"name",
					"model":"company"
				}
			]
		}`)
	})

	ctx := context.Background()
	includeArchived := true
	list, err := client.DataAttributes.List(ctx, &ListDataAttributesOptions{
		Model:           "company",
		IncludeArchived: &includeArchived,
	})
	if err != nil {
		t.Fatalf("DataAttributes.List returned error: %v", err)
	}
	if len(list.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(list.Data))
	}
}

func TestDataAttributesService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateDataAttributeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Mithril Shirt" {
			t.Errorf("body.Name = %v, want Mithril Shirt", body.Name)
		}
		if body.Model != "company" {
			t.Errorf("body.Model = %v, want company", body.Model)
		}
		if body.DataType != "string" {
			t.Errorf("body.DataType = %v, want string", body.DataType)
		}
		fmt.Fprint(w, `{
			"id":37,
			"type":"data_attribute",
			"name":"Mithril Shirt",
			"full_name":"custom_attributes.Mithril Shirt",
			"label":"Mithril Shirt",
			"data_type":"string",
			"api_writable":true,
			"ui_writable":false,
			"messenger_writable":false,
			"custom":true,
			"archived":false,
			"admin_id":"12345",
			"created_at":1734537756,
			"updated_at":1734537756,
			"model":"company"
		}`)
	})

	ctx := context.Background()
	attr, err := client.DataAttributes.Create(ctx, &CreateDataAttributeRequest{
		Name:     "Mithril Shirt",
		Model:    "company",
		DataType: "string",
	})
	if err != nil {
		t.Fatalf("DataAttributes.Create returned error: %v", err)
	}
	if attr.ID != 37 {
		t.Errorf("ID = %v, want 37", attr.ID)
	}
	if attr.Name != "Mithril Shirt" {
		t.Errorf("Name = %v, want Mithril Shirt", attr.Name)
	}
	if attr.Custom != true {
		t.Errorf("Custom = %v, want true", attr.Custom)
	}
	if attr.Model != "company" {
		t.Errorf("Model = %v, want company", attr.Model)
	}
}

func TestDataAttributesService_Create_WithOptions(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["data_type"] != "options" {
			t.Errorf("body.data_type = %v, want options", body["data_type"])
		}
		opts, ok := body["options"].([]any)
		if !ok {
			t.Fatal("body.options is not an array")
		}
		if len(opts) != 2 {
			t.Fatalf("options length = %d, want 2", len(opts))
		}
		fmt.Fprint(w, `{
			"id":38,
			"type":"data_attribute",
			"name":"Size Range",
			"data_type":"string",
			"options":["1-10","11-50"],
			"custom":true,
			"model":"company"
		}`)
	})

	ctx := context.Background()
	attr, err := client.DataAttributes.Create(ctx, &CreateDataAttributeRequest{
		Name:     "Size Range",
		Model:    "company",
		DataType: "options",
		Options:  []AttributeOption{{Value: "1-10"}, {Value: "11-50"}},
	})
	if err != nil {
		t.Fatalf("DataAttributes.Create returned error: %v", err)
	}
	if attr.ID != 38 {
		t.Errorf("ID = %v, want 38", attr.ID)
	}
	if len(attr.Options) != 2 {
		t.Fatalf("Options length = %d, want 2", len(attr.Options))
	}
}

func TestDataAttributesService_Update(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes/44", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateDataAttributeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Description != "Just a plain old ring" {
			t.Errorf("body.Description = %v, want Just a plain old ring", body.Description)
		}
		if *body.MessengerWritable != true {
			t.Errorf("body.MessengerWritable = %v, want true", *body.MessengerWritable)
		}
		fmt.Fprint(w, `{
			"id":44,
			"type":"data_attribute",
			"name":"The One Ring",
			"full_name":"custom_attributes.The One Ring",
			"label":"The One Ring",
			"description":"Just a plain old ring",
			"data_type":"string",
			"options":["1-10","11-20"],
			"api_writable":true,
			"ui_writable":false,
			"messenger_writable":true,
			"custom":true,
			"archived":false,
			"admin_id":"12345",
			"created_at":1734537762,
			"updated_at":1734537763,
			"model":"company"
		}`)
	})

	ctx := context.Background()
	messengerWritable := true
	attr, err := client.DataAttributes.Update(ctx, 44, &UpdateDataAttributeRequest{
		Description:       "Just a plain old ring",
		MessengerWritable: &messengerWritable,
	})
	if err != nil {
		t.Fatalf("DataAttributes.Update returned error: %v", err)
	}
	if attr.ID != 44 {
		t.Errorf("ID = %v, want 44", attr.ID)
	}
	if attr.Description != "Just a plain old ring" {
		t.Errorf("Description = %v, want Just a plain old ring", attr.Description)
	}
	if attr.MessengerWritable != true {
		t.Errorf("MessengerWritable = %v, want true", attr.MessengerWritable)
	}
}

func TestDataAttributesService_Update_Archive(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes/44", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["archived"] != true {
			t.Errorf("body.archived = %v, want true", body["archived"])
		}
		fmt.Fprint(w, `{
			"id":44,
			"type":"data_attribute",
			"name":"The One Ring",
			"archived":true,
			"model":"company"
		}`)
	})

	ctx := context.Background()
	archived := true
	attr, err := client.DataAttributes.Update(ctx, 44, &UpdateDataAttributeRequest{
		Archived: &archived,
	})
	if err != nil {
		t.Fatalf("DataAttributes.Update returned error: %v", err)
	}
	if attr.Archived != true {
		t.Errorf("Archived = %v, want true", attr.Archived)
	}
}

func TestDataAttributesService_Update_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"field_not_found","message":"We couldn't find that data attribute to update"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.DataAttributes.Update(ctx, 999, &UpdateDataAttributeRequest{
		Description: "test",
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

// --- Raw companion method tests ---

func TestDataAttributesService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-da-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"data_attribute","id":1,"name":"name","model":"company"}]}`)
	})

	ctx := context.Background()
	result, err := client.DataAttributes.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	data, err := ParseDataAttributeListResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(data.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(data.Data))
	}
	if data.Data[0].Name != "name" {
		t.Errorf("Data[0].Name = %v, want name", data.Data[0].Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-da-list" {
		t.Errorf("Header X-Request-Id = %q, want req-da-list", result.Header.Get("X-Request-Id"))
	}
}

func TestDataAttributesService_CreateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-da-create")
		fmt.Fprint(w, `{"id":37,"type":"data_attribute","name":"Mithril Shirt","data_type":"string","custom":true,"model":"company"}`)
	})

	ctx := context.Background()
	result, err := client.DataAttributes.CreateRaw(ctx, &CreateDataAttributeRequest{
		Name:     "Mithril Shirt",
		Model:    "company",
		DataType: "string",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	data, err := ParseDataAttributeCreateResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if data.ID != 37 {
		t.Errorf("ID = %v, want 37", data.ID)
	}
	if data.Name != "Mithril Shirt" {
		t.Errorf("Name = %v, want Mithril Shirt", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestDataAttributesService_UpdateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes/44", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-da-update")
		fmt.Fprint(w, `{"id":44,"type":"data_attribute","name":"The One Ring","description":"Just a plain old ring","model":"company"}`)
	})

	ctx := context.Background()
	result, err := client.DataAttributes.UpdateRaw(ctx, 44, &UpdateDataAttributeRequest{
		Description: "Just a plain old ring",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	data, err := ParseDataAttributeUpdateResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if data.ID != 44 {
		t.Errorf("ID = %v, want 44", data.ID)
	}
	if data.Description != "Just a plain old ring" {
		t.Errorf("Description = %v, want Just a plain old ring", data.Description)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-da-update" {
		t.Errorf("Header X-Request-Id = %q, want req-da-update", result.Header.Get("X-Request-Id"))
	}
}

func TestDataAttributesService_UpdateRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-123","errors":[{"code":"field_not_found","message":"We couldn't find that data attribute to update"}]}`)
	})

	ctx := context.Background()
	result, err := client.DataAttributes.UpdateRaw(ctx, 999, &UpdateDataAttributeRequest{
		Description: "test",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned Go error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Result.Error is nil, want non-nil")
	}
	if result.Error.Code != "field_not_found" {
		t.Errorf("Error.Code = %q, want field_not_found", result.Error.Code)
	}
	if result.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", result.StatusCode)
	}
}

func TestDataAttributesService_List_Unauthorized(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/data_attributes", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-456",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.DataAttributes.List(ctx, nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}
