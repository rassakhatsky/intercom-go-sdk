package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CompaniesService handles communication with the company related methods
// of the Intercom API.
type CompaniesService service

// Company represents an Intercom company.
type Company struct {
	Type             string              `json:"type"`
	ID               string              `json:"id"`
	AppID            string              `json:"app_id,omitempty"`
	CompanyID        string              `json:"company_id,omitempty"`
	Name             string              `json:"name,omitempty"`
	RemoteCreatedAt  int64               `json:"remote_created_at,omitempty"`
	CreatedAt        int64               `json:"created_at,omitempty"`
	UpdatedAt        int64               `json:"updated_at,omitempty"`
	LastRequestAt    int64               `json:"last_request_at,omitempty"`
	Size             int                 `json:"size,omitempty"`
	Website          string              `json:"website,omitempty"`
	Industry         string              `json:"industry,omitempty"`
	MonthlySpend     float64             `json:"monthly_spend,omitempty"`
	SessionCount     int                 `json:"session_count,omitempty"`
	UserCount        int                 `json:"user_count,omitempty"`
	Plan             *CompanyPlan        `json:"plan,omitempty"`
	CustomAttributes map[string]any      `json:"custom_attributes,omitempty"`
	Tags             *CompanyTagList     `json:"tags,omitempty"`
	Segments         *CompanySegmentList `json:"segments,omitempty"`
}

// CompanyPlan represents a company's plan.
type CompanyPlan struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// CompanyTagList is the tag list embedded in a company response.
type CompanyTagList struct {
	Type string   `json:"type,omitempty"`
	Tags []TagRef `json:"tags,omitempty"`
}

// CompanySegmentList is the segment list embedded in a company response.
type CompanySegmentList struct {
	Type     string       `json:"type,omitempty"`
	Segments []SegmentRef `json:"segments,omitempty"`
}

// CompanyDeleted represents the response from deleting a company.
type CompanyDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// CompanyScrollResponse represents the response from the scroll endpoint.
type CompanyScrollResponse struct {
	Type        string      `json:"type"`
	Data        []Company   `json:"data"`
	ScrollParam string      `json:"scroll_param,omitempty"`
	TotalCount  int         `json:"total_count,omitempty"`
	Pages       CursorPages `json:"pages,omitempty"`
}

