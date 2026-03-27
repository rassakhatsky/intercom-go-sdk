package intercom

import (
	"context"
	"fmt"
	"net/http"
)

// DataAttributesService handles communication with the data attribute related
// methods of the Intercom API.
type DataAttributesService service

// DataAttribute represents an Intercom data attribute.
type DataAttribute struct {
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

// DataAttributeList represents a list of data attributes.
type DataAttributeList struct {
	Type string          `json:"type"`
	Data []DataAttribute `json:"data"`
}

// ListDataAttributesOptions specifies optional parameters to the List method.
type ListDataAttributesOptions struct {
	Model           string `url:"model,omitempty"`
	IncludeArchived *bool  `url:"include_archived,omitempty"`
}

// AttributeOption represents an option value for list-type data attributes.
type AttributeOption struct {
	Value string `json:"value"`
}

// CreateDataAttributeRequest represents the request body for creating a data attribute.
type CreateDataAttributeRequest struct {
	Name              string            `json:"name"`
	Model             string            `json:"model"`
	DataType          string            `json:"data_type"`
	Description       string            `json:"description,omitempty"`
	MessengerWritable *bool             `json:"messenger_writable,omitempty"`
	Options           []AttributeOption `json:"options,omitempty"`
}

// UpdateDataAttributeRequest represents the request body for updating a data attribute.
type UpdateDataAttributeRequest struct {
	Description       string            `json:"description,omitempty"`
	Archived          *bool             `json:"archived,omitempty"`
	MessengerWritable *bool             `json:"messenger_writable,omitempty"`
	Options           []AttributeOption `json:"options,omitempty"`
}

// --- Parse Functions ---

// ParseDataAttributeListResult decodes a Result into a DataAttributeList.
func ParseDataAttributeListResult(r *Result) (*DataAttributeList, error) {
	return Decode[DataAttributeList](r)
}

// ParseDataAttributeCreateResult decodes a Result into a DataAttribute.
func ParseDataAttributeCreateResult(r *Result) (*DataAttribute, error) {
	return Decode[DataAttribute](r)
}

// ParseDataAttributeUpdateResult decodes a Result into a DataAttribute.
func ParseDataAttributeUpdateResult(r *Result) (*DataAttribute, error) {
	return Decode[DataAttribute](r)
}

// --- Regular Methods ---

// List returns all data attributes for the workspace.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/lisdataattributes
func (s *DataAttributesService) List(ctx context.Context, opts *ListDataAttributesOptions) (*DataAttributeList, error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseDataAttributeListResult(result)
}

// Create creates a new data attribute.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/createdataattribute
func (s *DataAttributesService) Create(ctx context.Context, body *CreateDataAttributeRequest) (*DataAttribute, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseDataAttributeCreateResult(result)
}

// Update updates a data attribute by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/updatedataattribute
func (s *DataAttributesService) Update(ctx context.Context, id int, body *UpdateDataAttributeRequest) (*DataAttribute, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseDataAttributeUpdateResult(result)
}

// --- Raw Methods ---

// ListRaw returns all data attributes for the workspace with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/lisdataattributes
func (s *DataAttributesService) ListRaw(ctx context.Context, opts *ListDataAttributesOptions) (*Result, error) {
	path, err := addQueryOptions("data_attributes", opts)
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
func (s *DataAttributesService) CreateRaw(ctx context.Context, body *CreateDataAttributeRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "data_attributes", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates a data attribute by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-attributes/updatedataattribute
func (s *DataAttributesService) UpdateRaw(ctx context.Context, id int, body *UpdateDataAttributeRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("data_attributes/%d", id), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
