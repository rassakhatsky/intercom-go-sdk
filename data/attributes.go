package data

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// AttributesService handles communication with the data attribute related
// methods of the Intercom API.
type AttributesService struct {
	client api.Caller
}

// NewAttributesService creates a new data attributes service.
func NewAttributesService(c api.Caller) *AttributesService {
	return &AttributesService{client: c}
}

// Attribute represents an Intercom data attribute.
type Attribute struct {
	Type              string   `json:"type"`
	ID                int      `json:"id,omitempty"`
	Model             string   `json:"model"`
	Name              string   `json:"name"`
	FullName          string   `json:"full_name,omitempty"`
	Label             string   `json:"label,omitempty"`
	Description       string   `json:"description,omitempty"`
	DataType          string   `json:"data_type,omitempty"`
	Options           []string `json:"options,omitempty"`
	APIWritable       bool     `json:"api_writable,omitempty"`
	UIWritable        bool     `json:"ui_writable,omitempty"`
	MessengerWritable bool     `json:"messenger_writable,omitempty"`
	Custom            bool     `json:"custom,omitempty"`
	Archived          bool     `json:"archived,omitempty"`
	CreatedAt         int64    `json:"created_at,omitempty"`
	UpdatedAt         int64    `json:"updated_at,omitempty"`
	AdminID           string   `json:"admin_id,omitempty"`
}

// AttributeList represents a list of data attributes.
type AttributeList struct {
	Type string      `json:"type"`
	Data []Attribute `json:"data"`
}

// ListAttributesOptions specifies optional parameters to the List method.
type ListAttributesOptions struct {
	Model           string `url:"model,omitempty"`
	IncludeArchived *bool  `url:"include_archived,omitempty"`
}

// AttributeOption represents an option value for list-type data attributes.
type AttributeOption struct {
	Value string `json:"value"`
}

// CreateAttributeRequest represents the request body for creating a data attribute.
type CreateAttributeRequest struct {
	Name              string            `json:"name"`
	Model             string            `json:"model"`
	DataType          string            `json:"data_type"`
	Description       string            `json:"description,omitempty"`
	MessengerWritable *bool             `json:"messenger_writable,omitempty"`
	Options           []AttributeOption `json:"options,omitempty"`
}

// UpdateAttributeRequest represents the request body for updating a data attribute.
type UpdateAttributeRequest struct {
	Description       string            `json:"description,omitempty"`
	Archived          *bool             `json:"archived,omitempty"`
	MessengerWritable *bool             `json:"messenger_writable,omitempty"`
	Options           []AttributeOption `json:"options,omitempty"`
}

// --- Parse Functions ---

// ParseListResult decodes a Result into an AttributeList.
func ParseAttributeListResult(r *api.Result) (*AttributeList, error) {
	return api.Decode[AttributeList](r)
}

// ParseCreateResult decodes a Result into an Attribute.
func ParseAttributeCreateResult(r *api.Result) (*Attribute, error) {
	return api.Decode[Attribute](r)
}

// ParseUpdateResult decodes a Result into an Attribute.
func ParseAttributeUpdateResult(r *api.Result) (*Attribute, error) {
	return api.Decode[Attribute](r)
}

// --- Regular Methods ---

// List returns all data attributes for the workspace.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/lisdataattributes
func (s *AttributesService) List(ctx context.Context, opts *ListAttributesOptions) (*AttributeList, error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAttributeListResult(result)
}

// Create creates a new data attribute.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/createdataattribute
func (s *AttributesService) Create(ctx context.Context, body *CreateAttributeRequest) (*Attribute, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAttributeCreateResult(result)
}

// Update updates a data attribute by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/updatedataattribute
func (s *AttributesService) Update(ctx context.Context, id int, body *UpdateAttributeRequest) (*Attribute, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAttributeUpdateResult(result)
}

// --- Raw Methods ---

// ListRaw returns all data attributes for the workspace with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/lisdataattributes
func (s *AttributesService) ListRaw(ctx context.Context, opts *ListAttributesOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("data_attributes", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates a new data attribute with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/createdataattribute
func (s *AttributesService) CreateRaw(ctx context.Context, body *CreateAttributeRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "data_attributes", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates a data attribute by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/updatedataattribute
func (s *AttributesService) UpdateRaw(ctx context.Context, id int, body *UpdateAttributeRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("data_attributes/%d", id), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
