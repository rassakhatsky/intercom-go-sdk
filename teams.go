package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// TeamsService handles communication with the team related methods
// of the Intercom API.
type TeamsService service

// Team represents an Intercom team.
type Team struct {
	Type               string  `json:"type"`
	ID                 string  `json:"id"`
	Name               string  `json:"name,omitempty"`
	AdminIDs           []int64 `json:"admin_ids,omitempty"`
	AssignmentLimit    *int    `json:"assignment_limit,omitempty"`
	DistributionMethod *string `json:"distribution_method,omitempty"`
}

// TeamList represents the response from the list teams endpoint.
type TeamList struct {
	Type  string `json:"type"`
	Teams []Team `json:"teams"`
}

// --- Parse Functions ---

// ParseTeamGetResult decodes a Result into a Team.
func ParseTeamGetResult(r *Result) (*Team, error) {
	return Decode[Team](r)
}

// ParseTeamListResult decodes a Result into a TeamList.
func ParseTeamListResult(r *Result) (*TeamList, error) {
	return Decode[TeamList](r)
}

// --- Regular Methods ---

// Get retrieves a team by ID.
func (s *TeamsService) Get(ctx context.Context, id string) (*Team, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTeamGetResult(result)
}

// List returns all teams.
func (s *TeamsService) List(ctx context.Context) (*TeamList, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseTeamListResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a team by ID with the full HTTP result.
func (s *TeamsService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("teams/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all teams with the full HTTP result.
func (s *TeamsService) ListRaw(ctx context.Context) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "teams", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
