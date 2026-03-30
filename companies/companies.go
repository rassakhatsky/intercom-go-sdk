package companies

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// Service handles communication with the company related methods
// of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new companies Service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Company represents an Intercom company.
type Company struct {
	Type             string         `json:"type"`
	ID               string         `json:"id"`
	AppID            string         `json:"app_id,omitempty"`
	CompanyID        string         `json:"company_id,omitempty"`
	Name             string         `json:"name,omitempty"`
	RemoteCreatedAt  int64          `json:"remote_created_at,omitempty"`
	CreatedAt        int64          `json:"created_at,omitempty"`
	UpdatedAt        int64          `json:"updated_at,omitempty"`
	LastRequestAt    int64          `json:"last_request_at,omitempty"`
	Size             int            `json:"size,omitempty"`
	Website          string         `json:"website,omitempty"`
	Industry         string         `json:"industry,omitempty"`
	MonthlySpend     float64        `json:"monthly_spend,omitempty"`
	SessionCount     int            `json:"session_count,omitempty"`
	UserCount        int            `json:"user_count,omitempty"`
	Plan             *Plan          `json:"plan,omitempty"`
	CustomAttributes map[string]any `json:"custom_attributes,omitempty"`
	Tags             *api.TagRefList `json:"tags,omitempty"`
	Segments         *SegmentList   `json:"segments,omitempty"`
}

// Plan represents a company's plan.
type Plan struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// SegmentList is the segment list embedded in a company response.
type SegmentList struct {
	Type     string           `json:"type,omitempty"`
	Segments []api.SegmentRef `json:"segments,omitempty"`
}


// ScrollResponse represents the response from the scroll endpoint.
type ScrollResponse struct {
	Type        string          `json:"type"`
	Data        []Company       `json:"data"`
	ScrollParam string          `json:"scroll_param,omitempty"`
	TotalCount  int             `json:"total_count,omitempty"`
	Pages       api.CursorPages `json:"pages,omitempty"`
}

// CreateOrUpdateRequest represents the body for creating or updating a company.
type CreateOrUpdateRequest struct {
	Name             string         `json:"name,omitempty"`
	CompanyID        string         `json:"company_id,omitempty"`
	Plan             string         `json:"plan,omitempty"`
	Size             int            `json:"size,omitempty"`
	Website          string         `json:"website,omitempty"`
	Industry         string         `json:"industry,omitempty"`
	MonthlySpend     float64        `json:"monthly_spend,omitempty"`
	RemoteCreatedAt  int64          `json:"remote_created_at,omitempty"`
	CustomAttributes map[string]any `json:"custom_attributes,omitempty"`
}

// UpdateRequest represents the body for updating a company by ID.
type UpdateRequest struct {
	Name             string         `json:"name,omitempty"`
	Plan             string         `json:"plan,omitempty"`
	Size             int            `json:"size,omitempty"`
	Website          string         `json:"website,omitempty"`
	Industry         string         `json:"industry,omitempty"`
	MonthlySpend     float64        `json:"monthly_spend,omitempty"`
	CustomAttributes map[string]any `json:"custom_attributes,omitempty"`
}

