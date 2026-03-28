package admins

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// Service handles communication with the admin related methods
// of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new admins Service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Admin represents an Intercom admin (teammate).
type Admin struct {
	Type               string         `json:"type"`
	ID                 string         `json:"id"`
	Name               string         `json:"name,omitempty"`
	Email              string         `json:"email,omitempty"`
	JobTitle           string         `json:"job_title,omitempty"`
	AwayModeEnabled    bool           `json:"away_mode_enabled,omitempty"`
	AwayModeReassign   bool           `json:"away_mode_reassign,omitempty"`
	AwayStatusReasonID *int           `json:"away_status_reason_id,omitempty"`
	HasInboxSeat       bool           `json:"has_inbox_seat,omitempty"`
	TeamIDs            []int          `json:"team_ids,omitempty"`
	Avatar             *Avatar        `json:"avatar,omitempty"`
	TeamPriorityLevel  map[string]any `json:"team_priority_level,omitempty"`
	EmailVerified      *bool          `json:"email_verified,omitempty"`
	App                *App           `json:"app,omitempty"`
}

// Avatar represents an admin's avatar.
type Avatar struct {
	Type     string `json:"type,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// App represents an Intercom workspace.
type App struct {
	Type                 string `json:"type"`
	IDCode               string `json:"id_code,omitempty"`
	Name                 string `json:"name,omitempty"`
	Region               string `json:"region,omitempty"`
	Timezone             string `json:"timezone,omitempty"`
	CreatedAt            int64  `json:"created_at,omitempty"`
	IdentityVerification bool   `json:"identity_verification,omitempty"`
}

// List represents a list of admins.
type List struct {
	Type   string  `json:"type"`
	Admins []Admin `json:"admins"`
}

// SetAwayRequest represents the request body for setting an admin's away status.
type SetAwayRequest struct {
	AwayModeEnabled    bool `json:"away_mode_enabled"`
	AwayModeReassign   bool `json:"away_mode_reassign"`
	AwayStatusReasonID *int `json:"away_status_reason_id,omitempty"`
}

// ActivityLog represents an admin activity log entry.
type ActivityLog struct {
	ID                  string           `json:"id"`
	PerformedBy         ActivityLogAdmin `json:"performed_by"`
	Metadata            map[string]any   `json:"metadata,omitempty"`
	CreatedAt           int64            `json:"created_at"`
	ActivityType        string           `json:"activity_type"`
	ActivityDescription string           `json:"activity_description,omitempty"`
}

// ActivityLogAdmin represents the admin who performed an activity.
type ActivityLogAdmin struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
	IP    string `json:"ip,omitempty"`
}

// ActivityLogList represents a paginated list of activity logs.
type ActivityLogList struct {
	Type         string           `json:"type"`
	Pages        *api.CursorPages `json:"pages,omitempty"`
	ActivityLogs []ActivityLog    `json:"activity_logs"`
}

// ActivityLogListOptions specifies the query parameters for ListActivityLogs.
type ActivityLogListOptions struct {
	CreatedAtAfter  string `url:"created_at_after,omitempty"`
	CreatedAtBefore string `url:"created_at_before,omitempty"`
}

// ParseMeResult decodes a Result into an Admin.
func ParseMeResult(r *api.Result) (*Admin, error) { return api.Decode[Admin](r) }

// ParseGetResult decodes a Result into an Admin.
func ParseGetResult(r *api.Result) (*Admin, error) { return api.Decode[Admin](r) }

// ParseListResult decodes a Result into a List.
func ParseListResult(r *api.Result) (*List, error) { return api.Decode[List](r) }

// ParseSetAwayResult decodes a Result into an Admin.
func ParseSetAwayResult(r *api.Result) (*Admin, error) { return api.Decode[Admin](r) }

// ParseListActivityLogsResult decodes a Result into an ActivityLogList.
func ParseListActivityLogsResult(r *api.Result) (*ActivityLogList, error) {
	return api.Decode[ActivityLogList](r)
}

// Me returns the currently authenticated admin along with the app.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/admins/identifyadmin
func (s *Service) Me(ctx context.Context) (*Admin, error) {
	result, err := s.MeRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseMeResult(result)
}

// Get retrieves an admin by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/admins/retrieveadmin
func (s *Service) Get(ctx context.Context, id string) (*Admin, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// List returns all admins for the workspace.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/admins/listadmins
func (s *Service) List(ctx context.Context) (*List, error) {
	result, err := s.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListResult(result)
}

// SetAway sets an admin's away status.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/admins/setawayadmin
func (s *Service) SetAway(ctx context.Context, id string, body *SetAwayRequest) (*Admin, error) {
	result, err := s.SetAwayRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseSetAwayResult(result)
}

// ListActivityLogs returns activity logs for the workspace.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/admins/listactivitylogs
func (s *Service) ListActivityLogs(ctx context.Context, opts *ActivityLogListOptions) (*ActivityLogList, error) {
	result, err := s.ListActivityLogsRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListActivityLogsResult(result)
}

// MeRaw returns the currently authenticated admin with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/admins/identifyadmin
func (s *Service) MeRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "me", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetRaw retrieves an admin by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/admins/retrieveadmin
func (s *Service) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("admins/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns all admins for the workspace with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/admins/listadmins
func (s *Service) ListRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "admins", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SetAwayRaw sets an admin's away status with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/admins/setawayadmin
func (s *Service) SetAwayRaw(ctx context.Context, id string, body *SetAwayRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("admins/%s/away", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListActivityLogsRaw returns activity logs for the workspace with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/admins/listactivitylogs
func (s *Service) ListActivityLogsRaw(ctx context.Context, opts *ActivityLogListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("admins/activity_logs", opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
