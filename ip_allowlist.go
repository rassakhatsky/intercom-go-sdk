package intercom

import (
	"context"
	"net/http"
)

// IPAllowlistService handles communication with the IP allowlist related methods
// of the Intercom API.
type IPAllowlistService service

// IPAllowlistSettings represents the Intercom IP allowlist configuration.
type IPAllowlistSettings struct {
	Type        string   `json:"type"`
	Enabled     bool     `json:"enabled"`
	IPAllowlist []string `json:"ip_allowlist"`
}

// UpdateIPAllowlistRequest represents a request to update the IP allowlist.
type UpdateIPAllowlistRequest struct {
	Type        string   `json:"type,omitempty"`
	Enabled     bool     `json:"enabled"`
	IPAllowlist []string `json:"ip_allowlist"`
}

// --- Parse Functions ---

// ParseIPAllowlistGetResult decodes a Result into IPAllowlistSettings.
func ParseIPAllowlistGetResult(r *Result) (*IPAllowlistSettings, error) {
	return Decode[IPAllowlistSettings](r)
}

// ParseIPAllowlistUpdateResult decodes a Result into IPAllowlistSettings.
func ParseIPAllowlistUpdateResult(r *Result) (*IPAllowlistSettings, error) {
	return Decode[IPAllowlistSettings](r)
}

// --- Regular Methods ---

// Get retrieves the current IP allowlist settings.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ip-allowlist/getipallowlist
func (s *IPAllowlistService) Get(ctx context.Context) (*IPAllowlistSettings, error) {
	result, err := s.GetRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseIPAllowlistGetResult(result)
}

// Update updates the IP allowlist settings.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ip-allowlist/updateipallowlist
func (s *IPAllowlistService) Update(ctx context.Context, body *UpdateIPAllowlistRequest) (*IPAllowlistSettings, error) {
	result, err := s.UpdateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseIPAllowlistUpdateResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves the current IP allowlist settings with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ip-allowlist/getipallowlist
func (s *IPAllowlistService) GetRaw(ctx context.Context) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "ip_allowlist", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates the IP allowlist settings with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ip-allowlist/updateipallowlist
func (s *IPAllowlistService) UpdateRaw(ctx context.Context, body *UpdateIPAllowlistRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, "ip_allowlist", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
