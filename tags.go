package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// TagsService handles communication with the tag related methods
// of the Intercom API.
type TagsService service

// Tag represents an Intercom tag.
type Tag struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	AppliedAt *int64 `json:"applied_at,omitempty"`
	AppliedBy *Admin `json:"applied_by,omitempty"`
}

// TagList represents a list of tags.
type TagList struct {
	Type string `json:"type"`
	Data []Tag  `json:"data"`
}

// CreateOrUpdateTagRequest represents the request body for creating or updating a tag.
type CreateOrUpdateTagRequest struct {
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

// ParseTagGetResult decodes a Result into a Tag.
func ParseTagGetResult(r *Result) (*Tag, error) {
	return Decode[Tag](r)
}

// ParseTagListResult decodes a Result into a TagList.
func ParseTagListResult(r *Result) (*TagList, error) {
	return Decode[TagList](r)
}

// ParseTagCreateOrUpdateResult decodes a Result into a Tag.
func ParseTagCreateOrUpdateResult(r *Result) (*Tag, error) {
	return Decode[Tag](r)
}

// ParseTagDeleteResult decodes a Result into an Empty.
func ParseTagDeleteResult(r *Result) (*Empty, error) {
	return Decode[Empty](r)
}

// ParseTagTagCompanyResult decodes a Result into a Tag.
func ParseTagTagCompanyResult(r *Result) (*Tag, error) {
	return Decode[Tag](r)
}

// ParseTagUntagCompanyResult decodes a Result into a Tag.
func ParseTagUntagCompanyResult(r *Result) (*Tag, error) {
	return Decode[Tag](r)
}

// --- Regular Methods ---

// Get retrieves a tag by ID.
func (s *TagsService) Get(ctx context.Context, id string) (*Tag, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTagGetResult(result)
}

// List returns all tags in the workspace.
func (s *TagsService) List(ctx context.Context) (*TagList, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTagListResult(result)
}

// CreateOrUpdate creates a new tag or updates an existing one.
// To update, include the ID field in the request.
func (s *TagsService) CreateOrUpdate(ctx context.Context, body *CreateOrUpdateTagRequest) (*Tag, error) {
	result, err := s.CreateOrUpdateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTagCreateOrUpdateResult(result)
}

// Delete deletes a tag by ID.
func (s *TagsService) Delete(ctx context.Context, id string) error {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return err
	}
	if result.Error != nil {
		return resultError(result)
	}
	return nil
}

// TagCompany tags one or more companies with a tag.
// The tag will be created if it doesn't already exist.
func (s *TagsService) TagCompany(ctx context.Context, body *TagCompanyRequest) (*Tag, error) {
	result, err := s.TagCompanyRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTagTagCompanyResult(result)
}

// UntagCompany removes a tag from one or more companies.
func (s *TagsService) UntagCompany(ctx context.Context, body *UntagCompanyRequest) (*Tag, error) {
	result, err := s.UntagCompanyRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTagUntagCompanyResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a tag by ID with the full HTTP result.
func (s *TagsService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("tags/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all tags with the full HTTP result.
func (s *TagsService) ListRaw(ctx context.Context) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "tags", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateOrUpdateRaw creates or updates a tag with the full HTTP result.
func (s *TagsService) CreateOrUpdateRaw(ctx context.Context, body *CreateOrUpdateTagRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tags", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes a tag by ID and returns the full HTTP result.
func (s *TagsService) DeleteRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("tags/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// TagCompanyRaw tags companies with the full HTTP result.
func (s *TagsService) TagCompanyRaw(ctx context.Context, body *TagCompanyRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tags", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UntagCompanyRaw untags companies with the full HTTP result.
func (s *TagsService) UntagCompanyRaw(ctx context.Context, body *UntagCompanyRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tags", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
