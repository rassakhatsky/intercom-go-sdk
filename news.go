package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// NewsService handles communication with the news related methods
// of the Intercom API.
type NewsService service

// NewsItem represents an Intercom news item.
type NewsItem struct {
	Type                string               `json:"type"`
	ID                  string               `json:"id"`
	WorkspaceID         string               `json:"workspace_id,omitempty"`
	Title               string               `json:"title,omitempty"`
	Body                string               `json:"body,omitempty"`
	SenderID            int                  `json:"sender_id,omitempty"`
	State               string               `json:"state,omitempty"`
	Labels              []string             `json:"labels,omitempty"`
	CoverImageURL       string               `json:"cover_image_url,omitempty"`
	Reactions           []string             `json:"reactions,omitempty"`
	DeliverSilently     bool                 `json:"deliver_silently,omitempty"`
	CreatedAt           int64                `json:"created_at,omitempty"`
	UpdatedAt           int64                `json:"updated_at,omitempty"`
	NewsfeedAssignments []NewsfeedAssignment `json:"newsfeed_assignments,omitempty"`
}

// NewsfeedAssignment assigns a news item to a newsfeed.
type NewsfeedAssignment struct {
	NewsfeedID  int   `json:"newsfeed_id"`
	PublishedAt int64 `json:"published_at,omitempty"`
}

// Newsfeed represents an Intercom newsfeed.
type Newsfeed struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

// NewsItemDeleted represents the response from deleting a news item.
type NewsItemDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// CreateNewsItemRequest represents the request body for creating a news item.
type CreateNewsItemRequest struct {
	Title               string               `json:"title"`
	Body                string               `json:"body,omitempty"`
	SenderID            int                  `json:"sender_id"`
	State               string               `json:"state,omitempty"`
	DeliverSilently     bool                 `json:"deliver_silently,omitempty"`
	Labels              []string             `json:"labels,omitempty"`
	Reactions           []string             `json:"reactions,omitempty"`
	NewsfeedAssignments []NewsfeedAssignment `json:"newsfeed_assignments,omitempty"`
}

// UpdateNewsItemRequest represents the request body for updating a news item.
type UpdateNewsItemRequest struct {
	Title               string               `json:"title,omitempty"`
	Body                string               `json:"body,omitempty"`
	SenderID            int                  `json:"sender_id,omitempty"`
	State               string               `json:"state,omitempty"`
	DeliverSilently     *bool                `json:"deliver_silently,omitempty"`
	Labels              []string             `json:"labels,omitempty"`
	Reactions           []string             `json:"reactions,omitempty"`
	NewsfeedAssignments []NewsfeedAssignment `json:"newsfeed_assignments,omitempty"`
}

// --- Parse Functions ---

// ParseNewsListNewsItemsResult decodes a Result into a PagedResult[NewsItem].
func ParseNewsListNewsItemsResult(r *Result) (*PagedResult[NewsItem], error) {
	return Decode[PagedResult[NewsItem]](r)
}

// ParseNewsGetNewsItemResult decodes a Result into a NewsItem.
func ParseNewsGetNewsItemResult(r *Result) (*NewsItem, error) {
	return Decode[NewsItem](r)
}

// ParseNewsCreateNewsItemResult decodes a Result into a NewsItem.
func ParseNewsCreateNewsItemResult(r *Result) (*NewsItem, error) {
	return Decode[NewsItem](r)
}

// ParseNewsUpdateNewsItemResult decodes a Result into a NewsItem.
func ParseNewsUpdateNewsItemResult(r *Result) (*NewsItem, error) {
	return Decode[NewsItem](r)
}

// ParseNewsDeleteNewsItemResult decodes a Result into a NewsItemDeleted.
func ParseNewsDeleteNewsItemResult(r *Result) (*NewsItemDeleted, error) {
	return Decode[NewsItemDeleted](r)
}

// ParseNewsListNewsfeedsResult decodes a Result into a PagedResult[Newsfeed].
func ParseNewsListNewsfeedsResult(r *Result) (*PagedResult[Newsfeed], error) {
	return Decode[PagedResult[Newsfeed]](r)
}

// ParseNewsGetNewsfeedResult decodes a Result into a Newsfeed.
func ParseNewsGetNewsfeedResult(r *Result) (*Newsfeed, error) {
	return Decode[Newsfeed](r)
}

