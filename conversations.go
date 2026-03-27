package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ConversationsService handles communication with the conversation related
// methods of the Intercom API.
type ConversationsService service

// Conversation represents an Intercom conversation.
type Conversation struct {
	Type              string                   `json:"type"`
	ID                string                   `json:"id"`
	CreatedAt         int64                    `json:"created_at,omitempty"`
	UpdatedAt         int64                    `json:"updated_at,omitempty"`
	WaitingSince      *int64                   `json:"waiting_since,omitempty"`
	SnoozedUntil      *int64                   `json:"snoozed_until,omitempty"`
	Title             string                   `json:"title,omitempty"`
	State             string                   `json:"state,omitempty"`
	Open              bool                     `json:"open,omitempty"`
	Read              bool                     `json:"read,omitempty"`
	Priority          string                   `json:"priority,omitempty"`
	AdminAssigneeID   string                   `json:"admin_assignee_id,omitempty"`
	TeamAssigneeID    string                   `json:"team_assignee_id,omitempty"`
	Source            *ConversationSource      `json:"source,omitempty"`
	Contacts          *ConversationContactList `json:"contacts,omitempty"`
	Tags              *ConversationTagList     `json:"tags,omitempty"`
	ConversationParts *ConversationPartList    `json:"conversation_parts,omitempty"`
	Statistics        map[string]any           `json:"statistics,omitempty"`
	CustomAttributes  map[string]any           `json:"custom_attributes,omitempty"`
	LinkedObjects     *LinkedObjectList        `json:"linked_objects,omitempty"`
}

// ConversationSource represents the originating message of a conversation.
type ConversationSource struct {
	Type        string              `json:"type"`
	ID          string              `json:"id"`
	DeliveredAs string              `json:"delivered_as,omitempty"`
	Subject     string              `json:"subject,omitempty"`
	Body        string              `json:"body,omitempty"`
	Author      *ConversationAuthor `json:"author,omitempty"`
	Attachments []any               `json:"attachments,omitempty"`
	URL         string              `json:"url,omitempty"`
	Redacted    bool                `json:"redacted,omitempty"`
}

// ConversationAuthor represents the author of a conversation part or source.
type ConversationAuthor struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// ConversationContactList is the contacts participating in a conversation.
type ConversationContactList struct {
	Type     string                `json:"type"`
	Contacts []ConversationContact `json:"contacts"`
}

// ConversationContact is a contact reference within a conversation.
type ConversationContact struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
}

// ConversationTagList is the list of tags on a conversation.
type ConversationTagList struct {
	Type string               `json:"type"`
	Tags []ConversationTagRef `json:"tags"`
}

// ConversationTagRef is a tag reference within a conversation.
type ConversationTagRef struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	AppliedAt int64  `json:"applied_at,omitempty"`
}

// ConversationPartList holds the list of conversation parts.
type ConversationPartList struct {
	Type       string             `json:"type"`
	Parts      []ConversationPart `json:"conversation_parts"`
	TotalCount int                `json:"total_count"`
}

// ConversationPart represents a single message/action in a conversation.
type ConversationPart struct {
	Type       string              `json:"type"`
	ID         string              `json:"id"`
	PartType   string              `json:"part_type,omitempty"`
	Body       string              `json:"body,omitempty"`
	CreatedAt  int64               `json:"created_at,omitempty"`
	UpdatedAt  int64               `json:"updated_at,omitempty"`
	NotifiedAt int64               `json:"notified_at,omitempty"`
	AssignedTo *ConversationAuthor `json:"assigned_to,omitempty"`
	Author     *ConversationAuthor `json:"author,omitempty"`
	ExternalID string              `json:"external_id,omitempty"`
	Redacted   bool                `json:"redacted,omitempty"`
}

// LinkedObjectList holds linked objects for a conversation.
type LinkedObjectList struct {
	Type       string `json:"type"`
	Data       []any  `json:"data"`
	TotalCount int    `json:"total_count"`
	HasMore    bool   `json:"has_more"`
}