// CreateOrUpdateCompanyRequest represents the body for creating or updating a company.
type CreateOrUpdateCompanyRequest struct {
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

// UpdateCompanyRequest represents the body for updating a company by ID.
type UpdateCompanyRequest struct {
	Name             string         `json:"name,omitempty"`
	Plan             string         `json:"plan,omitempty"`
	Size             int            `json:"size,omitempty"`
	Website          string         `json:"website,omitempty"`
	Industry         string         `json:"industry,omitempty"`
	MonthlySpend     float64        `json:"monthly_spend,omitempty"`
	CustomAttributes map[string]any `json:"custom_attributes,omitempty"`
}

// CompanyListOptions specifies parameters for the CompanyList (POST /companies/list) endpoint.
type CompanyListOptions struct {
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"per_page,omitempty"`
	Order   string `json:"order,omitempty"`
}

// ParseCompanyGetResult decodes a Result into a Company.
func ParseCompanyGetResult(r *Result) (*Company, error) { return Decode[Company](r) }

// ParseCompanyListResult decodes a Result into a PagedResult[Company].
func ParseCompanyListResult(r *Result) (*PagedResult[Company], error) {
	return Decode[PagedResult[Company]](r)
}

// ParseCompanyCreateResult decodes a Result into a Company.
func ParseCompanyCreateResult(r *Result) (*Company, error) { return Decode[Company](r) }

// ParseCompanyUpdateResult decodes a Result into a Company.
func ParseCompanyUpdateResult(r *Result) (*Company, error) { return Decode[Company](r) }

// ParseCompanyDeleteResult decodes a Result into a CompanyDeleted.
func ParseCompanyDeleteResult(r *Result) (*CompanyDeleted, error) { return Decode[CompanyDeleted](r) }

// ParseCompanyScrollResult decodes a Result into a CompanyScrollResponse.
func ParseCompanyScrollResult(r *Result) (*CompanyScrollResponse, error) {
	return Decode[CompanyScrollResponse](r)
}

// ParseCompanyListContactsResult decodes a Result into a PagedResult[Contact].
func ParseCompanyListContactsResult(r *Result) (*PagedResult[Contact], error) {
	return Decode[PagedResult[Contact]](r)
}

// ParseCompanyListSegmentsResult decodes a Result into a SegmentListResult.
func ParseCompanyListSegmentsResult(r *Result) (*SegmentListResult, error) {
	return Decode[SegmentListResult](r)
}

// ParseCompanyListNotesResult decodes a Result into a NoteListResult.
func ParseCompanyListNotesResult(r *Result) (*NoteListResult, error) {
	return Decode[NoteListResult](r)
}

// ParseCompanyCompanyListResult decodes a Result into a PagedResult[Company].
func ParseCompanyCompanyListResult(r *Result) (*PagedResult[Company], error) {
	return Decode[PagedResult[Company]](r)
}

// Get retrieves a company by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/retrieveacompanybyid
func (s *CompaniesService) Get(ctx context.Context, id string) (*Company, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCompanyGetResult(result)
}

// List returns a single page of companies using GET /companies.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/retrievecompany
func (s *CompaniesService) List(ctx context.Context, opts *ListOptions) (*PagedResult[Company], error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCompanyListResult(result)
}

// ListAll returns an iterator over all companies, handling pagination automatically.
func (s *CompaniesService) ListAll(ctx context.Context, opts *ListOptions) *Iter[Company] {
	return NewIter[Company](ctx, opts, s.List)
}

// Create creates or updates a company (upsert by company_id).
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/createorupdatecompany
func (s *CompaniesService) Create(ctx context.Context, body *CreateOrUpdateCompanyRequest) (*Company, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCompanyCreateResult(result)
}

// Update updates an existing company by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/updatecompany
func (s *CompaniesService) Update(ctx context.Context, id string, body *UpdateCompanyRequest) (*Company, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCompanyUpdateResult(result)
}

// Delete deletes a company by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/deletecompany
func (s *CompaniesService) Delete(ctx context.Context, id string) (*CompanyDeleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCompanyDeleteResult(result)
}

// Scroll iterates over all companies using scroll-based pagination.
// Pass an empty scrollParam for the first request. Use the returned
// ScrollParam for subsequent requests.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/scrolloverallcompanies
func (s *CompaniesService) Scroll(ctx context.Context, scrollParam string) (*CompanyScrollResponse, error) {
	result, err := s.ScrollRaw(ctx, scrollParam)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCompanyScrollResult(result)
}

// ListContacts returns the contacts attached to a company.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listattachedcontacts
func (s *CompaniesService) ListContacts(ctx context.Context, companyID string, opts *ListOptions) (*PagedResult[Contact], error) {
	result, err := s.ListContactsRaw(ctx, companyID, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCompanyListContactsResult(result)
}

// ListSegments returns the segments attached to a company.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listattachedsegmentsforcompanies
func (s *CompaniesService) ListSegments(ctx context.Context, companyID string) (*SegmentListResult, error) {
	result, err := s.ListSegmentsRaw(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCompanyListSegmentsResult(result)
}

// ListNotes returns the notes for a company.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listcompanynotes
func (s *CompaniesService) ListNotes(ctx context.Context, companyID string) (*NoteListResult, error) {
	result, err := s.ListNotesRaw(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCompanyListNotesResult(result)
}

// CompanyList lists all companies using the POST /companies/list endpoint.
// This endpoint supports page-based pagination with ordering.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listallcompanies
func (s *CompaniesService) CompanyList(ctx context.Context, opts *CompanyListOptions) (*PagedResult[Company], error) {
	result, err := s.CompanyListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCompanyCompanyListResult(result)
}

// GetRaw retrieves a company by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/retrieveacompanybyid
func (s *CompaniesService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("companies/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns a single page of companies with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/retrievecompany
func (s *CompaniesService) ListRaw(ctx context.Context, opts *ListOptions) (*Result, error) {
	path, err := addQueryOptions("companies", opts)
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
func (s *CompaniesService) CompanyListRaw(ctx context.Context, opts *CompanyListOptions) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "companies/list", opts)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates or updates a company with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/createorupdatecompany
func (s *CompaniesService) CreateRaw(ctx context.Context, body *CreateOrUpdateCompanyRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "companies", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing company by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/updatecompany
func (s *CompaniesService) UpdateRaw(ctx context.Context, id string, body *UpdateCompanyRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("companies/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes a company by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/deletecompany
func (s *CompaniesService) DeleteRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("companies/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ScrollRaw iterates over companies using scroll-based pagination with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/scrolloverallcompanies
func (s *CompaniesService) ScrollRaw(ctx context.Context, scrollParam string) (*Result, error) {
	opts := &ScrollOptions{ScrollParam: scrollParam}
	path, err := addQueryOptions("companies/scroll", opts)
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
func (s *CompaniesService) ListContactsRaw(ctx context.Context, companyID string, opts *ListOptions) (*Result, error) {
	path, err := addQueryOptions(fmt.Sprintf("companies/%s/contacts", url.PathEscape(companyID)), opts)
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
func (s *CompaniesService) ListSegmentsRaw(ctx context.Context, companyID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("companies/%s/segments", url.PathEscape(companyID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListNotesRaw returns the notes for a company with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/companies/listcompanynotes
func (s *CompaniesService) ListNotesRaw(ctx context.Context, companyID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("companies/%s/notes", url.PathEscape(companyID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
