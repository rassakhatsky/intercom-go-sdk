package contacts

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// VisitorsService handles communication with the visitor related
// methods of the Intercom API.
type VisitorsService struct {
	client api.Caller
}

// NewVisitorsService creates a new VisitorsService.
func NewVisitorsService(c api.Caller) *VisitorsService {
	return &VisitorsService{client: c}
}

// Visitor represents a visitor in Intercom. Visitors are anonymous people
// that have not yet been identified.
type Visitor struct {
	Type                   string                 `json:"type"`
	ID                     string                 `json:"id"`
	UserID                 string                 `json:"user_id,omitempty"`
	Anonymous              bool                   `json:"anonymous,omitempty"`
	Email                  string                 `json:"email,omitempty"`
	Phone                  string                 `json:"phone,omitempty"`
	Name                   string                 `json:"name,omitempty"`
	Pseudonym              string                 `json:"pseudonym,omitempty"`
	Avatar                 *VisitorAvatar         `json:"avatar,omitempty"`
	AppID                  string                 `json:"app_id,omitempty"`
	Companies              *VisitorCompanies      `json:"companies,omitempty"`
	LocationData           *VisitorLocation       `json:"location_data,omitempty"`
	LastRequestAt          int64                  `json:"last_request_at,omitempty"`
	CreatedAt              int64                  `json:"created_at,omitempty"`
	RemoteCreatedAt        int64                  `json:"remote_created_at,omitempty"`
	SignedUpAt             int64                  `json:"signed_up_at,omitempty"`
	UpdatedAt              int64                  `json:"updated_at,omitempty"`
	SessionCount           int                    `json:"session_count,omitempty"`
	SocialProfiles         *VisitorSocialProfiles `json:"social_profiles,omitempty"`
	OwnerID                *string                `json:"owner_id,omitempty"`
	UnsubscribedFromEmails bool                   `json:"unsubscribed_from_emails,omitempty"`
	MarkedEmailAsSpam      bool                   `json:"marked_email_as_spam,omitempty"`
	HasHardBounced         bool                   `json:"has_hard_bounced,omitempty"`
	Tags                   *VisitorTags           `json:"tags,omitempty"`
	Segments               *VisitorSegments       `json:"segments,omitempty"`
	CustomAttributes       map[string]any         `json:"custom_attributes,omitempty"`
	Referrer               string                 `json:"referrer,omitempty"`
	UTMCampaign            string                 `json:"utm_campaign,omitempty"`
	UTMContent             string                 `json:"utm_content,omitempty"`
	UTMMedium              string                 `json:"utm_medium,omitempty"`
	UTMSource              string                 `json:"utm_source,omitempty"`
	UTMTerm                string                 `json:"utm_term,omitempty"`
	DoNotTrack             *bool                  `json:"do_not_track,omitempty"`
}

// VisitorAvatar represents a visitor's avatar.
type VisitorAvatar struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url,omitempty"`
}

// VisitorCompanies represents the companies associated with a visitor.
type VisitorCompanies struct {
	Type      string `json:"type"`
	Companies []any  `json:"companies,omitempty"`
}

// VisitorLocation represents a visitor's location data.
type VisitorLocation struct {
	Type          string `json:"type,omitempty"`
	CityName      string `json:"city_name,omitempty"`
	ContinentCode string `json:"continent_code,omitempty"`
	CountryCode   string `json:"country_code,omitempty"`
	CountryName   string `json:"country_name,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	RegionName    string `json:"region_name,omitempty"`
	Timezone      string `json:"timezone,omitempty"`
}

// VisitorSocialProfiles represents a visitor's social profiles.
type VisitorSocialProfiles struct {
	Type           string `json:"type"`
	SocialProfiles []any  `json:"social_profiles,omitempty"`
}

// VisitorTags represents tags associated with a visitor.
type VisitorTags struct {
	Type string `json:"type"`
	Tags []any  `json:"tags,omitempty"`
}

// VisitorSegments represents segments associated with a visitor.
type VisitorSegments struct {
	Type     string `json:"type"`
	Segments []any  `json:"segments,omitempty"`
}

// UpdateVisitorRequest represents a request to update a visitor.
// At least one of ID or UserID must be provided.
type UpdateVisitorRequest struct {
	ID               string         `json:"id,omitempty"`
	UserID           string         `json:"user_id,omitempty"`
	Name             string         `json:"name,omitempty"`
	CustomAttributes map[string]any `json:"custom_attributes,omitempty"`
}

// ConvertVisitorRequest represents a request to convert a visitor to a contact.
type ConvertVisitorRequest struct {
	Type    string                   `json:"type"`
	Visitor ConvertVisitorIdentifier `json:"visitor"`
	User    ConvertVisitorUser       `json:"user"`
}

// ConvertVisitorIdentifier identifies the visitor to convert.
type ConvertVisitorIdentifier struct {
	ID     string `json:"id,omitempty"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
}

// ConvertVisitorUser specifies the contact identifiers retained after conversion.
type ConvertVisitorUser struct {
	ID     string `json:"id,omitempty"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
}

// --- Parse Functions ---

// ParseVisitorGetResult decodes a Result into a Visitor.
func ParseVisitorGetResult(r *api.Result) (*Visitor, error) {
	return api.Decode[Visitor](r)
}

// ParseVisitorUpdateResult decodes a Result into a Visitor.
func ParseVisitorUpdateResult(r *api.Result) (*Visitor, error) {
	return api.Decode[Visitor](r)
}

// ParseVisitorConvertResult decodes a Result into a Contact.
func ParseVisitorConvertResult(r *api.Result) (*Contact, error) {
	return api.Decode[Contact](r)
}

// --- Regular Methods ---

// Get retrieves a visitor by their user_id.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/visitors/retrievevisitorwithuserid
func (s *VisitorsService) Get(ctx context.Context, userID string) (*Visitor, error) {
	result, err := s.GetRaw(ctx, userID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseVisitorGetResult(result)
}

// Update updates a visitor. The request must include either ID or UserID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/visitors/updatevisitor
func (s *VisitorsService) Update(ctx context.Context, input *UpdateVisitorRequest) (*Visitor, error) {
	result, err := s.UpdateRaw(ctx, input)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseVisitorUpdateResult(result)
}

// Convert converts a visitor to a contact (user or lead).
// The returned Contact is the newly created or merged contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/visitors/convertvisitor
func (s *VisitorsService) Convert(ctx context.Context, input *ConvertVisitorRequest) (*Contact, error) {
	result, err := s.ConvertRaw(ctx, input)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseVisitorConvertResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a visitor by user_id with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/visitors/retrievevisitorwithuserid
func (s *VisitorsService) GetRaw(ctx context.Context, userID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("visitors?user_id=%s", url.QueryEscape(userID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates a visitor with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/visitors/updatevisitor
func (s *VisitorsService) UpdateRaw(ctx context.Context, input *UpdateVisitorRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, "visitors", input)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ConvertRaw converts a visitor to a contact with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/visitors/convertvisitor
func (s *VisitorsService) ConvertRaw(ctx context.Context, input *ConvertVisitorRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "visitors/convert", input)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
