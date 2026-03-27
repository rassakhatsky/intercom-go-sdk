package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// HelpCenterService handles communication with the help center related methods
// of the Intercom API.
type HelpCenterService service

// Collection represents an Intercom help center collection.
type Collection struct {
	Type              string                    `json:"type"`
	ID                string                    `json:"id"`
	WorkspaceID       string                    `json:"workspace_id,omitempty"`
	Name              string                    `json:"name,omitempty"`
	Description       string                    `json:"description,omitempty"`
	CreatedAt         int64                     `json:"created_at,omitempty"`
	UpdatedAt         int64                     `json:"updated_at,omitempty"`
	URL               string                    `json:"url,omitempty"`
	Icon              string                    `json:"icon,omitempty"`
	Order             int                       `json:"order,omitempty"`
	DefaultLocale     string                    `json:"default_locale,omitempty"`
	TranslatedContent *ArticleTranslatedContent `json:"translated_content,omitempty"`
	ParentID          string                    `json:"parent_id,omitempty"`
	HelpCenterID      int                       `json:"help_center_id,omitempty"`
}

// CollectionDeleted represents the response from deleting a collection.
type CollectionDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// CreateCollectionRequest represents the request body for creating a collection.
type CreateCollectionRequest struct {
	Name              string                    `json:"name"`
	Description       string                    `json:"description,omitempty"`
	TranslatedContent *ArticleTranslatedContent `json:"translated_content,omitempty"`
	ParentID          string                    `json:"parent_id,omitempty"`
	HelpCenterID      *int                      `json:"help_center_id,omitempty"`
}

// UpdateCollectionRequest represents the request body for updating a collection.
type UpdateCollectionRequest struct {
	Name              string                    `json:"name,omitempty"`
	Description       string                    `json:"description,omitempty"`
	TranslatedContent *ArticleTranslatedContent `json:"translated_content,omitempty"`
	ParentID          string                    `json:"parent_id,omitempty"`
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

// ParseHelpCenterListCollectionsResult decodes a Result into a PagedResult[Collection].
func ParseHelpCenterListCollectionsResult(r *Result) (*PagedResult[Collection], error) {
	return Decode[PagedResult[Collection]](r)
}

// ParseHelpCenterGetCollectionResult decodes a Result into a Collection.
func ParseHelpCenterGetCollectionResult(r *Result) (*Collection, error) {
	return Decode[Collection](r)
}

// ParseHelpCenterCreateCollectionResult decodes a Result into a Collection.
func ParseHelpCenterCreateCollectionResult(r *Result) (*Collection, error) {
	return Decode[Collection](r)
}

// ParseHelpCenterUpdateCollectionResult decodes a Result into a Collection.
func ParseHelpCenterUpdateCollectionResult(r *Result) (*Collection, error) {
	return Decode[Collection](r)
}

// ParseHelpCenterDeleteCollectionResult decodes a Result into a CollectionDeleted.
func ParseHelpCenterDeleteCollectionResult(r *Result) (*CollectionDeleted, error) {
	return Decode[CollectionDeleted](r)
}

// ParseHelpCenterListHelpCentersResult decodes a Result into a HelpCenterList.
func ParseHelpCenterListHelpCentersResult(r *Result) (*HelpCenterList, error) {
	return Decode[HelpCenterList](r)
}

// ParseHelpCenterGetHelpCenterResult decodes a Result into a HelpCenter.
func ParseHelpCenterGetHelpCenterResult(r *Result) (*HelpCenter, error) {
	return Decode[HelpCenter](r)
}

// --- Regular Methods ---

// ListCollections returns a single page of collections.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/listallcollections
func (s *HelpCenterService) ListCollections(ctx context.Context, opts *ListOptions) (*PagedResult[Collection], error) {
	result, err := s.ListCollectionsRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseHelpCenterListCollectionsResult(result)
}

// GetCollection retrieves a collection by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/retrievecollection
func (s *HelpCenterService) GetCollection(ctx context.Context, id string) (*Collection, error) {
	result, err := s.GetCollectionRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseHelpCenterGetCollectionResult(result)
}

// CreateCollection creates a new collection.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/createcollection
func (s *HelpCenterService) CreateCollection(ctx context.Context, body *CreateCollectionRequest) (*Collection, error) {
	result, err := s.CreateCollectionRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseHelpCenterCreateCollectionResult(result)
}

// UpdateCollection updates an existing collection by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/updatecollection
func (s *HelpCenterService) UpdateCollection(ctx context.Context, id string, body *UpdateCollectionRequest) (*Collection, error) {
	result, err := s.UpdateCollectionRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseHelpCenterUpdateCollectionResult(result)
}

// DeleteCollection deletes a collection by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/deletecollection
func (s *HelpCenterService) DeleteCollection(ctx context.Context, id string) (*CollectionDeleted, error) {
	result, err := s.DeleteCollectionRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseHelpCenterDeleteCollectionResult(result)
}

// ListHelpCenters returns all help centers.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/listhelpcenters
func (s *HelpCenterService) ListHelpCenters(ctx context.Context) (*HelpCenterList, error) {
	result, err := s.ListHelpCentersRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseHelpCenterListHelpCentersResult(result)
}

// GetHelpCenter retrieves a help center by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/retrievehelpcenter
func (s *HelpCenterService) GetHelpCenter(ctx context.Context, id string) (*HelpCenter, error) {
	result, err := s.GetHelpCenterRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseHelpCenterGetHelpCenterResult(result)
}

// --- Raw Methods ---

// ListCollectionsRaw returns a single page of collections with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/listallcollections
func (s *HelpCenterService) ListCollectionsRaw(ctx context.Context, opts *ListOptions) (*Result, error) {
	path, err := addQueryOptions("help_center/collections", opts)
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
func (s *HelpCenterService) GetCollectionRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("help_center/collections/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateCollectionRaw creates a new collection with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/createcollection
func (s *HelpCenterService) CreateCollectionRaw(ctx context.Context, body *CreateCollectionRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "help_center/collections", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateCollectionRaw updates an existing collection by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/updatecollection
func (s *HelpCenterService) UpdateCollectionRaw(ctx context.Context, id string, body *UpdateCollectionRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("help_center/collections/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteCollectionRaw deletes a collection by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/deletecollection
func (s *HelpCenterService) DeleteCollectionRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("help_center/collections/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListHelpCentersRaw returns all help centers with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/listhelpcenters
func (s *HelpCenterService) ListHelpCentersRaw(ctx context.Context) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "help_center/help_centers", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetHelpCenterRaw retrieves a help center by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/help-center/retrievehelpcenter
func (s *HelpCenterService) GetHelpCenterRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("help_center/help_centers/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
