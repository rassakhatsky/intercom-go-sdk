package articles

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// Service handles communication with the article related methods
// of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new articles Service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Article represents an Intercom help center article.
type Article struct {
	Type              string                        `json:"type"`
	ID                string                        `json:"id"`
	WorkspaceID       string                        `json:"workspace_id,omitempty"`
	Title             string                        `json:"title,omitempty"`
	Description       string                        `json:"description,omitempty"`
	Body              string                        `json:"body,omitempty"`
	AuthorID          int                           `json:"author_id,omitempty"`
	State             string                        `json:"state,omitempty"`
	CreatedAt         int64                         `json:"created_at,omitempty"`
	UpdatedAt         int64                         `json:"updated_at,omitempty"`
	URL               string                        `json:"url,omitempty"`
	ParentID          *int                          `json:"parent_id,omitempty"`
	ParentIDs         []int                         `json:"parent_ids,omitempty"`
	ParentType        *string                       `json:"parent_type,omitempty"`
	DefaultLocale     string                        `json:"default_locale,omitempty"`
	TranslatedContent *api.ArticleTranslatedContent `json:"translated_content,omitempty"`
	Statistics        *Statistics                   `json:"statistics,omitempty"`
}

// Statistics represents engagement statistics for an article.
type Statistics struct {
	Type                      string  `json:"type"`
	Views                     int     `json:"views"`
	Conversations             int     `json:"conversations"`
	Reactions                 int     `json:"reactions"`
	HappyReactionPercentage   float64 `json:"happy_reaction_percentage"`
	NeutralReactionPercentage float64 `json:"neutral_reaction_percentage"`
	SadReactionPercentage     float64 `json:"sad_reaction_percentage"`
}

// CreateRequest represents the request body for creating an article.
type CreateRequest struct {
	Title             string                        `json:"title"`
	AuthorID          int                           `json:"author_id"`
	Description       string                        `json:"description,omitempty"`
	Body              string                        `json:"body,omitempty"`
	State             string                        `json:"state,omitempty"`
	ParentID          *int                          `json:"parent_id,omitempty"`
	ParentType        string                        `json:"parent_type,omitempty"`
	TranslatedContent *api.ArticleTranslatedContent `json:"translated_content,omitempty"`
}

// UpdateRequest represents the request body for updating an article.
type UpdateRequest struct {
	Title             string                        `json:"title,omitempty"`
	Description       string                        `json:"description,omitempty"`
	Body              string                        `json:"body,omitempty"`
	AuthorID          int                           `json:"author_id,omitempty"`
	State             string                        `json:"state,omitempty"`
	ParentID          string                        `json:"parent_id,omitempty"`
	ParentType        string                        `json:"parent_type,omitempty"`
	TranslatedContent *api.ArticleTranslatedContent `json:"translated_content,omitempty"`
}

// SearchOptions specifies the query parameters for searching articles.
type SearchOptions struct {
	Phrase       string `url:"phrase,omitempty"`
	State        string `url:"state,omitempty"`
	HelpCenterID int    `url:"help_center_id,omitempty"`
	Highlight    *bool  `url:"highlight,omitempty"`
}

// SearchResponse represents the response from searching articles.
type SearchResponse struct {
	Type       string           `json:"type"`
	TotalCount int              `json:"total_count"`
	Data       SearchData       `json:"data"`
	Pages      *api.CursorPages `json:"pages,omitempty"`
}

// SearchData contains the articles and highlights from a search.
type SearchData struct {
	Articles   []Article         `json:"articles"`
	Highlights []SearchHighlight `json:"highlights,omitempty"`
}

// SearchHighlight contains highlighted text for a search result.
type SearchHighlight struct {
	ArticleID        string                `json:"article_id"`
	HighlightedTitle []HighlightedTextPart `json:"highlighted_title,omitempty"`
}

// HighlightedTextPart represents a segment of highlighted text.
type HighlightedTextPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// --- Parse Functions ---

// ParseGetResult decodes a Result into an Article.
func ParseGetResult(r *api.Result) (*Article, error) {
	return api.Decode[Article](r)
}

// ParseListResult decodes a Result into a PagedResult[Article].
func ParseListResult(r *api.Result) (*api.PagedResult[Article], error) {
	return api.Decode[api.PagedResult[Article]](r)
}

// ParseCreateResult decodes a Result into an Article.
func ParseCreateResult(r *api.Result) (*Article, error) {
	return api.Decode[Article](r)
}

// ParseUpdateResult decodes a Result into an Article.
func ParseUpdateResult(r *api.Result) (*Article, error) {
	return api.Decode[Article](r)
}

// ParseDeleteResult decodes a Result into a Deleted.
func ParseDeleteResult(r *api.Result) (*api.Deleted, error) {
	return api.Decode[api.Deleted](r)
}

// ParseSearchResult decodes a Result into a SearchResponse.
func ParseSearchResult(r *api.Result) (*SearchResponse, error) {
	return api.Decode[SearchResponse](r)
}

// --- Regular Methods ---

// Get retrieves an article by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/retrievearticle
func (s *Service) Get(ctx context.Context, id string) (*Article, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// List returns a single page of articles.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/listarticles
func (s *Service) List(ctx context.Context, opts *api.ListOptions) (*api.PagedResult[Article], error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListResult(result)
}

// ListAll returns an iterator over all articles, handling pagination automatically.
func (s *Service) ListAll(ctx context.Context, opts *api.ListOptions) *api.Iter[Article] {
	return api.NewIter[Article](ctx, opts, s.List)
}

// Create creates a new article.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/createarticle
func (s *Service) Create(ctx context.Context, body *CreateRequest) (*Article, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateResult(result)
}

// Update updates an existing article by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/updatearticle
func (s *Service) Update(ctx context.Context, id string, body *UpdateRequest) (*Article, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseUpdateResult(result)
}

// Delete deletes an article by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/deletearticle
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

// Search searches for articles using query parameters.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/searcharticles
func (s *Service) Search(ctx context.Context, opts *SearchOptions) (*SearchResponse, error) {
	result, err := s.SearchRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseSearchResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves an article by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/retrievearticle
func (s *Service) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("articles/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns a single page of articles with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/listarticles
func (s *Service) ListRaw(ctx context.Context, opts *api.ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("articles", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates a new article and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/createarticle
func (s *Service) CreateRaw(ctx context.Context, body *CreateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "articles", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing article by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/updatearticle
func (s *Service) UpdateRaw(ctx context.Context, id string, body *UpdateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("articles/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes an article by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/deletearticle
func (s *Service) DeleteRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("articles/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for articles using query parameters and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/articles/searcharticles
func (s *Service) SearchRaw(ctx context.Context, opts *SearchOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("articles/search", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
