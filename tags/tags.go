package tags

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// Service handles communication with the tag related methods
// of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new tags Service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Tag represents an Intercom tag.
type Tag = api.Tag

// List represents a list of tags.
type List = api.TagList

// CreateOrUpdateRequest represents the request body for creating or updating a tag.
type CreateOrUpdateRequest struct {
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`
}

// TagCompanyItem identifies a company to tag.
type TagCompanyItem struct {
	ID        string `json:"id,omitempty"`
	CompanyID string `json:"company_id,omitempty"`
}

// TagCompanyRequest represents the request body for tagging companies.
type TagCompanyRequest struct {
	Name      string           `json:"name"`
	Companies []TagCompanyItem `json:"companies"`
}

// UntagCompanyItem identifies a company to untag.
type UntagCompanyItem struct {
	ID        string `json:"id,omitempty"`
	CompanyID string `json:"company_id,omitempty"`
	Untag     bool   `json:"untag"`
}

// UntagCompanyRequest represents the request body for untagging companies.
type UntagCompanyRequest struct {
	Name      string             `json:"name"`
	Companies []UntagCompanyItem `json:"companies"`
}

// --- Parse Functions ---

// ParseGetResult decodes a Result into a Tag.
func ParseGetResult(r *api.Result) (*Tag, error) {
	return api.Decode[Tag](r)
}

// ParseListResult decodes a Result into a List.
func ParseListResult(r *api.Result) (*List, error) {
	return api.Decode[List](r)
}

// ParseCreateOrUpdateResult decodes a Result into a Tag.
func ParseCreateOrUpdateResult(r *api.Result) (*Tag, error) {
	return api.Decode[Tag](r)
}

// ParseDeleteResult decodes a Result into an Empty.
func ParseDeleteResult(r *api.Result) (*api.Empty, error) {
	return api.Decode[api.Empty](r)
}

// ParseTagCompanyResult decodes a Result into a Tag.
func ParseTagCompanyResult(r *api.Result) (*Tag, error) {
	return api.Decode[Tag](r)
}

// ParseUntagCompanyResult decodes a Result into a Tag.
func ParseUntagCompanyResult(r *api.Result) (*Tag, error) {
	return api.Decode[Tag](r)
}

// --- Regular Methods ---

// Get retrieves a tag by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/findtag
func (s *Service) Get(ctx context.Context, id string) (*Tag, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// List returns all tags in the workspace.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/listtags
func (s *Service) List(ctx context.Context) (*List, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListResult(result)
}

// CreateOrUpdate creates a new tag or updates an existing one.
// To update, include the ID field in the request.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/createtag
func (s *Service) CreateOrUpdate(ctx context.Context, body *CreateOrUpdateRequest) (*Tag, error) {
	result, err := s.CreateOrUpdateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateOrUpdateResult(result)
}

// Delete deletes a tag by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/deletetag
func (s *Service) Delete(ctx context.Context, id string) error {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return err
	}
	if result.Error != nil {
		return api.ResultError(result)
	}
	return nil
}

// TagCompany tags one or more companies with a tag.
// The tag will be created if it doesn't already exist.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/createtag
func (s *Service) TagCompany(ctx context.Context, body *TagCompanyRequest) (*Tag, error) {
	result, err := s.TagCompanyRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseTagCompanyResult(result)
}

// UntagCompany removes a tag from one or more companies.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/createtag
func (s *Service) UntagCompany(ctx context.Context, body *UntagCompanyRequest) (*Tag, error) {
	result, err := s.UntagCompanyRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseUntagCompanyResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a tag by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/findtag
func (s *Service) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("tags/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all tags with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/listtags
func (s *Service) ListRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "tags", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateOrUpdateRaw creates or updates a tag with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/createtag
func (s *Service) CreateOrUpdateRaw(ctx context.Context, body *CreateOrUpdateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tags", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes a tag by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/deletetag
func (s *Service) DeleteRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("tags/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// TagCompanyRaw tags companies with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/createtag
func (s *Service) TagCompanyRaw(ctx context.Context, body *TagCompanyRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tags", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UntagCompanyRaw untags companies with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/createtag
func (s *Service) UntagCompanyRaw(ctx context.Context, body *UntagCompanyRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tags", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
