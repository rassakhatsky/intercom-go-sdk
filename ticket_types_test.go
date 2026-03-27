package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestTicketTypesService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"ticket_type",
			"id":"tt-1",
			"category":"Customer",
			"name":"Bug Report",
			"description":"Report a bug",
			"icon":"🐛",
			"workspace_id":"ws-1",
			"archived":false,
			"created_at":1734537460,
			"updated_at":1734537460,
			"ticket_type_attributes":{
				"type":"list",
				"data":[
					{"type":"ticket_type_attribute","id":"attr-1","name":"Priority","description":"Bug priority","data_type":"list","order":0,"required_to_create":true,"visible_on_create":true,"visible_to_contacts":true,"default":false,"ticket_type_id":1,"archived":false}
				]
			}
		}`)
	})

	ctx := context.Background()
	tt, err := client.TicketTypes.Get(ctx, "tt-1")
	if err != nil {
		t.Fatalf("TicketTypes.Get returned error: %v", err)
	}
	if tt.ID != "tt-1" {
		t.Errorf("TicketType.ID = %v, want tt-1", tt.ID)
	}
	if tt.Name != "Bug Report" {
		t.Errorf("TicketType.Name = %v, want Bug Report", tt.Name)
	}
	if tt.Description != "Report a bug" {
		t.Errorf("TicketType.Description = %v, want Report a bug", tt.Description)
	}
	if tt.Category != "Customer" {
		t.Errorf("TicketType.Category = %v, want Customer", tt.Category)
	}
	if tt.Icon != "🐛" {
		t.Errorf("TicketType.Icon = %v, want 🐛", tt.Icon)
	}
	if tt.WorkspaceID != "ws-1" {
		t.Errorf("TicketType.WorkspaceID = %v, want ws-1", tt.WorkspaceID)
	}
	if tt.Archived {
		t.Error("TicketType.Archived = true, want false")
	}
	if tt.Attributes == nil || len(tt.Attributes.Data) != 1 {
		t.Fatal("TicketType.Attributes unexpected")
	}
	attr := tt.Attributes.Data[0]
	if attr.ID != "attr-1" {
		t.Errorf("Attribute.ID = %v, want attr-1", attr.ID)
	}
	if attr.Name != "Priority" {
		t.Errorf("Attribute.Name = %v, want Priority", attr.Name)
	}
	if attr.DataType != "list" {
		t.Errorf("Attribute.DataType = %v, want list", attr.DataType)
	}
	if !attr.RequiredToCreate {
		t.Error("Attribute.RequiredToCreate = false, want true")
	}
}

func TestTicketTypesService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"ticket_type","id":"tt-1","name":"Bug Report","category":"Customer"},
				{"type":"ticket_type","id":"tt-2","name":"Feature Request","category":"Customer"}
			]
		}`)
	})

	ctx := context.Background()
	list, err := client.TicketTypes.List(ctx)
	if err != nil {
		t.Fatalf("TicketTypes.List returned error: %v", err)
	}
	if len(list.Data) != 2 {
		t.Fatalf("List returned %d ticket types, want 2", len(list.Data))
	}
	if list.Data[0].ID != "tt-1" {
		t.Errorf("TicketType[0].ID = %v, want tt-1", list.Data[0].ID)
	}
	if list.Data[1].Name != "Feature Request" {
		t.Errorf("TicketType[1].Name = %v, want Feature Request", list.Data[1].Name)
	}
}

func TestTicketTypesService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateTicketTypeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Bug Report" {
			t.Errorf("Create body name = %v, want Bug Report", body.Name)
		}
		if body.Description != "Report a bug" {
			t.Errorf("Create body description = %v, want Report a bug", body.Description)
		}
		if body.Category != "Customer" {
			t.Errorf("Create body category = %v, want Customer", body.Category)
		}
		if body.Icon != "🐛" {
			t.Errorf("Create body icon = %v, want 🐛", body.Icon)
		}
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-new","name":"Bug Report","description":"Report a bug","category":"Customer","icon":"🐛","archived":false}`)
	})

	ctx := context.Background()
	tt, err := client.TicketTypes.Create(ctx, &CreateTicketTypeRequest{
		Name:        "Bug Report",
		Description: "Report a bug",
		Category:    "Customer",
		Icon:        "🐛",
	})
	if err != nil {
		t.Fatalf("TicketTypes.Create returned error: %v", err)
	}
	if tt.ID != "tt-new" {
		t.Errorf("TicketType.ID = %v, want tt-new", tt.ID)
	}
	if tt.Name != "Bug Report" {
		t.Errorf("TicketType.Name = %v, want Bug Report", tt.Name)
	}
}

