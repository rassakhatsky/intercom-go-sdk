package tickets

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// Service handles communication with the ticket related
// methods of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new tickets Service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Ticket represents an Intercom ticket.
type Ticket struct {
	Type             string            `json:"type"`
	ID               string            `json:"id"`
	TicketID         string            `json:"ticket_id,omitempty"`
	Category         string            `json:"category,omitempty"`
	TicketAttributes map[string]any    `json:"ticket_attributes,omitempty"`
	TicketState      *StateRef         `json:"ticket_state,omitempty"`
	TicketType       *TypeRef          `json:"ticket_type,omitempty"`
	Contacts         *ContactList      `json:"contacts,omitempty"`
	AdminAssigneeID  string            `json:"admin_assignee_id,omitempty"`
	TeamAssigneeID   string            `json:"team_assignee_id,omitempty"`
	CreatedAt        int64             `json:"created_at,omitempty"`
	UpdatedAt        int64             `json:"updated_at,omitempty"`
	Open             bool              `json:"open,omitempty"`
	SnoozedUntil     *int64            `json:"snoozed_until,omitempty"`
	LinkedObjects    *api.LinkedObjectList `json:"linked_objects,omitempty"`
	TicketParts      *PartList         `json:"ticket_parts,omitempty"`
	IsShared         bool              `json:"is_shared,omitempty"`
}

// StateRef is a reference to a ticket state within a ticket.
type StateRef struct {
	Type          string `json:"type"`
	ID            string `json:"id"`
	Category      string `json:"category,omitempty"`
	InternalLabel string `json:"internal_label,omitempty"`
	ExternalLabel string `json:"external_label,omitempty"`
}

// TypeRef is a reference to a ticket type within a ticket.
type TypeRef struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// ContactList holds the contacts associated with a ticket.
type ContactList struct {
	Type     string        `json:"type"`
	Contacts []ContactItem `json:"contacts"`
}

// ContactItem is a contact reference within a ticket.
type ContactItem struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
}

// PartList holds the list of ticket parts.
type PartList struct {
	Type       string     `json:"type"`
	Parts      []api.Part `json:"ticket_parts"`
	TotalCount int        `json:"total_count"`
}



// Deleted represents the response from deleting a ticket.
type Deleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// ContactRef identifies a contact when creating a ticket.
type ContactRef struct {
	ID         string `json:"id,omitempty"`
	ExternalID string `json:"external_id,omitempty"`
	Email      string `json:"email,omitempty"`
}

// Assignment specifies assignment when creating a ticket.
type Assignment struct {
	AdminAssigneeID string `json:"admin_assignee_id,omitempty"`
	TeamAssigneeID  string `json:"team_assignee_id,omitempty"`
}

// CreateRequest represents the body for creating a ticket.
type CreateRequest struct {
	TicketTypeID         string         `json:"ticket_type_id"`
	Contacts             []ContactRef   `json:"contacts"`
	ConversationToLinkID string         `json:"conversation_to_link_id,omitempty"`
	CompanyID            string         `json:"company_id,omitempty"`
	CreatedAt            *int64         `json:"created_at,omitempty"`
	TicketAttributes     map[string]any `json:"ticket_attributes,omitempty"`
	Assignment           *Assignment    `json:"assignment,omitempty"`
	SkipNotifications    *bool          `json:"skip_notifications,omitempty"`
}

// UpdateRequest represents the body for updating a ticket.
type UpdateRequest struct {
	TicketAttributes  map[string]any `json:"ticket_attributes,omitempty"`
	TicketStateID     string         `json:"ticket_state_id,omitempty"`
	CompanyID         string         `json:"company_id,omitempty"`
	Open              *bool          `json:"open,omitempty"`
	IsShared          *bool          `json:"is_shared,omitempty"`
	SnoozedUntil      *int64         `json:"snoozed_until,omitempty"`
	AdminID           string         `json:"admin_id,omitempty"`
	AssigneeID        string         `json:"assignee_id,omitempty"`
	SkipNotifications *bool          `json:"skip_notifications,omitempty"`
}

// ReplyRequest represents the body for replying to a ticket.
type ReplyRequest struct {
	MessageType    string        `json:"message_type"`
	Type           string        `json:"type"`
	Body           string        `json:"body,omitempty"`
	AdminID        string        `json:"admin_id,omitempty"`
	IntercomUserID string        `json:"intercom_user_id,omitempty"`
	Email          string        `json:"email,omitempty"`
	UserID         string        `json:"user_id,omitempty"`
	CreatedAt      *int64        `json:"created_at,omitempty"`
	AttachmentURLs []string      `json:"attachment_urls,omitempty"`
	ReplyOptions   []ReplyOption `json:"reply_options,omitempty"`
}

// ReplyOption represents a quick reply option.
type ReplyOption struct {
	Text string `json:"text"`
	UUID string `json:"uuid"`
}

// EnqueuedJob represents the response from an async enqueue operation.
type EnqueuedJob struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// listResponse is the raw API response for ticket search.
type listResponse struct {
	Type       string          `json:"type"`
	Tickets    []Ticket        `json:"tickets"`
	TotalCount int             `json:"total_count"`
	Pages      api.CursorPages `json:"pages"`
}

// toPagedResult converts to the standard PagedResult used by the iterator.
func (r *listResponse) toPagedResult() *api.PagedResult[Ticket] {
	return &api.PagedResult[Ticket]{
		Type:       r.Type,
		Data:       r.Tickets,
		TotalCount: r.TotalCount,
		Pages:      r.Pages,
	}
}

