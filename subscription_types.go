package intercom

import (
	"context"
	"net/http"
)

// SubscriptionTypesService handles communication with the subscription type
// related methods of the Intercom API.
type SubscriptionTypesService service

// SubscriptionType represents a subscription type in Intercom.
type SubscriptionType struct {
	Type               string        `json:"type"`
	ID                 string        `json:"id"`
	State              string        `json:"state,omitempty"`
	ConsentType        string        `json:"consent_type,omitempty"`
	DefaultTranslation *Translation  `json:"default_translation,omitempty"`
	Translations       []Translation `json:"translations,omitempty"`
	ContentTypes       []string      `json:"content_types,omitempty"`
}

// Translation represents a localised version of a subscription type.
type Translation struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Locale      string `json:"locale"`
}

// SubscriptionTypeList represents the response from listing subscription types.
type SubscriptionTypeList struct {
	Type string             `json:"type"`
	Data []SubscriptionType `json:"data"`
}

// --- Parse Functions ---

// ParseSubscriptionTypeListResult decodes a Result into a SubscriptionTypeList.
func ParseSubscriptionTypeListResult(r *Result) (*SubscriptionTypeList, error) {
	return Decode[SubscriptionTypeList](r)
}

// --- Regular Methods ---

// List returns all subscription types for the workspace.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/listsubscriptiontypes
func (s *SubscriptionTypesService) List(ctx context.Context) (*SubscriptionTypeList, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseSubscriptionTypeListResult(result)
}

// --- Raw Methods ---

// ListRaw returns all subscription types with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/listsubscriptiontypes
func (s *SubscriptionTypesService) ListRaw(ctx context.Context) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "subscription_types", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
