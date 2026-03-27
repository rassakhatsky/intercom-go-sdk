package intercom

import (
	"context"
	"net/http"
)

// DataEventsService handles communication with the data event related
// methods of the Intercom API.
type DataEventsService service

// DataEvent represents an Intercom data event.
type DataEvent struct {
	Type           string         `json:"type,omitempty"`
	EventName      string         `json:"event_name"`
	CreatedAt      int64          `json:"created_at"`
	UserID         string         `json:"user_id,omitempty"`
	ID             string         `json:"id,omitempty"`
	IntercomUserID string         `json:"intercom_user_id,omitempty"`
	Email          string         `json:"email,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// DataEventSummaryResponse represents the response from the list events endpoint.
type DataEventSummaryResponse struct {
	Type           string      `json:"type"`
	Events         []DataEvent `json:"events"`
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

// ListDataEventsOptions specifies parameters to the List method.
type ListDataEventsOptions struct {
	Type           string `url:"type"`
	UserID         string `url:"user_id,omitempty"`
	Email          string `url:"email,omitempty"`
	IntercomUserID string `url:"intercom_user_id,omitempty"`
	Summary        *bool  `url:"summary,omitempty"`
	PerPage        int    `url:"per_page,omitempty"`
}

// CreateDataEventRequest represents the request body for creating a data event.
type CreateDataEventRequest struct {
	EventName string         `json:"event_name"`
	CreatedAt int64          `json:"created_at"`
	UserID    string         `json:"user_id,omitempty"`
	ID        string         `json:"id,omitempty"`
	Email     string         `json:"email,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// CreateEventSummariesRequest represents the request body for creating event summaries.
type CreateEventSummariesRequest struct {
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

// ParseDataEventCreateResult validates a Result for the Create endpoint (empty body).
func ParseDataEventCreateResult(r *Result) error {
	_, err := Decode[Empty](r)
	return err
}

// ParseDataEventListResult decodes a Result into a DataEventSummaryResponse.
func ParseDataEventListResult(r *Result) (*DataEventSummaryResponse, error) {
	return Decode[DataEventSummaryResponse](r)
}

// ParseDataEventCreateSummariesResult validates a Result for the CreateSummaries endpoint (empty body).
func ParseDataEventCreateSummariesResult(r *Result) error {
	_, err := Decode[Empty](r)
	return err
}

// --- Regular Methods ---

// Create submits a new data event. Returns nil on success (API returns 202 with empty body).
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/createdataevent
func (s *DataEventsService) Create(ctx context.Context, body *CreateDataEventRequest) error {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return err
	}
	if result.Error != nil {
		return resultError(result)
	}
	return nil
}

// List returns data events for a user or lead.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/lisdataevents
func (s *DataEventsService) List(ctx context.Context, opts *ListDataEventsOptions) (*DataEventSummaryResponse, error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseDataEventListResult(result)
}

// CreateSummaries creates event summaries for a user.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/dataeventsummaries
func (s *DataEventsService) CreateSummaries(ctx context.Context, body *CreateEventSummariesRequest) error {
	result, err := s.CreateSummariesRaw(ctx, body)
	if err != nil {
		return err
	}
	if result.Error != nil {
		return resultError(result)
	}
	return nil
}

// --- Raw Methods ---

// CreateRaw submits a new data event and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/createdataevent
func (s *DataEventsService) CreateRaw(ctx context.Context, body *CreateDataEventRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "events", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns data events for a user or lead with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-events/lisdataevents
func (s *DataEventsService) ListRaw(ctx context.Context, opts *ListDataEventsOptions) (*Result, error) {
	path, err := addQueryOptions("events", opts)
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
func (s *DataEventsService) CreateSummariesRaw(ctx context.Context, body *CreateEventSummariesRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "events/summaries", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