// ConversationDeleted represents the response from deleting a conversation.
type ConversationDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// ConversationMessage represents the response from creating a conversation.
type ConversationMessage struct {
	Type           string `json:"type"`
	ID             string `json:"id"`
	CreatedAt      int64  `json:"created_at,omitempty"`
	Body           string `json:"body,omitempty"`
	MessageType    string `json:"message_type,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// ConversationFrom identifies the initiator of a conversation.
type ConversationFrom struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// CreateConversationRequest represents the body for creating a conversation.
type CreateConversationRequest struct {
	From           ConversationFrom `json:"from"`
	Body           string           `json:"body"`
	AttachmentURLs []string         `json:"attachment_urls,omitempty"`
	CreatedAt      *int64           `json:"created_at,omitempty"`
}

// UpdateConversationRequest represents the body for updating a conversation.
type UpdateConversationRequest struct {
	Read             *bool          `json:"read,omitempty"`
	Title            string         `json:"title,omitempty"`
	CustomAttributes map[string]any `json:"custom_attributes,omitempty"`
	CompanyID        string         `json:"company_id,omitempty"`
}

// ReplyConversationRequest represents the body for replying to a conversation.
type ReplyConversationRequest struct {
	MessageType    string   `json:"message_type"`
	Type           string   `json:"type"`
	Body           string   `json:"body,omitempty"`
	AdminID        string   `json:"admin_id,omitempty"`
	IntercomUserID string   `json:"intercom_user_id,omitempty"`
	Email          string   `json:"email,omitempty"`
	UserID         string   `json:"user_id,omitempty"`
	CreatedAt      *int64   `json:"created_at,omitempty"`
	AttachmentURLs []string `json:"attachment_urls,omitempty"`
}

// ManageConversationRequest is the body for close/open/snooze/assign operations.
type ManageConversationRequest struct {
	MessageType  string `json:"message_type"`
	Type         string `json:"type,omitempty"`
	AdminID      string `json:"admin_id"`
	Body         string `json:"body,omitempty"`
	AssigneeID   string `json:"assignee_id,omitempty"`
	SnoozedUntil int64  `json:"snoozed_until,omitempty"`
}

// ConvertConversationRequest is the body for converting a conversation to a ticket.
type ConvertConversationRequest struct {
	TicketTypeID string         `json:"ticket_type_id"`
	Attributes   map[string]any `json:"attributes,omitempty"`
}

// ConvertedTicket represents the response from converting a conversation to a ticket.
type ConvertedTicket struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	TicketID string `json:"ticket_id,omitempty"`
}

// RedactConversationRequest is the body for redacting a conversation part or source.
type RedactConversationRequest struct {
	Type               string `json:"type"`
	ConversationID     string `json:"conversation_id"`
	ConversationPartID string `json:"conversation_part_id,omitempty"`
	SourceID           string `json:"source_id,omitempty"`
}

// ConversationCustomer identifies a customer to attach/detach from a conversation.
type ConversationCustomer struct {
	IntercomUserID string `json:"intercom_user_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	Email          string `json:"email,omitempty"`
}

// AttachContactToConversationRequest is the body for adding a customer.
type AttachContactToConversationRequest struct {
	AdminID  string               `json:"admin_id"`
	Customer ConversationCustomer `json:"customer"`
}

// DetachContactFromConversationRequest is the body for removing a customer.
type DetachContactFromConversationRequest struct {
	AdminID string `json:"admin_id"`
}

// conversationListResponse is the raw API response for conversation list/search.
// The Intercom API returns conversations in a "conversations" field instead of "data".
type conversationListResponse struct {
	Type          string         `json:"type"`
	Conversations []Conversation `json:"conversations"`
	TotalCount    int            `json:"total_count"`
	Pages         CursorPages    `json:"pages"`
}

// toPagedResult converts to the standard PagedResult used by the iterator.
func (r *conversationListResponse) toPagedResult() *PagedResult[Conversation] {
	return &PagedResult[Conversation]{
		Type:       r.Type,
		Data:       r.Conversations,
		TotalCount: r.TotalCount,
		Pages:      r.Pages,
	}
}

