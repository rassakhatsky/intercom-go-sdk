package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// InternalArticlesService handles communication with the internal article
// related methods of the Intercom API.
type InternalArticlesService service

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

// InternalArticleDeleted represents the response from deleting an internal article.
type InternalArticleDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// CreateInternalArticleRequest represents the request body for creating an internal article.
type CreateInternalArticleRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body,omitempty"`
	AuthorID int    `json:"author_id"`
	OwnerID  int    `json:"owner_id"`
}

// UpdateInternalArticleRequest represents the request body for updating an internal article.
type UpdateInternalArticleRequest struct {
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	AuthorID int    `json:"author_id,omitempty"`
	OwnerID  int    `json:"owner_id,omitempty"`
}

// InternalArticleSearchOptions specifies the query parameters for searching internal articles.
type InternalArticleSearchOptions struct {
	FolderID string `url:"folder_id,omitempty"`
}

// InternalArticleSearchResponse represents the response from searching internal articles.
type InternalArticleSearchResponse struct {
	Type       string                    `json:"type"`
	TotalCount int                       `json:"total_count"`
	Data       InternalArticleSearchData `json:"data"`
	Pages      *CursorPages              `json:"pages,omitempty"`
}

// InternalArticleSearchData contains the internal articles from a search.
type InternalArticleSearchData struct {
	InternalArticles []InternalArticle `json:"internal_articles"`
}

// --- Parse Functions ---

// ParseInternalArticleGetResult decodes a Result into an InternalArticle.
func ParseInternalArticleGetResult(r *Result) (*InternalArticle, error) {
	return Decode[InternalArticle](r)
}

// ParseInternalArticleListResult decodes a Result into a PagedResult[InternalArticle].
func ParseInternalArticleListResult(r *Result) (*PagedResult[InternalArticle], error) {
	return Decode[PagedResult[InternalArticle]](r)
}

// ParseInternalArticleCreateResult decodes a Result into an InternalArticle.
func ParseInternalArticleCreateResult(r *Result) (*InternalArticle, error) {
	return Decode[InternalArticle](r)
}

// ParseInternalArticleUpdateResult decodes a Result into an InternalArticle.
func ParseInternalArticleUpdateResult(r *Result) (*InternalArticle, error) {
	return Decode[InternalArticle](r)
}

// ParseInternalArticleDeleteResult decodes a Result into an InternalArticleDeleted.
func ParseInternalArticleDeleteResult(r *Result) (*InternalArticleDeleted, error) {
	return Decode[InternalArticleDeleted](r)
}

// ParseInternalArticleSearchResult decodes a Result into an InternalArticleSearchResponse.
func ParseInternalArticleSearchResult(r *Result) (*InternalArticleSearchResponse, error) {
	return Decode[InternalArticleSearchResponse](r)
}

// --- Regular Methods ---

// Get retrieves an internal article by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/retrieveinternalarticle
func (s *InternalArticlesService) Get(ctx context.Context, id string) (*InternalArticle, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseInternalArticleGetResult(result)
}

// List returns a single page of internal articles.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/listinternalarticles
func (s *InternalArticlesService) List(ctx context.Context, opts *ListOptions) (*PagedResult[InternalArticle], error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseInternalArticleListResult(result)
}

// ListAll returns an iterator over all internal articles, handling pagination automatically.
func (s *InternalArticlesService) ListAll(ctx context.Context, opts *ListOptions) *Iter[InternalArticle] {
	return NewIter[InternalArticle](ctx, opts, s.List)
}

// Create creates a new internal article.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/createinternalarticle
func (s *InternalArticlesService) Create(ctx context.Context, body *CreateInternalArticleRequest) (*InternalArticle, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseInternalArticleCreateResult(result)
}

// Update updates an existing internal article by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/updateinternalarticle
func (s *InternalArticlesService) Update(ctx context.Context, id string, body *UpdateInternalArticleRequest) (*InternalArticle, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseInternalArticleUpdateResult(result)
}

// Delete deletes an internal article by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/deleteinternalarticle
func (s *InternalArticlesService) Delete(ctx context.Context, id string) (*InternalArticleDeleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseInternalArticleDeleteResult(result)
}

// Search searches for internal articles using query parameters.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/searchinternalarticles
func (s *InternalArticlesService) Search(ctx context.Context, opts *InternalArticleSearchOptions) (*InternalArticleSearchResponse, error) {
	result, err := s.SearchRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseInternalArticleSearchResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves an internal article by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/retrieveinternalarticle
func (s *InternalArticlesService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("internal_articles/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns a single page of internal articles with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/listinternalarticles
func (s *InternalArticlesService) ListRaw(ctx context.Context, opts *ListOptions) (*Result, error) {
	path, err := addQueryOptions("internal_articles", opts)
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
func (s *InternalArticlesService) CreateRaw(ctx context.Context, body *CreateInternalArticleRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "internal_articles", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing internal article by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/updateinternalarticle
func (s *InternalArticlesService) UpdateRaw(ctx context.Context, id string, body *UpdateInternalArticleRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("internal_articles/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes an internal article by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/deleteinternalarticle
func (s *InternalArticlesService) DeleteRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("internal_articles/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for internal articles using query parameters with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/internal-articles/searchinternalarticles
func (s *InternalArticlesService) SearchRaw(ctx context.Context, opts *InternalArticleSearchOptions) (*Result, error) {
	path, err := addQueryOptions("internal_articles/search", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
