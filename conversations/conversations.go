package conversations

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// Service handles communication with the conversation related
// methods of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new conversations Service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Conversation represents an Intercom conversation.
type Conversation struct {
	Type              string                `json:"type"`
	ID                string                `json:"id"`
	CreatedAt         int64                 `json:"created_at,omitempty"`
	UpdatedAt         int64                 `json:"updated_at,omitempty"`
	WaitingSince      *int64                `json:"waiting_since,omitempty"`
	SnoozedUntil      *int64                `json:"snoozed_until,omitempty"`
	Title             string                `json:"title,omitempty"`
	State             string                `json:"state,omitempty"`
	Open              bool                  `json:"open,omitempty"`
	Read              bool                  `json:"read,omitempty"`
	Priority          string                `json:"priority,omitempty"`
	AdminAssigneeID   string                `json:"admin_assignee_id,omitempty"`
	TeamAssigneeID    string                `json:"team_assignee_id,omitempty"`
	Source            *Source               `json:"source,omitempty"`
	Contacts          *api.ContactRefList   `json:"contacts,omitempty"`
	Tags              *api.TagRefList       `json:"tags,omitempty"`
	ConversationParts *PartList             `json:"conversation_parts,omitempty"`
	Statistics        map[string]any        `json:"statistics,omitempty"`
	CustomAttributes  map[string]any        `json:"custom_attributes,omitempty"`
	LinkedObjects     *api.LinkedObjectList `json:"linked_objects,omitempty"`
}

// Source represents the originating message of a conversation.
type Source struct {
	Type        string      `json:"type"`
	ID          string      `json:"id"`
	DeliveredAs string      `json:"delivered_as,omitempty"`
	Subject     string      `json:"subject,omitempty"`
	Body        string      `json:"body,omitempty"`
	Author      *api.Author `json:"author,omitempty"`
	Attachments []any       `json:"attachments,omitempty"`
	URL         string      `json:"url,omitempty"`
	Redacted    bool        `json:"redacted,omitempty"`
}

// PartList holds the list of conversation parts.
type PartList struct {
	Type       string     `json:"type"`
	Parts      []api.Part `json:"conversation_parts"`
	TotalCount int        `json:"total_count"`
}

// Message represents the response from creating a conversation.
type Message struct {
	Type           string `json:"type"`
	ID             string `json:"id"`
	CreatedAt      int64  `json:"created_at,omitempty"`
	Body           string `json:"body,omitempty"`
	MessageType    string `json:"message_type,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// From identifies the initiator of a conversation.
type From struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// CreateRequest represents the body for creating a conversation.
type CreateRequest struct {
	From           From     `json:"from"`
	Body           string   `json:"body"`
	AttachmentURLs []string `json:"attachment_urls,omitempty"`
	CreatedAt      *int64   `json:"created_at,omitempty"`
}

// UpdateRequest represents the body for updating a conversation.
type UpdateRequest struct {
	Read             *bool          `json:"read,omitempty"`
	Title            string         `json:"title,omitempty"`
	CustomAttributes map[string]any `json:"custom_attributes,omitempty"`
	CompanyID        string         `json:"company_id,omitempty"`
}

// ReplyRequest represents the body for replying to a conversation.
type ReplyRequest struct {
	MessageType       string             `json:"message_type"`
	Type              string             `json:"type"`
	Body              string             `json:"body,omitempty"`
	AdminID           string             `json:"admin_id,omitempty"`
	IntercomUserID    string             `json:"intercom_user_id,omitempty"`
	Email             string             `json:"email,omitempty"`
	UserID            string             `json:"user_id,omitempty"`
	CreatedAt         *int64             `json:"created_at,omitempty"`
	AttachmentURLs    []string           `json:"attachment_urls,omitempty"`
	AttachmentFiles   []AttachmentFile   `json:"attachment_files,omitempty"`
	ReplyOptions      []QuickReplyOption `json:"reply_options,omitempty"`
	SkipNotifications *bool              `json:"skip_notifications,omitempty"`
}

// AttachmentFile represents a file attachment in a conversation reply.
type AttachmentFile struct {
	ContentType string `json:"content_type,omitempty"`
	Data        string `json:"data,omitempty"`
	Name        string `json:"name,omitempty"`
}

// QuickReplyOption represents a quick reply option in a conversation.
type QuickReplyOption struct {
	Text string `json:"text"`
	UUID string `json:"uuid"`
}

// ManageRequest is the body for close/open/snooze/assign operations.
type ManageRequest struct {
	MessageType  string `json:"message_type"`
	Type         string `json:"type,omitempty"`
	AdminID      string `json:"admin_id"`
	Body         string `json:"body,omitempty"`
	AssigneeID   string `json:"assignee_id,omitempty"`
	SnoozedUntil *int64 `json:"snoozed_until,omitempty"`
}

// ConvertRequest is the body for converting a conversation to a ticket.
type ConvertRequest struct {
	TicketTypeID string         `json:"ticket_type_id"`
	Attributes   map[string]any `json:"attributes,omitempty"`
}

// ConvertedTicket represents the response from converting a conversation to a ticket.
type ConvertedTicket struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	TicketID string `json:"ticket_id,omitempty"`
}

// RedactRequest is the body for redacting a conversation part or source.
type RedactRequest struct {
	Type               string `json:"type"`
	ConversationID     string `json:"conversation_id"`
	ConversationPartID string `json:"conversation_part_id,omitempty"`
	SourceID           string `json:"source_id,omitempty"`
}

// Customer identifies a customer to attach/detach from a conversation.
type Customer struct {
	IntercomUserID string `json:"intercom_user_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	Email          string `json:"email,omitempty"`
}

