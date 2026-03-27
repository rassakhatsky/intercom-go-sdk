package intercom

import (
	"context"
	"net/http"
)

// CustomChannelEventsService handles communication with the custom channel
// event related methods of the Intercom API.
type CustomChannelEventsService service

// CustomChannelContact represents a contact in a custom channel event.
type CustomChannelContact struct {
	Type       string `json:"type"`
	ExternalID string `json:"external_id"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty"`
}

// CustomChannelBaseEvent represents the base fields for a custom channel event.
type CustomChannelBaseEvent struct {
	EventID                string               `json:"event_id"`
	ExternalConversationID string               `json:"external_conversation_id"`
	Contact                CustomChannelContact `json:"contact"`
}

// CustomChannelMessageEvent represents a new message event.
type CustomChannelMessageEvent struct {
	CustomChannelBaseEvent
	Body string `json:"body"`
}

// CustomChannelQuickReplyEvent represents a quick reply selected event.
type CustomChannelQuickReplyEvent struct {
	CustomChannelBaseEvent
	QuickReplyOptionID string `json:"quick_reply_option_id"`
}

// CustomChannelAttribute represents an attribute collected in a custom channel event.
type CustomChannelAttribute struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// CustomChannelAttributeEvent represents an attribute collected event.
type CustomChannelAttributeEvent struct {
	CustomChannelBaseEvent
	Attribute CustomChannelAttribute `json:"attribute"`
}

// CustomChannelNotificationResponse represents the response from a custom channel notification.
type CustomChannelNotificationResponse struct {
	ExternalConversationID string `json:"external_conversation_id"`
	ConversationID         string `json:"conversation_id"`
	ExternalContactID      string `json:"external_contact_id"`
	ContactID              string `json:"contact_id"`
}

// --- Parse Functions ---

// ParseCustomChannelEventNotifyNewConversationResult decodes a Result into a CustomChannelNotificationResponse.
func ParseCustomChannelEventNotifyNewConversationResult(r *Result) (*CustomChannelNotificationResponse, error) {
	return Decode[CustomChannelNotificationResponse](r)
}

// ParseCustomChannelEventNotifyNewMessageResult decodes a Result into a CustomChannelNotificationResponse.
func ParseCustomChannelEventNotifyNewMessageResult(r *Result) (*CustomChannelNotificationResponse, error) {
	return Decode[CustomChannelNotificationResponse](r)
}

// ParseCustomChannelEventNotifyQuickReplyResult decodes a Result into a CustomChannelNotificationResponse.
func ParseCustomChannelEventNotifyQuickReplyResult(r *Result) (*CustomChannelNotificationResponse, error) {
	return Decode[CustomChannelNotificationResponse](r)
}

// ParseCustomChannelEventNotifyAttributeCollectedResult decodes a Result into a CustomChannelNotificationResponse.
func ParseCustomChannelEventNotifyAttributeCollectedResult(r *Result) (*CustomChannelNotificationResponse, error) {
	return Decode[CustomChannelNotificationResponse](r)
}

// --- Regular Methods ---

// NotifyNewConversation notifies Intercom of a new conversation on a custom channel.
func (s *CustomChannelEventsService) NotifyNewConversation(ctx context.Context, event *CustomChannelBaseEvent) (*CustomChannelNotificationResponse, error) {
	result, err := s.NotifyNewConversationRaw(ctx, event)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCustomChannelEventNotifyNewConversationResult(result)
}

// NotifyNewMessage notifies Intercom of a new message on a custom channel.
func (s *CustomChannelEventsService) NotifyNewMessage(ctx context.Context, event *CustomChannelMessageEvent) (*CustomChannelNotificationResponse, error) {
	result, err := s.NotifyNewMessageRaw(ctx, event)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCustomChannelEventNotifyNewMessageResult(result)
}

// NotifyQuickReply notifies Intercom of a quick reply selection on a custom channel.
func (s *CustomChannelEventsService) NotifyQuickReply(ctx context.Context, event *CustomChannelQuickReplyEvent) (*CustomChannelNotificationResponse, error) {
	result, err := s.NotifyQuickReplyRaw(ctx, event)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCustomChannelEventNotifyQuickReplyResult(result)
}

// NotifyAttributeCollected notifies Intercom of an attribute collected on a custom channel.
func (s *CustomChannelEventsService) NotifyAttributeCollected(ctx context.Context, event *CustomChannelAttributeEvent) (*CustomChannelNotificationResponse, error) {
	result, err := s.NotifyAttributeCollectedRaw(ctx, event)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCustomChannelEventNotifyAttributeCollectedResult(result)
}

// --- Raw Methods ---

// NotifyNewConversationRaw notifies Intercom of a new conversation on a custom channel with the full HTTP result.
func (s *CustomChannelEventsService) NotifyNewConversationRaw(ctx context.Context, event *CustomChannelBaseEvent) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "custom_channel_events/notify_new_conversation", event)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// NotifyNewMessageRaw notifies Intercom of a new message on a custom channel with the full HTTP result.
func (s *CustomChannelEventsService) NotifyNewMessageRaw(ctx context.Context, event *CustomChannelMessageEvent) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "custom_channel_events/notify_new_message", event)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// NotifyQuickReplyRaw notifies Intercom of a quick reply selection on a custom channel with the full HTTP result.
func (s *CustomChannelEventsService) NotifyQuickReplyRaw(ctx context.Context, event *CustomChannelQuickReplyEvent) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "custom_channel_events/notify_quick_reply_selected", event)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// NotifyAttributeCollectedRaw notifies Intercom of an attribute collected on a custom channel with the full HTTP result.
func (s *CustomChannelEventsService) NotifyAttributeCollectedRaw(ctx context.Context, event *CustomChannelAttributeEvent) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "custom_channel_events/notify_attribute_collected", event)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
