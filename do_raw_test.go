package intercom

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// Tests for Client.DoRaw (non-generic method)

func TestClientDoRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-abc")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"1","name":"Alice"}`))
	})

	req, err := client.NewRequest(http.MethodGet, "contacts/1", nil)
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}

	result, err := client.DoRaw(context.Background(), req)
	if err != nil {
		t.Fatalf("DoRaw returned Go error: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if got := result.Header.Get("X-Request-Id"); got != "req-abc" {
		t.Errorf("Header X-Request-Id = %q, want %q", got, "req-abc")
	}
	if result.URL == "" {
		t.Error("URL is empty, want non-empty")
	}
	if string(result.Body) != `{"id":"1","name":"Alice"}` {
		t.Errorf("Body = %q, want %q", string(result.Body), `{"id":"1","name":"Alice"}`)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
}

func TestClientDoRaw_APIError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/999", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"type":"error.list","request_id":"req-err","errors":[{"code":"not_found","message":"Resource not found"}]}`))
	})

	req, err := client.NewRequest(http.MethodGet, "contacts/999", nil)
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}

	result, err := client.DoRaw(context.Background(), req)
	if err != nil {
		t.Fatalf("DoRaw returned Go error: %v, want nil (API errors should not be Go errors)", err)
	}

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusNotFound)
	}
	if result.Error == nil {
		t.Fatal("Error is nil, want non-nil")
	}
	if result.Error.Code != "not_found" {
		t.Errorf("Error.Code = %q, want %q", result.Error.Code, "not_found")
	}
	if result.Error.Message != "Resource not found" {
		t.Errorf("Error.Message = %q, want %q", result.Error.Message, "Resource not found")
	}
	if len(result.Body) == 0 {
		t.Error("Body is empty, want raw error JSON body")
	}
}

func TestClientDoRaw_5xxError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal server error`))
	})

	req, err := client.NewRequest(http.MethodGet, "contacts", nil)
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}

	result, err := client.DoRaw(context.Background(), req)
	if err != nil {
		t.Fatalf("DoRaw returned Go error: %v", err)
	}

	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusInternalServerError)
	}
	if result.Error == nil {
		t.Fatal("Error is nil, want non-nil")
	}
	if !strings.Contains(result.Error.Message, "internal server error") {
		t.Errorf("Error.Message = %q, want to contain 'internal server error'", result.Error.Message)
	}
}

func TestClientDoRaw_TransportError(t *testing.T) {
	client := NewClient("test-token", WithBaseURL("http://127.0.0.1:1/"))

	req, err := client.NewRequest(http.MethodGet, "contacts/1", nil)
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}

	result, err := client.DoRaw(context.Background(), req)
	if err == nil {
		t.Fatal("DoRaw returned nil error, want transport error")
	}
	if result != nil {
		t.Errorf("result = %v, want nil on transport error", result)
	}
}

func TestClientDoRaw_NoContent(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req, err := client.NewRequest(http.MethodDelete, "contacts/1", nil)
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}

	result, err := client.DoRaw(context.Background(), req)
	if err != nil {
		t.Fatalf("DoRaw returned Go error: %v", err)
	}

	if result.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusNoContent)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
	if len(result.Body) != 0 {
		t.Errorf("Body = %q, want empty", string(result.Body))
	}
}

// Tests for Do delegating to DoRaw

func TestDo_DelegatesToDoRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"1","name":"Alice"}`))
	})

	type contact struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	req, _ := client.NewRequest(http.MethodGet, "contacts/1", nil)
	var got contact
	resp, err := client.Do(context.Background(), req, &got)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if got.ID != "1" || got.Name != "Alice" {
		t.Errorf("Do decoded = %+v, want {ID:1, Name:Alice}", got)
	}
	if resp.Result == nil {
		t.Fatal("Response.Result is nil")
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Response.StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
