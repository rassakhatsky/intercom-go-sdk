package intercom

import (
	"context"
	"net/http"
)

// PhoneCallRedirectsService handles communication with the phone call redirect
// (phone switch) related methods of the Intercom API.
type PhoneCallRedirectsService service

// PhoneCallRedirect represents an Intercom phone call redirect (phone switch).
type PhoneCallRedirect struct {
	Type  string `json:"type"`
	Phone string `json:"phone"`
}

// CreatePhoneCallRedirectRequest represents a request to create a phone call redirect.
type CreatePhoneCallRedirectRequest struct {
	Phone            string         `json:"phone"`
	CustomAttributes map[string]any `json:"custom_attributes,omitempty"`
}

// --- Parse Functions ---

// ParsePhoneCallRedirectCreateResult decodes a Result into a PhoneCallRedirect.
func ParsePhoneCallRedirectCreateResult(r *Result) (*PhoneCallRedirect, error) {
	return Decode[PhoneCallRedirect](r)
}

// --- Regular Methods ---

// Create creates a new phone call redirect (sends SMS to initiate a phone switch).
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/switch/createphoneswitch
func (s *PhoneCallRedirectsService) Create(ctx context.Context, body *CreatePhoneCallRedirectRequest) (*PhoneCallRedirect, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParsePhoneCallRedirectCreateResult(result)
}

// --- Raw Methods ---

// CreateRaw creates a new phone call redirect with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/switch/createphoneswitch
func (s *PhoneCallRedirectsService) CreateRaw(ctx context.Context, body *CreatePhoneCallRedirectRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "phone_call_redirects", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
