package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ArticlesService handles communication with the article related methods
// of the Intercom API.
type ArticlesService service

// Article represents an Intercom help center article.
type Article struct {
	Type              string                    `json:"type"`
	ID                string                    `json:"id"`
	WorkspaceID       string                    `json:"workspace_id,omitempty"`
	Title             string                    `json:"title,omitempty"`
	Description       string                    `json:"description,omitempty"`
	Body              string                    `json:"body,omitempty"`
	AuthorID          int                       `json:"author_id,omitempty"`
	State             string                    `json:"state,omitempty"`
	CreatedAt         int64                     `json:"created_at,omitempty"`
	UpdatedAt         int64                     `json:"updated_at,omitempty"`
	URL               string                    `json:"url,omitempty"`
	ParentID          *int                      `json:"parent_id,omitempty"`
	ParentIDs         []int                     `json:"parent_ids,omitempty"`
	ParentType        *string                   `json:"parent_type,omitempty"`
	DefaultLocale     string                    `json:"default_locale,omitempty"`
	TranslatedContent *ArticleTranslatedContent `json:"translated_content,omitempty"`
	Statistics        *ArticleStatistics        `json:"statistics,omitempty"`
}

// ArticleStatistics represents engagement statistics for an article.
type ArticleStatistics struct {
	Type                      string  `json:"type"`
	Views                     int     `json:"views"`
	Conversations             int     `json:"conversations"`
	Reactions                 int     `json:"reactions"`
	HappyReactionPercentage   float64 `json:"happy_reaction_percentage"`
	NeutralReactionPercentage float64 `json:"neutral_reaction_percentage"`
	SadReactionPercentage     float64 `json:"sad_reaction_percentage"`
}

// ArticleContent represents translated content for a single locale.
type ArticleContent struct {
	Type        string `json:"type,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Body        string `json:"body,omitempty"`
	AuthorID    int    `json:"author_id,omitempty"`
	State       string `json:"state,omitempty"`
	CreatedAt   int64  `json:"created_at,omitempty"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
	URL         string `json:"url,omitempty"`
}

// ArticleTranslatedContent holds translations for an article keyed by locale.
type ArticleTranslatedContent struct {
	Type string          `json:"type,omitempty"`
	AR   *ArticleContent `json:"ar,omitempty"`
	BG   *ArticleContent `json:"bg,omitempty"`
	BS   *ArticleContent `json:"bs,omitempty"`
	CA   *ArticleContent `json:"ca,omitempty"`
	CS   *ArticleContent `json:"cs,omitempty"`
	DA   *ArticleContent `json:"da,omitempty"`
	DE   *ArticleContent `json:"de,omitempty"`
	EL   *ArticleContent `json:"el,omitempty"`
	EN   *ArticleContent `json:"en,omitempty"`
	ES   *ArticleContent `json:"es,omitempty"`
	ET   *ArticleContent `json:"et,omitempty"`
	FI   *ArticleContent `json:"fi,omitempty"`
	FR   *ArticleContent `json:"fr,omitempty"`
	HE   *ArticleContent `json:"he,omitempty"`
	HR   *ArticleContent `json:"hr,omitempty"`
	HU   *ArticleContent `json:"hu,omitempty"`
	ID   *ArticleContent `json:"id,omitempty"`
	IT   *ArticleContent `json:"it,omitempty"`
	JA   *ArticleContent `json:"ja,omitempty"`
	KO   *ArticleContent `json:"ko,omitempty"`
	LT   *ArticleContent `json:"lt,omitempty"`
	LV   *ArticleContent `json:"lv,omitempty"`
	MN   *ArticleContent `json:"mn,omitempty"`
	NB   *ArticleContent `json:"nb,omitempty"`
	NL   *ArticleContent `json:"nl,omitempty"`
	PL   *ArticleContent `json:"pl,omitempty"`
	PT   *ArticleContent `json:"pt,omitempty"`
	PtBR *ArticleContent `json:"pt-BR,omitempty"`
	RO   *ArticleContent `json:"ro,omitempty"`
	RU   *ArticleContent `json:"ru,omitempty"`
	SL   *ArticleContent `json:"sl,omitempty"`
	SR   *ArticleContent `json:"sr,omitempty"`
	SV   *ArticleContent `json:"sv,omitempty"`
	TR   *ArticleContent `json:"tr,omitempty"`
	VI   *ArticleContent `json:"vi,omitempty"`
	ZhCN *ArticleContent `json:"zh-CN,omitempty"`
	ZhTW *ArticleContent `json:"zh-TW,omitempty"`
}

// ArticleDeleted represents the response from deleting an article.
type ArticleDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// CreateArticleRequest represents the request body for creating an article.
type CreateArticleRequest struct {
	Title             string                    `json:"title"`
	AuthorID          int                       `json:"author_id"`
	Description       string                    `json:"description,omitempty"`
	Body              string                    `json:"body,omitempty"`
	State             string                    `json:"state,omitempty"`
	ParentID          *int                      `json:"parent_id,omitempty"`
	ParentType        string                    `json:"parent_type,omitempty"`
	TranslatedContent *ArticleTranslatedContent `json:"translated_content,omitempty"`
}

