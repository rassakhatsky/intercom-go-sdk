package tickets

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// TypesService handles communication with the ticket type related
// methods of the Intercom API.
type TypesService struct {
	client api.Caller
}

// NewTypesService creates a new ticket types TypesService.
func NewTypesService(c api.Caller) *TypesService {
	return &TypesService{client: c}
}

// Type represents an Intercom ticket type.
type Type struct {
	Type        string             `json:"type"`
	ID          string             `json:"id"`
	Category    string             `json:"category,omitempty"`
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	Icon        string             `json:"icon,omitempty"`
	WorkspaceID string             `json:"workspace_id,omitempty"`
	Attributes  *TypeAttributeList `json:"ticket_type_attributes,omitempty"`
	Archived    bool               `json:"archived,omitempty"`
	CreatedAt   int64              `json:"created_at,omitempty"`
	UpdatedAt   int64              `json:"updated_at,omitempty"`
}

// TypeAttribute represents an attribute on a ticket type.
type TypeAttribute struct {
	Type                        string        `json:"type"`
	ID                          string        `json:"id"`
	WorkspaceID                 string        `json:"workspace_id,omitempty"`
	Name                        string        `json:"name,omitempty"`
	Description                 string        `json:"description,omitempty"`
	DataType                    string        `json:"data_type,omitempty"`
	InputOptions                *InputOptions `json:"input_options,omitempty"`
	Order                       int           `json:"order,omitempty"`
	RequiredToCreate            bool          `json:"required_to_create,omitempty"`
	RequiredToCreateForContacts bool          `json:"required_to_create_for_contacts,omitempty"`
	VisibleOnCreate             bool          `json:"visible_on_create,omitempty"`
	VisibleToContacts           bool          `json:"visible_to_contacts,omitempty"`
	Default                     bool          `json:"default,omitempty"`
	TicketTypeID                int           `json:"ticket_type_id,omitempty"`
	Archived                    bool          `json:"archived,omitempty"`
	CreatedAt                   int64         `json:"created_at,omitempty"`
	UpdatedAt                   int64         `json:"updated_at,omitempty"`
}

// InputOptions holds configuration for a ticket type attribute's input.
type InputOptions struct {
	Multiline bool `json:"multiline,omitempty"`
}

// TypeAttributeList holds a list of ticket type attributes.
type TypeAttributeList struct {
	Type string          `json:"type"`
	Data []TypeAttribute `json:"data"`
}

// TypeList holds a list of ticket types.
type TypeList struct {
	Type string `json:"type"`
	Data []Type `json:"data"`
}

// CreateTypeRequest represents the body for creating a ticket type.
type CreateTypeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Icon        string `json:"icon,omitempty"`
	IsInternal  bool   `json:"is_internal,omitempty"`
}

// UpdateTypeRequest represents the body for updating a ticket type.
type UpdateTypeRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Archived    *bool  `json:"archived,omitempty"`
	IsInternal  *bool  `json:"is_internal,omitempty"`
}

// CreateTypeAttributeRequest represents the body for creating a ticket type attribute.
type CreateTypeAttributeRequest struct {
	Name                        string `json:"name"`
	Description                 string `json:"description"`
	DataType                    string `json:"data_type"`
	RequiredToCreate            *bool  `json:"required_to_create,omitempty"`
	RequiredToCreateForContacts *bool  `json:"required_to_create_for_contacts,omitempty"`
	VisibleOnCreate             *bool  `json:"visible_on_create,omitempty"`
	VisibleToContacts           *bool  `json:"visible_to_contacts,omitempty"`
	Multiline                   *bool  `json:"multiline,omitempty"`
	ListItems                   string `json:"list_items,omitempty"`
	AllowMultipleValues         *bool  `json:"allow_multiple_values,omitempty"`
}

// UpdateTypeAttributeRequest represents the body for updating a ticket type attribute.
type UpdateTypeAttributeRequest struct {
	Name                        string `json:"name,omitempty"`
	Description                 string `json:"description,omitempty"`
	RequiredToCreate            *bool  `json:"required_to_create,omitempty"`
	RequiredToCreateForContacts *bool  `json:"required_to_create_for_contacts,omitempty"`
	VisibleOnCreate             *bool  `json:"visible_on_create,omitempty"`
	VisibleToContacts           *bool  `json:"visible_to_contacts,omitempty"`
	Multiline                   *bool  `json:"multiline,omitempty"`
	ListItems                   string `json:"list_items,omitempty"`
	AllowMultipleValues         *bool  `json:"allow_multiple_values,omitempty"`
	Archived                    *bool  `json:"archived,omitempty"`
}