// AttachContactRequest is the body for adding a customer.
type AttachContactRequest struct {
	AdminID  string   `json:"admin_id"`
	Customer Customer `json:"customer"`
}

// DetachContactRequest is the body for removing a customer.
type DetachContactRequest struct {
	AdminID string `json:"admin_id"`
}

// listResponse is the raw API response for conversation list/search.
// The Intercom API returns conversations in a "conversations" field instead of "data".
type listResponse struct {
	Type          string          `json:"type"`
	Conversations []Conversation  `json:"conversations"`
	TotalCount    int             `json:"total_count"`
	Pages         api.CursorPages `json:"pages"`
}

// toPagedResult converts to the standard PagedResult used by the iterator.
func (r *listResponse) toPagedResult() *api.PagedResult[Conversation] {
	return &api.PagedResult[Conversation]{
		Type:       r.Type,
		Data:       r.Conversations,
		TotalCount: r.TotalCount,
		Pages:      r.Pages,
	}
}

// ParseGetResult decodes a Result into a Conversation.
func ParseGetResult(r *api.Result) (*Conversation, error) { return api.Decode[Conversation](r) }

// ParseListResult decodes a Result into a PagedResult[Conversation].
// Handles the Intercom API's "conversations" field instead of "data".
func ParseListResult(r *api.Result) (*api.PagedResult[Conversation], error) {
	raw, err := api.Decode[listResponse](r)
	if err != nil {
		return nil, err
	}
	return raw.toPagedResult(), nil
}

// ParseCreateResult decodes a Result into a Message.
func ParseCreateResult(r *api.Result) (*Message, error) {
	return api.Decode[Message](r)
}

// ParseUpdateResult decodes a Result into a Conversation.
func ParseUpdateResult(r *api.Result) (*Conversation, error) { return api.Decode[Conversation](r) }

// ParseDeleteResult decodes a Result into a Deleted.
func ParseDeleteResult(r *api.Result) (*api.Deleted, error) {
	return api.Decode[api.Deleted](r)
}

// ParseSearchResult decodes a Result into a PagedResult[Conversation].
// Handles the Intercom API's "conversations" field instead of "data".
func ParseSearchResult(r *api.Result) (*api.PagedResult[Conversation], error) {
	raw, err := api.Decode[listResponse](r)
	if err != nil {
		return nil, err
	}
	return raw.toPagedResult(), nil
}

// ParseReplyResult decodes a Result into a Conversation.
func ParseReplyResult(r *api.Result) (*Conversation, error) { return api.Decode[Conversation](r) }

// ParseCloseResult decodes a Result into a Conversation.
func ParseCloseResult(r *api.Result) (*Conversation, error) { return api.Decode[Conversation](r) }

// ParseOpenResult decodes a Result into a Conversation.
func ParseOpenResult(r *api.Result) (*Conversation, error) { return api.Decode[Conversation](r) }

// ParseSnoozeResult decodes a Result into a Conversation.
func ParseSnoozeResult(r *api.Result) (*Conversation, error) { return api.Decode[Conversation](r) }

