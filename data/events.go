package data

import (
	"context"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// EventsService handles communication with the data event related
// methods of the Intercom API.
type EventsService struct {
	client api.Caller
}

// NewEventsService creates a new data events service.
func NewEventsService(c api.Caller) *EventsService {
	return &EventsService{client: c}
}

// Event represents an Intercom data event.
type Event struct {
	Type           string         `json:"type,omitempty"`
	EventName      string         `json:"event_name"`
	CreatedAt      int64          `json:"created_at"`
	UserID         string         `json:"user_id,omitempty"`
	ID             string         `json:"id,omitempty"`
	IntercomUserID string         `json:"intercom_user_id,omitempty"`
	Email          string         `json:"email,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// SummaryResponse represents the response from the list events endpoint.
type SummaryResponse struct {
	Type           string      `json:"type"`
	Events         []Event     `json:"events"`
	Pages          *EventPages `json:"pages,omitempty"`
	Email          string      `json:"email,omitempty"`
	IntercomUserID string      `json:"intercom_user_id,omitempty"`
	UserID         string      `json:"user_id,omitempty"`
}

// EventPages represents pagination info for events.
type EventPages struct {
	Next  string `json:"next,omitempty"`
	Since string `json:"since,omitempty"`
}

// ListOptions specifies parameters to the List method.
type ListOptions struct {
	Type           string `url:"type"`
	UserID         string `url:"user_id,omitempty"`
	Email          string `url:"email,omitempty"`
	IntercomUserID string `url:"intercom_user_id,omitempty"`
	Summary        *bool  `url:"summary,omitempty"`
	PerPage        int    `url:"per_page,omitempty"`
}

// CreateEventRequest represents the request body for creating a data event.
type CreateEventRequest struct {
	EventName string         `json:"event_name"`
	CreatedAt int64          `json:"created_at"`
	UserID    string         `json:"user_id,omitempty"`
	ID        string         `json:"id,omitempty"`
	Email     string         `json:"email,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// CreateSummariesRequest represents the request body for creating event summaries.
type CreateSummariesRequest struct {
	UserID         string         `json:"user_id"`
	EventSummaries []EventSummary `json:"event_summaries"`
}

// EventSummary represents a single event summary.
type EventSummary struct {
	EventName string `json:"event_name"`
	Count     int    `json:"count"`
	First     int64  `json:"first"`
	Last      int64  `json:"last"`
}

// --- Parse Functions ---

// ParseCreateResult validates a Result for the Create endpoint (empty body).
func ParseCreateResult(r *api.Result) error {
	_, err := api.Decode[api.Empty](r)
	return err
}

// ParseListResult decodes a Result into a SummaryResponse.
func ParseListResult(r *api.Result) (*SummaryResponse, error) {
	return api.Decode[SummaryResponse](r)
}

// ParseCreateSummariesResult validates a Result for the CreateSummaries endpoint (empty body).
func ParseCreateSummariesResult(r *api.Result) error {
	_, err := api.Decode[api.Empty](r)
	return err
}

// --- Regular Methods ---

// Create submits a new data event. Returns nil on success (API returns 202 with empty body).
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/createdataevent
func (s *EventsService) Create(ctx context.Context, body *CreateEventRequest) error {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return err
	}
	if result.Error != nil {
		return api.ResultError(result)
	}
	return nil
}

// List returns data events for a user or lead.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/listdataevents
func (s *EventsService) List(ctx context.Context, opts *ListOptions) (*SummaryResponse, error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListResult(result)
}

// CreateSummaries creates event summaries for a user.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/dataeventsummaries
func (s *EventsService) CreateSummaries(ctx context.Context, body *CreateSummariesRequest) error {
	result, err := s.CreateSummariesRaw(ctx, body)
	if err != nil {
		return err
	}
	if result.Error != nil {
		return api.ResultError(result)
	}
	return nil
}

// --- Raw Methods ---

// CreateRaw submits a new data event and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/createdataevent
func (s *EventsService) CreateRaw(ctx context.Context, body *CreateEventRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "events", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns data events for a user or lead with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/listdataevents
func (s *EventsService) ListRaw(ctx context.Context, opts *ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("events", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateSummariesRaw creates event summaries for a user and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/dataeventsummaries
func (s *EventsService) CreateSummariesRaw(ctx context.Context, body *CreateSummariesRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "events/summaries", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
