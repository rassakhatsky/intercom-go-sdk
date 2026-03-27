package intercom

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorResponse represents an error response from the Intercom API.
type ErrorResponse struct {
	Response  *http.Response `json:"-"`
	Type      string         `json:"type"`
	RequestID string         `json:"request_id,omitempty"`
	Errors    []ErrorDetail  `json:"errors"`
	RawBody   string         `json:"-"`
}

// ErrorDetail represents a single error within an ErrorResponse.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
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

// resultError converts Result.Error into an *ErrorResponse, preserving
// compatibility with IsNotFound, IsRateLimited, and IsUnauthorized.
func resultError(r *Result) error {
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
	return errResp
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

func hasStatusCode(err error, code int) bool {
	var errResp *ErrorResponse
	if errors.As(err, &errResp) && errResp.Response != nil {
		return errResp.Response.StatusCode == code
	}
	return false
}
