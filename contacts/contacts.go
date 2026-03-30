package contacts

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// Service handles communication with the contact related methods
// of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new contacts Service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Contact represents an Intercom contact (user or lead).
type Contact struct {
	Type                   string             `json:"type"`
	ID                     string             `json:"id"`
	WorkspaceID            string             `json:"workspace_id,omitempty"`
	ExternalID             string             `json:"external_id,omitempty"`
	Role                   string             `json:"role,omitempty"`
	Email                  string             `json:"email,omitempty"`
	Phone                  string             `json:"phone,omitempty"`
	Name                   string             `json:"name,omitempty"`
	Avatar                 string             `json:"avatar,omitempty"`
	OwnerID                *int               `json:"owner_id,omitempty"`
	HasHardBounced         bool               `json:"has_hard_bounced,omitempty"`
	MarkedEmailAsSpam      bool               `json:"marked_email_as_spam,omitempty"`
	UnsubscribedFromEmails bool               `json:"unsubscribed_from_emails,omitempty"`
	CreatedAt              int64              `json:"created_at,omitempty"`
	UpdatedAt              int64              `json:"updated_at,omitempty"`
	SignedUpAt             *int64             `json:"signed_up_at,omitempty"`
	LastSeenAt             *int64             `json:"last_seen_at,omitempty"`
	LastRepliedAt          *int64             `json:"last_replied_at,omitempty"`
	LastContactedAt        *int64             `json:"last_contacted_at,omitempty"`
	LastEmailOpenedAt      *int64             `json:"last_email_opened_at,omitempty"`
	LastEmailClickedAt     *int64             `json:"last_email_clicked_at,omitempty"`
	LanguageOverride       string             `json:"language_override,omitempty"`
	Browser                string             `json:"browser,omitempty"`
	BrowserVersion         string             `json:"browser_version,omitempty"`
	BrowserLanguage        string             `json:"browser_language,omitempty"`
	OS                     string             `json:"os,omitempty"`
	Location               *Location          `json:"location,omitempty"`
	AndroidAppName         string             `json:"android_app_name,omitempty"`
	AndroidAppVersion      string             `json:"android_app_version,omitempty"`
	AndroidDevice          string             `json:"android_device,omitempty"`
	AndroidOSVersion       string             `json:"android_os_version,omitempty"`
	AndroidSDKVersion      string             `json:"android_sdk_version,omitempty"`
	AndroidLastSeenAt      *int64             `json:"android_last_seen_at,omitempty"`
	IOSAppName             string             `json:"ios_app_name,omitempty"`
	IOSAppVersion          string             `json:"ios_app_version,omitempty"`
	IOSDevice              string             `json:"ios_device,omitempty"`
	IOSOSVersion           string             `json:"ios_os_version,omitempty"`
	IOSSDKVersion          string             `json:"ios_sdk_version,omitempty"`
	IOSLastSeenAt          *int64             `json:"ios_last_seen_at,omitempty"`
	CustomAttributes       map[string]any     `json:"custom_attributes,omitempty"`
	Tags                   *ListRef           `json:"tags,omitempty"`
	Notes                  *ListRef           `json:"notes,omitempty"`
	Companies              *ListRef           `json:"companies,omitempty"`
	SocialProfiles         *SocialProfileList `json:"social_profiles,omitempty"`
	UTMCampaign            string             `json:"utm_campaign,omitempty"`
	UTMContent             string             `json:"utm_content,omitempty"`
	UTMMedium              string             `json:"utm_medium,omitempty"`
	UTMSource              string             `json:"utm_source,omitempty"`
	UTMTerm                string             `json:"utm_term,omitempty"`
	Referrer               string             `json:"referrer,omitempty"`
}

// Location represents a contact's geographic location.
type Location struct {
	Type          string `json:"type,omitempty"`
	Country       string `json:"country,omitempty"`
	Region        string `json:"region,omitempty"`
	City          string `json:"city,omitempty"`
	CountryCode   string `json:"country_code,omitempty"`
	ContinentCode string `json:"continent_code,omitempty"`
}

// ListRef is a reference to a sub-resource list on a contact.
type ListRef struct {
	Type       string `json:"type,omitempty"`
	Data       []any  `json:"data,omitempty"`
	URL        string `json:"url,omitempty"`
	TotalCount int    `json:"total_count,omitempty"`
	HasMore    bool   `json:"has_more,omitempty"`
}

// SocialProfileList is a list of social profiles.
type SocialProfileList struct {
	Type string          `json:"type,omitempty"`
	Data []SocialProfile `json:"data,omitempty"`
}

