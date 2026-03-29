package api

import (
	"encoding/json"
	"net/http"
)

// Result contains raw HTTP response metadata, body, and any API error.
type Result struct {
	StatusCode int
	Header     http.Header
	URL        string
	Body       []byte
	Error      *ErrorResult // non-nil for 4xx/5xx
}

// ErrorResult represents a typed API error response.
type ErrorResult struct {
	Type      string        `json:"type"`
	RequestID string        `json:"request_id,omitempty"`
	Code      ErrorCode     `json:"-"`
	Message   string        `json:"-"`
	Errors    []ErrorDetail `json:"errors"`
}

// ErrorDetail represents a single error within an API error response.
type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// Error returns a human-readable description of the API error.
func (e *ErrorResult) Error() string {
	if e.Code != "" {
		return string(e.Code) + ": " + e.Message
	}
	if e.Message != "" {
		return e.Message
	}
	return "unknown API error"
}

// Empty is used as the type parameter for Decode when the API returns no body.
type Empty struct{}

// Decode unmarshals the JSON body of a Result into a value of type T.
// For 204 No Content or empty bodies, it returns a zero-value T.
func Decode[T any](r *Result) (*T, error) {
	if r.StatusCode == http.StatusNoContent || len(r.Body) == 0 {
		return new(T), nil
	}
	var data T
	if err := json.Unmarshal(r.Body, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// BuildResult constructs a Result from an HTTP response and body,
// populating Error for 4xx/5xx responses.
func BuildResult(resp *http.Response, body []byte) *Result {
	result := &Result{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		URL:        resp.Request.URL.String(),
		Body:       body,
	}

	if resp.StatusCode >= 400 {
		var errResult ErrorResult
		if len(body) > 0 {
			if err := json.Unmarshal(body, &errResult); err != nil {
				errResult.Message = string(body)
			}
		}
		if len(errResult.Errors) > 0 {
			errResult.Code = errResult.Errors[0].Code
			errResult.Message = errResult.Errors[0].Message
		} else if errResult.Message == "" && len(body) > 0 {
			errResult.Message = string(body)
		}
		result.Error = &errResult
	}

	return result
}
