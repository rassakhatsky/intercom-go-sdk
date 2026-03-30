package articles

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// InternalService handles communication with the internal article
// related methods of the Intercom API.
type InternalService struct {
	client api.Caller
}

// NewInternalService creates a new internal articles InternalService.
func NewInternalService(c api.Caller) *InternalService {
	return &InternalService{client: c}
}

// InternalArticle represents an Intercom internal (team-only) article.
type InternalArticle struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Title     string `json:"title,omitempty"`
	Body      string `json:"body,omitempty"`
	OwnerID   int    `json:"owner_id,omitempty"`
	AuthorID  int    `json:"author_id,omitempty"`
	Locale    string `json:"locale,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

// InternalDeleted represents the response from deleting an internal article.
type InternalDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// CreateInternalRequest represents the request body for creating an internal article.
type CreateInternalRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body,omitempty"`
	AuthorID int    `json:"author_id"`
	OwnerID  int    `json:"owner_id"`
}

// UpdateInternalRequest represents the request body for updating an internal article.
type UpdateInternalRequest struct {
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	AuthorID int    `json:"author_id,omitempty"`
	OwnerID  int    `json:"owner_id,omitempty"`
}

// InternalSearchOptions specifies the query parameters for searching internal articles.
type InternalSearchOptions struct {
	FolderID string `url:"folder_id,omitempty"`
}

// InternalSearchResponse represents the response from searching internal articles.
type InternalSearchResponse struct {
	Type       string             `json:"type"`
	TotalCount int                `json:"total_count"`
	Data       InternalSearchData `json:"data"`
	Pages      *api.CursorPages   `json:"pages,omitempty"`
}

// InternalSearchData contains the internal articles from a search.
type InternalSearchData struct {
	InternalArticles []InternalArticle `json:"internal_articles"`
}

// --- Parse Functions ---

// ParseInternalGetResult decodes a Result into an InternalArticle.
func ParseInternalGetResult(r *api.Result) (*InternalArticle, error) {
	return api.Decode[InternalArticle](r)
}

// ParseInternalListResult decodes a Result into a PagedResult[InternalArticle].
func ParseInternalListResult(r *api.Result) (*api.PagedResult[InternalArticle], error) {
	return api.Decode[api.PagedResult[InternalArticle]](r)
}

// ParseInternalCreateResult decodes a Result into an InternalArticle.
func ParseInternalCreateResult(r *api.Result) (*InternalArticle, error) {
	return api.Decode[InternalArticle](r)
}

// ParseInternalUpdateResult decodes a Result into an InternalArticle.
func ParseInternalUpdateResult(r *api.Result) (*InternalArticle, error) {
	return api.Decode[InternalArticle](r)
}

// ParseInternalDeleteResult decodes a Result into an InternalDeleted.
func ParseInternalDeleteResult(r *api.Result) (*InternalDeleted, error) {
	return api.Decode[InternalDeleted](r)
}

// ParseInternalSearchResult decodes a Result into an InternalSearchResponse.
func ParseInternalSearchResult(r *api.Result) (*InternalSearchResponse, error) {
	return api.Decode[InternalSearchResponse](r)
}

// --- Regular Methods ---

// Get retrieves an internal article by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/retrieveinternalarticle
func (s *InternalService) Get(ctx context.Context, id string) (*InternalArticle, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseInternalGetResult(result)
}

// List returns a single page of internal articles.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/listinternalarticles
func (s *InternalService) List(ctx context.Context, opts *api.ListOptions) (*api.PagedResult[InternalArticle], error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseInternalListResult(result)
}

// ListAll returns an iterator over all internal articles, handling pagination automatically.
func (s *InternalService) ListAll(ctx context.Context, opts *api.ListOptions) *api.Iter[InternalArticle] {
	return api.NewIter[InternalArticle](ctx, opts, s.List)
}

// Create creates a new internal article.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/createinternalarticle
func (s *InternalService) Create(ctx context.Context, body *CreateInternalRequest) (*InternalArticle, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseInternalCreateResult(result)
}

// Update updates an existing internal article by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/updateinternalarticle
func (s *InternalService) Update(ctx context.Context, id string, body *UpdateInternalRequest) (*InternalArticle, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseInternalUpdateResult(result)
}

// Delete deletes an internal article by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/deleteinternalarticle
func (s *InternalService) Delete(ctx context.Context, id string) (*InternalDeleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseInternalDeleteResult(result)
}

// Search searches for internal articles using query parameters.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/searchinternalarticles
func (s *InternalService) Search(ctx context.Context, opts *InternalSearchOptions) (*InternalSearchResponse, error) {
	result, err := s.SearchRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseInternalSearchResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves an internal article by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/retrieveinternalarticle
func (s *InternalService) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("internal_articles/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns a single page of internal articles with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/listinternalarticles
func (s *InternalService) ListRaw(ctx context.Context, opts *api.ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("internal_articles", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates a new internal article with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/createinternalarticle
func (s *InternalService) CreateRaw(ctx context.Context, body *CreateInternalRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "internal_articles", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing internal article by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/updateinternalarticle
func (s *InternalService) UpdateRaw(ctx context.Context, id string, body *UpdateInternalRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("internal_articles/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes an internal article by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/deleteinternalarticle
func (s *InternalService) DeleteRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("internal_articles/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for internal articles using query parameters with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/searchinternalarticles
func (s *InternalService) SearchRaw(ctx context.Context, opts *InternalSearchOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("internal_articles/search", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