// ListOptions specifies parameters for the CompanyList (POST /companies/list) endpoint.
type ListOptions struct {
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"per_page,omitempty"`
	Order   string `json:"order,omitempty"`
}

// Contact represents a contact returned in company sub-resource responses.
type Contact struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// ParseGetResult decodes a Result into a Company.
func ParseGetResult(r *api.Result) (*Company, error) { return api.Decode[Company](r) }

// ParseListResult decodes a Result into a PagedResult[Company].
func ParseListResult(r *api.Result) (*api.PagedResult[Company], error) {
	return api.Decode[api.PagedResult[Company]](r)
}

// ParseCreateResult decodes a Result into a Company.
func ParseCreateResult(r *api.Result) (*Company, error) { return api.Decode[Company](r) }

// ParseUpdateResult decodes a Result into a Company.
func ParseUpdateResult(r *api.Result) (*Company, error) { return api.Decode[Company](r) }

// ParseDeleteResult decodes a Result into a Deleted.
func ParseDeleteResult(r *api.Result) (*api.Deleted, error) { return api.Decode[api.Deleted](r) }

// ParseScrollResult decodes a Result into a ScrollResponse.
func ParseScrollResult(r *api.Result) (*ScrollResponse, error) {
	return api.Decode[ScrollResponse](r)
}

// ParseListContactsResult decodes a Result into a PagedResult[Contact].
func ParseListContactsResult(r *api.Result) (*api.PagedResult[Contact], error) {
	return api.Decode[api.PagedResult[Contact]](r)
}

// ParseListSegmentsResult decodes a Result into a SegmentListResult.
func ParseListSegmentsResult(r *api.Result) (*api.SegmentListResult, error) {
	return api.Decode[api.SegmentListResult](r)
}

// ParseListNotesResult decodes a Result into a NoteListResult.
func ParseListNotesResult(r *api.Result) (*api.NoteListResult, error) {
	return api.Decode[api.NoteListResult](r)
}

// ParseCompanyListResult decodes a Result into a PagedResult[Company].
func ParseCompanyListResult(r *api.Result) (*api.PagedResult[Company], error) {
	return api.Decode[api.PagedResult[Company]](r)
}

// Get retrieves a company by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/retrieveacompanybyid
func (s *Service) Get(ctx context.Context, id string) (*Company, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// List returns a single page of companies using GET /companies.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/retrievecompany
func (s *Service) List(ctx context.Context, opts *api.ListOptions) (*api.PagedResult[Company], error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListResult(result)
}

// ListAll returns an iterator over all companies, handling pagination automatically.
func (s *Service) ListAll(ctx context.Context, opts *api.ListOptions) *api.Iter[Company] {
	return api.NewIter[Company](ctx, opts, s.List)
}

// Create creates or updates a company (upsert by company_id).
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/createorupdatecompany
func (s *Service) Create(ctx context.Context, body *CreateOrUpdateRequest) (*Company, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateResult(result)
}

// Update updates an existing company by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/updatecompany
func (s *Service) Update(ctx context.Context, id string, body *UpdateRequest) (*Company, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseUpdateResult(result)
}

// Delete deletes a company by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/deletecompany
func (s *Service) Delete(ctx context.Context, id string) (*api.Deleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseDeleteResult(result)
}

// Scroll iterates over all companies using scroll-based pagination.
// Pass an empty scrollParam for the first request. Use the returned
// ScrollParam for subsequent requests.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/scrolloverallcompanies
func (s *Service) Scroll(ctx context.Context, scrollParam string) (*ScrollResponse, error) {
	result, err := s.ScrollRaw(ctx, scrollParam)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseScrollResult(result)
}

// ListContacts returns the contacts attached to a company.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listattachedcontacts
func (s *Service) ListContacts(ctx context.Context, companyID string, opts *api.ListOptions) (*api.PagedResult[Contact], error) {
	result, err := s.ListContactsRaw(ctx, companyID, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListContactsResult(result)
}

// ListSegments returns the segments attached to a company.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listattachedsegmentsforcompanies
func (s *Service) ListSegments(ctx context.Context, companyID string) (*api.SegmentListResult, error) {
	result, err := s.ListSegmentsRaw(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListSegmentsResult(result)
}

// ListNotes returns the notes for a company.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listcompanynotes
func (s *Service) ListNotes(ctx context.Context, companyID string) (*api.NoteListResult, error) {
	result, err := s.ListNotesRaw(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListNotesResult(result)
}

// CompanyList lists all companies using the POST /companies/list endpoint.
// This endpoint supports page-based pagination with ordering.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listallcompanies
func (s *Service) CompanyList(ctx context.Context, opts *ListOptions) (*api.PagedResult[Company], error) {
	result, err := s.CompanyListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCompanyListResult(result)
}

// GetRaw retrieves a company by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/retrieveacompanybyid
func (s *Service) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("companies/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns a single page of companies with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/retrievecompany
func (s *Service) ListRaw(ctx context.Context, opts *api.ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("companies", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CompanyListRaw lists all companies using POST /companies/list with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listallcompanies
func (s *Service) CompanyListRaw(ctx context.Context, opts *ListOptions) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "companies/list", opts)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates or updates a company with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/createorupdatecompany
func (s *Service) CreateRaw(ctx context.Context, body *CreateOrUpdateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "companies", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing company by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/updatecompany
func (s *Service) UpdateRaw(ctx context.Context, id string, body *UpdateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("companies/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes a company by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/deletecompany
func (s *Service) DeleteRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("companies/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ScrollRaw iterates over companies using scroll-based pagination with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/scrolloverallcompanies
func (s *Service) ScrollRaw(ctx context.Context, scrollParam string) (*api.Result, error) {
	opts := &api.ScrollOptions{ScrollParam: scrollParam}
	path, err := api.AddQueryOptions("companies/scroll", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListContactsRaw returns the contacts attached to a company with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listattachedcontacts
func (s *Service) ListContactsRaw(ctx context.Context, companyID string, opts *api.ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions(fmt.Sprintf("companies/%s/contacts", url.PathEscape(companyID)), opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListSegmentsRaw returns the segments attached to a company with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listattachedsegmentsforcompanies
func (s *Service) ListSegmentsRaw(ctx context.Context, companyID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("companies/%s/segments", url.PathEscape(companyID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListNotesRaw returns the notes for a company with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listcompanynotes
func (s *Service) ListNotesRaw(ctx context.Context, companyID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("companies/%s/notes", url.PathEscape(companyID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
