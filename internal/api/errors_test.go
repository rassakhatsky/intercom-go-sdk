package api

import (
	"errors"
	"net/http"
	"testing"
)

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

func TestErrorResponse_ErrorFormat_RawBody(t *testing.T) {
	err := &ErrorResponse{
		Response: &http.Response{StatusCode: 502},
		RawBody:  "Bad Gateway",
	}

	got := err.Error()
	want := "HTTP 502: Bad Gateway"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestErrorResponse_ErrorFormat_NoResponse(t *testing.T) {
	err := &ErrorResponse{}

	got := err.Error()
	want := "unknown API error"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
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

	err := ResultError(r)
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

	err := ResultError(r)
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

	err := ResultError(r)
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

	err := ResultError(r)
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

	err := ResultError(r)
	if err != nil {
		t.Errorf("expected nil error for result without Error, got %v", err)
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

func TestIsRateLimited_FalseForOther(t *testing.T) {
	if IsRateLimited(errors.New("random error")) {
		t.Error("IsRateLimited() = true for non-ErrorResponse")
	}
}

func TestIsUnauthorized_FalseForOther(t *testing.T) {
	if IsUnauthorized(errors.New("random error")) {
		t.Error("IsUnauthorized() = true for non-ErrorResponse")
	}
}