// SocialProfile represents a social media profile.
type SocialProfile struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Deleted represents the response from deleting a contact.
type Deleted struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
	Deleted    bool   `json:"deleted"`
}

// Archived represents the response from archiving a contact.
type Archived struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
	Archived   bool   `json:"archived"`
}

// Unarchived represents the response from unarchiving a contact.
type Unarchived struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
	Archived   bool   `json:"archived"`
}

// Blocked represents the response from blocking a contact.
type Blocked struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
	Blocked    bool   `json:"blocked"`
}

// CreateRequest represents the body for creating a contact.
type CreateRequest struct {
	Role                   string         `json:"role,omitempty"`
	ExternalID             string         `json:"external_id,omitempty"`
	Email                  string         `json:"email,omitempty"`
	Phone                  string         `json:"phone,omitempty"`
	Name                   string         `json:"name,omitempty"`
	Avatar                 string         `json:"avatar,omitempty"`
	SignedUpAt             *int64         `json:"signed_up_at,omitempty"`
	LastSeenAt             *int64         `json:"last_seen_at,omitempty"`
	OwnerID                *int           `json:"owner_id,omitempty"`
	UnsubscribedFromEmails *bool          `json:"unsubscribed_from_emails,omitempty"`
	CustomAttributes       map[string]any `json:"custom_attributes,omitempty"`
}

// UpdateRequest represents the body for updating a contact.
type UpdateRequest struct {
	Role                   string         `json:"role,omitempty"`
	ExternalID             string         `json:"external_id,omitempty"`
	Email                  string         `json:"email,omitempty"`
	Phone                  string         `json:"phone,omitempty"`
	Name                   string         `json:"name,omitempty"`
	Avatar                 string         `json:"avatar,omitempty"`
	SignedUpAt             *int64         `json:"signed_up_at,omitempty"`
	LastSeenAt             *int64         `json:"last_seen_at,omitempty"`
	OwnerID                *int           `json:"owner_id,omitempty"`
	UnsubscribedFromEmails *bool          `json:"unsubscribed_from_emails,omitempty"`
	CustomAttributes       map[string]any `json:"custom_attributes,omitempty"`
}

// MergeRequest represents the body for merging a lead into a user.
type MergeRequest struct {
	From string `json:"from"`
	Into string `json:"into"`
}

// CompanyRef represents a company object in sub-resource responses.
type CompanyRef struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	CompanyID string `json:"company_id,omitempty"`
	Name      string `json:"name,omitempty"`
}

// CompanyListResult is the response for listing companies attached to a contact.
type CompanyListResult struct {
	Type       string          `json:"type"`
	Data       []CompanyRef    `json:"data"`
	TotalCount int             `json:"total_count"`
	Pages      api.CursorPages `json:"pages"`
}

// CreateNoteRequest represents the body for creating a note on a contact.
type CreateNoteRequest struct {
	Body    string `json:"body"`
	AdminID string `json:"admin_id,omitempty"`
}

// SubscriptionListResult is the response for listing subscriptions for a contact.
type SubscriptionListResult struct {
	Type string                 `json:"type"`
	Data []api.SubscriptionType `json:"data"`
}

// AddSubscriptionRequest represents the body for adding a subscription to a contact.
type AddSubscriptionRequest struct {
	ID          string `json:"id"`
	ConsentType string `json:"consent_type"`
}

// --- Parse Functions ---

// ParseGetResult decodes a Result into a Contact.
func ParseGetResult(r *api.Result) (*Contact, error) { return api.Decode[Contact](r) }

// ParseListResult decodes a Result into a PagedResult[Contact].
func ParseListResult(r *api.Result) (*api.PagedResult[Contact], error) {
	return api.Decode[api.PagedResult[Contact]](r)
}

// ParseCreateResult decodes a Result into a Contact.
func ParseCreateResult(r *api.Result) (*Contact, error) { return api.Decode[Contact](r) }

// ParseUpdateResult decodes a Result into a Contact.
func ParseUpdateResult(r *api.Result) (*Contact, error) { return api.Decode[Contact](r) }

// ParseDeleteResult decodes a Result into a Deleted.
func ParseDeleteResult(r *api.Result) (*Deleted, error) { return api.Decode[Deleted](r) }

// ParseSearchResult decodes a Result into a PagedResult[Contact].
func ParseSearchResult(r *api.Result) (*api.PagedResult[Contact], error) {
	return api.Decode[api.PagedResult[Contact]](r)
}

