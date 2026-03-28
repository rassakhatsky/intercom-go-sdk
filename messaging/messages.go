package messaging

import (
	"context"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// MessagesService handles communication with the message related
// methods of the Intercom API.
type MessagesService struct {
	client api.Caller
}

// NewMessagesService creates a new MessagesService.
func NewMessagesService(c api.Caller) *MessagesService {
	return &MessagesService{client: c}
}

// Message represents an Intercom message.
type Message struct {
	Type           string `json:"type,omitempty"`
	ID             string `json:"id,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty"`
	Subject        string `json:"subject,omitempty"`
	Body           string `json:"body,omitempty"`
	MessageType    string `json:"message_type,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// MessageSender represents the sender of a message.
type MessageSender struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// MessageRecipient represents a recipient of a message.
type MessageRecipient struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// CreateMessageRequest represents the request body for creating a message.
type CreateMessageRequest struct {
	MessageType                           string           `json:"message_type"`
	Subject                               string           `json:"subject,omitempty"`
	Body                                  string           `json:"body,omitempty"`
	Template                              string           `json:"template,omitempty"`
	From                                  MessageSender    `json:"from"`
	To                                    MessageRecipient `json:"to"`
	CC                                    any              `json:"cc,omitempty"`
	BCC                                   any              `json:"bcc,omitempty"`
	CreatedAt                             int64            `json:"created_at,omitempty"`
	CreateConversationWithoutContactReply *bool            `json:"create_conversation_without_contact_reply,omitempty"`
}

// --- Parse Functions ---

// ParseCreateResult decodes a Result into a Message.
func ParseCreateResult(r *api.Result) (*Message, error) {
	return api.Decode[Message](r)
}

// --- Regular Methods ---

// Create creates a new message initiated by an admin.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/messages/createmessage
func (s *MessagesService) Create(ctx context.Context, body *CreateMessageRequest) (*Message, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateResult(result)
}

// --- Raw Methods ---

// CreateRaw creates a new message initiated by an admin and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/messages/createmessage
func (s *MessagesService) CreateRaw(ctx context.Context, body *CreateMessageRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "messages", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
