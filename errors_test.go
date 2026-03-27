package intercom

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestCheckResponse_ParsesErrorList(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/999", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{
			"type": "error.list",
			"request_id": "req-123",
			"errors": [
				{"code": "not_found", "message": "User not found"}
			]
		}`))
	})

	req, _ := client.NewRequest(http.MethodGet, "contacts/999", nil)
	_, err := client.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}

	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}

	if errResp.Type != "error.list" {
		t.Errorf("Type = %q, want %q", errResp.Type, "error.list")
	}
	if errResp.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want %q", errResp.RequestID, "req-123")
	}
	if len(errResp.Errors) != 1 {
		t.Fatalf("len(Errors) = %d, want 1", len(errResp.Errors))
	}
	if errResp.Errors[0].Code != "not_found" {
		t.Errorf("Errors[0].Code = %q, want %q", errResp.Errors[0].Code, "not_found")
	}
	if errResp.Errors[0].Message != "User not found" {
		t.Errorf("Errors[0].Message = %q, want %q", errResp.Errors[0].Message, "User not found")
	}
}

func TestErrorResponse_ErrorFormat(t *testing.T) {
	err := &ErrorResponse{
		Response: &http.Response{StatusCode: 422},
		Errors: []ErrorDetail{
			{Code: "parameter_invalid", Message: "email is required"},
		},
	}

	got := err.Error()
	want := "parameter_invalid: email is required"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestErrorResponse_ErrorFormat_NoErrors(t *testing.T) {
	err := &ErrorResponse{
		Response: &http.Response{StatusCode: 500},
		Errors:   nil,
	}

	got := err.Error()
	want := "HTTP 500"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestIsNotFound_True(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/999", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"not_found","message":"not found"}]}`))
	})

	req, _ := client.NewRequest(http.MethodGet, "contacts/999", nil)
	_, err := client.Do(context.Background(), req, nil)

	if !IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true for 404")
	}
}

func TestIsNotFound_FalseForOther(t *testing.T) {
	if IsNotFound(errors.New("random error")) {
		t.Error("IsNotFound() = true for non-ErrorResponse")
	}
	if IsNotFound(nil) {
		t.Error("IsNotFound() = true for nil")
	}
}

func TestIsRateLimited_True(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"rate_limit_exceeded","message":"rate limit exceeded"}]}`))
	})

	req, _ := client.NewRequest(http.MethodGet, "contacts", nil)
	_, err := client.Do(context.Background(), req, nil)

	if !IsRateLimited(err) {
		t.Errorf("IsRateLimited() = false, want true for 429")
	}
}

func TestIsRateLimited_FalseForOther(t *testing.T) {
	if IsRateLimited(errors.New("random error")) {
		t.Error("IsRateLimited() = true for non-ErrorResponse")
	}
}

func TestIsUnauthorized_True(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"unauthorized","message":"invalid token"}]}`))
	})

	req, _ := client.NewRequest(http.MethodGet, "contacts", nil)
	_, err := client.Do(context.Background(), req, nil)

	if !IsUnauthorized(err) {
		t.Errorf("IsUnauthorized() = false, want true for 401")
	}
}

func TestIsUnauthorized_FalseForOther(t *testing.T) {
	if IsUnauthorized(errors.New("random error")) {
		t.Error("IsUnauthorized() = true for non-ErrorResponse")
	}
}

func TestCheckResponse_NonJSONBody(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/bad", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	})

	req, _ := client.NewRequest(http.MethodGet, "bad", nil)
	_, err := client.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("expected error for 502 response")
	}

	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	// Should include raw body since JSON parsing failed on non-JSON content
	if got := errResp.Error(); got != "HTTP 502: Bad Gateway" {
		t.Errorf("Error() = %q, want %q", got, "HTTP 502: Bad Gateway")
	}
}

func TestCheckResponse_EmptyBody(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	req, _ := client.NewRequest(http.MethodGet, "empty", nil)
	_, err := client.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if got := errResp.Error(); got != "HTTP 500" {
		t.Errorf("Error() = %q, want %q", got, "HTTP 500")
	}
}

func TestCheckResponse_2xxSuccess(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	req, _ := client.NewRequest(http.MethodGet, "ok", nil)
	_, err := client.Do(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("expected no error for 200 response, got %v", err)
	}
}

func TestResultError_404(t *testing.T) {
	r := &Result{
		StatusCode: http.StatusNotFound,
		Error: &ErrorResult{
			Type:      "error.list",
			RequestID: "req-456",
			Code:      "not_found",
			Message:   "Contact not found",
			Errors:    []ErrorDetail{{Code: "not_found", Message: "Contact not found"}},
		},
	}

	err := resultError(r)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.Type != "error.list" {
		t.Errorf("Type = %q, want %q", errResp.Type, "error.list")
	}
	if errResp.RequestID != "req-456" {
		t.Errorf("RequestID = %q, want %q", errResp.RequestID, "req-456")
	}
	if len(errResp.Errors) != 1 || errResp.Errors[0].Code != "not_found" {
		t.Errorf("Errors = %v, want [{not_found Contact not found}]", errResp.Errors)
	}
	if !IsNotFound(err) {
		t.Error("IsNotFound() = false, want true")
	}
}

func TestResultError_429(t *testing.T) {
	r := &Result{
		StatusCode: http.StatusTooManyRequests,
		Error: &ErrorResult{
			Type:   "error.list",
			Code:   "rate_limit_exceeded",
			Errors: []ErrorDetail{{Code: "rate_limit_exceeded", Message: "rate limit exceeded"}},
		},
	}

	err := resultError(r)
	if !IsRateLimited(err) {
		t.Error("IsRateLimited() = false, want true")
	}
}

func TestResultError_401(t *testing.T) {
	r := &Result{
		StatusCode: http.StatusUnauthorized,
		Error: &ErrorResult{
			Type:   "error.list",
			Code:   "unauthorized",
			Errors: []ErrorDetail{{Code: "unauthorized", Message: "invalid token"}},
		},
	}

	err := resultError(r)
	if !IsUnauthorized(err) {
		t.Error("IsUnauthorized() = false, want true")
	}
}

func TestResultError_500_NonJSON(t *testing.T) {
	r := &Result{
		StatusCode: http.StatusInternalServerError,
		Error: &ErrorResult{
			Message: "Internal Server Error",
		},
	}

	err := resultError(r)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.RawBody != "Internal Server Error" {
		t.Errorf("RawBody = %q, want %q", errResp.RawBody, "Internal Server Error")
	}
	if got := errResp.Error(); got != "HTTP 500: Internal Server Error" {
		t.Errorf("Error() = %q, want %q", got, "HTTP 500: Internal Server Error")
	}
}

func TestResultError_NilError(t *testing.T) {
	r := &Result{StatusCode: http.StatusOK}

	err := resultError(r)
	if err != nil {
		t.Errorf("expected nil error for result without Error, got %v", err)
	}
}
