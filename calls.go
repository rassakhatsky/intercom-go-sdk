package intercom

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// CallsService handles communication with the call related methods
// of the Intercom API.
type CallsService service

// Call represents an Intercom call.
type Call struct {
	Type                string                `json:"type"`
	ID                  string                `json:"id"`
	ConversationID      string                `json:"conversation_id,omitempty"`
	AdminID             string                `json:"admin_id,omitempty"`
	ContactID           string                `json:"contact_id,omitempty"`
	State               string                `json:"state,omitempty"`
	InitiatedAt         int64                 `json:"initiated_at,omitempty"`
	AnsweredAt          int64                 `json:"answered_at,omitempty"`
	EndedAt             int64                 `json:"ended_at,omitempty"`
	CreatedAt           int64                 `json:"created_at,omitempty"`
	UpdatedAt           int64                 `json:"updated_at,omitempty"`
	RecordingURL        string                `json:"recording_url,omitempty"`
	TranscriptionURL    string                `json:"transcription_url,omitempty"`
	CallType            string                `json:"call_type,omitempty"`
	Direction           string                `json:"direction,omitempty"`
	EndedReason         string                `json:"ended_reason,omitempty"`
	Phone               string                `json:"phone,omitempty"`
	FinRecordingURL     string                `json:"fin_recording_url,omitempty"`
	FinTranscriptionURL string                `json:"fin_transcription_url,omitempty"`
	TranscriptStatus    string                `json:"transcript_status,omitempty"`
	Transcript          []CallTranscriptEntry `json:"transcript,omitempty"`
}

// CallTranscriptEntry represents a single entry in a call transcript.
type CallTranscriptEntry struct {
	Speaker string `json:"speaker,omitempty"`
	Text    string `json:"text,omitempty"`
}

// CallList represents the response from the list calls endpoint.
type CallList struct {
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

// CallListOptions specifies optional parameters to the List method.
type CallListOptions struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}

// CallSearchRequest represents a request to search calls with transcripts.
type CallSearchRequest struct {
	ConversationIDs []string `json:"conversation_ids"`
}

// ParseCallGetResult decodes a Result into a Call.
func ParseCallGetResult(r *Result) (*Call, error) { return Decode[Call](r) }

// ParseCallListResult decodes a Result into a CallList.
func ParseCallListResult(r *Result) (*CallList, error) { return Decode[CallList](r) }

// ParseCallSearchResult decodes a Result into a CallList.
func ParseCallSearchResult(r *Result) (*CallList, error) { return Decode[CallList](r) }

// ParseCallGetRecordingURLResult extracts the recording URL from a redirect Result.
// Returns the Location header for 302/301 responses, or an error for unexpected status codes.
func ParseCallGetRecordingURLResult(r *Result) (string, error) {
	if r.StatusCode == http.StatusFound || r.StatusCode == http.StatusMovedPermanently {
		loc := r.Header.Get("Location")
		if loc == "" {
			return "", fmt.Errorf("intercom: redirect response %d missing Location header", r.StatusCode)
		}
		return loc, nil
	}
	return "", fmt.Errorf("intercom: expected redirect for recording URL, got %d", r.StatusCode)
}

// ParseCallGetTranscriptResult extracts the plain text transcript from a Result.
func ParseCallGetTranscriptResult(r *Result) (string, error) {
	return string(r.Body), nil
}

// Get retrieves a call by ID.
func (s *CallsService) Get(ctx context.Context, callID string) (*Call, error) {
	result, err := s.GetRaw(ctx, callID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCallGetResult(result)
}

// List returns calls with optional pagination.
func (s *CallsService) List(ctx context.Context, opts *CallListOptions) (*CallList, error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCallListResult(result)
}

// Search searches for calls with transcripts by conversation IDs.
func (s *CallsService) Search(ctx context.Context, searchReq *CallSearchRequest) (*CallList, error) {
	result, err := s.SearchRaw(ctx, searchReq)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCallSearchResult(result)
}

// GetRecordingURL returns the signed URL for a call recording.
// The API responds with a 302 redirect; this method returns the Location header.
func (s *CallsService) GetRecordingURL(ctx context.Context, callID string) (string, error) {
	result, err := s.GetRecordingURLRaw(ctx, callID)
	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", resultError(result)
	}
	return ParseCallGetRecordingURLResult(result)
}

// GetTranscript returns the plain text transcript of a call.
func (s *CallsService) GetTranscript(ctx context.Context, callID string) (string, error) {
	result, err := s.GetTranscriptRaw(ctx, callID)
	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", resultError(result)
	}
	return ParseCallGetTranscriptResult(result)
}

// GetRaw retrieves a call by ID with the full HTTP result.
func (s *CallsService) GetRaw(ctx context.Context, callID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("calls/%s", url.PathEscape(callID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns calls with optional pagination and the full HTTP result.
func (s *CallsService) ListRaw(ctx context.Context, opts *CallListOptions) (*Result, error) {
	path, err := addQueryOptions("calls", opts)
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
func (s *CallsService) SearchRaw(ctx context.Context, searchReq *CallSearchRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "calls/search", searchReq)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetRecordingURLRaw returns the full HTTP result for a call recording URL request.
// The recording URL is available in Result.Header.Get("Location") for redirect responses.
func (s *CallsService) GetRecordingURLRaw(ctx context.Context, callID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("calls/%s/recording", url.PathEscape(callID)), nil)
	if err != nil {
		return nil, err
	}

	noRedirectClient := &http.Client{
		Transport: s.client.httpClient.Transport,
		Timeout:   s.client.httpClient.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req2 := req.WithContext(ctx)
	s.client.logger.Debug("http request", "method", req2.Method, "url", req2.URL)
	resp, err := noRedirectClient.Do(req2)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return buildResult(resp, body), nil
}

// GetTranscriptRaw returns the plain text transcript of a call with the full HTTP result.
// The transcript text is available in string(Result.Body).
func (s *CallsService) GetTranscriptRaw(ctx context.Context, callID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("calls/%s/transcript", url.PathEscape(callID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