// --- Parse Functions ---

// ParseGetResult decodes a Result into a Ticket.
func ParseGetResult(r *api.Result) (*Ticket, error) {
	return api.Decode[Ticket](r)
}

// ParseCreateResult decodes a Result into a Ticket.
func ParseCreateResult(r *api.Result) (*Ticket, error) {
	return api.Decode[Ticket](r)
}

// ParseUpdateResult decodes a Result into a Ticket.
func ParseUpdateResult(r *api.Result) (*Ticket, error) {
	return api.Decode[Ticket](r)
}

// ParseDeleteResult decodes a Result into a Deleted.
func ParseDeleteResult(r *api.Result) (*Deleted, error) {
	return api.Decode[Deleted](r)
}

// ParseSearchResult decodes a Result into a PagedResult[Ticket].
// Handles the Intercom API's "tickets" field instead of "data".
func ParseSearchResult(r *api.Result) (*api.PagedResult[Ticket], error) {
	raw, err := api.Decode[listResponse](r)
	if err != nil {
		return nil, err
	}
	return raw.toPagedResult(), nil
}

// ParseReplyResult decodes a Result into a Part.
func ParseReplyResult(r *api.Result) (*api.Part, error) {
	return api.Decode[api.Part](r)
}

// ParseAddTagResult decodes a Result into a TagRef.
func ParseAddTagResult(r *api.Result) (*api.TagRef, error) {
	return api.Decode[api.TagRef](r)
}

// ParseRemoveTagResult decodes a Result into a TagRef.
func ParseRemoveTagResult(r *api.Result) (*api.TagRef, error) {
	return api.Decode[api.TagRef](r)
}

// ParseEnqueueResult decodes a Result into an EnqueuedJob.
func ParseEnqueueResult(r *api.Result) (*EnqueuedJob, error) {
	return api.Decode[EnqueuedJob](r)
}

// --- Regular Methods ---

// Get retrieves a ticket by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/getticket
func (s *Service) Get(ctx context.Context, id string) (*Ticket, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// Create creates a new ticket.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/createticket
func (s *Service) Create(ctx context.Context, body *CreateRequest) (*Ticket, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateResult(result)
}

// Update updates an existing ticket.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/updateticket
func (s *Service) Update(ctx context.Context, id string, body *UpdateRequest) (*Ticket, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseUpdateResult(result)
}

// Delete permanently deletes a ticket.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/deleteticket
func (s *Service) Delete(ctx context.Context, id string) (*Deleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseDeleteResult(result)
}

// Search searches for tickets using the provided query filters.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/searchtickets
func (s *Service) Search(ctx context.Context, body *api.SearchRequest) (*api.PagedResult[Ticket], error) {
	result, err := s.SearchRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseSearchResult(result)
}

// Reply adds a reply to a ticket.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/replyticket
func (s *Service) Reply(ctx context.Context, id string, body *ReplyRequest) (*api.Part, error) {
	result, err := s.ReplyRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseReplyResult(result)
}

// AddTag adds a tag to a ticket. Requires both the tag ID and admin ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/attachtagtoticket
func (s *Service) AddTag(ctx context.Context, ticketID, tagID, adminID string) (*api.TagRef, error) {
	result, err := s.AddTagRaw(ctx, ticketID, tagID, adminID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAddTagResult(result)
}

// RemoveTag removes a tag from a ticket. Requires the admin ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/detachtagfromticket
func (s *Service) RemoveTag(ctx context.Context, ticketID, tagID, adminID string) (*api.TagRef, error) {
	result, err := s.RemoveTagRaw(ctx, ticketID, tagID, adminID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseRemoveTagResult(result)
}

// Enqueue asynchronously creates a ticket, returning a job reference.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/enqueuecreateticket
func (s *Service) Enqueue(ctx context.Context, body *CreateRequest) (*EnqueuedJob, error) {
	result, err := s.EnqueueRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseEnqueueResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a ticket by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/getticket
func (s *Service) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("tickets/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates a new ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/createticket
func (s *Service) CreateRaw(ctx context.Context, body *CreateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tickets", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/updateticket
func (s *Service) UpdateRaw(ctx context.Context, id string, body *UpdateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("tickets/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw permanently deletes a ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/deleteticket
func (s *Service) DeleteRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("tickets/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for tickets and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/searchtickets
func (s *Service) SearchRaw(ctx context.Context, body *api.SearchRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tickets/search", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ReplyRaw adds a reply to a ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/replyticket
func (s *Service) ReplyRaw(ctx context.Context, id string, body *ReplyRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("tickets/%s/reply", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddTagRaw adds a tag to a ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/attachtagtoticket
func (s *Service) AddTagRaw(ctx context.Context, ticketID, tagID, adminID string) (*api.Result, error) {
	body := struct {
		ID      string `json:"id"`
		AdminID string `json:"admin_id"`
	}{ID: tagID, AdminID: adminID}
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("tickets/%s/tags", url.PathEscape(ticketID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// RemoveTagRaw removes a tag from a ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/detachtagfromticket
func (s *Service) RemoveTagRaw(ctx context.Context, ticketID, tagID, adminID string) (*api.Result, error) {
	body := struct {
		AdminID string `json:"admin_id"`
	}{AdminID: adminID}
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("tickets/%s/tags/%s", url.PathEscape(ticketID), url.PathEscape(tagID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// EnqueueRaw asynchronously creates a ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/enqueuecreateticket
func (s *Service) EnqueueRaw(ctx context.Context, body *CreateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tickets/enqueue", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
