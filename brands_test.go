package intercom

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestBrandsService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands/brand_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"brand",
			"id":"brand_123",
			"name":"My Brand",
			"is_default":true,
			"created_at":1700000000,
			"updated_at":1700000100,
			"help_center_id":"hc_1",
			"default_address_settings_id":"das_1"
		}`)
	})

	ctx := context.Background()
	brand, err := client.Brands.Get(ctx, "brand_123")
	if err != nil {
		t.Fatalf("Brands.Get returned error: %v", err)
	}
	if brand.ID != "brand_123" {
		t.Errorf("Brand.ID = %v, want brand_123", brand.ID)
	}
	if brand.Name != "My Brand" {
		t.Errorf("Brand.Name = %v, want My Brand", brand.Name)
	}
	if !brand.IsDefault {
		t.Error("Brand.IsDefault = false, want true")
	}
	if brand.HelpCenterID != "hc_1" {
		t.Errorf("Brand.HelpCenterID = %v, want hc_1", brand.HelpCenterID)
	}
}

func TestBrandsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Brand not found"}]}`)
	})

	ctx := context.Background()
	_, err := client.Brands.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestBrandsService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"brand","id":"brand_1","name":"Default Brand","is_default":true},
				{"type":"brand","id":"brand_2","name":"Secondary","is_default":false}
			]
		}`)
	})

	ctx := context.Background()
	result, err := client.Brands.List(ctx)
	if err != nil {
		t.Fatalf("Brands.List returned error: %v", err)
	}
	if result.Type != "list" {
		t.Errorf("Type = %v, want list", result.Type)
	}
	if len(result.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(result.Data))
	}
	if result.Data[0].Name != "Default Brand" {
		t.Errorf("Data[0].Name = %v, want Default Brand", result.Data[0].Name)
	}
	if result.Data[1].IsDefault {
		t.Error("Data[1].IsDefault = true, want false")
	}
}

// --- Raw companion method tests ---

func TestBrandsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands/brand_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-brand-get")
		fmt.Fprint(w, `{"type":"brand","id":"brand_123","name":"My Brand","is_default":true}`)
	})

	ctx := context.Background()
	result, err := client.Brands.GetRaw(ctx, "brand_123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-brand-get" {
		t.Errorf("Header X-Request-Id = %q, want req-brand-get", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	brand, err := ParseBrandGetResult(result)
	if err != nil {
		t.Fatalf("ParseBrandGetResult returned error: %v", err)
	}
	if brand.ID != "brand_123" {
		t.Errorf("Brand.ID = %v, want brand_123", brand.ID)
	}
	if brand.Name != "My Brand" {
		t.Errorf("Brand.Name = %v, want My Brand", brand.Name)
	}
}

func TestBrandsService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Brand not found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Brands.GetRaw(ctx, "nonexistent")
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

func TestBrandsService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-brand-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"brand","id":"brand_1","name":"Default Brand"},{"type":"brand","id":"brand_2","name":"Secondary"}]}`)
	})

	ctx := context.Background()
	result, err := client.Brands.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-brand-list" {
		t.Errorf("Header X-Request-Id = %q, want req-brand-list", result.Header.Get("X-Request-Id"))
	}
	brands, err := ParseBrandListResult(result)
	if err != nil {
		t.Fatalf("ParseBrandListResult returned error: %v", err)
	}
	if len(brands.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(brands.Data))
	}
	if brands.Data[0].Name != "Default Brand" {
		t.Errorf("Data[0].Name = %v, want Default Brand", brands.Data[0].Name)
	}
}
