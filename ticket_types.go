package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// TicketTypesService handles communication with the ticket type related
// methods of the Intercom API.
type TicketTypesService service

// TicketType represents an Intercom ticket type.
type TicketType struct {
	Type        string                   `json:"type"`
	ID          string                   `json:"id"`
	Category    string                   `json:"category,omitempty"`
	Name        string                   `json:"name,omitempty"`
	Description string                   `json:"description,omitempty"`
	Icon        string                   `json:"icon,omitempty"`
	WorkspaceID string                   `json:"workspace_id,omitempty"`
	Attributes  *TicketTypeAttributeList `json:"ticket_type_attributes,omitempty"`
	Archived    bool                     `json:"archived,omitempty"`
	CreatedAt   int64                    `json:"created_at,omitempty"`
	UpdatedAt   int64                    `json:"updated_at,omitempty"`
}

// TicketTypeAttribute represents an attribute on a ticket type.
type TicketTypeAttribute struct {
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

// TicketTypeAttributeList holds a list of ticket type attributes.
type TicketTypeAttributeList struct {
	Type string                `json:"type"`
	Data []TicketTypeAttribute `json:"data"`
}

// TicketTypeList holds a list of ticket types.
type TicketTypeList struct {
	Type string       `json:"type"`
	Data []TicketType `json:"data"`
}

// CreateTicketTypeRequest represents the body for creating a ticket type.
type CreateTicketTypeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Icon        string `json:"icon,omitempty"`
	IsInternal  bool   `json:"is_internal,omitempty"`
}

// UpdateTicketTypeRequest represents the body for updating a ticket type.
type UpdateTicketTypeRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Archived    *bool  `json:"archived,omitempty"`
	IsInternal  *bool  `json:"is_internal,omitempty"`
}

// CreateTicketTypeAttributeRequest represents the body for creating a ticket type attribute.
type CreateTicketTypeAttributeRequest struct {
	Name                        string `json:"name"`
	Description                 string `json:"description"`
	DataType                    string `json:"data_type"`
	RequiredToCreate            bool   `json:"required_to_create,omitempty"`
	RequiredToCreateForContacts bool   `json:"required_to_create_for_contacts,omitempty"`
	VisibleOnCreate             bool   `json:"visible_on_create,omitempty"`
	VisibleToContacts           bool   `json:"visible_to_contacts,omitempty"`
	Multiline                   bool   `json:"multiline,omitempty"`
	ListItems                   string `json:"list_items,omitempty"`
	AllowMultipleValues         bool   `json:"allow_multiple_values,omitempty"`
}

// UpdateTicketTypeAttributeRequest represents the body for updating a ticket type attribute.
type UpdateTicketTypeAttributeRequest struct {
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

// ParseTicketTypeGetResult decodes a Result into a TicketType.
func ParseTicketTypeGetResult(r *Result) (*TicketType, error) {
	return Decode[TicketType](r)
}

// ParseTicketTypeListResult decodes a Result into a TicketTypeList.
func ParseTicketTypeListResult(r *Result) (*TicketTypeList, error) {
	return Decode[TicketTypeList](r)
}

// ParseTicketTypeCreateResult decodes a Result into a TicketType.
func ParseTicketTypeCreateResult(r *Result) (*TicketType, error) {
	return Decode[TicketType](r)
}

// ParseTicketTypeUpdateResult decodes a Result into a TicketType.
func ParseTicketTypeUpdateResult(r *Result) (*TicketType, error) {
	return Decode[TicketType](r)
}

// ParseTicketTypeCreateAttributeResult decodes a Result into a TicketTypeAttribute.
func ParseTicketTypeCreateAttributeResult(r *Result) (*TicketTypeAttribute, error) {
	return Decode[TicketTypeAttribute](r)
}

// ParseTicketTypeUpdateAttributeResult decodes a Result into a TicketTypeAttribute.
func ParseTicketTypeUpdateAttributeResult(r *Result) (*TicketTypeAttribute, error) {
	return Decode[TicketTypeAttribute](r)
}

// --- Regular Methods ---

// Get retrieves a ticket type by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/gettickettype
func (s *TicketTypesService) Get(ctx context.Context, id string) (*TicketType, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketTypeGetResult(result)
}

// List returns all ticket types.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/listtickettypes
func (s *TicketTypesService) List(ctx context.Context) (*TicketTypeList, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketTypeListResult(result)
}

// Create creates a new ticket type.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/createtickettype
func (s *TicketTypesService) Create(ctx context.Context, body *CreateTicketTypeRequest) (*TicketType, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketTypeCreateResult(result)
}

// Update updates an existing ticket type.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/updatetickettype
func (s *TicketTypesService) Update(ctx context.Context, id string, body *UpdateTicketTypeRequest) (*TicketType, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketTypeUpdateResult(result)
}

// CreateAttribute creates a new attribute for a ticket type.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-type-attributes/createtickettypeattribute
func (s *TicketTypesService) CreateAttribute(ctx context.Context, ticketTypeID string, body *CreateTicketTypeAttributeRequest) (*TicketTypeAttribute, error) {
	result, err := s.CreateAttributeRaw(ctx, ticketTypeID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketTypeCreateAttributeResult(result)
}

// UpdateAttribute updates an existing attribute for a ticket type.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-type-attributes/updatetickettypeattribute
func (s *TicketTypesService) UpdateAttribute(ctx context.Context, ticketTypeID, attributeID string, body *UpdateTicketTypeAttributeRequest) (*TicketTypeAttribute, error) {
	result, err := s.UpdateAttributeRaw(ctx, ticketTypeID, attributeID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketTypeUpdateAttributeResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a ticket type by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/gettickettype
func (s *TicketTypesService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("ticket_types/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all ticket types with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/listtickettypes
func (s *TicketTypesService) ListRaw(ctx context.Context) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "ticket_types", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates a new ticket type with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/createtickettype
func (s *TicketTypesService) CreateRaw(ctx context.Context, body *CreateTicketTypeRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "ticket_types", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing ticket type with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-types/updatetickettype
func (s *TicketTypesService) UpdateRaw(ctx context.Context, id string, body *UpdateTicketTypeRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("ticket_types/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateAttributeRaw creates a new attribute for a ticket type with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-type-attributes/createtickettypeattribute
func (s *TicketTypesService) CreateAttributeRaw(ctx context.Context, ticketTypeID string, body *CreateTicketTypeAttributeRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("ticket_types/%s/attributes", url.PathEscape(ticketTypeID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateAttributeRaw updates an existing attribute for a ticket type with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-type-attributes/updatetickettypeattribute
func (s *TicketTypesService) UpdateAttributeRaw(ctx context.Context, ticketTypeID, attributeID string, body *UpdateTicketTypeAttributeRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("ticket_types/%s/attributes/%s", url.PathEscape(ticketTypeID), url.PathEscape(attributeID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
