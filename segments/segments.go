package segments

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// Service handles communication with the segment related methods
// of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new segments Service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Segment represents an Intercom segment.
type Segment struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	UpdatedAt  int64  `json:"updated_at,omitempty"`
	PersonType string `json:"person_type,omitempty"`
	Count      int    `json:"count,omitempty"`
}

// List represents the response from the list segments endpoint.
type List struct {
	Type     string    `json:"type"`
	Segments []Segment `json:"segments"`
}

// ListOptions specifies optional parameters to the List method.
type ListOptions struct {
	IncludeCount *bool `url:"include_count,omitempty"`
}

// --- Parse Functions ---

// ParseGetResult decodes a Result into a Segment.
func ParseGetResult(r *api.Result) (*Segment, error) {
	return api.Decode[Segment](r)
}

// ParseListResult decodes a Result into a List.
func ParseListResult(r *api.Result) (*List, error) {
	return api.Decode[List](r)
}

// --- Regular Methods ---

// Get retrieves a segment by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/segments/retrievesegment
func (s *Service) Get(ctx context.Context, id string) (*Segment, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// List returns all segments.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/segments/listsegments
func (s *Service) List(ctx context.Context, opts *ListOptions) (*List, error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a segment by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/segments/retrievesegment
func (s *Service) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("segments/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all segments with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/segments/listsegments
func (s *Service) ListRaw(ctx context.Context, opts *ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("segments", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
