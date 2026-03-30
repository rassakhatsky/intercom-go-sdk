package messaging

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// EmailsService handles communication with the email setting related methods
// of the Intercom API.
type EmailsService struct {
	client api.Caller
}

// NewEmailsService creates a new EmailsService.
func NewEmailsService(c api.Caller) *EmailsService {
	return &EmailsService{client: c}
}

// EmailSetting represents an Intercom email setting.
type EmailSetting struct {
	Type                         string `json:"type"`
	ID                           string `json:"id"`
	Email                        string `json:"email,omitempty"`
	Verified                     bool   `json:"verified,omitempty"`
	Domain                       string `json:"domain,omitempty"`
	BrandID                      string `json:"brand_id,omitempty"`
	ForwardingEnabled            bool   `json:"forwarding_enabled,omitempty"`
	ForwardedEmailLastReceivedAt *int64 `json:"forwarded_email_last_received_at,omitempty"`
	CreatedAt                    int64  `json:"created_at,omitempty"`
	UpdatedAt                    int64  `json:"updated_at,omitempty"`
}

// EmailSettingList represents the response from the list emails endpoint.
type EmailSettingList struct {
	Type string         `json:"type"`
	Data []EmailSetting `json:"data"`
}

// --- Parse Functions ---

// ParseGetResult decodes a Result into an EmailSetting.
func ParseGetResult(r *api.Result) (*EmailSetting, error) {
	return api.Decode[EmailSetting](r)
}

// ParseListResult decodes a Result into an EmailSettingList.
func ParseListResult(r *api.Result) (*EmailSettingList, error) {
	return api.Decode[EmailSettingList](r)
}

// --- Regular Methods ---

// Get retrieves an email setting by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/emails/retrieveemail
func (s *EmailsService) Get(ctx context.Context, id string) (*EmailSetting, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// List returns all email settings.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/emails/listemails
func (s *EmailsService) List(ctx context.Context) (*EmailSettingList, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves an email setting by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/emails/retrieveemail
func (s *EmailsService) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("emails/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all email settings with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/emails/listemails
func (s *EmailsService) ListRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "emails", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