func TestTicketTypesService_Update(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateTicketTypeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Critical Bug" {
			t.Errorf("Update body name = %v, want Critical Bug", body.Name)
		}
		if body.Description != "Report a critical bug" {
			t.Errorf("Update body description = %v, want Report a critical bug", body.Description)
		}
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-1","name":"Critical Bug","description":"Report a critical bug","category":"Customer"}`)
	})

	ctx := context.Background()
	tt, err := client.TicketTypes.Update(ctx, "tt-1", &UpdateTicketTypeRequest{
		Name:        "Critical Bug",
		Description: "Report a critical bug",
	})
	if err != nil {
		t.Fatalf("TicketTypes.Update returned error: %v", err)
	}
	if tt.Name != "Critical Bug" {
		t.Errorf("TicketType.Name = %v, want Critical Bug", tt.Name)
	}
}

func TestTicketTypesService_Update_Archive(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["archived"] != true {
			t.Errorf("Update body archived = %v, want true", body["archived"])
		}
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-1","name":"Bug Report","archived":true}`)
	})

	ctx := context.Background()
	archived := true
	tt, err := client.TicketTypes.Update(ctx, "tt-1", &UpdateTicketTypeRequest{
		Archived: &archived,
	})
	if err != nil {
		t.Fatalf("TicketTypes.Update returned error: %v", err)
	}
	if !tt.Archived {
		t.Error("TicketType.Archived = false, want true")
	}
}

func TestTicketTypesService_CreateAttribute(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1/attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateTicketTypeAttributeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Priority" {
			t.Errorf("CreateAttribute body name = %v, want Priority", body.Name)
		}
		if body.Description != "Bug priority level" {
			t.Errorf("CreateAttribute body description = %v, want Bug priority level", body.Description)
		}
		if body.DataType != "list" {
			t.Errorf("CreateAttribute body data_type = %v, want list", body.DataType)
		}
		if !body.RequiredToCreate {
			t.Error("CreateAttribute body required_to_create = false, want true")
		}
		if !body.VisibleOnCreate {
			t.Error("CreateAttribute body visible_on_create = false, want true")
		}
		fmt.Fprint(w, `{
			"type":"ticket_type_attribute",
			"id":"attr-new",
			"name":"Priority",
			"description":"Bug priority level",
			"data_type":"list",
			"required_to_create":true,
			"visible_on_create":true,
			"visible_to_contacts":false,
			"default":false,
			"order":1,
			"ticket_type_id":1,
			"archived":false,
			"created_at":1734537460,
			"updated_at":1734537460
		}`)
	})

	ctx := context.Background()
	attr, err := client.TicketTypes.CreateAttribute(ctx, "tt-1", &CreateTicketTypeAttributeRequest{
		Name:             "Priority",
		Description:      "Bug priority level",
		DataType:         "list",
		RequiredToCreate: true,
		VisibleOnCreate:  true,
	})
	if err != nil {
		t.Fatalf("TicketTypes.CreateAttribute returned error: %v", err)
	}
	if attr.ID != "attr-new" {
		t.Errorf("Attribute.ID = %v, want attr-new", attr.ID)
	}
	if attr.Name != "Priority" {
		t.Errorf("Attribute.Name = %v, want Priority", attr.Name)
	}
	if attr.DataType != "list" {
		t.Errorf("Attribute.DataType = %v, want list", attr.DataType)
	}
	if !attr.RequiredToCreate {
		t.Error("Attribute.RequiredToCreate = false, want true")
	}
}

func TestTicketTypesService_UpdateAttribute(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1/attributes/attr-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateTicketTypeAttributeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Severity" {
			t.Errorf("UpdateAttribute body name = %v, want Severity", body.Name)
		}
		if body.Description != "Bug severity level" {
			t.Errorf("UpdateAttribute body description = %v, want Bug severity level", body.Description)
		}
		fmt.Fprint(w, `{
			"type":"ticket_type_attribute",
			"id":"attr-1",
			"name":"Severity",
			"description":"Bug severity level",
			"data_type":"list",
			"required_to_create":true,
			"visible_on_create":true,
			"archived":false
		}`)
	})

	ctx := context.Background()
	attr, err := client.TicketTypes.UpdateAttribute(ctx, "tt-1", "attr-1", &UpdateTicketTypeAttributeRequest{
		Name:        "Severity",
		Description: "Bug severity level",
	})
	if err != nil {
		t.Fatalf("TicketTypes.UpdateAttribute returned error: %v", err)
	}
	if attr.Name != "Severity" {
		t.Errorf("Attribute.Name = %v, want Severity", attr.Name)
	}
	if attr.Description != "Bug severity level" {
		t.Errorf("Attribute.Description = %v, want Bug severity level", attr.Description)
	}
}

func TestTicketTypesService_UpdateAttribute_Archive(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1/attributes/attr-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["archived"] != true {
			t.Errorf("UpdateAttribute body archived = %v, want true", body["archived"])
		}
		fmt.Fprint(w, `{"type":"ticket_type_attribute","id":"attr-1","name":"Priority","archived":true}`)
	})

	ctx := context.Background()
	archived := true
	attr, err := client.TicketTypes.UpdateAttribute(ctx, "tt-1", "attr-1", &UpdateTicketTypeAttributeRequest{
		Archived: &archived,
	})
	if err != nil {
		t.Fatalf("TicketTypes.UpdateAttribute returned error: %v", err)
	}
	if !attr.Archived {
		t.Error("Attribute.Archived = false, want true")
	}
}

