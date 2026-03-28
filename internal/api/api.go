package api

import (
	"context"
	"net/http"
)

// Caller abstracts the HTTP operations needed by sub-package services.
// The root intercom.Client implements this interface.
type Caller interface {
	NewRequest(method, urlStr string, body any) (*http.Request, error)
	DoRaw(ctx context.Context, req *http.Request) (*Result, error)
	Do(ctx context.Context, req *http.Request, v any) (*Response, error)
}

// Response wraps a Result to provide additional API-specific data.
type Response struct {
	*Result
}
