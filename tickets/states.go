package tickets

import (
	"context"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// StatesService handles communication with the ticket state related
// methods of the Intercom API.
type StatesService struct {
	client api.Caller
}

// NewStatesService creates a new ticket states StatesService.
func NewStatesService(c api.Caller) *StatesService {
	return &StatesService{client: c}
}

// State represents a ticket state in Intercom.
type State struct {
	Type          string `json:"type"`
	ID            string `json:"id"`
	Category      string `json:"category"`
	InternalLabel string `json:"internal_label"`
	ExternalLabel string `json:"external_label"`
	Archived      bool   `json:"archived"`
}

// StateList represents the response from listing ticket states.
type StateList struct {
	Type string  `json:"type"`
	Data []State `json:"data"`
}

// --- Parse Functions ---

// ParseStateListResult decodes a Result into a StateList.
func ParseStateListResult(r *api.Result) (*StateList, error) {
	return api.Decode[StateList](r)
}

// --- Regular Methods ---

// List returns all ticket states for the workspace.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-states/listticketstates
func (s *StatesService) List(ctx context.Context) (*StateList, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseStateListResult(result)
}

// --- Raw Methods ---

// ListRaw returns all ticket states with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-states/listticketstates
func (s *StatesService) ListRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "ticket_states", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
