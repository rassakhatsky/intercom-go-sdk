package helpcenter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// Service handles communication with the help center related methods
// of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new help center Service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Collection represents an Intercom help center collection.
type Collection struct {
	Type              string                        `json:"type"`
	ID                string                        `json:"id"`
	WorkspaceID       string                        `json:"workspace_id,omitempty"`
	Name              string                        `json:"name,omitempty"`
	Description       string                        `json:"description,omitempty"`
	CreatedAt         int64                         `json:"created_at,omitempty"`
	UpdatedAt         int64                         `json:"updated_at,omitempty"`
	URL               string                        `json:"url,omitempty"`
	Icon              string                        `json:"icon,omitempty"`
	Order             int                           `json:"order,omitempty"`
	DefaultLocale     string                        `json:"default_locale,omitempty"`
	TranslatedContent *api.ArticleTranslatedContent `json:"translated_content,omitempty"`
	ParentID          string                        `json:"parent_id,omitempty"`
	HelpCenterID      int                           `json:"help_center_id,omitempty"`
}

// CollectionDeleted represents the response from deleting a collection.
type CollectionDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// CreateCollectionRequest represents the request body for creating a collection.
type CreateCollectionRequest struct {
	Name              string                        `json:"name"`
	Description       string                        `json:"description,omitempty"`
	TranslatedContent *api.ArticleTranslatedContent `json:"translated_content,omitempty"`
	ParentID          string                        `json:"parent_id,omitempty"`
	HelpCenterID      *int                          `json:"help_center_id,omitempty"`
}

// UpdateCollectionRequest represents the request body for updating a collection.
type UpdateCollectionRequest struct {
	Name              string                        `json:"name,omitempty"`
	Description       string                        `json:"description,omitempty"`
	TranslatedContent *api.ArticleTranslatedContent `json:"translated_content,omitempty"`
	ParentID          string                        `json:"parent_id,omitempty"`
}

// HelpCenter represents an Intercom help center.
type HelpCenter struct {
	Type            string `json:"type"`
	ID              string `json:"id"`
	WorkspaceID     string `json:"workspace_id,omitempty"`
	CreatedAt       int64  `json:"created_at,omitempty"`
	UpdatedAt       int64  `json:"updated_at,omitempty"`
	Identifier      string `json:"identifier,omitempty"`
	WebsiteTurnedOn bool   `json:"website_turned_on,omitempty"`
	DisplayName     string `json:"display_name,omitempty"`
	URL             string `json:"url,omitempty"`
	CustomDomain    string `json:"custom_domain,omitempty"`
}

// HelpCenterList represents a list of help centers.
type HelpCenterList struct {
	Type string       `json:"type"`
	Data []HelpCenter `json:"data"`
}

// --- Parse Functions ---

// ParseListCollectionsResult decodes a Result into a PagedResult[Collection].
func ParseListCollectionsResult(r *api.Result) (*api.PagedResult[Collection], error) {
	return api.Decode[api.PagedResult[Collection]](r)
}

// ParseGetCollectionResult decodes a Result into a Collection.
func ParseGetCollectionResult(r *api.Result) (*Collection, error) {
	return api.Decode[Collection](r)
}

// ParseCreateCollectionResult decodes a Result into a Collection.
func ParseCreateCollectionResult(r *api.Result) (*Collection, error) {
	return api.Decode[Collection](r)
}

// ParseUpdateCollectionResult decodes a Result into a Collection.
func ParseUpdateCollectionResult(r *api.Result) (*Collection, error) {
	return api.Decode[Collection](r)
}

// ParseDeleteCollectionResult decodes a Result into a CollectionDeleted.
func ParseDeleteCollectionResult(r *api.Result) (*CollectionDeleted, error) {
	return api.Decode[CollectionDeleted](r)
}

// ParseListHelpCentersResult decodes a Result into a HelpCenterList.
func ParseListHelpCentersResult(r *api.Result) (*HelpCenterList, error) {
	return api.Decode[HelpCenterList](r)
}

// ParseGetHelpCenterResult decodes a Result into a HelpCenter.
func ParseGetHelpCenterResult(r *api.Result) (*HelpCenter, error) {
	return api.Decode[HelpCenter](r)
}

// --- Regular Methods ---

// ListCollections returns a single page of collections.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/listallcollections
func (s *Service) ListCollections(ctx context.Context, opts *api.ListOptions) (*api.PagedResult[Collection], error) {
	result, err := s.ListCollectionsRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListCollectionsResult(result)
}

// GetCollection retrieves a collection by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/retrievecollection
func (s *Service) GetCollection(ctx context.Context, id string) (*Collection, error) {
	result, err := s.GetCollectionRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetCollectionResult(result)
}

// CreateCollection creates a new collection.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/createcollection
func (s *Service) CreateCollection(ctx context.Context, body *CreateCollectionRequest) (*Collection, error) {
	result, err := s.CreateCollectionRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateCollectionResult(result)
}

// UpdateCollection updates an existing collection by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/updatecollection
func (s *Service) UpdateCollection(ctx context.Context, id string, body *UpdateCollectionRequest) (*Collection, error) {
	result, err := s.UpdateCollectionRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseUpdateCollectionResult(result)
}

// DeleteCollection deletes a collection by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/deletecollection
func (s *Service) DeleteCollection(ctx context.Context, id string) (*CollectionDeleted, error) {
	result, err := s.DeleteCollectionRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseDeleteCollectionResult(result)
}

// ListHelpCenters returns all help centers.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/listhelpcenters
func (s *Service) ListHelpCenters(ctx context.Context) (*HelpCenterList, error) {
	result, err := s.ListHelpCentersRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListHelpCentersResult(result)
}

// GetHelpCenter retrieves a help center by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/retrievehelpcenter
func (s *Service) GetHelpCenter(ctx context.Context, id string) (*HelpCenter, error) {
	result, err := s.GetHelpCenterRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetHelpCenterResult(result)
}

// --- Raw Methods ---

// ListCollectionsRaw returns a single page of collections with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/listallcollections
func (s *Service) ListCollectionsRaw(ctx context.Context, opts *api.ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("help_center/collections", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetCollectionRaw retrieves a collection by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/retrievecollection
func (s *Service) GetCollectionRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("help_center/collections/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateCollectionRaw creates a new collection with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/createcollection
func (s *Service) CreateCollectionRaw(ctx context.Context, body *CreateCollectionRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "help_center/collections", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateCollectionRaw updates an existing collection by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/updatecollection
func (s *Service) UpdateCollectionRaw(ctx context.Context, id string, body *UpdateCollectionRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("help_center/collections/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteCollectionRaw deletes a collection by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/deletecollection
func (s *Service) DeleteCollectionRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("help_center/collections/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListHelpCentersRaw returns all help centers with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/listhelpcenters
func (s *Service) ListHelpCentersRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "help_center/help_centers", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetHelpCenterRaw retrieves a help center by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/retrievehelpcenter
func (s *Service) GetHelpCenterRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("help_center/help_centers/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
