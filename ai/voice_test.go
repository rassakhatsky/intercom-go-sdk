package ai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/ai"
	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

func TestVoiceService_Register(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/register", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["phone_number"] != "+15551234567" {
			t.Errorf("phone_number = %v, want +15551234567", body["phone_number"])
		}
		if body["call_id"] != "ext_call_1" {
			t.Errorf("call_id = %v, want ext_call_1", body["call_id"])
		}
		fmt.Fprint(w, `{
			"id":42,
			"app_id":100,
			"user_phone_number":"+15551234567",
			"status":"registered",
			"external_call_id":"ext_call_1",
			"intercom_call_id":null,
			"intercom_conversation_id":null,
			"call_transcript":[],
			"call_summary":"",
			"intent":[]
		}`)
	})

	ctx := context.Background()
	resp, err := svc.Register(ctx, &ai.RegisterCallRequest{
		PhoneNumber: "+15551234567",
		CallID:      "ext_call_1",
		Source:      "aws_connect",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if resp.ID != 42 {
		t.Errorf("ID = %v, want 42", resp.ID)
	}
	if resp.Status != "registered" {
		t.Errorf("Status = %v, want registered", resp.Status)
	}
	if resp.ExternalCallID != "ext_call_1" {
		t.Errorf("ExternalCallID = %v, want ext_call_1", resp.ExternalCallID)
	}
}

func TestVoiceService_Collect(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/collect/42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id":42,
			"app_id":100,
			"user_phone_number":"+15551234567",
			"status":"completed",
			"external_call_id":"ext_call_1",
			"intercom_call_id":"call_123",
			"intercom_conversation_id":"conv_456",
			"call_transcript":[{"speaker":"agent","text":"Hello"}],
			"call_summary":"Customer asked about billing",
			"intent":[{"name":"billing_inquiry"}]
		}`)
	})

	ctx := context.Background()
	resp, err := svc.Collect(ctx, 42)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("Status = %v, want completed", resp.Status)
	}
	if resp.IntercomCallID != "call_123" {
		t.Errorf("IntercomCallID = %v, want call_123", resp.IntercomCallID)
	}
	if resp.CallSummary != "Customer asked about billing" {
		t.Errorf("CallSummary = %v, want Customer asked about billing", resp.CallSummary)
	}
}

func TestVoiceService_Collect_NotFound(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/collect/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Call not found"}]}`)
	})

	ctx := context.Background()
	_, err := svc.Collect(ctx, 999)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestVoiceService_GetByExternalID(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/external_id/ext_call_1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id":42,
			"app_id":100,
			"user_phone_number":"+15551234567",
			"status":"completed",
			"external_call_id":"ext_call_1"
		}`)
	})

	ctx := context.Background()
	resp, err := svc.GetByExternalID(ctx, "ext_call_1")
	if err != nil {
		t.Fatalf("GetByExternalID returned error: %v", err)
	}
	if resp.ExternalCallID != "ext_call_1" {
		t.Errorf("ExternalCallID = %v, want ext_call_1", resp.ExternalCallID)
	}
}

func TestVoiceService_GetByConversation(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/conversation/conv_456", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
			{"id":42,"status":"completed","external_call_id":"ext_call_1"},
			{"id":43,"status":"registered","external_call_id":"ext_call_2"}
		]`)
	})

	ctx := context.Background()
	calls, err := svc.GetByConversation(ctx, "conv_456")
	if err != nil {
		t.Fatalf("GetByConversation returned error: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("length = %d, want 2", len(calls))
	}
	if calls[0].ID != 42 {
		t.Errorf("calls[0].ID = %v, want 42", calls[0].ID)
	}
}

