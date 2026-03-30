package admins

import (
	"context"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// AwayStatusReasonsService handles communication with the away status reason
// related methods of the Intercom API.
type AwayStatusReasonsService struct {
	client api.Caller
}

// NewAwayStatusReasonsService creates a new AwayStatusReasonsService.
func NewAwayStatusReasonsService(c api.Caller) *AwayStatusReasonsService {
	return &AwayStatusReasonsService{client: c}
}

// AwayStatusReason represents an Intercom away status reason.
type AwayStatusReason struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Label     string `json:"label,omitempty"`
	Emoji     string `json:"emoji,omitempty"`
	Order     int    `json:"order,omitempty"`
	Deleted   bool   `json:"deleted,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

// --- Parse Functions ---

// ParseAwayStatusReasonListResult decodes a Result into a slice of AwayStatusReason.
func ParseAwayStatusReasonListResult(r *api.Result) ([]AwayStatusReason, error) {
	v, err := api.Decode[[]AwayStatusReason](r)
	if err != nil {
		return nil, err
	}
	return *v, nil
}

// --- Regular Methods ---

// List returns all away status reasons (including deleted ones).
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/away-status-reasons/listawaystatusreasons
func (s *AwayStatusReasonsService) List(ctx context.Context) ([]AwayStatusReason, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAwayStatusReasonListResult(result)
}

// --- Raw Methods ---

// ListRaw returns all away status reasons with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/away-status-reasons/listawaystatusreasons
func (s *AwayStatusReasonsService) ListRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "away_status_reasons", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
