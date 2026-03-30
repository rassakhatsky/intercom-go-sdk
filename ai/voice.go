package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// VoiceService handles communication with the Fin Voice related methods
// of the Intercom API.
type VoiceService struct {
	client api.Caller
}

// NewVoiceService creates a new Fin Voice service.
func NewVoiceService(c api.Caller) *VoiceService {
	return &VoiceService{client: c}
}

// CallResponse represents the response from Fin Voice endpoints.
type CallResponse struct {
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

// RegisterCallRequest represents a request to register a Fin Voice call.
type RegisterCallRequest struct {
	PhoneNumber string         `json:"phone_number"`
	CallID      string         `json:"call_id"`
	Source      string         `json:"source,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

// --- Parse Functions ---

// ParseRegisterResult decodes a Result into a CallResponse.
func ParseRegisterResult(r *api.Result) (*CallResponse, error) {
	return api.Decode[CallResponse](r)
}

// ParseCollectResult decodes a Result into a CallResponse.
func ParseCollectResult(r *api.Result) (*CallResponse, error) {
	return api.Decode[CallResponse](r)
}

// ParseGetByExternalIDResult decodes a Result into a CallResponse.
func ParseGetByExternalIDResult(r *api.Result) (*CallResponse, error) {
	return api.Decode[CallResponse](r)
}

// ParseGetByConversationResult decodes a Result into a slice of CallResponse.
func ParseGetByConversationResult(r *api.Result) (*[]CallResponse, error) {
	return api.Decode[[]CallResponse](r)
}

// ParseGetByPhoneNumberResult decodes a Result into a CallResponse.
func ParseGetByPhoneNumberResult(r *api.Result) (*CallResponse, error) {
	return api.Decode[CallResponse](r)
}

// --- Regular Methods ---

// Register registers a new Fin Voice call.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/registerfinvoicecall
func (s *VoiceService) Register(ctx context.Context, req *RegisterCallRequest) (*CallResponse, error) {
	result, err := s.RegisterRaw(ctx, req)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseRegisterResult(result)
}

// Collect retrieves a Fin Voice call by its Intercom ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyid
func (s *VoiceService) Collect(ctx context.Context, id int) (*CallResponse, error) {
	result, err := s.CollectRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCollectResult(result)
}

// GetByExternalID retrieves a Fin Voice call by its external call ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyexternalid
func (s *VoiceService) GetByExternalID(ctx context.Context, externalID string) (*CallResponse, error) {
	result, err := s.GetByExternalIDRaw(ctx, externalID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetByExternalIDResult(result)
}

// GetByConversation retrieves all Fin Voice calls for a conversation.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallsbyconversationid
func (s *VoiceService) GetByConversation(ctx context.Context, conversationID string) ([]CallResponse, error) {
	result, err := s.GetByConversationRaw(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	data, err := ParseGetByConversationResult(result)
	if err != nil {
		return nil, err
	}
	return *data, nil
}

// GetByPhoneNumber retrieves the most recent Fin Voice call for a phone number.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyphonenumber
func (s *VoiceService) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*CallResponse, error) {
	result, err := s.GetByPhoneNumberRaw(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetByPhoneNumberResult(result)
}

// --- Raw Methods ---

// RegisterRaw registers a new Fin Voice call with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/registerfinvoicecall
func (s *VoiceService) RegisterRaw(ctx context.Context, req *RegisterCallRequest) (*api.Result, error) {
	httpReq, err := s.client.NewRequest(http.MethodPost, "fin_voice/register", req)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, httpReq)
}

// CollectRaw retrieves a Fin Voice call by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyid
func (s *VoiceService) CollectRaw(ctx context.Context, id int) (*api.Result, error) {
	httpReq, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("fin_voice/collect/%d", id), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, httpReq)
}

// GetByExternalIDRaw retrieves a Fin Voice call by external ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyexternalid
func (s *VoiceService) GetByExternalIDRaw(ctx context.Context, externalID string) (*api.Result, error) {
	httpReq, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("fin_voice/external_id/%s", url.PathEscape(externalID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, httpReq)
}

// GetByConversationRaw retrieves all Fin Voice calls for a conversation with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallsbyconversationid
func (s *VoiceService) GetByConversationRaw(ctx context.Context, conversationID string) (*api.Result, error) {
	httpReq, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("fin_voice/conversation/%s", url.PathEscape(conversationID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, httpReq)
}

// GetByPhoneNumberRaw retrieves the most recent Fin Voice call for a phone number with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/collectfinvoicecallbyphonenumber
func (s *VoiceService) GetByPhoneNumberRaw(ctx context.Context, phoneNumber string) (*api.Result, error) {
	httpReq, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("fin_voice/phone_number/%s", url.PathEscape(phoneNumber)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, httpReq)
}
