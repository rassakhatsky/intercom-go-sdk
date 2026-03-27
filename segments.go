package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// SegmentsService handles communication with the segment related methods
// of the Intercom API.
type SegmentsService service

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

// SegmentList represents the response from the list segments endpoint.
type SegmentList struct {
	Type     string    `json:"type"`
	Segments []Segment `json:"segments"`
}

// SegmentListOptions specifies optional parameters to the List method.
type SegmentListOptions struct {
	IncludeCount *bool `url:"include_count,omitempty"`
}

// --- Parse Functions ---

// ParseSegmentGetResult decodes a Result into a Segment.
func ParseSegmentGetResult(r *Result) (*Segment, error) {
	return Decode[Segment](r)
}

// ParseSegmentListResult decodes a Result into a SegmentList.
func ParseSegmentListResult(r *Result) (*SegmentList, error) {
	return Decode[SegmentList](r)
}

// --- Regular Methods ---

// Get retrieves a segment by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/segments/retrievesegment
func (s *SegmentsService) Get(ctx context.Context, id string) (*Segment, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseSegmentGetResult(result)
}

// List returns all segments.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/segments/listsegments
func (s *SegmentsService) List(ctx context.Context, opts *SegmentListOptions) (*SegmentList, error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseSegmentListResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a segment by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/segments/retrievesegment
func (s *SegmentsService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("segments/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all segments with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/segments/listsegments
func (s *SegmentsService) ListRaw(ctx context.Context, opts *SegmentListOptions) (*Result, error) {
	path, err := addQueryOptions("segments", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
