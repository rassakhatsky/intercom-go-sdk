package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCustomObjectsService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order/obj_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"id":"obj_123",
			"external_id":"ext_123",
			"type":"Order",
			"created_at":1700000000,
			"updated_at":1700000100,
			"custom_attributes":{"status":"shipped"}
		}`)
	})

	ctx := context.Background()
	obj, err := client.CustomObjects.Get(ctx, "Order", "obj_123")
	if err != nil {
		t.Fatalf("CustomObjects.Get returned error: %v", err)
	}
	if obj.ID != "obj_123" {
		t.Errorf("ID = %v, want obj_123", obj.ID)
	}
	if obj.ExternalID != "ext_123" {
		t.Errorf("ExternalID = %v, want ext_123", obj.ExternalID)
	}
	if obj.Type != "Order" {
		t.Errorf("Type = %v, want Order", obj.Type)
	}
	if obj.CustomAttributes["status"] != "shipped" {
		t.Errorf("CustomAttributes[status] = %v, want shipped", obj.CustomAttributes["status"])
	}
}

func TestCustomObjectsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Object not found"}]}`)
	})

	ctx := context.Background()
	_, err := client.CustomObjects.Get(ctx, "Order", "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestCustomObjectsService_GetByExternalID(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if r.URL.Query().Get("external_id") != "ext_123" {
			t.Errorf("external_id = %v, want ext_123", r.URL.Query().Get("external_id"))
		}
		fmt.Fprint(w, `{
			"id":"obj_123",
			"external_id":"ext_123",
			"type":"Order",
			"custom_attributes":{"status":"pending"}
		}`)
	})

	ctx := context.Background()
	obj, err := client.CustomObjects.GetByExternalID(ctx, "Order", "ext_123")
	if err != nil {
		t.Fatalf("CustomObjects.GetByExternalID returned error: %v", err)
	}
	if obj.ExternalID != "ext_123" {
		t.Errorf("ExternalID = %v, want ext_123", obj.ExternalID)
	}
}

func TestCustomObjectsService_CreateOrUpdate(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["external_id"] != "ext_456" {
			t.Errorf("external_id = %v, want ext_456", body["external_id"])
		}
		fmt.Fprint(w, `{
			"id":"obj_456",
			"external_id":"ext_456",
			"type":"Order",
			"created_at":1700000000,
			"updated_at":1700000000,
			"custom_attributes":{"status":"new"}
		}`)
	})

	ctx := context.Background()
	obj, err := client.CustomObjects.CreateOrUpdate(ctx, "Order", &CreateOrUpdateCustomObjectRequest{
		ExternalID:       "ext_456",
		CustomAttributes: map[string]string{"status": "new"},
	})
	if err != nil {
		t.Fatalf("CustomObjects.CreateOrUpdate returned error: %v", err)
	}
	if obj.ID != "obj_456" {
		t.Errorf("ID = %v, want obj_456", obj.ID)
	}
}

func TestCustomObjectsService_Delete(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order/obj_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"obj_123","object":"Order","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.CustomObjects.Delete(ctx, "Order", "obj_123")
	if err != nil {
		t.Fatalf("CustomObjects.Delete returned error: %v", err)
	}
	if !result.Deleted {
		t.Error("Deleted = false, want true")
	}
	if result.ID != "obj_123" {
		t.Errorf("ID = %v, want obj_123", result.ID)
	}
}

func TestCustomObjectsService_DeleteByExternalID(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		if r.URL.Query().Get("external_id") != "ext_123" {
			t.Errorf("external_id = %v, want ext_123", r.URL.Query().Get("external_id"))
		}
		fmt.Fprint(w, `{"id":"obj_123","object":"Order","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.CustomObjects.DeleteByExternalID(ctx, "Order", "ext_123")
	if err != nil {
		t.Fatalf("CustomObjects.DeleteByExternalID returned error: %v", err)
	}
	if !result.Deleted {
		t.Error("Deleted = false, want true")
	}
}

// --- Raw companion method tests ---

func TestCustomObjectsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order/obj_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-co-get")
		fmt.Fprint(w, `{"id":"obj_123","external_id":"ext_123","type":"Order","custom_attributes":{"status":"shipped"}}`)
	})

	ctx := context.Background()
	result, err := client.CustomObjects.GetRaw(ctx, "Order", "obj_123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	data, err := ParseCustomObjectGetResult(result)
	if err != nil {
		t.Fatalf("ParseCustomObjectGetResult returned error: %v", err)
	}
	if data.ID != "obj_123" {
		t.Errorf("Data.ID = %v, want obj_123", data.ID)
	}
	if data.CustomAttributes["status"] != "shipped" {
		t.Errorf("Data.CustomAttributes[status] = %v, want shipped", data.CustomAttributes["status"])
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-co-get" {
		t.Errorf("Header X-Request-Id = %q, want req-co-get", result.Header.Get("X-Request-Id"))
	}
}

func TestCustomObjectsService_GetByExternalIDRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-co-extid")
		fmt.Fprint(w, `{"id":"obj_123","external_id":"ext_123","type":"Order"}`)
	})

	ctx := context.Background()
	result, err := client.CustomObjects.GetByExternalIDRaw(ctx, "Order", "ext_123")
	if err != nil {
		t.Fatalf("GetByExternalIDRaw returned error: %v", err)
	}
	data, err := ParseCustomObjectGetByExternalIDResult(result)
	if err != nil {
		t.Fatalf("ParseCustomObjectGetByExternalIDResult returned error: %v", err)
	}
	if data.ExternalID != "ext_123" {
		t.Errorf("Data.ExternalID = %v, want ext_123", data.ExternalID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCustomObjectsService_CreateOrUpdateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-co-create")
		fmt.Fprint(w, `{"id":"obj_456","external_id":"ext_456","type":"Order","custom_attributes":{"status":"new"}}`)
	})

	ctx := context.Background()
	result, err := client.CustomObjects.CreateOrUpdateRaw(ctx, "Order", &CreateOrUpdateCustomObjectRequest{
		ExternalID:       "ext_456",
		CustomAttributes: map[string]string{"status": "new"},
	})
	if err != nil {
		t.Fatalf("CreateOrUpdateRaw returned error: %v", err)
	}
	data, err := ParseCustomObjectCreateOrUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseCustomObjectCreateOrUpdateResult returned error: %v", err)
	}
	if data.ID != "obj_456" {
		t.Errorf("Data.ID = %v, want obj_456", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCustomObjectsService_DeleteRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order/obj_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-co-del")
		fmt.Fprint(w, `{"id":"obj_123","object":"Order","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.CustomObjects.DeleteRaw(ctx, "Order", "obj_123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	data, err := ParseCustomObjectDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseCustomObjectDeleteResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCustomObjectsService_DeleteByExternalIDRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_object_instances/Order", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("X-Request-Id", "req-co-del-ext")
		fmt.Fprint(w, `{"id":"obj_123","object":"Order","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.CustomObjects.DeleteByExternalIDRaw(ctx, "Order", "ext_123")
	if err != nil {
		t.Fatalf("DeleteByExternalIDRaw returned error: %v", err)
	}
	data, err := ParseCustomObjectDeleteByExternalIDResult(result)
	if err != nil {
		t.Fatalf("ParseCustomObjectDeleteByExternalIDResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-co-del-ext" {
		t.Errorf("Header X-Request-Id = %q, want req-co-del-ext", result.Header.Get("X-Request-Id"))
	}
}
