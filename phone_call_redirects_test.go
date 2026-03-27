package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestPhoneCallRedirectsService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/phone_call_redirects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["phone"] != "+15551234567" {
			t.Errorf("phone = %v, want +15551234567", body["phone"])
		}
		fmt.Fprint(w, `{
			"type":"phone_call_redirect",
			"phone":"+15551234567"
		}`)
	})

	ctx := context.Background()
	result, err := client.PhoneCallRedirects.Create(ctx, &CreatePhoneCallRedirectRequest{
		Phone: "+15551234567",
	})
	if err != nil {
		t.Fatalf("PhoneCallRedirects.Create returned error: %v", err)
	}
	if result.Type != "phone_call_redirect" {
		t.Errorf("Type = %v, want phone_call_redirect", result.Type)
	}
	if result.Phone != "+15551234567" {
		t.Errorf("Phone = %v, want +15551234567", result.Phone)
	}
}

func TestPhoneCallRedirectsService_Create_WithCustomAttributes(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/phone_call_redirects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		attrs := body["custom_attributes"].(map[string]any)
		if attrs["plan"] != "premium" {
			t.Errorf("custom_attributes.plan = %v, want premium", attrs["plan"])
		}
		fmt.Fprint(w, `{
			"type":"phone_call_redirect",
			"phone":"+15551234567"
		}`)
	})

	ctx := context.Background()
	result, err := client.PhoneCallRedirects.Create(ctx, &CreatePhoneCallRedirectRequest{
		Phone:            "+15551234567",
		CustomAttributes: map[string]any{"plan": "premium"},
	})
	if err != nil {
		t.Fatalf("PhoneCallRedirects.Create returned error: %v", err)
	}
	if result.Phone != "+15551234567" {
		t.Errorf("Phone = %v, want +15551234567", result.Phone)
	}
}

func TestPhoneCallRedirectsService_CreateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/phone_call_redirects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-pcr")
		fmt.Fprint(w, `{"type":"phone_call_redirect","phone":"+15551234567"}`)
	})

	ctx := context.Background()
	result, err := client.PhoneCallRedirects.CreateRaw(ctx, &CreatePhoneCallRedirectRequest{
		Phone: "+15551234567",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-pcr" {
		t.Errorf("Header X-Request-Id = %q, want req-pcr", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	pcr, err := ParsePhoneCallRedirectCreateResult(result)
	if err != nil {
		t.Fatalf("ParsePhoneCallRedirectCreateResult returned error: %v", err)
	}
	if pcr.Type != "phone_call_redirect" {
		t.Errorf("Type = %v, want phone_call_redirect", pcr.Type)
	}
	if pcr.Phone != "+15551234567" {
		t.Errorf("Phone = %v, want +15551234567", pcr.Phone)
	}
}
