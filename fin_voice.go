package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// FinVoiceService handles communication with the Fin Voice related methods
// of the Intercom API.
type FinVoiceService service

// AICallResponse represents the response from Fin Voice endpoints.
type AICallResponse struct {
	ID                     int              `json:"id"`
	AppID                  int              `json:"app_id,omitempty"`
	UserPhoneNumber        string           `json:"user_phone_number,omitempty"`
	Status                 string           `json:"status,omitempty"`
	IntercomCallID         string           `json:"intercom_call_id,omitempty"`
	ExternalCallID         string           `json:"external_call_id,omitempty"`
	IntercomConversationID string           `json:"intercom_conversation_id,omitempty"`
	CallTranscript         []map[string]any `json:"call_transcript,omitempty"`
	CallSummary            string           `json:"call_summary,omitempty"`
	Intent                 []map[string]any `json:"intent,omitempty"`
}

// RegisterFinVoiceCallRequest represents a request to register a Fin Voice call.
type RegisterFinVoiceCallRequest struct {
	PhoneNumber string         `json:"phone_number"`
	CallID      string         `json:"call_id"`
	Source      string         `json:"source,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

// --- Parse Functions ---

// ParseFinVoiceRegisterResult decodes a Result into an AICallResponse.
func ParseFinVoiceRegisterResult(r *Result) (*AICallResponse, error) {
	return Decode[AICallResponse](r)
}

// ParseFinVoiceCollectResult decodes a Result into an AICallResponse.
func ParseFinVoiceCollectResult(r *Result) (*AICallResponse, error) {
	return Decode[AICallResponse](r)
}

// ParseFinVoiceGetByExternalIDResult decodes a Result into an AICallResponse.
func ParseFinVoiceGetByExternalIDResult(r *Result) (*AICallResponse, error) {
	return Decode[AICallResponse](r)
}

// ParseFinVoiceGetByConversationResult decodes a Result into a slice of AICallResponse.
func ParseFinVoiceGetByConversationResult(r *Result) (*[]AICallResponse, error) {
	return Decode[[]AICallResponse](r)
}

// ParseFinVoiceGetByPhoneNumberResult decodes a Result into an AICallResponse.
func ParseFinVoiceGetByPhoneNumberResult(r *Result) (*AICallResponse, error) {
	return Decode[AICallResponse](r)
}

// --- Regular Methods ---

// Register registers a new Fin Voice call.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/registerfinvoicecall
func (s *FinVoiceService) Register(ctx context.Context, req *RegisterFinVoiceCallRequest) (*AICallResponse, error) {
	result, err := s.RegisterRaw(ctx, req)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseFinVoiceRegisterResult(result)
}

// Collect retrieves a Fin Voice call by its external reference ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyid
func (s *FinVoiceService) Collect(ctx context.Context, id int) (*AICallResponse, error) {
	result, err := s.CollectRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseFinVoiceCollectResult(result)
}

// GetByExternalID retrieves a Fin Voice call by its external call ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyexternalid
func (s *FinVoiceService) GetByExternalID(ctx context.Context, externalID string) (*AICallResponse, error) {
	result, err := s.GetByExternalIDRaw(ctx, externalID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseFinVoiceGetByExternalIDResult(result)
}

// GetByConversation retrieves all Fin Voice calls for a conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallsbyconversationid
func (s *FinVoiceService) GetByConversation(ctx context.Context, conversationID string) ([]AICallResponse, error) {
	result, err := s.GetByConversationRaw(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	data, err := ParseFinVoiceGetByConversationResult(result)
	if err != nil {
		return nil, err
	}
	return *data, nil
}

// GetByPhoneNumber retrieves the most recent Fin Voice call for a phone number.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyphonenumber
func (s *FinVoiceService) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*AICallResponse, error) {
	result, err := s.GetByPhoneNumberRaw(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseFinVoiceGetByPhoneNumberResult(result)
}

// --- Raw Methods ---

// RegisterRaw registers a new Fin Voice call with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/registerfinvoicecall
func (s *FinVoiceService) RegisterRaw(ctx context.Context, req *RegisterFinVoiceCallRequest) (*Result, error) {
	httpReq, err := s.client.NewRequest(http.MethodPost, "fin_voice/register", req)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, httpReq)
}

// CollectRaw retrieves a Fin Voice call by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyid
func (s *FinVoiceService) CollectRaw(ctx context.Context, id int) (*Result, error) {
	httpReq, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("fin_voice/collect/%d", id), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, httpReq)
}

// GetByExternalIDRaw retrieves a Fin Voice call by external ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyexternalid
func (s *FinVoiceService) GetByExternalIDRaw(ctx context.Context, externalID string) (*Result, error) {
	httpReq, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("fin_voice/external_id/%s", url.PathEscape(externalID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, httpReq)
}

// GetByConversationRaw retrieves all Fin Voice calls for a conversation with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallsbyconversationid
func (s *FinVoiceService) GetByConversationRaw(ctx context.Context, conversationID string) (*Result, error) {
	httpReq, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("fin_voice/conversation/%s", url.PathEscape(conversationID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, httpReq)
}

// GetByPhoneNumberRaw retrieves the most recent Fin Voice call for a phone number with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyphonenumber
func (s *FinVoiceService) GetByPhoneNumberRaw(ctx context.Context, phoneNumber string) (*Result, error) {
	httpReq, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("fin_voice/phone_number/%s", url.PathEscape(phoneNumber)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, httpReq)
}
