package calls

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// Service handles communication with the call related methods
// of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new calls service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Call represents an Intercom call.
type Call struct {
	Type                string            `json:"type"`
	ID                  string            `json:"id"`
	ConversationID      string            `json:"conversation_id,omitempty"`
	AdminID             string            `json:"admin_id,omitempty"`
	ContactID           string            `json:"contact_id,omitempty"`
	State               string            `json:"state,omitempty"`
	InitiatedAt         int64             `json:"initiated_at,omitempty"`
	AnsweredAt          int64             `json:"answered_at,omitempty"`
	EndedAt             int64             `json:"ended_at,omitempty"`
	CreatedAt           int64             `json:"created_at,omitempty"`
	UpdatedAt           int64             `json:"updated_at,omitempty"`
	RecordingURL        string            `json:"recording_url,omitempty"`
	TranscriptionURL    string            `json:"transcription_url,omitempty"`
	CallType            string            `json:"call_type,omitempty"`
	Direction           string            `json:"direction,omitempty"`
	EndedReason         string            `json:"ended_reason,omitempty"`
	Phone               string            `json:"phone,omitempty"`
	FinRecordingURL     string            `json:"fin_recording_url,omitempty"`
	FinTranscriptionURL string            `json:"fin_transcription_url,omitempty"`
	TranscriptStatus    string            `json:"transcript_status,omitempty"`
	Transcript          []TranscriptEntry `json:"transcript,omitempty"`
}

// TranscriptEntry represents a single entry in a call transcript.
type TranscriptEntry struct {
	Speaker string `json:"speaker,omitempty"`
	Text    string `json:"text,omitempty"`
}

// List represents the response from the list calls endpoint.
type List struct {
	Type       string `json:"type"`
	Data       []Call `json:"data"`
	TotalCount int    `json:"total_count,omitempty"`
	Pages      *Pages `json:"pages,omitempty"`
}

// Pages represents pagination info in call list responses.
type Pages struct {
	Type       string `json:"type,omitempty"`
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	TotalPages int    `json:"total_pages,omitempty"`
}

// ListOptions specifies optional parameters to the List method.
type ListOptions struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}

// SearchRequest represents a request to search calls with transcripts.
type SearchRequest struct {
	ConversationIDs []string `json:"conversation_ids"`
}

// ParseGetResult decodes a Result into a Call.
func ParseGetResult(r *api.Result) (*Call, error) { return api.Decode[Call](r) }

// ParseListResult decodes a Result into a List.
func ParseListResult(r *api.Result) (*List, error) { return api.Decode[List](r) }

// ParseSearchResult decodes a Result into a List.
func ParseSearchResult(r *api.Result) (*List, error) { return api.Decode[List](r) }

// ParseGetRecordingURLResult extracts the recording URL from a redirect Result.
// Returns the Location header for 302/301 responses, or an error for unexpected status codes.
func ParseGetRecordingURLResult(r *api.Result) (string, error) {
	if r == nil {
		return "", fmt.Errorf("intercom: nil result")
	}
	if r.StatusCode == http.StatusFound || r.StatusCode == http.StatusMovedPermanently {
		loc := r.Header.Get("Location")
		if loc == "" {
			return "", fmt.Errorf("intercom: redirect response %d missing Location header", r.StatusCode)
		}
		return loc, nil
	}
	body := string(r.Body)
	if len(body) > 512 {
		body = body[:512]
	}
	return "", fmt.Errorf("intercom: expected redirect for recording URL, got status %d: %s", r.StatusCode, body)
}

// ParseGetTranscriptResult extracts the plain text transcript from a Result.
func ParseGetTranscriptResult(r *api.Result) (string, error) {
	if r == nil {
		return "", fmt.Errorf("intercom: nil result")
	}
	if len(r.Body) == 0 {
		return "", fmt.Errorf("intercom: empty transcript body")
	}
	return string(r.Body), nil
}

// Get retrieves a call by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/showcall
func (s *Service) Get(ctx context.Context, callID string) (*Call, error) {
	result, err := s.GetRaw(ctx, callID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// List returns calls with optional pagination.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/listcalls
func (s *Service) List(ctx context.Context, opts *ListOptions) (*List, error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListResult(result)
}

// Search searches for calls with transcripts by conversation IDs.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/listcallswithtranscripts
func (s *Service) Search(ctx context.Context, searchReq *SearchRequest) (*List, error) {
	result, err := s.SearchRaw(ctx, searchReq)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseSearchResult(result)
}

// GetRecordingURL returns the signed URL for a call recording.
// The API responds with a 302 redirect; this method returns the Location header.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/showcallrecording
func (s *Service) GetRecordingURL(ctx context.Context, callID string) (string, error) {
	result, err := s.GetRecordingURLRaw(ctx, callID)
	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", api.ResultError(result)
	}
	return ParseGetRecordingURLResult(result)
}

// GetTranscript returns the plain text transcript of a call.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/showcalltranscript
func (s *Service) GetTranscript(ctx context.Context, callID string) (string, error) {
	result, err := s.GetTranscriptRaw(ctx, callID)
	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", api.ResultError(result)
	}
	return ParseGetTranscriptResult(result)
}

// GetRaw retrieves a call by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/showcall
func (s *Service) GetRaw(ctx context.Context, callID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("calls/%s", url.PathEscape(callID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns calls with optional pagination and the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/listcalls
func (s *Service) ListRaw(ctx context.Context, opts *ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("calls", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for calls with transcripts by conversation IDs with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/listcallswithtranscripts
func (s *Service) SearchRaw(ctx context.Context, searchReq *SearchRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "calls/search", searchReq)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetRecordingURLRaw returns the full HTTP result for a call recording URL request.
// The recording URL is available in Result.Header.Get("Location") for redirect responses.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/showcallrecording
func (s *Service) GetRecordingURLRaw(ctx context.Context, callID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("calls/%s/recording", url.PathEscape(callID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRawNoRedirect(ctx, req)
}

// GetTranscriptRaw returns the plain text transcript of a call with the full HTTP result.
// The transcript text is available in string(Result.Body).
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/calls/showcalltranscript
func (s *Service) GetTranscriptRaw(ctx context.Context, callID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("calls/%s/transcript", url.PathEscape(callID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
