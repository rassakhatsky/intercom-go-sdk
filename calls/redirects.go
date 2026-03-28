package calls

import (
	"context"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// RedirectsService handles communication with the phone call redirect
// (phone switch) related methods of the Intercom API.
type RedirectsService struct {
	client api.Caller
}

// NewRedirectsService creates a new phone call redirects service.
func NewRedirectsService(c api.Caller) *RedirectsService {
	return &RedirectsService{client: c}
}

// Redirect represents an Intercom phone call redirect (phone switch).
type Redirect struct {
	Type  string `json:"type"`
	Phone string `json:"phone"`
}

// CreateRedirectRequest represents a request to create a phone call redirect.
type CreateRedirectRequest struct {
	Phone            string         `json:"phone"`
	CustomAttributes map[string]any `json:"custom_attributes,omitempty"`
}

// --- Parse Functions ---

// ParseRedirectCreateResult decodes a Result into a Redirect.
func ParseRedirectCreateResult(r *api.Result) (*Redirect, error) {
	return api.Decode[Redirect](r)
}

// --- Regular Methods ---

// Create creates a new phone call redirect (sends SMS to initiate a phone switch).
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/switch/createphoneswitch
func (s *RedirectsService) Create(ctx context.Context, body *CreateRedirectRequest) (*Redirect, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseRedirectCreateResult(result)
}

// --- Raw Methods ---

// CreateRaw creates a new phone call redirect with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/switch/createphoneswitch
func (s *RedirectsService) CreateRaw(ctx context.Context, body *CreateRedirectRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "phone_call_redirects", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
