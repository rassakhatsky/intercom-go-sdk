package api

import (
	"errors"
	"net/http"
	"testing"
	"time"
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

	// With nil Header, RateLimit should be nil (no panic)
	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.RateLimit != nil {
		t.Errorf("expected nil RateLimit when Header is nil, got %+v", errResp.RateLimit)
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

func TestResultError_NilResult(t *testing.T) {
	err := ResultError(nil)
	if err != nil {
		t.Errorf("expected nil error for nil Result, got %v", err)
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

func TestErrorCodeConstants(t *testing.T) {
	tests := []struct {
		constant ErrorCode
		want     string
	}{
		{ErrServerError, "server_error"},
		{ErrClientError, "client_error"},
		{ErrTypeMismatch, "type_mismatch"},
		{ErrParameterNotFound, "parameter_not_found"},
		{ErrParameterInvalid, "parameter_invalid"},
		{ErrActionForbidden, "action_forbidden"},
		{ErrConflict, "conflict"},
		{ErrAPIPlanRestricted, "api_plan_restricted"},
		{ErrRateLimitExceeded, "rate_limit_exceeded"},
		{ErrUnsupported, "unsupported"},
		{ErrTokenRevoked, "token_revoked"},
		{ErrTokenBlocked, "token_blocked"},
		{ErrTokenNotFound, "token_not_found"},
		{ErrTokenUnauthorized, "token_unauthorized"},
		{ErrTokenExpired, "token_expired"},
		{ErrMissingAuth, "missing_authorization"},
		{ErrRetryAfter, "retry_after"},
		{ErrJobClosed, "job_closed"},
		{ErrNotRestorable, "not_restorable"},
		{ErrTeamNotFound, "team_not_found"},
		{ErrTeamUnavailable, "team_unavailable"},
		{ErrAdminNotFound, "admin_not_found"},
	}

	for _, tt := range tests {
		if string(tt.constant) != tt.want {
			t.Errorf("ErrorCode constant %q != %q", tt.constant, tt.want)
		}
	}

	// Verify we have exactly 22 constants by checking all are distinct
	seen := make(map[ErrorCode]bool)
	for _, tt := range tests {
		if seen[tt.constant] {
			t.Errorf("duplicate ErrorCode constant: %q", tt.constant)
		}
		seen[tt.constant] = true
	}
	if len(seen) != 22 {
		t.Errorf("expected 22 error code constants, got %d", len(seen))
	}
}

func TestParseRateLimitInfo_AllHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit", "100")
	h.Set("X-RateLimit-Remaining", "42")
	h.Set("X-RateLimit-Reset", "1711584000") // 2024-03-28T00:00:00Z
	h.Set("Retry-After", "30")

	info := parseRateLimitInfo(h)
	if info == nil {
		t.Fatal("expected non-nil RateLimitInfo")
	}
	if info.Limit != 100 {
		t.Errorf("Limit = %d, want 100", info.Limit)
	}
	if info.Remaining != 42 {
		t.Errorf("Remaining = %d, want 42", info.Remaining)
	}
	wantReset := time.Unix(1711584000, 0)
	if !info.Reset.Equal(wantReset) {
		t.Errorf("Reset = %v, want %v", info.Reset, wantReset)
	}
	if info.RetryAfter != 30*time.Second {
		t.Errorf("RetryAfter = %v, want %v", info.RetryAfter, 30*time.Second)
	}
}

func TestParseRateLimitInfo_PartialHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit", "200")
	// Missing: Remaining, Reset, Retry-After

	info := parseRateLimitInfo(h)
	if info == nil {
		t.Fatal("expected non-nil RateLimitInfo when at least one header is present")
	}
	if info.Limit != 200 {
		t.Errorf("Limit = %d, want 200", info.Limit)
	}
	if info.Remaining != 0 {
		t.Errorf("Remaining = %d, want 0", info.Remaining)
	}
	if !info.Reset.IsZero() {
		t.Errorf("Reset = %v, want zero", info.Reset)
	}
	if info.RetryAfter != 0 {
		t.Errorf("RetryAfter = %v, want 0", info.RetryAfter)
	}
}

func TestParseRateLimitInfo_NoHeaders(t *testing.T) {
	h := http.Header{}

	info := parseRateLimitInfo(h)
	if info != nil {
		t.Errorf("expected nil RateLimitInfo when no rate limit headers present, got %+v", info)
	}
}

func TestParseRateLimitInfo_InvalidValues(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit", "not-a-number")
	h.Set("X-RateLimit-Remaining", "abc")
	h.Set("X-RateLimit-Reset", "xyz")
	h.Set("Retry-After", "bad")

	info := parseRateLimitInfo(h)
	if info == nil {
		t.Fatal("expected non-nil RateLimitInfo when headers are present (even invalid)")
	}
	// Invalid values should result in zero values, not panics
	if info.Limit != 0 {
		t.Errorf("Limit = %d, want 0 for invalid header", info.Limit)
	}
	if info.Remaining != 0 {
		t.Errorf("Remaining = %d, want 0 for invalid header", info.Remaining)
	}
	if !info.Reset.IsZero() {
		t.Errorf("Reset should be zero for invalid header")
	}
	if info.RetryAfter != 0 {
		t.Errorf("RetryAfter should be zero for invalid header")
	}
}

func TestResultError_429_WithRateLimitInfo(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit", "100")
	h.Set("X-RateLimit-Remaining", "0")
	h.Set("X-RateLimit-Reset", "1711584000")
	h.Set("Retry-After", "15")

	r := &Result{
		StatusCode: http.StatusTooManyRequests,
		Header:     h,
		Error: &ErrorResult{
			Type:   "error.list",
			Code:   "rate_limit_exceeded",
			Errors: []ErrorDetail{{Code: "rate_limit_exceeded", Message: "rate limit exceeded"}},
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
	if errResp.RateLimit == nil {
		t.Fatal("expected non-nil RateLimit on 429 response")
	}
	if errResp.RateLimit.Limit != 100 {
		t.Errorf("RateLimit.Limit = %d, want 100", errResp.RateLimit.Limit)
	}
	if errResp.RateLimit.Remaining != 0 {
		t.Errorf("RateLimit.Remaining = %d, want 0", errResp.RateLimit.Remaining)
	}
	if errResp.RateLimit.RetryAfter != 15*time.Second {
		t.Errorf("RateLimit.RetryAfter = %v, want %v", errResp.RateLimit.RetryAfter, 15*time.Second)
	}
}

func TestResultError_NonRateLimited_NoRateLimitInfo(t *testing.T) {
	r := &Result{
		StatusCode: http.StatusNotFound,
		Error: &ErrorResult{
			Type:   "error.list",
			Code:   "not_found",
			Errors: []ErrorDetail{{Code: "not_found", Message: "not found"}},
		},
	}

	err := ResultError(r)
	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.RateLimit != nil {
		t.Errorf("expected nil RateLimit for non-429 response, got %+v", errResp.RateLimit)
	}
}

func TestIsBadRequest(t *testing.T) {
	err := &ErrorResponse{
		Response: &http.Response{StatusCode: 400},
		Errors:   []ErrorDetail{{Code: "client_error", Message: "bad request"}},
	}
	if !IsBadRequest(err) {
		t.Error("IsBadRequest() = false, want true")
	}
}

func TestIsBadRequest_FalseForOther(t *testing.T) {
	if IsBadRequest(errors.New("random error")) {
		t.Error("IsBadRequest() = true for non-ErrorResponse")
	}
	if IsBadRequest(nil) {
		t.Error("IsBadRequest() = true for nil")
	}
}

func TestIsForbidden(t *testing.T) {
	err := &ErrorResponse{
		Response: &http.Response{StatusCode: 403},
		Errors:   []ErrorDetail{{Code: "action_forbidden", Message: "forbidden"}},
	}
	if !IsForbidden(err) {
		t.Error("IsForbidden() = false, want true")
	}
}

func TestIsForbidden_FalseForOther(t *testing.T) {
	if IsForbidden(errors.New("random error")) {
		t.Error("IsForbidden() = true for non-ErrorResponse")
	}
	if IsForbidden(nil) {
		t.Error("IsForbidden() = true for nil")
	}
}

func TestIsConflict(t *testing.T) {
	err := &ErrorResponse{
		Response: &http.Response{StatusCode: 409},
		Errors:   []ErrorDetail{{Code: "conflict", Message: "conflict"}},
	}
	if !IsConflict(err) {
		t.Error("IsConflict() = false, want true")
	}
}

func TestIsConflict_FalseForOther(t *testing.T) {
	if IsConflict(errors.New("random error")) {
		t.Error("IsConflict() = true for non-ErrorResponse")
	}
	if IsConflict(nil) {
		t.Error("IsConflict() = true for nil")
	}
}

func TestIsUnprocessableEntity(t *testing.T) {
	err := &ErrorResponse{
		Response: &http.Response{StatusCode: 422},
		Errors:   []ErrorDetail{{Code: "parameter_invalid", Message: "invalid"}},
	}
	if !IsUnprocessableEntity(err) {
		t.Error("IsUnprocessableEntity() = false, want true")
	}
}

func TestIsUnprocessableEntity_FalseForOther(t *testing.T) {
	if IsUnprocessableEntity(errors.New("random error")) {
		t.Error("IsUnprocessableEntity() = true for non-ErrorResponse")
	}
	if IsUnprocessableEntity(nil) {
		t.Error("IsUnprocessableEntity() = true for nil")
	}
}

func TestIsServerError(t *testing.T) {
	codes := []int{500, 502, 503, 504, 599}
	for _, code := range codes {
		err := &ErrorResponse{
			Response: &http.Response{StatusCode: code},
			Errors:   []ErrorDetail{{Code: "server_error", Message: "server error"}},
		}
		if !IsServerError(err) {
			t.Errorf("IsServerError() = false for status %d, want true", code)
		}
	}
}

func TestIsServerError_FalseForClientErrors(t *testing.T) {
	codes := []int{400, 401, 403, 404, 422, 429, 499, 600}
	for _, code := range codes {
		err := &ErrorResponse{
			Response: &http.Response{StatusCode: code},
			Errors:   []ErrorDetail{{Code: "client_error", Message: "client error"}},
		}
		if IsServerError(err) {
			t.Errorf("IsServerError() = true for status %d, want false", code)
		}
	}
}

func TestIsServerError_FalseForOther(t *testing.T) {
	if IsServerError(errors.New("random error")) {
		t.Error("IsServerError() = true for non-ErrorResponse")
	}
	if IsServerError(nil) {
		t.Error("IsServerError() = true for nil")
	}
}

func TestErrorCode_IsStringType(t *testing.T) {
	// ErrorCode should be usable as a string
	var code ErrorCode = "test_code"
	if string(code) != "test_code" {
		t.Errorf("ErrorCode string conversion failed")
	}
}

func TestErrorDetail_Code_IsErrorCode(t *testing.T) {
	// ErrorDetail.Code should be typed as ErrorCode
	detail := ErrorDetail{Code: ErrParameterInvalid, Message: "bad param"}
	if detail.Code != ErrParameterInvalid {
		t.Errorf("ErrorDetail.Code = %q, want %q", detail.Code, ErrParameterInvalid)
	}
	// Should also accept plain string values (backward compat)
	detail2 := ErrorDetail{Code: "custom_code", Message: "custom"}
	if string(detail2.Code) != "custom_code" {
		t.Errorf("ErrorDetail.Code string assignment failed")
	}
}

func TestHasErrorCode(t *testing.T) {
	tests := []struct {
		name string
		resp *ErrorResponse
		code ErrorCode
		want bool
	}{
		{
			name: "matches first error",
			resp: &ErrorResponse{
				Response: &http.Response{StatusCode: 422},
				Errors:   []ErrorDetail{{Code: ErrParameterInvalid, Message: "bad"}},
			},
			code: ErrParameterInvalid,
			want: true,
		},
		{
			name: "matches second error",
			resp: &ErrorResponse{
				Response: &http.Response{StatusCode: 422},
				Errors: []ErrorDetail{
					{Code: ErrClientError, Message: "first"},
					{Code: ErrParameterNotFound, Message: "second"},
				},
			},
			code: ErrParameterNotFound,
			want: true,
		},
		{
			name: "no match",
			resp: &ErrorResponse{
				Response: &http.Response{StatusCode: 409},
				Errors:   []ErrorDetail{{Code: ErrConflict, Message: "conflict"}},
			},
			code: ErrParameterInvalid,
			want: false,
		},
		{
			name: "empty errors slice",
			resp: &ErrorResponse{
				Response: &http.Response{StatusCode: 500},
				Errors:   nil,
			},
			code: ErrServerError,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.HasErrorCode(tt.code); got != tt.want {
				t.Errorf("HasErrorCode(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
