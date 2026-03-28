package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// TicketsService handles communication with the ticket related
// methods of the Intercom API.
type TicketsService service

// Ticket represents an Intercom ticket.
type Ticket struct {
	Type             string             `json:"type"`
	ID               string             `json:"id"`
	TicketID         string             `json:"ticket_id,omitempty"`
	Category         string             `json:"category,omitempty"`
	TicketAttributes map[string]any     `json:"ticket_attributes,omitempty"`
	TicketState      *TicketStateRef    `json:"ticket_state,omitempty"`
	TicketType       *TicketTypeRef     `json:"ticket_type,omitempty"`
	Contacts         *TicketContactList `json:"contacts,omitempty"`
	AdminAssigneeID  string             `json:"admin_assignee_id,omitempty"`
	TeamAssigneeID   string             `json:"team_assignee_id,omitempty"`
	CreatedAt        int64              `json:"created_at,omitempty"`
	UpdatedAt        int64              `json:"updated_at,omitempty"`
	Open             bool               `json:"open,omitempty"`
	SnoozedUntil     *int64             `json:"snoozed_until,omitempty"`
	LinkedObjects    *LinkedObjectList  `json:"linked_objects,omitempty"`
	TicketParts      *TicketPartList    `json:"ticket_parts,omitempty"`
	IsShared         bool               `json:"is_shared,omitempty"`
}

// TicketStateRef is a reference to a ticket state within a ticket.
type TicketStateRef struct {
	Type          string `json:"type"`
	ID            string `json:"id"`
	Category      string `json:"category,omitempty"`
	InternalLabel string `json:"internal_label,omitempty"`
	ExternalLabel string `json:"external_label,omitempty"`
}

// TicketTypeRef is a reference to a ticket type within a ticket.
type TicketTypeRef struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// TicketContactList holds the contacts associated with a ticket.
type TicketContactList struct {
	Type     string              `json:"type"`
	Contacts []TicketContactItem `json:"contacts"`
}

// TicketContactItem is a contact reference within a ticket.
type TicketContactItem struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
}

// TicketPartList holds the list of ticket parts.
type TicketPartList struct {
	Type       string       `json:"type"`
	Parts      []TicketPart `json:"ticket_parts"`
	TotalCount int          `json:"total_count"`
}

// TicketPart represents a single message/event in a ticket.
type TicketPart struct {
	Type       string              `json:"type"`
	ID         string              `json:"id"`
	PartType   string              `json:"part_type,omitempty"`
	Body       string              `json:"body,omitempty"`
	CreatedAt  int64               `json:"created_at,omitempty"`
	UpdatedAt  int64               `json:"updated_at,omitempty"`
	AssignedTo *ConversationAuthor `json:"assigned_to,omitempty"`
	Author     *ConversationAuthor `json:"author,omitempty"`
	ExternalID string              `json:"external_id,omitempty"`
	Redacted   bool                `json:"redacted,omitempty"`
}

// TicketDeleted represents the response from deleting a ticket.
type TicketDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// TicketContactRef identifies a contact when creating a ticket.
type TicketContactRef struct {
	ID         string `json:"id,omitempty"`
	ExternalID string `json:"external_id,omitempty"`
	Email      string `json:"email,omitempty"`
}

// TicketAssignment specifies assignment when creating a ticket.
type TicketAssignment struct {
	AdminAssigneeID string `json:"admin_assignee_id,omitempty"`
	TeamAssigneeID  string `json:"team_assignee_id,omitempty"`
}

// CreateTicketRequest represents the body for creating a ticket.
type CreateTicketRequest struct {
	TicketTypeID         string             `json:"ticket_type_id"`
	Contacts             []TicketContactRef `json:"contacts"`
	ConversationToLinkID string             `json:"conversation_to_link_id,omitempty"`
	CompanyID            string             `json:"company_id,omitempty"`
	CreatedAt            *int64             `json:"created_at,omitempty"`
	TicketAttributes     map[string]any     `json:"ticket_attributes,omitempty"`
	Assignment           *TicketAssignment  `json:"assignment,omitempty"`
}

// UpdateTicketRequest represents the body for updating a ticket.
type UpdateTicketRequest struct {
	TicketAttributes map[string]any `json:"ticket_attributes,omitempty"`
	TicketStateID    string         `json:"ticket_state_id,omitempty"`
	CompanyID        string         `json:"company_id,omitempty"`
	Open             *bool          `json:"open,omitempty"`
	IsShared         *bool          `json:"is_shared,omitempty"`
	SnoozedUntil     *int64         `json:"snoozed_until,omitempty"`
	AdminID          *int           `json:"admin_id,omitempty"`
	AssigneeID       string         `json:"assignee_id,omitempty"`
}

