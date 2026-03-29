package news

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// Service handles communication with the news related methods
// of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new news Service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Item represents an Intercom news item.
type Item struct {
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

// ItemDeleted represents the response from deleting a news item.
type ItemDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// CreateItemRequest represents the request body for creating a news item.
type CreateItemRequest struct {
	Title               string               `json:"title"`
	Body                string               `json:"body,omitempty"`
	SenderID            int                  `json:"sender_id"`
	State               string               `json:"state,omitempty"`
	DeliverSilently     *bool                `json:"deliver_silently,omitempty"`
	Labels              []string             `json:"labels,omitempty"`
	Reactions           []string             `json:"reactions,omitempty"`
	NewsfeedAssignments []NewsfeedAssignment `json:"newsfeed_assignments,omitempty"`
}

// UpdateItemRequest represents the request body for updating a news item.
type UpdateItemRequest struct {
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

// ParseListItemsResult decodes a Result into a PagedResult[Item].
func ParseListItemsResult(r *api.Result) (*api.PagedResult[Item], error) {
	return api.Decode[api.PagedResult[Item]](r)
}

// ParseGetItemResult decodes a Result into an Item.
func ParseGetItemResult(r *api.Result) (*Item, error) {
	return api.Decode[Item](r)
}

// ParseCreateItemResult decodes a Result into an Item.
func ParseCreateItemResult(r *api.Result) (*Item, error) {
	return api.Decode[Item](r)
}

// ParseUpdateItemResult decodes a Result into an Item.
func ParseUpdateItemResult(r *api.Result) (*Item, error) {
	return api.Decode[Item](r)
}

// ParseDeleteItemResult decodes a Result into an ItemDeleted.
func ParseDeleteItemResult(r *api.Result) (*ItemDeleted, error) {
	return api.Decode[ItemDeleted](r)
}

// ParseListNewsfeedsResult decodes a Result into a PagedResult[Newsfeed].
func ParseListNewsfeedsResult(r *api.Result) (*api.PagedResult[Newsfeed], error) {
	return api.Decode[api.PagedResult[Newsfeed]](r)
}

// ParseGetNewsfeedResult decodes a Result into a Newsfeed.
func ParseGetNewsfeedResult(r *api.Result) (*Newsfeed, error) {
	return api.Decode[Newsfeed](r)
}

// ParseListNewsfeedItemsResult decodes a Result into a PagedResult[Item].
func ParseListNewsfeedItemsResult(r *api.Result) (*api.PagedResult[Item], error) {
	return api.Decode[api.PagedResult[Item]](r)
}

// --- Regular Methods ---

// ListItems returns a single page of news items.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/listnewsitems
func (s *Service) ListItems(ctx context.Context, opts *api.ListOptions) (*api.PagedResult[Item], error) {
	result, err := s.ListItemsRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListItemsResult(result)
}

// GetItem retrieves a news item by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/retrievenewsitem
func (s *Service) GetItem(ctx context.Context, id string) (*Item, error) {
	result, err := s.GetItemRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetItemResult(result)
}

// CreateItem creates a new news item.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/createnewsitem
func (s *Service) CreateItem(ctx context.Context, body *CreateItemRequest) (*Item, error) {
	result, err := s.CreateItemRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateItemResult(result)
}

// UpdateItem updates an existing news item by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/updatenewsitem
func (s *Service) UpdateItem(ctx context.Context, id string, body *UpdateItemRequest) (*Item, error) {
	result, err := s.UpdateItemRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseUpdateItemResult(result)
}

// DeleteItem deletes a news item by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/deletenewsitem
func (s *Service) DeleteItem(ctx context.Context, id string) (*ItemDeleted, error) {
	result, err := s.DeleteItemRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseDeleteItemResult(result)
}

// ListNewsfeeds returns a single page of newsfeeds.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/listnewsfeeds
func (s *Service) ListNewsfeeds(ctx context.Context, opts *api.ListOptions) (*api.PagedResult[Newsfeed], error) {
	result, err := s.ListNewsfeedsRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListNewsfeedsResult(result)
}

// GetNewsfeed retrieves a newsfeed by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/retrievenewsfeed
func (s *Service) GetNewsfeed(ctx context.Context, id string) (*Newsfeed, error) {
	result, err := s.GetNewsfeedRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetNewsfeedResult(result)
}

// ListNewsfeedItems returns a single page of live news items for a specific newsfeed.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/listnewsfeeditems
func (s *Service) ListNewsfeedItems(ctx context.Context, newsfeedID string, opts *api.ListOptions) (*api.PagedResult[Item], error) {
	result, err := s.ListNewsfeedItemsRaw(ctx, newsfeedID, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListNewsfeedItemsResult(result)
}

// --- Raw Methods ---

// ListItemsRaw returns a single page of news items with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/listnewsitems
func (s *Service) ListItemsRaw(ctx context.Context, opts *api.ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("news/news_items", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetItemRaw retrieves a news item by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/retrievenewsitem
func (s *Service) GetItemRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("news/news_items/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateItemRaw creates a new news item with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/createnewsitem
func (s *Service) CreateItemRaw(ctx context.Context, body *CreateItemRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "news/news_items", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateItemRaw updates an existing news item by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/updatenewsitem
func (s *Service) UpdateItemRaw(ctx context.Context, id string, body *UpdateItemRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("news/news_items/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteItemRaw deletes a news item by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/deletenewsitem
func (s *Service) DeleteItemRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("news/news_items/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListNewsfeedsRaw returns a single page of newsfeeds with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/listnewsfeeds
func (s *Service) ListNewsfeedsRaw(ctx context.Context, opts *api.ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("news/newsfeeds", opts)
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
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/retrievenewsfeed
func (s *Service) GetNewsfeedRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("news/newsfeeds/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListNewsfeedItemsRaw returns a single page of live news items for a specific newsfeed with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/news/listnewsfeeditems
func (s *Service) ListNewsfeedItemsRaw(ctx context.Context, newsfeedID string, opts *api.ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions(fmt.Sprintf("news/newsfeeds/%s/items", url.PathEscape(newsfeedID)), opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
