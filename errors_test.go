package intercom

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/api"
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

	var errResp *api.ErrorResponse
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
	err := &api.ErrorResponse{
		StatusCode: 422,
		Errors: []api.ErrorDetail{
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
	err := &api.ErrorResponse{
		StatusCode: 500,
		Errors:     nil,
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

	if !api.IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true for 404")
	}
}

func TestIsNotFound_FalseForOther(t *testing.T) {
	if api.IsNotFound(errors.New("random error")) {
		t.Error("IsNotFound() = true for non-ErrorResponse")
	}
	if api.IsNotFound(nil) {
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

	if !api.IsRateLimited(err) {
		t.Errorf("IsRateLimited() = false, want true for 429")
	}
}

func TestIsRateLimited_FalseForOther(t *testing.T) {
	if api.IsRateLimited(errors.New("random error")) {
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

	if !api.IsUnauthorized(err) {
		t.Errorf("IsUnauthorized() = false, want true for 401")
	}
}

func TestIsUnauthorized_FalseForOther(t *testing.T) {
	if api.IsUnauthorized(errors.New("random error")) {
		t.Error("IsUnauthorized() = true for non-ErrorResponse")
	}
}

func TestIsBadRequest_True(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/bad", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"client_error","message":"bad request"}]}`))
	})

	req, _ := client.NewRequest("GET", "contacts/bad", nil)
	_, err := client.Do(context.Background(), req, nil)

	if !api.IsBadRequest(err) {
		t.Errorf("IsBadRequest() = false, want true for 400")
	}
}

func TestIsForbidden_True(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/forbidden", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"action_forbidden","message":"forbidden"}]}`))
	})

	req, _ := client.NewRequest("GET", "contacts/forbidden", nil)
	_, err := client.Do(context.Background(), req, nil)

	if !api.IsForbidden(err) {
		t.Errorf("IsForbidden() = false, want true for 403")
	}
}

func TestIsConflict_True(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/conflict", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"conflict","message":"conflict"}]}`))
	})

	req, _ := client.NewRequest("GET", "contacts/conflict", nil)
	_, err := client.Do(context.Background(), req, nil)

	if !api.IsConflict(err) {
		t.Errorf("IsConflict() = false, want true for 409")
	}
}

func TestIsUnprocessableEntity_True(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/invalid", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"parameter_invalid","message":"invalid"}]}`))
	})

	req, _ := client.NewRequest("GET", "contacts/invalid", nil)
	_, err := client.Do(context.Background(), req, nil)

	if !api.IsUnprocessableEntity(err) {
		t.Errorf("IsUnprocessableEntity() = false, want true for 422")
	}
}

func TestIsServerError_True(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"server_error","message":"internal error"}]}`))
	})

	req, _ := client.NewRequest("GET", "contacts/error", nil)
	_, err := client.Do(context.Background(), req, nil)

	if !api.IsServerError(err) {
		t.Errorf("IsServerError() = false, want true for 500")
	}
}

func TestIsRateLimited_RateLimitInfo(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/ratelimit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "1711584000")
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"rate_limit_exceeded","message":"rate limit exceeded"}]}`))
	})

	req, _ := client.NewRequest("GET", "contacts/ratelimit", nil)
	_, err := client.Do(context.Background(), req, nil)

	if !api.IsRateLimited(err) {
		t.Fatal("IsRateLimited() = false, want true")
	}

	var errResp *api.ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.RateLimit == nil {
		t.Fatal("expected non-nil RateLimit on 429 response with rate limit headers")
	}
	if errResp.RateLimit.Limit != 100 {
		t.Errorf("RateLimit.Limit = %d, want 100", errResp.RateLimit.Limit)
	}
	if errResp.RateLimit.Remaining != 0 {
		t.Errorf("RateLimit.Remaining = %d, want 0", errResp.RateLimit.Remaining)
	}
	if errResp.RateLimit.RetryAfter != 30*1e9 { // 30 seconds in nanoseconds
		t.Errorf("RateLimit.RetryAfter = %v, want 30s", errResp.RateLimit.RetryAfter)
	}
}