// ParseConversationGetResult decodes a Result into a Conversation.
func ParseConversationGetResult(r *Result) (*Conversation, error) { return Decode[Conversation](r) }

// ParseConversationListResult decodes a Result into a PagedResult[Conversation].
// Handles the Intercom API's "conversations" field instead of "data".
func ParseConversationListResult(r *Result) (*PagedResult[Conversation], error) {
	raw, err := Decode[conversationListResponse](r)
	if err != nil {
		return nil, err
	}
	return raw.toPagedResult(), nil
}

// ParseConversationCreateResult decodes a Result into a ConversationMessage.
func ParseConversationCreateResult(r *Result) (*ConversationMessage, error) {
	return Decode[ConversationMessage](r)
}

// ParseConversationUpdateResult decodes a Result into a Conversation.
func ParseConversationUpdateResult(r *Result) (*Conversation, error) { return Decode[Conversation](r) }

// ParseConversationDeleteResult decodes a Result into a ConversationDeleted.
func ParseConversationDeleteResult(r *Result) (*ConversationDeleted, error) {
	return Decode[ConversationDeleted](r)
}

// ParseConversationSearchResult decodes a Result into a PagedResult[Conversation].
// Handles the Intercom API's "conversations" field instead of "data".
func ParseConversationSearchResult(r *Result) (*PagedResult[Conversation], error) {
	raw, err := Decode[conversationListResponse](r)
	if err != nil {
		return nil, err
	}
	return raw.toPagedResult(), nil
}

// ParseConversationReplyResult decodes a Result into a Conversation.
func ParseConversationReplyResult(r *Result) (*Conversation, error) { return Decode[Conversation](r) }

// ParseConversationCloseResult decodes a Result into a Conversation.
func ParseConversationCloseResult(r *Result) (*Conversation, error) { return Decode[Conversation](r) }

// ParseConversationOpenResult decodes a Result into a Conversation.
func ParseConversationOpenResult(r *Result) (*Conversation, error) { return Decode[Conversation](r) }

// ParseConversationSnoozeResult decodes a Result into a Conversation.
func ParseConversationSnoozeResult(r *Result) (*Conversation, error) { return Decode[Conversation](r) }

// ParseConversationAssignResult decodes a Result into a Conversation.
func ParseConversationAssignResult(r *Result) (*Conversation, error) { return Decode[Conversation](r) }

// ParseConversationConvertResult decodes a Result into a ConvertedTicket.
func ParseConversationConvertResult(r *Result) (*ConvertedTicket, error) {
	return Decode[ConvertedTicket](r)
}

// ParseConversationRedactResult decodes a Result into a Conversation.
func ParseConversationRedactResult(r *Result) (*Conversation, error) { return Decode[Conversation](r) }

// ParseConversationAddCustomerResult decodes a Result into a Conversation.
func ParseConversationAddCustomerResult(r *Result) (*Conversation, error) {
	return Decode[Conversation](r)
}

// ParseConversationRemoveCustomerResult decodes a Result into a Conversation.
func ParseConversationRemoveCustomerResult(r *Result) (*Conversation, error) {
	return Decode[Conversation](r)
}

// ParseConversationAddTagResult decodes a Result into a TagRef.
func ParseConversationAddTagResult(r *Result) (*TagRef, error) { return Decode[TagRef](r) }

// ParseConversationRemoveTagResult decodes a Result into a TagRef.
func ParseConversationRemoveTagResult(r *Result) (*TagRef, error) { return Decode[TagRef](r) }

// Get retrieves a conversation by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/retrieveconversation
func (s *ConversationsService) Get(ctx context.Context, id string) (*Conversation, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationGetResult(result)
}

// List returns a single page of conversations.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/listconversations
func (s *ConversationsService) List(ctx context.Context, opts *ListOptions) (*PagedResult[Conversation], error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationListResult(result)
}

