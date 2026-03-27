package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// NotesService handles communication with the note related methods
// of the Intercom API.
type NotesService service

// --- Parse Functions ---

// ParseNoteGetResult decodes a Result into a Note.
func ParseNoteGetResult(r *Result) (*Note, error) {
	return Decode[Note](r)
}

// --- Regular Methods ---

// Get retrieves a note by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/notes/retrievenote
func (s *NotesService) Get(ctx context.Context, id string) (*Note, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseNoteGetResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a note by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/notes/retrievenote
func (s *NotesService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("notes/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