// ParseNewsListNewsfeedItemsResult decodes a Result into a PagedResult[NewsItem].
func ParseNewsListNewsfeedItemsResult(r *Result) (*PagedResult[NewsItem], error) {
	return Decode[PagedResult[NewsItem]](r)
}

// --- Regular Methods ---

// ListNewsItems returns a single page of news items.
func (s *NewsService) ListNewsItems(ctx context.Context, opts *ListOptions) (*PagedResult[NewsItem], error) {
	result, err := s.ListNewsItemsRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseNewsListNewsItemsResult(result)
}

// GetNewsItem retrieves a news item by ID.
func (s *NewsService) GetNewsItem(ctx context.Context, id string) (*NewsItem, error) {
	result, err := s.GetNewsItemRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseNewsGetNewsItemResult(result)
}

// CreateNewsItem creates a new news item.
func (s *NewsService) CreateNewsItem(ctx context.Context, body *CreateNewsItemRequest) (*NewsItem, error) {
	result, err := s.CreateNewsItemRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseNewsCreateNewsItemResult(result)
}

// UpdateNewsItem updates an existing news item by ID.
func (s *NewsService) UpdateNewsItem(ctx context.Context, id string, body *UpdateNewsItemRequest) (*NewsItem, error) {
	result, err := s.UpdateNewsItemRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseNewsUpdateNewsItemResult(result)
}

// DeleteNewsItem deletes a news item by ID.
func (s *NewsService) DeleteNewsItem(ctx context.Context, id string) (*NewsItemDeleted, error) {
	result, err := s.DeleteNewsItemRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseNewsDeleteNewsItemResult(result)
}

// ListNewsfeeds returns a single page of newsfeeds.
func (s *NewsService) ListNewsfeeds(ctx context.Context, opts *ListOptions) (*PagedResult[Newsfeed], error) {
	result, err := s.ListNewsfeedsRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseNewsListNewsfeedsResult(result)
}

// GetNewsfeed retrieves a newsfeed by ID.
func (s *NewsService) GetNewsfeed(ctx context.Context, id string) (*Newsfeed, error) {
	result, err := s.GetNewsfeedRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseNewsGetNewsfeedResult(result)
}

// ListNewsfeedItems returns a single page of live news items for a specific newsfeed.
func (s *NewsService) ListNewsfeedItems(ctx context.Context, newsfeedID string, opts *ListOptions) (*PagedResult[NewsItem], error) {
	result, err := s.ListNewsfeedItemsRaw(ctx, newsfeedID, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseNewsListNewsfeedItemsResult(result)
}

// --- Raw Methods ---

// ListNewsItemsRaw returns a single page of news items with the full HTTP result.
func (s *NewsService) ListNewsItemsRaw(ctx context.Context, opts *ListOptions) (*Result, error) {
	path, err := addQueryOptions("news/news_items", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetNewsItemRaw retrieves a news item by ID with the full HTTP result.
func (s *NewsService) GetNewsItemRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("news/news_items/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateNewsItemRaw creates a new news item with the full HTTP result.
func (s *NewsService) CreateNewsItemRaw(ctx context.Context, body *CreateNewsItemRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "news/news_items", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateNewsItemRaw updates an existing news item by ID with the full HTTP result.
func (s *NewsService) UpdateNewsItemRaw(ctx context.Context, id string, body *UpdateNewsItemRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("news/news_items/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteNewsItemRaw deletes a news item by ID with the full HTTP result.
func (s *NewsService) DeleteNewsItemRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("news/news_items/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListNewsfeedsRaw returns a single page of newsfeeds with the full HTTP result.
func (s *NewsService) ListNewsfeedsRaw(ctx context.Context, opts *ListOptions) (*Result, error) {
	path, err := addQueryOptions("news/newsfeeds", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetNewsfeedRaw retrieves a newsfeed by ID with the full HTTP result.
func (s *NewsService) GetNewsfeedRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("news/newsfeeds/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListNewsfeedItemsRaw returns a single page of live news items for a specific newsfeed with the full HTTP result.
func (s *NewsService) ListNewsfeedItemsRaw(ctx context.Context, newsfeedID string, opts *ListOptions) (*Result, error) {
	path, err := addQueryOptions(fmt.Sprintf("news/newsfeeds/%s/items", url.PathEscape(newsfeedID)), opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
