package admins

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// TeamsService handles communication with the team related methods
// of the Intercom API.
type TeamsService struct {
	client api.Caller
}

// NewTeamsService creates a new TeamsService.
func NewTeamsService(c api.Caller) *TeamsService {
	return &TeamsService{client: c}
}

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
func ParseTeamGetResult(r *api.Result) (*Team, error) {
	return api.Decode[Team](r)
}

// ParseTeamListResult decodes a Result into a TeamList.
func ParseTeamListResult(r *api.Result) (*TeamList, error) {
	return api.Decode[TeamList](r)
}

// --- Regular Methods ---

// Get retrieves a team by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/teams/retrieveteam
func (s *TeamsService) Get(ctx context.Context, id string) (*Team, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseTeamGetResult(result)
}

// List returns all teams.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/teams/listteams
func (s *TeamsService) List(ctx context.Context) (*TeamList, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseTeamListResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a team by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/teams/retrieveteam
func (s *TeamsService) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("teams/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all teams with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/teams/listteams
func (s *TeamsService) ListRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "teams", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