// --- Parse Functions ---

// ParseTypeGetResult decodes a Result into a Type.
func ParseTypeGetResult(r *api.Result) (*Type, error) {
	return api.Decode[Type](r)
}

// ParseTypeListResult decodes a Result into a TypeList.
func ParseTypeListResult(r *api.Result) (*TypeList, error) {
	return api.Decode[TypeList](r)
}

// ParseTypeCreateResult decodes a Result into a Type.
func ParseTypeCreateResult(r *api.Result) (*Type, error) {
	return api.Decode[Type](r)
}

// ParseTypeUpdateResult decodes a Result into a Type.
func ParseTypeUpdateResult(r *api.Result) (*Type, error) {
	return api.Decode[Type](r)
}

// ParseTypeCreateAttributeResult decodes a Result into a TypeAttribute.
func ParseTypeCreateAttributeResult(r *api.Result) (*TypeAttribute, error) {
	return api.Decode[TypeAttribute](r)
}

// ParseTypeUpdateAttributeResult decodes a Result into a TypeAttribute.
func ParseTypeUpdateAttributeResult(r *api.Result) (*TypeAttribute, error) {
	return api.Decode[TypeAttribute](r)
}

// --- Regular Methods ---

// Get retrieves a ticket type by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/gettickettype
func (s *TypesService) Get(ctx context.Context, id string) (*Type, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseTypeGetResult(result)
}

// List returns all ticket types.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/listtickettypes
func (s *TypesService) List(ctx context.Context) (*TypeList, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseTypeListResult(result)
}

// Create creates a new ticket type.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/createtickettype
func (s *TypesService) Create(ctx context.Context, body *CreateTypeRequest) (*Type, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseTypeCreateResult(result)
}

// Update updates an existing ticket type.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/updatetickettype
func (s *TypesService) Update(ctx context.Context, id string, body *UpdateTypeRequest) (*Type, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseTypeUpdateResult(result)
}

// CreateAttribute creates a new attribute for a ticket type.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-type-attributes/createtickettypeattribute
func (s *TypesService) CreateAttribute(ctx context.Context, ticketTypeID string, body *CreateTypeAttributeRequest) (*TypeAttribute, error) {
	result, err := s.CreateAttributeRaw(ctx, ticketTypeID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseTypeCreateAttributeResult(result)
}

// UpdateAttribute updates an existing attribute for a ticket type.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-type-attributes/updatetickettypeattribute
func (s *TypesService) UpdateAttribute(ctx context.Context, ticketTypeID, attributeID string, body *UpdateTypeAttributeRequest) (*TypeAttribute, error) {
	result, err := s.UpdateAttributeRaw(ctx, ticketTypeID, attributeID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseTypeUpdateAttributeResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a ticket type by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/gettickettype
func (s *TypesService) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("ticket_types/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all ticket types with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/listtickettypes
func (s *TypesService) ListRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "ticket_types", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates a new ticket type with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/createtickettype
func (s *TypesService) CreateRaw(ctx context.Context, body *CreateTypeRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "ticket_types", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing ticket type with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/updatetickettype
func (s *TypesService) UpdateRaw(ctx context.Context, id string, body *UpdateTypeRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("ticket_types/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateAttributeRaw creates a new attribute for a ticket type with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-type-attributes/createtickettypeattribute
func (s *TypesService) CreateAttributeRaw(ctx context.Context, ticketTypeID string, body *CreateTypeAttributeRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("ticket_types/%s/attributes", url.PathEscape(ticketTypeID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateAttributeRaw updates an existing attribute for a ticket type with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-type-attributes/updatetickettypeattribute
func (s *TypesService) UpdateAttributeRaw(ctx context.Context, ticketTypeID, attributeID string, body *UpdateTypeAttributeRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("ticket_types/%s/attributes/%s", url.PathEscape(ticketTypeID), url.PathEscape(attributeID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
