package messaging

import (
	"context"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// SubscriptionsService handles communication with the subscription type
// related methods of the Intercom API.
type SubscriptionsService struct {
	client api.Caller
}

// NewSubscriptionsService creates a new SubscriptionsService.
func NewSubscriptionsService(c api.Caller) *SubscriptionsService {
	return &SubscriptionsService{client: c}
}

// SubscriptionType represents a subscription type in Intercom.
type SubscriptionType = api.SubscriptionType

// Translation represents a localised version of a subscription type.
type Translation = api.Translation

// SubscriptionTypeList represents the response from listing subscription types.
type SubscriptionTypeList struct {
	Type string             `json:"type"`
	Data []SubscriptionType `json:"data"`
}

// --- Parse Functions ---

// ParseSubscriptionListResult decodes a Result into a SubscriptionTypeList.
func ParseSubscriptionListResult(r *api.Result) (*SubscriptionTypeList, error) {
	return api.Decode[SubscriptionTypeList](r)
}

// --- Regular Methods ---

// List returns all subscription types for the workspace.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/listsubscriptiontypes
func (s *SubscriptionsService) List(ctx context.Context) (*SubscriptionTypeList, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseSubscriptionListResult(result)
}

// --- Raw Methods ---

// ListRaw returns all subscription types with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/listsubscriptiontypes
func (s *SubscriptionsService) ListRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "subscription_types", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