// ParseAssignResult decodes a Result into a Conversation.
func ParseAssignResult(r *api.Result) (*Conversation, error) { return api.Decode[Conversation](r) }

// ParseConvertResult decodes a Result into a ConvertedTicket.
func ParseConvertResult(r *api.Result) (*ConvertedTicket, error) {
	return api.Decode[ConvertedTicket](r)
}

// ParseRedactResult decodes a Result into a Conversation.
func ParseRedactResult(r *api.Result) (*Conversation, error) { return api.Decode[Conversation](r) }

// ParseAddCustomerResult decodes a Result into a Conversation.
func ParseAddCustomerResult(r *api.Result) (*Conversation, error) {
	return api.Decode[Conversation](r)
}

// ParseRemoveCustomerResult decodes a Result into a Conversation.
func ParseRemoveCustomerResult(r *api.Result) (*Conversation, error) {
	return api.Decode[Conversation](r)
}

// ParseAddTagResult decodes a Result into a TagRef.
func ParseAddTagResult(r *api.Result) (*api.TagRef, error) { return api.Decode[api.TagRef](r) }

// ParseRemoveTagResult decodes a Result into a TagRef.
func ParseRemoveTagResult(r *api.Result) (*api.TagRef, error) { return api.Decode[api.TagRef](r) }

// Get retrieves a conversation by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/retrieveconversation
func (s *Service) Get(ctx context.Context, id string) (*Conversation, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// List returns a single page of conversations.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/listconversations
func (s *Service) List(ctx context.Context, opts *api.ListOptions) (*api.PagedResult[Conversation], error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListResult(result)
}

// ListAll returns an iterator over all conversations, handling pagination automatically.
func (s *Service) ListAll(ctx context.Context, opts *api.ListOptions) *api.Iter[Conversation] {
	return api.NewIter[Conversation](ctx, opts, s.List)
}

// Create creates a new conversation initiated by a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/createconversation
func (s *Service) Create(ctx context.Context, body *CreateRequest) (*Message, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateResult(result)
}

// Update updates an existing conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/updateconversation
func (s *Service) Update(ctx context.Context, id string, body *UpdateRequest) (*Conversation, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseUpdateResult(result)
}

// Delete permanently deletes a conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/deleteconversation
func (s *Service) Delete(ctx context.Context, id string) (*api.Deleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseDeleteResult(result)
}

// Search searches for conversations using the provided query filters.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/searchconversations
func (s *Service) Search(ctx context.Context, body *api.SearchRequest) (*api.PagedResult[Conversation], error) {
	result, err := s.SearchRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseSearchResult(result)
}

// Reply adds a reply to a conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/replyconversation
func (s *Service) Reply(ctx context.Context, id string, body *ReplyRequest) (*Conversation, error) {
	result, err := s.ReplyRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseReplyResult(result)
}

// Close closes a conversation. The request body should have MessageType "close".
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *Service) Close(ctx context.Context, id string, body *ManageRequest) (*Conversation, error) {
	result, err := s.CloseRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCloseResult(result)
}

// Open opens a snoozed or closed conversation. The request body should have MessageType "open".
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *Service) Open(ctx context.Context, id string, body *ManageRequest) (*Conversation, error) {
	result, err := s.OpenRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseOpenResult(result)
}

// Snooze snoozes a conversation until a given time. The request body should have MessageType "snoozed".
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *Service) Snooze(ctx context.Context, id string, body *ManageRequest) (*Conversation, error) {
	result, err := s.SnoozeRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseSnoozeResult(result)
}

// Assign assigns a conversation to an admin or team. The request body should have MessageType "assignment".
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *Service) Assign(ctx context.Context, id string, body *ManageRequest) (*Conversation, error) {
	result, err := s.AssignRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAssignResult(result)
}

// Convert converts a conversation to a ticket.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/convertconversationtoticket
func (s *Service) Convert(ctx context.Context, id string, body *ConvertRequest) (*ConvertedTicket, error) {
	result, err := s.ConvertRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseConvertResult(result)
}

// Redact redacts a conversation part or source message.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/redactconversation
func (s *Service) Redact(ctx context.Context, body *RedactRequest) (*Conversation, error) {
	result, err := s.RedactRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseRedactResult(result)
}

