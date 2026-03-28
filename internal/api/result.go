package api

import "net/http"

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
	Code      string        `json:"-"`
	Message   string        `json:"-"`
	Errors    []ErrorDetail `json:"errors"`
}

// ErrorDetail represents a single error within an API error response.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error returns a human-readable description of the API error.
func (e *ErrorResult) Error() string {
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	if e.Message != "" {
		return e.Message
	}
	return "unknown API error"
}

// Empty is used as the type parameter for Decode when the API returns no body.
type Empty struct{}