func TestTicketTypesService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-tt-get")
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-1","name":"Bug Report","category":"Customer"}`)
	})

	ctx := context.Background()
	result, err := client.TicketTypes.GetRaw(ctx, "tt-1")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	data, err := ParseTicketTypeGetResult(result)
	if err != nil {
		t.Fatalf("ParseTicketTypeGetResult returned error: %v", err)
	}
	if data.ID != "tt-1" {
		t.Errorf("Data.ID = %v, want tt-1", data.ID)
	}
	if data.Name != "Bug Report" {
		t.Errorf("Data.Name = %v, want Bug Report", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-tt-get" {
		t.Errorf("Header X-Request-Id = %q, want req-tt-get", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestTicketTypesService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Not found"}]}`)
	})

	ctx := context.Background()
	result, err := client.TicketTypes.GetRaw(ctx, "tt-999")
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

func TestTicketTypesService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-tt-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"ticket_type","id":"tt-1","name":"Bug Report"},{"type":"ticket_type","id":"tt-2","name":"Feature Request"}]}`)
	})

	ctx := context.Background()
	result, err := client.TicketTypes.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	data, err := ParseTicketTypeListResult(result)
	if err != nil {
		t.Fatalf("ParseTicketTypeListResult returned error: %v", err)
	}
	if len(data.Data) != 2 {
		t.Fatalf("Data.Data length = %d, want 2", len(data.Data))
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTicketTypesService_CreateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-tt-create")
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-new","name":"Bug Report"}`)
	})

	ctx := context.Background()
	result, err := client.TicketTypes.CreateRaw(ctx, &CreateTicketTypeRequest{Name: "Bug Report"})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	data, err := ParseTicketTypeCreateResult(result)
	if err != nil {
		t.Fatalf("ParseTicketTypeCreateResult returned error: %v", err)
	}
	if data.ID != "tt-new" {
		t.Errorf("Data.ID = %v, want tt-new", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTicketTypesService_UpdateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-tt-update")
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-1","name":"Critical Bug"}`)
	})

	ctx := context.Background()
	result, err := client.TicketTypes.UpdateRaw(ctx, "tt-1", &UpdateTicketTypeRequest{Name: "Critical Bug"})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	data, err := ParseTicketTypeUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseTicketTypeUpdateResult returned error: %v", err)
	}
	if data.Name != "Critical Bug" {
		t.Errorf("Data.Name = %v, want Critical Bug", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTicketTypesService_CreateAttributeRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1/attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-tt-attr-create")
		fmt.Fprint(w, `{"type":"ticket_type_attribute","id":"attr-new","name":"Priority","data_type":"list"}`)
	})

	ctx := context.Background()
	result, err := client.TicketTypes.CreateAttributeRaw(ctx, "tt-1", &CreateTicketTypeAttributeRequest{
		Name:     "Priority",
		DataType: "list",
	})
	if err != nil {
		t.Fatalf("CreateAttributeRaw returned error: %v", err)
	}
	data, err := ParseTicketTypeCreateAttributeResult(result)
	if err != nil {
		t.Fatalf("ParseTicketTypeCreateAttributeResult returned error: %v", err)
	}
	if data.ID != "attr-new" {
		t.Errorf("Data.ID = %v, want attr-new", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTicketTypesService_UpdateAttributeRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1/attributes/attr-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-tt-attr-update")
		fmt.Fprint(w, `{"type":"ticket_type_attribute","id":"attr-1","name":"Severity","description":"Bug severity level"}`)
	})

	ctx := context.Background()
	result, err := client.TicketTypes.UpdateAttributeRaw(ctx, "tt-1", "attr-1", &UpdateTicketTypeAttributeRequest{
		Name:        "Severity",
		Description: "Bug severity level",
	})
	if err != nil {
		t.Fatalf("UpdateAttributeRaw returned error: %v", err)
	}
	data, err := ParseTicketTypeUpdateAttributeResult(result)
	if err != nil {
		t.Fatalf("ParseTicketTypeUpdateAttributeResult returned error: %v", err)
	}
	if data.Name != "Severity" {
		t.Errorf("Data.Name = %v, want Severity", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTicketTypesService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	_, err := client.TicketTypes.Get(ctx, "tt-999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true")
	}
}

func TestTicketTypesService_Create_Unauthorized(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ticket_types", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"type":"error.list","errors":[{"code":"unauthorized","message":"Unauthorized"}]}`)
	})

	ctx := context.Background()
	_, err := client.TicketTypes.Create(ctx, &CreateTicketTypeRequest{
		Name: "Test",
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("IsUnauthorized = false, want true")
	}
}
