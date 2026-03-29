package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ErrorCode represents an Intercom API error code.
type ErrorCode string

const (
	ErrServerError       ErrorCode = "server_error"
	ErrClientError       ErrorCode = "client_error"
	ErrTypeMismatch      ErrorCode = "type_mismatch"
	ErrParameterNotFound ErrorCode = "parameter_not_found"
	ErrParameterInvalid  ErrorCode = "parameter_invalid"
	ErrActionForbidden   ErrorCode = "action_forbidden"
	ErrConflict          ErrorCode = "conflict"
	ErrAPIPlanRestricted ErrorCode = "api_plan_restricted"
	ErrRateLimitExceeded ErrorCode = "rate_limit_exceeded"
	ErrUnsupported       ErrorCode = "unsupported"
	ErrTokenRevoked      ErrorCode = "token_revoked"
	ErrTokenBlocked      ErrorCode = "token_blocked"
	ErrTokenNotFound     ErrorCode = "token_not_found"
	ErrTokenUnauthorized ErrorCode = "token_unauthorized"
	ErrTokenExpired      ErrorCode = "token_expired"
	ErrMissingAuth       ErrorCode = "missing_authorization"
	ErrRetryAfter        ErrorCode = "retry_after"
	ErrJobClosed         ErrorCode = "job_closed"
	ErrNotRestorable     ErrorCode = "not_restorable"
	ErrTeamNotFound      ErrorCode = "team_not_found"
	ErrTeamUnavailable   ErrorCode = "team_unavailable"
	ErrAdminNotFound     ErrorCode = "admin_not_found"
)

// RateLimitInfo contains rate limit metadata parsed from HTTP response headers.
type RateLimitInfo struct {
	Limit      int           // X-RateLimit-Limit: total requests allowed per window
	Remaining  int           // X-RateLimit-Remaining: requests remaining in current window
	Reset      time.Time     // X-RateLimit-Reset: when the rate limit window resets (Unix timestamp)
	RetryAfter time.Duration // Retry-After: how long to wait before retrying
}

// ErrorResponse represents an error response from the Intercom API.
type ErrorResponse struct {
	Response  *http.Response `json:"-"`
	Type      string         `json:"type"`
	RequestID string         `json:"request_id,omitempty"`
	Errors    []ErrorDetail  `json:"errors"`
	RawBody   string         `json:"-"`
	RateLimit *RateLimitInfo `json:"-"`
}

// Error returns a human-readable description of the API error.
func (e *ErrorResponse) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("%s: %s", e.Errors[0].Code, e.Errors[0].Message)
	}
	if e.Response != nil {
		if e.RawBody != "" {
			return fmt.Sprintf("HTTP %d: %s", e.Response.StatusCode, e.RawBody)
		}
		return fmt.Sprintf("HTTP %d", e.Response.StatusCode)
	}
	return "unknown API error"
}

// ResultError converts Result.Error into an *ErrorResponse, preserving
// compatibility with IsNotFound, IsRateLimited, and IsUnauthorized.
func ResultError(r *Result) error {
	if r.Error == nil {
		return nil
	}
	errResp := &ErrorResponse{
		Response:  &http.Response{StatusCode: r.StatusCode, Header: r.Header},
		Type:      r.Error.Type,
		RequestID: r.Error.RequestID,
		Errors:    r.Error.Errors,
	}
	if len(r.Error.Errors) == 0 && r.Error.Message != "" {
		errResp.RawBody = r.Error.Message
	}
	if r.StatusCode == http.StatusTooManyRequests && r.Header != nil {
		errResp.RateLimit = parseRateLimitInfo(r.Header)
	}
	return errResp
}

// parseRateLimitInfo extracts rate limit metadata from HTTP response headers.
// Returns nil if none of the rate limit headers are present.
func parseRateLimitInfo(h http.Header) *RateLimitInfo {
	limit := h.Get("X-RateLimit-Limit")
	remaining := h.Get("X-RateLimit-Remaining")
	reset := h.Get("X-RateLimit-Reset")
	retryAfter := h.Get("Retry-After")

	if limit == "" && remaining == "" && reset == "" && retryAfter == "" {
		return nil
	}

	info := &RateLimitInfo{}
	if v, err := strconv.Atoi(limit); err == nil {
		info.Limit = v
	}
	if v, err := strconv.Atoi(remaining); err == nil {
		info.Remaining = v
	}
	if v, err := strconv.ParseInt(reset, 10, 64); err == nil {
		info.Reset = time.Unix(v, 0)
	}
	if v, err := strconv.Atoi(retryAfter); err == nil {
		info.RetryAfter = time.Duration(v) * time.Second
	}
	return info
}

// IsNotFound returns true if the error is an Intercom 404 response.
func IsNotFound(err error) bool {
	return hasStatusCode(err, http.StatusNotFound)
}

// IsRateLimited returns true if the error is an Intercom 429 response.
func IsRateLimited(err error) bool {
	return hasStatusCode(err, http.StatusTooManyRequests)
}

// IsUnauthorized returns true if the error is an Intercom 401 response.
func IsUnauthorized(err error) bool {
	return hasStatusCode(err, http.StatusUnauthorized)
}

// IsBadRequest returns true if the error is an Intercom 400 response.
func IsBadRequest(err error) bool {
	return hasStatusCode(err, http.StatusBadRequest)
}

// IsForbidden returns true if the error is an Intercom 403 response.
func IsForbidden(err error) bool {
	return hasStatusCode(err, http.StatusForbidden)
}

// IsConflict returns true if the error is an Intercom 409 response.
func IsConflict(err error) bool {
	return hasStatusCode(err, http.StatusConflict)
}

// IsUnprocessableEntity returns true if the error is an Intercom 422 response.
func IsUnprocessableEntity(err error) bool {
	return hasStatusCode(err, http.StatusUnprocessableEntity)
}

// IsServerError returns true if the error is an Intercom 5xx response.
func IsServerError(err error) bool {
	var errResp *ErrorResponse
	if errors.As(err, &errResp) && errResp.Response != nil {
		return errResp.Response.StatusCode >= 500 && errResp.Response.StatusCode < 600
	}
	return false
}

func hasStatusCode(err error, code int) bool {
	var errResp *ErrorResponse
	if errors.As(err, &errResp) && errResp.Response != nil {
		return errResp.Response.StatusCode == code
	}
	return false
}