// ListAll returns an iterator over all conversations, handling pagination automatically.
func (s *ConversationsService) ListAll(ctx context.Context, opts *ListOptions) *Iter[Conversation] {
	return NewIter[Conversation](ctx, opts, s.List)
}

// Create creates a new conversation initiated by a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/createconversation
func (s *ConversationsService) Create(ctx context.Context, body *CreateConversationRequest) (*ConversationMessage, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationCreateResult(result)
}

// Update updates an existing conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/updateconversation
func (s *ConversationsService) Update(ctx context.Context, id string, body *UpdateConversationRequest) (*Conversation, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationUpdateResult(result)
}

// Delete permanently deletes a conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/deleteconversation
func (s *ConversationsService) Delete(ctx context.Context, id string) (*ConversationDeleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationDeleteResult(result)
}

// Search searches for conversations using the provided query filters.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/searchconversations
func (s *ConversationsService) Search(ctx context.Context, body *SearchRequest) (*PagedResult[Conversation], error) {
	result, err := s.SearchRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationSearchResult(result)
}

// Reply adds a reply to a conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/replyconversation
func (s *ConversationsService) Reply(ctx context.Context, id string, body *ReplyConversationRequest) (*Conversation, error) {
	result, err := s.ReplyRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationReplyResult(result)
}

// Close closes a conversation. The request body should have MessageType "close".
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *ConversationsService) Close(ctx context.Context, id string, body *ManageConversationRequest) (*Conversation, error) {
	result, err := s.CloseRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationCloseResult(result)
}

// Open opens a snoozed or closed conversation. The request body should have MessageType "open".
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *ConversationsService) Open(ctx context.Context, id string, body *ManageConversationRequest) (*Conversation, error) {
	result, err := s.OpenRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationOpenResult(result)
}

// Snooze snoozes a conversation until a given time. The request body should have MessageType "snoozed".
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *ConversationsService) Snooze(ctx context.Context, id string, body *ManageConversationRequest) (*Conversation, error) {
	result, err := s.SnoozeRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationSnoozeResult(result)
}

// Assign assigns a conversation to an admin or team. The request body should have MessageType "assignment".
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *ConversationsService) Assign(ctx context.Context, id string, body *ManageConversationRequest) (*Conversation, error) {
	result, err := s.AssignRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationAssignResult(result)
}

// Convert converts a conversation to a ticket.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/convertconversationtoticket
func (s *ConversationsService) Convert(ctx context.Context, id string, body *ConvertConversationRequest) (*ConvertedTicket, error) {
	result, err := s.ConvertRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationConvertResult(result)
}

// Redact redacts a conversation part or source message.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/redactconversation
func (s *ConversationsService) Redact(ctx context.Context, body *RedactConversationRequest) (*Conversation, error) {
	result, err := s.RedactRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationRedactResult(result)
}