func TestHasErrorCode_Integration(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/hascode", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"parameter_invalid","message":"email is required"}]}`))
	})

	req, _ := client.NewRequest("GET", "contacts/hascode", nil)
	_, err := client.Do(context.Background(), req, nil)

	var errResp *api.ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if !errResp.HasErrorCode(api.ErrParameterInvalid) {
		t.Error("HasErrorCode(api.ErrParameterInvalid) = false, want true")
	}
	if errResp.HasErrorCode(api.ErrConflict) {
		t.Error("HasErrorCode(api.ErrConflict) = true, want false")
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

	var errResp *api.ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	// Non-JSON body gets a synthetic server_error code with the raw body as message
	if got := errResp.Error(); got != "server_error: Bad Gateway" {
		t.Errorf("Error() = %q, want %q", got, "server_error: Bad Gateway")
	}
	if !errResp.HasErrorCode(api.ErrServerError) {
		t.Error("HasErrorCode(ErrServerError) = false, want true for non-JSON 5xx")
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

	var errResp *api.ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	// Empty body gets a synthetic server_error code with no message
	if got := errResp.Error(); got != "server_error" {
		t.Errorf("Error() = %q, want %q", got, "server_error")
	}
	if !errResp.HasErrorCode(api.ErrServerError) {
		t.Error("HasErrorCode(ErrServerError) = false, want true for empty-body 5xx")
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
	r := &api.Result{
		StatusCode: http.StatusNotFound,
		Error: &api.ErrorResult{
			Type:      "error.list",
			RequestID: "req-456",
			Code:      "not_found",
			Message:   "Contact not found",
			Errors:    []api.ErrorDetail{{Code: "not_found", Message: "Contact not found"}},
		},
	}

	err := api.ResultError(r)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var errResp *api.ErrorResponse
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
	if !api.IsNotFound(err) {
		t.Error("IsNotFound() = false, want true")
	}
}

func TestResultError_429(t *testing.T) {
	r := &api.Result{
		StatusCode: http.StatusTooManyRequests,
		Error: &api.ErrorResult{
			Type:   "error.list",
			Code:   "rate_limit_exceeded",
			Errors: []api.ErrorDetail{{Code: "rate_limit_exceeded", Message: "rate limit exceeded"}},
		},
	}

	err := api.ResultError(r)
	if !api.IsRateLimited(err) {
		t.Error("IsRateLimited() = false, want true")
	}
}

func TestResultError_401(t *testing.T) {
	r := &api.Result{
		StatusCode: http.StatusUnauthorized,
		Error: &api.ErrorResult{
			Type:   "error.list",
			Code:   "unauthorized",
			Errors: []api.ErrorDetail{{Code: "unauthorized", Message: "invalid token"}},
		},
	}

	err := api.ResultError(r)
	if !api.IsUnauthorized(err) {
		t.Error("IsUnauthorized() = false, want true")
	}
}

func TestResultError_500_NonJSON(t *testing.T) {
	r := &api.Result{
		StatusCode: http.StatusInternalServerError,
		Error: &api.ErrorResult{
			Message: "Internal Server Error",
		},
	}

	err := api.ResultError(r)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var errResp *api.ErrorResponse
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
	r := &api.Result{StatusCode: http.StatusOK}

	err := api.ResultError(r)
	if err != nil {
		t.Errorf("expected nil error for result without Error, got %v", err)
	}
}

func TestErrorAliases(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		predicate  func(error) bool
		wantTrue   bool
	}{
		{"IsBadRequest/400", 400, api.IsBadRequest, true},
		{"IsBadRequest/404", 404, api.IsBadRequest, false},
		{"IsForbidden/403", 403, api.IsForbidden, true},
		{"IsForbidden/401", 401, api.IsForbidden, false},
		{"IsConflict/409", 409, api.IsConflict, true},
		{"IsConflict/400", 400, api.IsConflict, false},
		{"IsUnprocessableEntity/422", 422, api.IsUnprocessableEntity, true},
		{"IsUnprocessableEntity/400", 400, api.IsUnprocessableEntity, false},
		{"IsServerError/500", 500, api.IsServerError, true},
		{"IsServerError/502", 502, api.IsServerError, true},
		{"IsServerError/599", 599, api.IsServerError, true},
		{"IsServerError/400", 400, api.IsServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &api.ErrorResponse{StatusCode: tt.statusCode}
			if got := tt.predicate(err); got != tt.wantTrue {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.wantTrue)
			}
		})
	}

	// All predicates return false for nil and non-ErrorResponse errors.
	predicates := []struct {
		name string
		fn   func(error) bool
	}{
		{"IsBadRequest", api.IsBadRequest},
		{"IsForbidden", api.IsForbidden},
		{"IsConflict", api.IsConflict},
		{"IsUnprocessableEntity", api.IsUnprocessableEntity},
		{"IsServerError", api.IsServerError},
	}
	for _, p := range predicates {
		if p.fn(nil) {
			t.Errorf("%s(nil) = true, want false", p.name)
		}
		if p.fn(errors.New("random")) {
			t.Errorf("%s(non-ErrorResponse) = true, want false", p.name)
		}
	}
}
