package settings

import (
	"context"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// ChannelEventsService handles communication with the custom channel
// event related methods of the Intercom API.
type ChannelEventsService struct {
	client api.Caller
}

// NewChannelEventsService creates a new ChannelEventsService.
func NewChannelEventsService(c api.Caller) *ChannelEventsService {
	return &ChannelEventsService{client: c}
}

// Contact represents a contact in a custom channel event.
type Contact struct {
	Type       string `json:"type"`
	ExternalID string `json:"external_id"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty"`
}

// BaseEvent represents the base fields for a custom channel event.
type BaseEvent struct {
	EventID                string  `json:"event_id"`
	ExternalConversationID string  `json:"external_conversation_id"`
	Contact                Contact `json:"contact"`
}

// MessageEvent represents a new message event.
type MessageEvent struct {
	BaseEvent
	Body string `json:"body"`
}

// QuickReplyEvent represents a quick reply selected event.
type QuickReplyEvent struct {
	BaseEvent
	QuickReplyOptionID string `json:"quick_reply_option_id"`
}

// Attribute represents an attribute collected in a custom channel event.
type Attribute struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// AttributeEvent represents an attribute collected event.
type AttributeEvent struct {
	BaseEvent
	Attribute Attribute `json:"attribute"`
}

// NotificationResponse represents the response from a custom channel notification.
type NotificationResponse struct {
	ExternalConversationID string `json:"external_conversation_id"`
	ConversationID         string `json:"conversation_id"`
	ExternalContactID      string `json:"external_contact_id"`
	ContactID              string `json:"contact_id"`
}

// --- Parse Functions ---

// ParseNotifyNewConversationResult decodes a Result into a NotificationResponse.
func ParseNotifyNewConversationResult(r *api.Result) (*NotificationResponse, error) {
	return api.Decode[NotificationResponse](r)
}

// ParseNotifyNewMessageResult decodes a Result into a NotificationResponse.
func ParseNotifyNewMessageResult(r *api.Result) (*NotificationResponse, error) {
	return api.Decode[NotificationResponse](r)
}

// ParseNotifyQuickReplyResult decodes a Result into a NotificationResponse.
func ParseNotifyQuickReplyResult(r *api.Result) (*NotificationResponse, error) {
	return api.Decode[NotificationResponse](r)
}

// ParseNotifyAttributeCollectedResult decodes a Result into a NotificationResponse.
func ParseNotifyAttributeCollectedResult(r *api.Result) (*NotificationResponse, error) {
	return api.Decode[NotificationResponse](r)
}

// --- Regular Methods ---

// NotifyNewConversation notifies Intercom of a new conversation on a custom channel.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/custom-channel-events/notifynewconversation
func (s *ChannelEventsService) NotifyNewConversation(ctx context.Context, event *BaseEvent) (*NotificationResponse, error) {
	result, err := s.NotifyNewConversationRaw(ctx, event)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseNotifyNewConversationResult(result)
}

// NotifyNewMessage notifies Intercom of a new message on a custom channel.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/custom-channel-events/notifynewmessage
func (s *ChannelEventsService) NotifyNewMessage(ctx context.Context, event *MessageEvent) (*NotificationResponse, error) {
	result, err := s.NotifyNewMessageRaw(ctx, event)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseNotifyNewMessageResult(result)
}

// NotifyQuickReply notifies Intercom of a quick reply selection on a custom channel.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/custom-channel-events/notifyquickreplyselected
func (s *ChannelEventsService) NotifyQuickReply(ctx context.Context, event *QuickReplyEvent) (*NotificationResponse, error) {
	result, err := s.NotifyQuickReplyRaw(ctx, event)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseNotifyQuickReplyResult(result)
}

// NotifyAttributeCollected notifies Intercom of an attribute collected on a custom channel.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/custom-channel-events/notifyattributecollected
func (s *ChannelEventsService) NotifyAttributeCollected(ctx context.Context, event *AttributeEvent) (*NotificationResponse, error) {
	result, err := s.NotifyAttributeCollectedRaw(ctx, event)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseNotifyAttributeCollectedResult(result)
}

// --- Raw Methods ---

// NotifyNewConversationRaw notifies Intercom of a new conversation on a custom channel with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/custom-channel-events/notifynewconversation
func (s *ChannelEventsService) NotifyNewConversationRaw(ctx context.Context, event *BaseEvent) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "custom_channel_events/notify_new_conversation", event)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// NotifyNewMessageRaw notifies Intercom of a new message on a custom channel with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/custom-channel-events/notifynewmessage
func (s *ChannelEventsService) NotifyNewMessageRaw(ctx context.Context, event *MessageEvent) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "custom_channel_events/notify_new_message", event)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// NotifyQuickReplyRaw notifies Intercom of a quick reply selection on a custom channel with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/custom-channel-events/notifyquickreplyselected
func (s *ChannelEventsService) NotifyQuickReplyRaw(ctx context.Context, event *QuickReplyEvent) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "custom_channel_events/notify_quick_reply_selected", event)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// NotifyAttributeCollectedRaw notifies Intercom of an attribute collected on a custom channel with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/custom-channel-events/notifyattributecollected
func (s *ChannelEventsService) NotifyAttributeCollectedRaw(ctx context.Context, event *AttributeEvent) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "custom_channel_events/notify_attribute_collected", event)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