// AddCustomer attaches a contact as a participant to a conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/attachcontacttoconversation
func (s *Service) AddCustomer(ctx context.Context, conversationID string, body *AttachContactRequest) (*Conversation, error) {
	result, err := s.AddCustomerRaw(ctx, conversationID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAddCustomerResult(result)
}

// RemoveCustomer detaches a contact from a conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/detachcontactfromconversation
func (s *Service) RemoveCustomer(ctx context.Context, conversationID, contactID string, body *DetachContactRequest) (*Conversation, error) {
	result, err := s.RemoveCustomerRaw(ctx, conversationID, contactID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseRemoveCustomerResult(result)
}

// AddTag adds a tag to a conversation. Requires both the tag ID and admin ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/attachtagtoconversation
func (s *Service) AddTag(ctx context.Context, conversationID, tagID, adminID string) (*api.TagRef, error) {
	result, err := s.AddTagRaw(ctx, conversationID, tagID, adminID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAddTagResult(result)
}

// RemoveTag removes a tag from a conversation. Requires the admin ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/detachtagfromconversation
func (s *Service) RemoveTag(ctx context.Context, conversationID, tagID, adminID string) (*api.TagRef, error) {
	result, err := s.RemoveTagRaw(ctx, conversationID, tagID, adminID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseRemoveTagResult(result)
}

// GetRaw retrieves a conversation by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/retrieveconversation
func (s *Service) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("conversations/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns a single page of conversations with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/listconversations
func (s *Service) ListRaw(ctx context.Context, opts *api.ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("conversations", opts)
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
func (s *Service) CreateRaw(ctx context.Context, body *CreateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "conversations", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/updateconversation
func (s *Service) UpdateRaw(ctx context.Context, id string, body *UpdateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("conversations/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw permanently deletes a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/deleteconversation
func (s *Service) DeleteRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("conversations/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for conversations and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/searchconversations
func (s *Service) SearchRaw(ctx context.Context, body *api.SearchRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "conversations/search", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ReplyRaw adds a reply to a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/replyconversation
func (s *Service) ReplyRaw(ctx context.Context, id string, body *ReplyRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("conversations/%s/reply", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CloseRaw closes a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *Service) CloseRaw(ctx context.Context, id string, body *ManageRequest) (*api.Result, error) {
	return s.managePartsRaw(ctx, id, body)
}

// OpenRaw opens a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *Service) OpenRaw(ctx context.Context, id string, body *ManageRequest) (*api.Result, error) {
	return s.managePartsRaw(ctx, id, body)
}

// SnoozeRaw snoozes a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *Service) SnoozeRaw(ctx context.Context, id string, body *ManageRequest) (*api.Result, error) {
	return s.managePartsRaw(ctx, id, body)
}

// AssignRaw assigns a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/manageconversation
func (s *Service) AssignRaw(ctx context.Context, id string, body *ManageRequest) (*api.Result, error) {
	return s.managePartsRaw(ctx, id, body)
}

// managePartsRaw sends a POST to the /parts endpoint and returns the full HTTP result.
func (s *Service) managePartsRaw(ctx context.Context, id string, body *ManageRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("conversations/%s/parts", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ConvertRaw converts a conversation to a ticket and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/convertconversationtoticket
func (s *Service) ConvertRaw(ctx context.Context, id string, body *ConvertRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("conversations/%s/convert", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// RedactRaw redacts a conversation part or source and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/redactconversation
func (s *Service) RedactRaw(ctx context.Context, body *RedactRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "conversations/redact", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddCustomerRaw attaches a contact to a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/attachcontacttoconversation
func (s *Service) AddCustomerRaw(ctx context.Context, conversationID string, body *AttachContactRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("conversations/%s/customers", url.PathEscape(conversationID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// RemoveCustomerRaw detaches a contact from a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/detachcontactfromconversation
func (s *Service) RemoveCustomerRaw(ctx context.Context, conversationID, contactID string, body *DetachContactRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("conversations/%s/customers/%s", url.PathEscape(conversationID), url.PathEscape(contactID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddTagRaw adds a tag to a conversation and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/attachtagtoconversation
func (s *Service) AddTagRaw(ctx context.Context, conversationID, tagID, adminID string) (*api.Result, error) {
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
func (s *Service) RemoveTagRaw(ctx context.Context, conversationID, tagID, adminID string) (*api.Result, error) {
	body := struct {
		AdminID string `json:"admin_id"`
	}{AdminID: adminID}
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("conversations/%s/tags/%s", url.PathEscape(conversationID), url.PathEscape(tagID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