// ReplyTicketRequest represents the body for replying to a ticket.
type ReplyTicketRequest struct {
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

// ticketListResponse is the raw API response for ticket search.
type ticketListResponse struct {
	Type       string      `json:"type"`
	Tickets    []Ticket    `json:"tickets"`
	TotalCount int         `json:"total_count"`
	Pages      CursorPages `json:"pages"`
}

// toPagedResult converts to the standard PagedResult used by the iterator.
func (r *ticketListResponse) toPagedResult() *PagedResult[Ticket] {
	return &PagedResult[Ticket]{
		Type:       r.Type,
		Data:       r.Tickets,
		TotalCount: r.TotalCount,
		Pages:      r.Pages,
	}
}

// --- Parse Functions ---

// ParseTicketGetResult decodes a Result into a Ticket.
func ParseTicketGetResult(r *Result) (*Ticket, error) {
	return Decode[Ticket](r)
}

// ParseTicketCreateResult decodes a Result into a Ticket.
func ParseTicketCreateResult(r *Result) (*Ticket, error) {
	return Decode[Ticket](r)
}

// ParseTicketUpdateResult decodes a Result into a Ticket.
func ParseTicketUpdateResult(r *Result) (*Ticket, error) {
	return Decode[Ticket](r)
}

// ParseTicketDeleteResult decodes a Result into a TicketDeleted.
func ParseTicketDeleteResult(r *Result) (*TicketDeleted, error) {
	return Decode[TicketDeleted](r)
}

// ParseTicketSearchResult decodes a Result into a PagedResult[Ticket].
// Handles the Intercom API's "tickets" field instead of "data".
func ParseTicketSearchResult(r *Result) (*PagedResult[Ticket], error) {
	raw, err := Decode[ticketListResponse](r)
	if err != nil {
		return nil, err
	}
	return raw.toPagedResult(), nil
}

// ParseTicketReplyResult decodes a Result into a TicketPart.
func ParseTicketReplyResult(r *Result) (*TicketPart, error) {
	return Decode[TicketPart](r)
}

// ParseTicketAddTagResult decodes a Result into a TagRef.
func ParseTicketAddTagResult(r *Result) (*TagRef, error) {
	return Decode[TagRef](r)
}

// ParseTicketRemoveTagResult decodes a Result into a TagRef.
func ParseTicketRemoveTagResult(r *Result) (*TagRef, error) {
	return Decode[TagRef](r)
}

// ParseTicketEnqueueResult decodes a Result into an EnqueuedJob.
func ParseTicketEnqueueResult(r *Result) (*EnqueuedJob, error) {
	return Decode[EnqueuedJob](r)
}

// --- Regular Methods ---

// Get retrieves a ticket by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/getticket
func (s *TicketsService) Get(ctx context.Context, id string) (*Ticket, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketGetResult(result)
}

// Create creates a new ticket.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/createticket
func (s *TicketsService) Create(ctx context.Context, body *CreateTicketRequest) (*Ticket, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketCreateResult(result)
}

// Update updates an existing ticket.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/updateticket
func (s *TicketsService) Update(ctx context.Context, id string, body *UpdateTicketRequest) (*Ticket, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketUpdateResult(result)
}

// Delete permanently deletes a ticket.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/deleteticket
func (s *TicketsService) Delete(ctx context.Context, id string) (*TicketDeleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketDeleteResult(result)
}

// Search searches for tickets using the provided query filters.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/searchtickets
func (s *TicketsService) Search(ctx context.Context, body *SearchRequest) (*PagedResult[Ticket], error) {
	result, err := s.SearchRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketSearchResult(result)
}

// Reply adds a reply to a ticket.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/replyticket
func (s *TicketsService) Reply(ctx context.Context, id string, body *ReplyTicketRequest) (*TicketPart, error) {
	result, err := s.ReplyRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketReplyResult(result)
}

// AddTag adds a tag to a ticket. Requires both the tag ID and admin ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/attachtagtoticket
func (s *TicketsService) AddTag(ctx context.Context, ticketID, tagID, adminID string) (*TagRef, error) {
	result, err := s.AddTagRaw(ctx, ticketID, tagID, adminID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketAddTagResult(result)
}

// RemoveTag removes a tag from a ticket. Requires the admin ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/detachtagfromticket
func (s *TicketsService) RemoveTag(ctx context.Context, ticketID, tagID, adminID string) (*TagRef, error) {
	result, err := s.RemoveTagRaw(ctx, ticketID, tagID, adminID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketRemoveTagResult(result)
}

// Enqueue asynchronously creates a ticket, returning a job reference.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/enqueuecreateticket
func (s *TicketsService) Enqueue(ctx context.Context, body *CreateTicketRequest) (*EnqueuedJob, error) {
	result, err := s.EnqueueRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketEnqueueResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a ticket by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/getticket
func (s *TicketsService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("tickets/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates a new ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/createticket
func (s *TicketsService) CreateRaw(ctx context.Context, body *CreateTicketRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tickets", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/updateticket
func (s *TicketsService) UpdateRaw(ctx context.Context, id string, body *UpdateTicketRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("tickets/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw permanently deletes a ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/deleteticket
func (s *TicketsService) DeleteRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("tickets/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for tickets and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/searchtickets
func (s *TicketsService) SearchRaw(ctx context.Context, body *SearchRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tickets/search", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ReplyRaw adds a reply to a ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tickets/replyticket
func (s *TicketsService) ReplyRaw(ctx context.Context, id string, body *ReplyTicketRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("tickets/%s/reply", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddTagRaw adds a tag to a ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/attachtagtoticket
func (s *TicketsService) AddTagRaw(ctx context.Context, ticketID, tagID, adminID string) (*Result, error) {
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
func (s *TicketsService) RemoveTagRaw(ctx context.Context, ticketID, tagID, adminID string) (*Result, error) {
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
func (s *TicketsService) EnqueueRaw(ctx context.Context, body *CreateTicketRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "tickets/enqueue", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
