package intercom

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestEmailsService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/emails/email_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"email_setting",
			"id":"email_123",
			"email":"support@example.com",
			"verified":true,
			"domain":"example.com",
			"brand_id":"brand_1",
			"forwarding_enabled":true,
			"created_at":1700000000,
			"updated_at":1700000100
		}`)
	})

	ctx := context.Background()
	email, err := client.Emails.Get(ctx, "email_123")
	if err != nil {
		t.Fatalf("Emails.Get returned error: %v", err)
	}
	if email.ID != "email_123" {
		t.Errorf("ID = %v, want email_123", email.ID)
	}
	if email.Email != "support@example.com" {
		t.Errorf("Email = %v, want support@example.com", email.Email)
	}
	if !email.Verified {
		t.Error("Verified = false, want true")
	}
	if email.Domain != "example.com" {
		t.Errorf("Domain = %v, want example.com", email.Domain)
	}
	if !email.ForwardingEnabled {
		t.Error("ForwardingEnabled = false, want true")
	}
}

func TestEmailsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/emails/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Email not found"}]}`)
	})

	ctx := context.Background()
	_, err := client.Emails.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestEmailsService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/emails", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"email_setting","id":"email_1","email":"support@example.com","verified":true},
				{"type":"email_setting","id":"email_2","email":"sales@example.com","verified":false}
			]
		}`)
	})

	ctx := context.Background()
	result, err := client.Emails.List(ctx)
	if err != nil {
		t.Fatalf("Emails.List returned error: %v", err)
	}
	if result.Type != "list" {
		t.Errorf("Type = %v, want list", result.Type)
	}
	if len(result.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(result.Data))
	}
	if result.Data[0].Email != "support@example.com" {
		t.Errorf("Data[0].Email = %v, want support@example.com", result.Data[0].Email)
	}
}

func TestEmailsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/emails/email_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-email")
		fmt.Fprint(w, `{"type":"email_setting","id":"email_123","email":"support@example.com","verified":true}`)
	})

	ctx := context.Background()
	result, err := client.Emails.GetRaw(ctx, "email_123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-email" {
		t.Errorf("Header X-Request-Id = %q, want req-email", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	email, err := ParseEmailGetResult(result)
	if err != nil {
		t.Fatalf("ParseEmailGetResult returned error: %v", err)
	}
	if email.ID != "email_123" {
		t.Errorf("ID = %v, want email_123", email.ID)
	}
	if email.Email != "support@example.com" {
		t.Errorf("Email = %v, want support@example.com", email.Email)
	}
}

func TestEmailsService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/emails/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Email not found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Emails.GetRaw(ctx, "nonexistent")
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

func TestEmailsService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/emails", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"email_setting","id":"email_1","email":"support@example.com"}]}`)
	})

	ctx := context.Background()
	result, err := client.Emails.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	emails, err := ParseEmailListResult(result)
	if err != nil {
		t.Fatalf("ParseEmailListResult returned error: %v", err)
	}
	if len(emails.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(emails.Data))
	}
	if emails.Data[0].Email != "support@example.com" {
		t.Errorf("Data[0].Email = %v, want support@example.com", emails.Data[0].Email)
	}
}