// UpdateArticleRequest represents the request body for updating an article.
type UpdateArticleRequest struct {
	Title             string                    `json:"title,omitempty"`
	Description       string                    `json:"description,omitempty"`
	Body              string                    `json:"body,omitempty"`
	AuthorID          int                       `json:"author_id,omitempty"`
	State             string                    `json:"state,omitempty"`
	ParentID          string                    `json:"parent_id,omitempty"`
	ParentType        string                    `json:"parent_type,omitempty"`
	TranslatedContent *ArticleTranslatedContent `json:"translated_content,omitempty"`
}

// ArticleSearchOptions specifies the query parameters for searching articles.
type ArticleSearchOptions struct {
	Phrase       string `url:"phrase,omitempty"`
	State        string `url:"state,omitempty"`
	HelpCenterID int    `url:"help_center_id,omitempty"`
	Highlight    *bool  `url:"highlight,omitempty"`
}

// ArticleSearchResponse represents the response from searching articles.
type ArticleSearchResponse struct {
	Type       string            `json:"type"`
	TotalCount int               `json:"total_count"`
	Data       ArticleSearchData `json:"data"`
	Pages      *CursorPages      `json:"pages,omitempty"`
}

// ArticleSearchData contains the articles and highlights from a search.
type ArticleSearchData struct {
	Articles   []Article                `json:"articles"`
	Highlights []ArticleSearchHighlight `json:"highlights,omitempty"`
}

// ArticleSearchHighlight contains highlighted text for a search result.
type ArticleSearchHighlight struct {
	ArticleID        string                `json:"article_id"`
	HighlightedTitle []HighlightedTextPart `json:"highlighted_title,omitempty"`
}

// HighlightedTextPart represents a segment of highlighted text.
type HighlightedTextPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// --- Parse Functions ---

// ParseArticleGetResult decodes a Result into an Article.
func ParseArticleGetResult(r *Result) (*Article, error) {
	return Decode[Article](r)
}

// ParseArticleListResult decodes a Result into a PagedResult[Article].
func ParseArticleListResult(r *Result) (*PagedResult[Article], error) {
	return Decode[PagedResult[Article]](r)
}

// ParseArticleCreateResult decodes a Result into an Article.
func ParseArticleCreateResult(r *Result) (*Article, error) {
	return Decode[Article](r)
}

// ParseArticleUpdateResult decodes a Result into an Article.
func ParseArticleUpdateResult(r *Result) (*Article, error) {
	return Decode[Article](r)
}

// ParseArticleDeleteResult decodes a Result into an ArticleDeleted.
func ParseArticleDeleteResult(r *Result) (*ArticleDeleted, error) {
	return Decode[ArticleDeleted](r)
}

// ParseArticleSearchResult decodes a Result into an ArticleSearchResponse.
func ParseArticleSearchResult(r *Result) (*ArticleSearchResponse, error) {
	return Decode[ArticleSearchResponse](r)
}

// --- Regular Methods ---

// Get retrieves an article by ID.
func (s *ArticlesService) Get(ctx context.Context, id string) (*Article, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseArticleGetResult(result)
}

// List returns a single page of articles.
func (s *ArticlesService) List(ctx context.Context, opts *ListOptions) (*PagedResult[Article], error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseArticleListResult(result)
}

// ListAll returns an iterator over all articles, handling pagination automatically.
func (s *ArticlesService) ListAll(ctx context.Context, opts *ListOptions) *Iter[Article] {
	return NewIter[Article](ctx, opts, s.List)
}

// Create creates a new article.
func (s *ArticlesService) Create(ctx context.Context, body *CreateArticleRequest) (*Article, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseArticleCreateResult(result)
}

// Update updates an existing article by ID.
func (s *ArticlesService) Update(ctx context.Context, id string, body *UpdateArticleRequest) (*Article, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseArticleUpdateResult(result)
}

// Delete deletes an article by ID.
func (s *ArticlesService) Delete(ctx context.Context, id string) (*ArticleDeleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseArticleDeleteResult(result)
}

// Search searches for articles using query parameters.
func (s *ArticlesService) Search(ctx context.Context, opts *ArticleSearchOptions) (*ArticleSearchResponse, error) {
	result, err := s.SearchRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseArticleSearchResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves an article by ID and returns the full HTTP result.
func (s *ArticlesService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("articles/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns a single page of articles with the full HTTP result.
func (s *ArticlesService) ListRaw(ctx context.Context, opts *ListOptions) (*Result, error) {
	path, err := addQueryOptions("articles", opts)
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
func (s *ArticlesService) CreateRaw(ctx context.Context, body *CreateArticleRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "articles", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing article by ID and returns the full HTTP result.
func (s *ArticlesService) UpdateRaw(ctx context.Context, id string, body *UpdateArticleRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("articles/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes an article by ID and returns the full HTTP result.
func (s *ArticlesService) DeleteRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("articles/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for articles using query parameters and returns the full HTTP result.
func (s *ArticlesService) SearchRaw(ctx context.Context, opts *ArticleSearchOptions) (*Result, error) {
	path, err := addQueryOptions("articles/search", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
