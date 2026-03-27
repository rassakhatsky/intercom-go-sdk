package intercom

import (
	"context"
	"net/http"
)

// TicketStatesService handles communication with the ticket state related
// methods of the Intercom API.
type TicketStatesService service

// TicketState represents a ticket state in Intercom.
type TicketState struct {
	Type          string `json:"type"`
	ID            string `json:"id"`
	Category      string `json:"category"`
	InternalLabel string `json:"internal_label"`
	ExternalLabel string `json:"external_label"`
	Archived      bool   `json:"archived"`
}

// TicketStateList represents the response from listing ticket states.
type TicketStateList struct {
	Type string        `json:"type"`
	Data []TicketState `json:"data"`
}

// --- Parse Functions ---

// ParseTicketStateListResult decodes a Result into a TicketStateList.
func ParseTicketStateListResult(r *Result) (*TicketStateList, error) {
	return Decode[TicketStateList](r)
}

// --- Regular Methods ---

// List returns all ticket states for the workspace.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-states/listticketstates
func (s *TicketStatesService) List(ctx context.Context) (*TicketStateList, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTicketStateListResult(result)
}

// --- Raw Methods ---

// ListRaw returns all ticket states with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ticket-states/listticketstates
func (s *TicketStatesService) ListRaw(ctx context.Context) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "ticket_states", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