// AddCustomer attaches a contact as a participant to a conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/attachcontacttoconversation
func (s *ConversationsService) AddCustomer(ctx context.Context, conversationID string, body *AttachContactToConversationRequest) (*Conversation, error) {
	result, err := s.AddCustomerRaw(ctx, conversationID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationAddCustomerResult(result)
}

// RemoveCustomer detaches a contact from a conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/detachcontactfromconversation
func (s *ConversationsService) RemoveCustomer(ctx context.Context, conversationID, contactID string, body *DetachContactFromConversationRequest) (*Conversation, error) {
	result, err := s.RemoveCustomerRaw(ctx, conversationID, contactID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationRemoveCustomerResult(result)
}

// AddTag adds a tag to a conversation. Requires both the tag ID and admin ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/attachtagtoconversation
func (s *ConversationsService) AddTag(ctx context.Context, conversationID, tagID, adminID string) (*TagRef, error) {
	result, err := s.AddTagRaw(ctx, conversationID, tagID, adminID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationAddTagResult(result)
}

// RemoveTag removes a tag from a conversation. Requires the admin ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/detachtagfromconversation
func (s *ConversationsService) RemoveTag(ctx context.Context, conversationID, tagID, adminID string) (*TagRef, error) {
	result, err := s.RemoveTagRaw(ctx, conversationID, tagID, adminID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseConversationRemoveTagResult(result)
}

// GetRaw retrieves a conversation by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/retrieveconversation
func (s *ConversationsService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("conversations/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns a single page of conversations with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/listconversations
func (s *ConversationsService) ListRaw(ctx context.Context, opts *ListOptions) (*Result, error) {
	path, err := addQueryOptions("conversations", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates a new conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/createconversation
func (s *ConversationsService) CreateRaw(ctx context.Context, body *CreateConversationRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "conversations", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/updateconversation
func (s *ConversationsService) UpdateRaw(ctx context.Context, id string, body *UpdateConversationRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("conversations/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw permanently deletes a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/deleteconversation
func (s *ConversationsService) DeleteRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("conversations/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for conversations and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/searchconversations
func (s *ConversationsService) SearchRaw(ctx context.Context, body *SearchRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "conversations/search", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ReplyRaw adds a reply to a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/replyconversation
func (s *ConversationsService) ReplyRaw(ctx context.Context, id string, body *ReplyConversationRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("conversations/%s/reply", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CloseRaw closes a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *ConversationsService) CloseRaw(ctx context.Context, id string, body *ManageConversationRequest) (*Result, error) {
	return s.managePartsRaw(ctx, id, body)
}

// OpenRaw opens a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *ConversationsService) OpenRaw(ctx context.Context, id string, body *ManageConversationRequest) (*Result, error) {
	return s.managePartsRaw(ctx, id, body)
}

// SnoozeRaw snoozes a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *ConversationsService) SnoozeRaw(ctx context.Context, id string, body *ManageConversationRequest) (*Result, error) {
	return s.managePartsRaw(ctx, id, body)
}

// AssignRaw assigns a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *ConversationsService) AssignRaw(ctx context.Context, id string, body *ManageConversationRequest) (*Result, error) {
	return s.managePartsRaw(ctx, id, body)
}

// managePartsRaw sends a POST to the /parts endpoint and returns the full HTTP result.
func (s *ConversationsService) managePartsRaw(ctx context.Context, id string, body *ManageConversationRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("conversations/%s/parts", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ConvertRaw converts a conversation to a ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/convertconversationtoticket
func (s *ConversationsService) ConvertRaw(ctx context.Context, id string, body *ConvertConversationRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("conversations/%s/convert", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// RedactRaw redacts a conversation part or source and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/redactconversation
func (s *ConversationsService) RedactRaw(ctx context.Context, body *RedactConversationRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "conversations/redact", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddCustomerRaw attaches a contact to a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/attachcontacttoconversation
func (s *ConversationsService) AddCustomerRaw(ctx context.Context, conversationID string, body *AttachContactToConversationRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("conversations/%s/customers", url.PathEscape(conversationID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// RemoveCustomerRaw detaches a contact from a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/detachcontactfromconversation
func (s *ConversationsService) RemoveCustomerRaw(ctx context.Context, conversationID, contactID string, body *DetachContactFromConversationRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("conversations/%s/customers/%s", url.PathEscape(conversationID), url.PathEscape(contactID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddTagRaw adds a tag to a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/attachtagtoconversation
func (s *ConversationsService) AddTagRaw(ctx context.Context, conversationID, tagID, adminID string) (*Result, error) {
	body := struct {
		ID      string `json:"id"`
		AdminID string `json:"admin_id"`
	}{ID: tagID, AdminID: adminID}
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("conversations/%s/tags", url.PathEscape(conversationID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// RemoveTagRaw removes a tag from a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/detachtagfromconversation
func (s *ConversationsService) RemoveTagRaw(ctx context.Context, conversationID, tagID, adminID string) (*Result, error) {
	body := struct {
		AdminID string `json:"admin_id"`
	}{AdminID: adminID}
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("conversations/%s/tags/%s", url.PathEscape(conversationID), url.PathEscape(tagID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

