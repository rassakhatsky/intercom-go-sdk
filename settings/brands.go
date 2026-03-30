package settings

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// BrandsService handles communication with the brand related methods
// of the Intercom API.
type BrandsService struct {
	client api.Caller
}

// NewBrandsService creates a new BrandsService.
func NewBrandsService(c api.Caller) *BrandsService {
	return &BrandsService{client: c}
}

// Brand represents an Intercom brand.
type Brand struct {
	Type                     string `json:"type"`
	ID                       string `json:"id"`
	Name                     string `json:"name,omitempty"`
	IsDefault                bool   `json:"is_default,omitempty"`
	CreatedAt                int64  `json:"created_at,omitempty"`
	UpdatedAt                int64  `json:"updated_at,omitempty"`
	HelpCenterID             string `json:"help_center_id,omitempty"`
	DefaultAddressSettingsID string `json:"default_address_settings_id,omitempty"`
}

// BrandList represents the response from the list brands endpoint.
type BrandList struct {
	Type string  `json:"type"`
	Data []Brand `json:"data"`
}

// --- Parse Functions ---

// ParseGetResult decodes a Result into a Brand.
func ParseGetResult(r *api.Result) (*Brand, error) {
	return api.Decode[Brand](r)
}

// ParseListResult decodes a Result into a BrandList.
func ParseListResult(r *api.Result) (*BrandList, error) {
	return api.Decode[BrandList](r)
}

// --- Regular Methods ---

// Get retrieves a brand by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/brands/retrievebrand
func (s *BrandsService) Get(ctx context.Context, id string) (*Brand, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// List returns all brands.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/brands/listbrands
func (s *BrandsService) List(ctx context.Context) (*BrandList, error) {
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

// GetRaw retrieves a brand by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/brands/retrievebrand
func (s *BrandsService) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("brands/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all brands with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/brands/listbrands
func (s *BrandsService) ListRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "brands", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