// ParseMergeResult decodes a Result into a Contact.
func ParseMergeResult(r *api.Result) (*Contact, error) { return api.Decode[Contact](r) }

// ParseArchiveResult decodes a Result into an Archived.
func ParseArchiveResult(r *api.Result) (*Archived, error) {
	return api.Decode[Archived](r)
}

// ParseUnarchiveResult decodes a Result into an Unarchived.
func ParseUnarchiveResult(r *api.Result) (*Unarchived, error) {
	return api.Decode[Unarchived](r)
}

// ParseBlockResult decodes a Result into a Blocked.
func ParseBlockResult(r *api.Result) (*Blocked, error) {
	return api.Decode[Blocked](r)
}

// ParseFindByExternalIDResult decodes a Result into a Contact.
func ParseFindByExternalIDResult(r *api.Result) (*Contact, error) { return api.Decode[Contact](r) }

// ParseListCompaniesResult decodes a Result into a CompanyListResult.
func ParseListCompaniesResult(r *api.Result) (*CompanyListResult, error) {
	return api.Decode[CompanyListResult](r)
}

// ParseAddCompanyResult decodes a Result into a CompanyRef.
func ParseAddCompanyResult(r *api.Result) (*CompanyRef, error) { return api.Decode[CompanyRef](r) }

// ParseRemoveCompanyResult decodes a Result into a CompanyRef.
func ParseRemoveCompanyResult(r *api.Result) (*CompanyRef, error) { return api.Decode[CompanyRef](r) }

// ParseListNotesResult decodes a Result into a NoteListResult.
func ParseListNotesResult(r *api.Result) (*api.NoteListResult, error) {
	return api.Decode[api.NoteListResult](r)
}

// ParseCreateNoteResult decodes a Result into a Note.
func ParseCreateNoteResult(r *api.Result) (*api.Note, error) { return api.Decode[api.Note](r) }

// ParseListSegmentsResult decodes a Result into a SegmentListResult.
func ParseListSegmentsResult(r *api.Result) (*api.SegmentListResult, error) {
	return api.Decode[api.SegmentListResult](r)
}

// ParseListSubscriptionsResult decodes a Result into a SubscriptionListResult.
func ParseListSubscriptionsResult(r *api.Result) (*SubscriptionListResult, error) {
	return api.Decode[SubscriptionListResult](r)
}

// ParseAddSubscriptionResult decodes a Result into a SubscriptionType.
func ParseAddSubscriptionResult(r *api.Result) (*api.SubscriptionType, error) {
	return api.Decode[api.SubscriptionType](r)
}

// ParseRemoveSubscriptionResult decodes a Result into a SubscriptionType.
func ParseRemoveSubscriptionResult(r *api.Result) (*api.SubscriptionType, error) {
	return api.Decode[api.SubscriptionType](r)
}

// ParseAddTagResult decodes a Result into a TagRef.
func ParseAddTagResult(r *api.Result) (*api.TagRef, error) { return api.Decode[api.TagRef](r) }

// ParseRemoveTagResult decodes a Result into a TagRef.
func ParseRemoveTagResult(r *api.Result) (*api.TagRef, error) { return api.Decode[api.TagRef](r) }

// ParseListTagsResult decodes a Result into an api.TagList.
func ParseListTagsResult(r *api.Result) (*api.TagList, error) {
	return api.Decode[api.TagList](r)
}

// --- Regular Methods ---

// Get retrieves a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/showcontact
func (s *Service) Get(ctx context.Context, id string) (*Contact, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetResult(result)
}

// List returns a single page of contacts.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listcontacts
func (s *Service) List(ctx context.Context, opts *api.ListOptions) (*api.PagedResult[Contact], error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListResult(result)
}

// ListAll returns an iterator over all contacts, handling pagination automatically.
func (s *Service) ListAll(ctx context.Context, opts *api.ListOptions) *api.Iter[Contact] {
	return api.NewIter[Contact](ctx, opts, s.List)
}

// Create creates a new contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/createcontact
func (s *Service) Create(ctx context.Context, body *CreateRequest) (*Contact, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateResult(result)
}

// Update updates an existing contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/updatecontact
func (s *Service) Update(ctx context.Context, id string, body *UpdateRequest) (*Contact, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseUpdateResult(result)
}

// Delete deletes a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/deletecontact
func (s *Service) Delete(ctx context.Context, id string) (*Deleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseDeleteResult(result)
}