func TestVoiceService_GetByPhoneNumber(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/phone_number/+15551234567", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id":42,
			"user_phone_number":"+15551234567",
			"status":"completed"
		}`)
	})

	ctx := context.Background()
	resp, err := svc.GetByPhoneNumber(ctx, "+15551234567")
	if err != nil {
		t.Fatalf("GetByPhoneNumber returned error: %v", err)
	}
	if resp.UserPhoneNumber != "+15551234567" {
		t.Errorf("UserPhoneNumber = %v, want +15551234567", resp.UserPhoneNumber)
	}
}

// --- Raw Methods ---

func TestVoiceService_RegisterRaw_Success(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/register", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-fv-register")
		fmt.Fprint(w, `{"id":42,"status":"registered","external_call_id":"ext_call_1","user_phone_number":"+15551234567"}`)
	})

	ctx := context.Background()
	result, err := svc.RegisterRaw(ctx, &ai.RegisterCallRequest{
		PhoneNumber: "+15551234567",
		CallID:      "ext_call_1",
	})
	if err != nil {
		t.Fatalf("RegisterRaw returned error: %v", err)
	}
	data, err := ai.ParseRegisterResult(result)
	if err != nil {
		t.Fatalf("ParseRegisterResult returned error: %v", err)
	}
	if data.ID != 42 {
		t.Errorf("Data.ID = %v, want 42", data.ID)
	}
	if data.Status != "registered" {
		t.Errorf("Data.Status = %v, want registered", data.Status)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-fv-register" {
		t.Errorf("Header X-Request-Id = %q, want req-fv-register", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestVoiceService_CollectRaw_Success(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/collect/42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-fv-collect")
		fmt.Fprint(w, `{"id":42,"status":"completed","intercom_call_id":"call_123","call_summary":"Customer asked about billing"}`)
	})

	ctx := context.Background()
	result, err := svc.CollectRaw(ctx, 42)
	if err != nil {
		t.Fatalf("CollectRaw returned error: %v", err)
	}
	data, err := ai.ParseCollectResult(result)
	if err != nil {
		t.Fatalf("ParseCollectResult returned error: %v", err)
	}
	if data.Status != "completed" {
		t.Errorf("Data.Status = %v, want completed", data.Status)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestVoiceService_CollectRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/collect/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Call not found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.CollectRaw(ctx, 999)
	if err != nil {
		t.Fatalf("CollectRaw returned Go error: %v", err)
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

func TestVoiceService_GetByExternalIDRaw_Success(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/external_id/ext_call_1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-fv-ext")
		fmt.Fprint(w, `{"id":42,"external_call_id":"ext_call_1","status":"completed"}`)
	})

	ctx := context.Background()
	result, err := svc.GetByExternalIDRaw(ctx, "ext_call_1")
	if err != nil {
		t.Fatalf("GetByExternalIDRaw returned error: %v", err)
	}
	data, err := ai.ParseGetByExternalIDResult(result)
	if err != nil {
		t.Fatalf("ParseGetByExternalIDResult returned error: %v", err)
	}
	if data.ExternalCallID != "ext_call_1" {
		t.Errorf("Data.ExternalCallID = %v, want ext_call_1", data.ExternalCallID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestVoiceService_GetByConversationRaw_Success(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/conversation/conv_456", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-fv-conv")
		fmt.Fprint(w, `[{"id":42,"status":"completed","external_call_id":"ext_call_1"},{"id":43,"status":"registered","external_call_id":"ext_call_2"}]`)
	})

	ctx := context.Background()
	result, err := svc.GetByConversationRaw(ctx, "conv_456")
	if err != nil {
		t.Fatalf("GetByConversationRaw returned error: %v", err)
	}
	data, err := ai.ParseGetByConversationResult(result)
	if err != nil {
		t.Fatalf("ParseGetByConversationResult returned error: %v", err)
	}
	calls := *data
	if len(calls) != 2 {
		t.Fatalf("Data length = %d, want 2", len(calls))
	}
	if calls[0].ID != 42 {
		t.Errorf("Data[0].ID = %v, want 42", calls[0].ID)
	}
	if calls[1].Status != "registered" {
		t.Errorf("Data[1].Status = %v, want registered", calls[1].Status)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-fv-conv" {
		t.Errorf("Header X-Request-Id = %q, want req-fv-conv", result.Header.Get("X-Request-Id"))
	}
}

func TestVoiceService_GetByPhoneNumberRaw_Success(t *testing.T) {
	svc, mux, teardown := setupVoice()
	defer teardown()

	mux.HandleFunc("/fin_voice/phone_number/+15551234567", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-fv-phone")
		fmt.Fprint(w, `{"id":42,"user_phone_number":"+15551234567","status":"completed"}`)
	})

	ctx := context.Background()
	result, err := svc.GetByPhoneNumberRaw(ctx, "+15551234567")
	if err != nil {
		t.Fatalf("GetByPhoneNumberRaw returned error: %v", err)
	}
	data, err := ai.ParseGetByPhoneNumberResult(result)
	if err != nil {
		t.Fatalf("ParseGetByPhoneNumberResult returned error: %v", err)
	}
	if data.UserPhoneNumber != "+15551234567" {
		t.Errorf("Data.UserPhoneNumber = %v, want +15551234567", data.UserPhoneNumber)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-fv-phone" {
		t.Errorf("Header X-Request-Id = %q, want req-fv-phone", result.Header.Get("X-Request-Id"))
	}
}