// Search searches for contacts using the provided query filters.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/searchcontacts
func (s *Service) Search(ctx context.Context, body *api.SearchRequest) (*api.PagedResult[Contact], error) {
	result, err := s.SearchRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseSearchResult(result)
}

// Merge merges a lead into a user. The From field must be a lead, and
// the Into field must be a user.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/mergecontact
func (s *Service) Merge(ctx context.Context, body *MergeRequest) (*Contact, error) {
	result, err := s.MergeRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseMergeResult(result)
}

// Archive archives a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/archivecontact
func (s *Service) Archive(ctx context.Context, id string) (*Archived, error) {
	result, err := s.ArchiveRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseArchiveResult(result)
}

// Unarchive unarchives a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/unarchivecontact
func (s *Service) Unarchive(ctx context.Context, id string) (*Unarchived, error) {
	result, err := s.UnarchiveRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseUnarchiveResult(result)
}

// Block blocks a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/blockcontact
func (s *Service) Block(ctx context.Context, id string) (*Blocked, error) {
	result, err := s.BlockRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseBlockResult(result)
}

// FindByExternalID retrieves a contact by its external ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/showcontactbyexternalid
func (s *Service) FindByExternalID(ctx context.Context, externalID string) (*Contact, error) {
	result, err := s.FindByExternalIDRaw(ctx, externalID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseFindByExternalIDResult(result)
}

// ListCompanies returns the companies attached to a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listcompaniesforacontact
func (s *Service) ListCompanies(ctx context.Context, contactID string) (*CompanyListResult, error) {
	result, err := s.ListCompaniesRaw(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListCompaniesResult(result)
}

// AddCompany attaches a company to a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/attachcontacttoacompany
func (s *Service) AddCompany(ctx context.Context, contactID, companyID string) (*CompanyRef, error) {
	result, err := s.AddCompanyRaw(ctx, contactID, companyID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAddCompanyResult(result)
}

// RemoveCompany detaches a company from a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/detachcontactfromacompany
func (s *Service) RemoveCompany(ctx context.Context, contactID, companyID string) (*CompanyRef, error) {
	result, err := s.RemoveCompanyRaw(ctx, contactID, companyID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseRemoveCompanyResult(result)
}

// ListNotes returns the notes for a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listnotes
func (s *Service) ListNotes(ctx context.Context, contactID string) (*api.NoteListResult, error) {
	result, err := s.ListNotesRaw(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListNotesResult(result)
}

// CreateNote creates a note on a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/createnote
func (s *Service) CreateNote(ctx context.Context, contactID string, body *CreateNoteRequest) (*api.Note, error) {
	result, err := s.CreateNoteRaw(ctx, contactID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateNoteResult(result)
}

// ListSegments returns the segments for a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listsegmentsforacontact
func (s *Service) ListSegments(ctx context.Context, contactID string) (*api.SegmentListResult, error) {
	result, err := s.ListSegmentsRaw(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListSegmentsResult(result)
}

// ListSubscriptions returns the subscriptions for a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/listsubscriptionsforacontact
func (s *Service) ListSubscriptions(ctx context.Context, contactID string) (*SubscriptionListResult, error) {
	result, err := s.ListSubscriptionsRaw(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListSubscriptionsResult(result)
}

// AddSubscription adds a subscription to a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/attachsubscriptiontypetocontact
func (s *Service) AddSubscription(ctx context.Context, contactID string, body *AddSubscriptionRequest) (*api.SubscriptionType, error) {
	result, err := s.AddSubscriptionRaw(ctx, contactID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAddSubscriptionResult(result)
}

// RemoveSubscription removes a subscription from a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/detachsubscriptiontypetocontact
func (s *Service) RemoveSubscription(ctx context.Context, contactID, subscriptionID string) (*api.SubscriptionType, error) {
	result, err := s.RemoveSubscriptionRaw(ctx, contactID, subscriptionID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseRemoveSubscriptionResult(result)
}

// AddTag adds a tag to a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/attachtagtocontact
func (s *Service) AddTag(ctx context.Context, contactID, tagID string) (*api.TagRef, error) {
	result, err := s.AddTagRaw(ctx, contactID, tagID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseAddTagResult(result)
}

// RemoveTag removes a tag from a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/detachtagfromcontact
func (s *Service) RemoveTag(ctx context.Context, contactID, tagID string) (*api.TagRef, error) {
	result, err := s.RemoveTagRaw(ctx, contactID, tagID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseRemoveTagResult(result)
}

// ListTags returns the tags attached to a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listtagsforacontact
func (s *Service) ListTags(ctx context.Context, contactID string) (*api.TagList, error) {
	result, err := s.ListTagsRaw(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseListTagsResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/showcontact
func (s *Service) GetRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns a single page of contacts with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listcontacts
func (s *Service) ListRaw(ctx context.Context, opts *api.ListOptions) (*api.Result, error) {
	path, err := api.AddQueryOptions("contacts", opts)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateRaw creates a new contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/createcontact
func (s *Service) CreateRaw(ctx context.Context, body *CreateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "contacts", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/updatecontact
func (s *Service) UpdateRaw(ctx context.Context, id string, body *UpdateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("contacts/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes a contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/deletecontact
func (s *Service) DeleteRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("contacts/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for contacts and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/searchcontacts
func (s *Service) SearchRaw(ctx context.Context, body *api.SearchRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "contacts/search", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// MergeRaw merges a lead into a user and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/mergecontact
func (s *Service) MergeRaw(ctx context.Context, body *MergeRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "contacts/merge", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ArchiveRaw archives a contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/archivecontact
func (s *Service) ArchiveRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/archive", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UnarchiveRaw unarchives a contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/unarchivecontact
func (s *Service) UnarchiveRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/unarchive", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// BlockRaw blocks a contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/blockcontact
func (s *Service) BlockRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/block", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// FindByExternalIDRaw retrieves a contact by external ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/showcontactbyexternalid
func (s *Service) FindByExternalIDRaw(ctx context.Context, externalID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/find_by_external_id/%s", url.PathEscape(externalID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListCompaniesRaw returns the companies attached to a contact with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listcompaniesforacontact
func (s *Service) ListCompaniesRaw(ctx context.Context, contactID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s/companies", url.PathEscape(contactID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddCompanyRaw attaches a company to a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/attachcontacttoacompany
func (s *Service) AddCompanyRaw(ctx context.Context, contactID, companyID string) (*api.Result, error) {
	body := struct {
		ID string `json:"id"`
	}{ID: companyID}
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/companies", url.PathEscape(contactID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// RemoveCompanyRaw detaches a company from a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/detachcontactfromacompany
func (s *Service) RemoveCompanyRaw(ctx context.Context, contactID, companyID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("contacts/%s/companies/%s", url.PathEscape(contactID), url.PathEscape(companyID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListNotesRaw returns the notes for a contact with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listnotes
func (s *Service) ListNotesRaw(ctx context.Context, contactID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s/notes", url.PathEscape(contactID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateNoteRaw creates a note on a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/createnote
func (s *Service) CreateNoteRaw(ctx context.Context, contactID string, body *CreateNoteRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/notes", url.PathEscape(contactID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListSegmentsRaw returns the segments for a contact with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listsegmentsforacontact
func (s *Service) ListSegmentsRaw(ctx context.Context, contactID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s/segments", url.PathEscape(contactID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListSubscriptionsRaw returns the subscriptions for a contact with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/listsubscriptionsforacontact
func (s *Service) ListSubscriptionsRaw(ctx context.Context, contactID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s/subscriptions", url.PathEscape(contactID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddSubscriptionRaw adds a subscription to a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/attachsubscriptiontypetocontact
func (s *Service) AddSubscriptionRaw(ctx context.Context, contactID string, body *AddSubscriptionRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/subscriptions", url.PathEscape(contactID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// RemoveSubscriptionRaw removes a subscription from a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/detachsubscriptiontypetocontact
func (s *Service) RemoveSubscriptionRaw(ctx context.Context, contactID, subscriptionID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("contacts/%s/subscriptions/%s", url.PathEscape(contactID), url.PathEscape(subscriptionID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddTagRaw adds a tag to a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/attachtagtocontact
func (s *Service) AddTagRaw(ctx context.Context, contactID, tagID string) (*api.Result, error) {
	body := struct {
		ID string `json:"id"`
	}{ID: tagID}
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/tags", url.PathEscape(contactID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// RemoveTagRaw removes a tag from a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/detachtagfromcontact
func (s *Service) RemoveTagRaw(ctx context.Context, contactID, tagID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("contacts/%s/tags/%s", url.PathEscape(contactID), url.PathEscape(tagID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListTagsRaw returns the tags attached to a contact with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listtagsforacontact
func (s *Service) ListTagsRaw(ctx context.Context, contactID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s/tags", url.PathEscape(contactID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
