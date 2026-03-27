package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ContactsService handles communication with the contact related methods
// of the Intercom API.
type ContactsService service

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
	Location               *ContactLocation   `json:"location,omitempty"`
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
	Tags                   *ContactListRef    `json:"tags,omitempty"`
	Notes                  *ContactListRef    `json:"notes,omitempty"`
	Companies              *ContactListRef    `json:"companies,omitempty"`
	SocialProfiles         *SocialProfileList `json:"social_profiles,omitempty"`
	UTMCampaign            string             `json:"utm_campaign,omitempty"`
	UTMContent             string             `json:"utm_content,omitempty"`
	UTMMedium              string             `json:"utm_medium,omitempty"`
	UTMSource              string             `json:"utm_source,omitempty"`
	UTMTerm                string             `json:"utm_term,omitempty"`
	Referrer               string             `json:"referrer,omitempty"`
}

// ContactLocation represents a contact's geographic location.
type ContactLocation struct {
	Type          string `json:"type,omitempty"`
	Country       string `json:"country,omitempty"`
	Region        string `json:"region,omitempty"`
	City          string `json:"city,omitempty"`
	CountryCode   string `json:"country_code,omitempty"`
	ContinentCode string `json:"continent_code,omitempty"`
}

// ContactListRef is a reference to a sub-resource list on a contact.
type ContactListRef struct {
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

// ContactDeleted represents the response from deleting a contact.
type ContactDeleted struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
	Deleted    bool   `json:"deleted"`
}

// ContactArchived represents the response from archiving a contact.
type ContactArchived struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
	Archived   bool   `json:"archived"`
}

// ContactUnarchived represents the response from unarchiving a contact.
type ContactUnarchived struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
	Archived   bool   `json:"archived"`
}

// ContactBlocked represents the response from blocking a contact.
type ContactBlocked struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
	Blocked    bool   `json:"blocked"`
}

// CreateContactRequest represents the body for creating a contact.
type CreateContactRequest struct {
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

// UpdateContactRequest represents the body for updating a contact.
type UpdateContactRequest struct {
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

// MergeContactsRequest represents the body for merging a lead into a user.
type MergeContactsRequest struct {
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
	Type       string       `json:"type"`
	Data       []CompanyRef `json:"data"`
	TotalCount int          `json:"total_count"`
	Pages      CursorPages  `json:"pages"`
}

// Note represents an Intercom note on a contact.
type Note struct {
	Type      string      `json:"type"`
	ID        string      `json:"id"`
	CreatedAt int64       `json:"created_at,omitempty"`
	Body      string      `json:"body,omitempty"`
	Contact   *ContactRef `json:"contact,omitempty"`
	Author    *NoteAuthor `json:"author,omitempty"`
}

// ContactRef is a lightweight reference to a contact.
type ContactRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// NoteAuthor represents the admin who authored a note.
type NoteAuthor struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// NoteListResult is the response for listing notes on a contact.
type NoteListResult struct {
	Type       string      `json:"type"`
	Data       []Note      `json:"data"`
	TotalCount int         `json:"total_count"`
	Pages      CursorPages `json:"pages"`
}

// CreateNoteRequest represents the body for creating a note on a contact.
type CreateNoteRequest struct {
	Body    string `json:"body"`
	AdminID string `json:"admin_id,omitempty"`
}

// SegmentRef represents a segment in sub-resource responses.
type SegmentRef struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	UpdatedAt  int64  `json:"updated_at,omitempty"`
	PersonType string `json:"person_type,omitempty"`
}

// SegmentListResult is the response for listing segments for a contact.
type SegmentListResult struct {
	Type string       `json:"type"`
	Data []SegmentRef `json:"data"`
}

// SubscriptionListResult is the response for listing subscriptions for a contact.
type SubscriptionListResult struct {
	Type string             `json:"type"`
	Data []SubscriptionType `json:"data"`
}

// AddSubscriptionRequest represents the body for adding a subscription to a contact.
type AddSubscriptionRequest struct {
	ID          string `json:"id"`
	ConsentType string `json:"consent_type"`
}

// TagRef represents a tag in sub-resource responses.
type TagRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// ParseContactGetResult decodes a Result into a Contact.
func ParseContactGetResult(r *Result) (*Contact, error) { return Decode[Contact](r) }

// ParseContactListResult decodes a Result into a PagedResult[Contact].
func ParseContactListResult(r *Result) (*PagedResult[Contact], error) {
	return Decode[PagedResult[Contact]](r)
}

// ParseContactCreateResult decodes a Result into a Contact.
func ParseContactCreateResult(r *Result) (*Contact, error) { return Decode[Contact](r) }

// ParseContactUpdateResult decodes a Result into a Contact.
func ParseContactUpdateResult(r *Result) (*Contact, error) { return Decode[Contact](r) }

// ParseContactDeleteResult decodes a Result into a ContactDeleted.
func ParseContactDeleteResult(r *Result) (*ContactDeleted, error) { return Decode[ContactDeleted](r) }

// ParseContactSearchResult decodes a Result into a PagedResult[Contact].
func ParseContactSearchResult(r *Result) (*PagedResult[Contact], error) {
	return Decode[PagedResult[Contact]](r)
}

// ParseContactMergeResult decodes a Result into a Contact.
func ParseContactMergeResult(r *Result) (*Contact, error) { return Decode[Contact](r) }

// ParseContactArchiveResult decodes a Result into a ContactArchived.
func ParseContactArchiveResult(r *Result) (*ContactArchived, error) {
	return Decode[ContactArchived](r)
}

// ParseContactUnarchiveResult decodes a Result into a ContactUnarchived.
func ParseContactUnarchiveResult(r *Result) (*ContactUnarchived, error) {
	return Decode[ContactUnarchived](r)
}

// ParseContactBlockResult decodes a Result into a ContactBlocked.
func ParseContactBlockResult(r *Result) (*ContactBlocked, error) {
	return Decode[ContactBlocked](r)
}

// ParseContactFindByExternalIDResult decodes a Result into a Contact.
func ParseContactFindByExternalIDResult(r *Result) (*Contact, error) { return Decode[Contact](r) }

// ParseContactListCompaniesResult decodes a Result into a CompanyListResult.
func ParseContactListCompaniesResult(r *Result) (*CompanyListResult, error) {
	return Decode[CompanyListResult](r)
}

// ParseContactAddCompanyResult decodes a Result into a CompanyRef.
func ParseContactAddCompanyResult(r *Result) (*CompanyRef, error) { return Decode[CompanyRef](r) }

// ParseContactRemoveCompanyResult decodes a Result into a CompanyRef.
func ParseContactRemoveCompanyResult(r *Result) (*CompanyRef, error) { return Decode[CompanyRef](r) }

// ParseContactListNotesResult decodes a Result into a NoteListResult.
func ParseContactListNotesResult(r *Result) (*NoteListResult, error) {
	return Decode[NoteListResult](r)
}

// ParseContactCreateNoteResult decodes a Result into a Note.
func ParseContactCreateNoteResult(r *Result) (*Note, error) { return Decode[Note](r) }

// ParseContactListSegmentsResult decodes a Result into a SegmentListResult.
func ParseContactListSegmentsResult(r *Result) (*SegmentListResult, error) {
	return Decode[SegmentListResult](r)
}

// ParseContactListSubscriptionsResult decodes a Result into a SubscriptionListResult.
func ParseContactListSubscriptionsResult(r *Result) (*SubscriptionListResult, error) {
	return Decode[SubscriptionListResult](r)
}

// ParseContactAddSubscriptionResult decodes a Result into a SubscriptionType.
func ParseContactAddSubscriptionResult(r *Result) (*SubscriptionType, error) {
	return Decode[SubscriptionType](r)
}

// ParseContactRemoveSubscriptionResult decodes a Result into a SubscriptionType.
func ParseContactRemoveSubscriptionResult(r *Result) (*SubscriptionType, error) {
	return Decode[SubscriptionType](r)
}

// ParseContactAddTagResult decodes a Result into a TagRef.
func ParseContactAddTagResult(r *Result) (*TagRef, error) { return Decode[TagRef](r) }

// ParseContactRemoveTagResult decodes a Result into a TagRef.
func ParseContactRemoveTagResult(r *Result) (*TagRef, error) { return Decode[TagRef](r) }

// Get retrieves a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/showcontact
func (s *ContactsService) Get(ctx context.Context, id string) (*Contact, error) {
	result, err := s.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactGetResult(result)
}

// List returns a single page of contacts.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listcontacts
func (s *ContactsService) List(ctx context.Context, opts *ListOptions) (*PagedResult[Contact], error) {
	result, err := s.ListRaw(ctx, opts)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactListResult(result)
}

// ListAll returns an iterator over all contacts, handling pagination automatically.
func (s *ContactsService) ListAll(ctx context.Context, opts *ListOptions) *Iter[Contact] {
	return NewIter[Contact](ctx, opts, s.List)
}

// Create creates a new contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/createcontact
func (s *ContactsService) Create(ctx context.Context, body *CreateContactRequest) (*Contact, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactCreateResult(result)
}

// Update updates an existing contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/updatecontact
func (s *ContactsService) Update(ctx context.Context, id string, body *UpdateContactRequest) (*Contact, error) {
	result, err := s.UpdateRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactUpdateResult(result)
}

// Delete deletes a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/deletecontact
func (s *ContactsService) Delete(ctx context.Context, id string) (*ContactDeleted, error) {
	result, err := s.DeleteRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactDeleteResult(result)
}

// Search searches for contacts using the provided query filters.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/searchcontacts
func (s *ContactsService) Search(ctx context.Context, body *SearchRequest) (*PagedResult[Contact], error) {
	result, err := s.SearchRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactSearchResult(result)
}

// Merge merges a lead into a user. The From field must be a lead, and
// the Into field must be a user.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/mergecontact
func (s *ContactsService) Merge(ctx context.Context, body *MergeContactsRequest) (*Contact, error) {
	result, err := s.MergeRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactMergeResult(result)
}

// Archive archives a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/archivecontact
func (s *ContactsService) Archive(ctx context.Context, id string) (*ContactArchived, error) {
	result, err := s.ArchiveRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactArchiveResult(result)
}

// Unarchive unarchives a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/unarchivecontact
func (s *ContactsService) Unarchive(ctx context.Context, id string) (*ContactUnarchived, error) {
	result, err := s.UnarchiveRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactUnarchiveResult(result)
}

// Block blocks a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/blockcontact
func (s *ContactsService) Block(ctx context.Context, id string) (*ContactBlocked, error) {
	result, err := s.BlockRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactBlockResult(result)
}

// FindByExternalID retrieves a contact by its external ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/showcontactbyexternalid
func (s *ContactsService) FindByExternalID(ctx context.Context, externalID string) (*Contact, error) {
	result, err := s.FindByExternalIDRaw(ctx, externalID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactFindByExternalIDResult(result)
}

// ListCompanies returns the companies attached to a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listcompaniesforacontact
func (s *ContactsService) ListCompanies(ctx context.Context, contactID string) (*CompanyListResult, error) {
	result, err := s.ListCompaniesRaw(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactListCompaniesResult(result)
}

// AddCompany attaches a company to a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/attachcontacttoacompany
func (s *ContactsService) AddCompany(ctx context.Context, contactID, companyID string) (*CompanyRef, error) {
	result, err := s.AddCompanyRaw(ctx, contactID, companyID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactAddCompanyResult(result)
}

// RemoveCompany detaches a company from a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/detachcontactfromacompany
func (s *ContactsService) RemoveCompany(ctx context.Context, contactID, companyID string) (*CompanyRef, error) {
	result, err := s.RemoveCompanyRaw(ctx, contactID, companyID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactRemoveCompanyResult(result)
}

// ListNotes returns the notes for a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listnotes
func (s *ContactsService) ListNotes(ctx context.Context, contactID string) (*NoteListResult, error) {
	result, err := s.ListNotesRaw(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactListNotesResult(result)
}

// CreateNote creates a note on a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/createnote
func (s *ContactsService) CreateNote(ctx context.Context, contactID string, body *CreateNoteRequest) (*Note, error) {
	result, err := s.CreateNoteRaw(ctx, contactID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactCreateNoteResult(result)
}

// ListSegments returns the segments for a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listsegmentsforacontact
func (s *ContactsService) ListSegments(ctx context.Context, contactID string) (*SegmentListResult, error) {
	result, err := s.ListSegmentsRaw(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactListSegmentsResult(result)
}

// ListSubscriptions returns the subscriptions for a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/listsubscriptionsforacontact
func (s *ContactsService) ListSubscriptions(ctx context.Context, contactID string) (*SubscriptionListResult, error) {
	result, err := s.ListSubscriptionsRaw(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactListSubscriptionsResult(result)
}

// AddSubscription adds a subscription to a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/attachsubscriptiontypetocontact
func (s *ContactsService) AddSubscription(ctx context.Context, contactID string, body *AddSubscriptionRequest) (*SubscriptionType, error) {
	result, err := s.AddSubscriptionRaw(ctx, contactID, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactAddSubscriptionResult(result)
}

// RemoveSubscription removes a subscription from a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/detachsubscriptiontypetocontact
func (s *ContactsService) RemoveSubscription(ctx context.Context, contactID, subscriptionID string) (*SubscriptionType, error) {
	result, err := s.RemoveSubscriptionRaw(ctx, contactID, subscriptionID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactRemoveSubscriptionResult(result)
}

// AddTag adds a tag to a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/attachtagtocontact
func (s *ContactsService) AddTag(ctx context.Context, contactID, tagID string) (*TagRef, error) {
	result, err := s.AddTagRaw(ctx, contactID, tagID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactAddTagResult(result)
}

// RemoveTag removes a tag from a contact.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/detachtagfromcontact
func (s *ContactsService) RemoveTag(ctx context.Context, contactID, tagID string) (*TagRef, error) {
	result, err := s.RemoveTagRaw(ctx, contactID, tagID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseContactRemoveTagResult(result)
}

// GetRaw retrieves a contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/showcontact
func (s *ContactsService) GetRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListRaw returns a single page of contacts with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listcontacts
func (s *ContactsService) ListRaw(ctx context.Context, opts *ListOptions) (*Result, error) {
	path, err := addQueryOptions("contacts", opts)
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
func (s *ContactsService) CreateRaw(ctx context.Context, body *CreateContactRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "contacts", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateRaw updates an existing contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/updatecontact
func (s *ContactsService) UpdateRaw(ctx context.Context, id string, body *UpdateContactRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("contacts/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes a contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/deletecontact
func (s *ContactsService) DeleteRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("contacts/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// SearchRaw searches for contacts and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/searchcontacts
func (s *ContactsService) SearchRaw(ctx context.Context, body *SearchRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "contacts/search", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// MergeRaw merges a lead into a user and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/mergecontact
func (s *ContactsService) MergeRaw(ctx context.Context, body *MergeContactsRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "contacts/merge", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ArchiveRaw archives a contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/archivecontact
func (s *ContactsService) ArchiveRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/archive", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UnarchiveRaw unarchives a contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/unarchivecontact
func (s *ContactsService) UnarchiveRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/unarchive", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// BlockRaw blocks a contact by ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/blockcontact
func (s *ContactsService) BlockRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/block", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// FindByExternalIDRaw retrieves a contact by external ID and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/showcontactbyexternalid
func (s *ContactsService) FindByExternalIDRaw(ctx context.Context, externalID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/find_by_external_id/%s", url.PathEscape(externalID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListCompaniesRaw returns the companies attached to a contact with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listcompaniesforacontact
func (s *ContactsService) ListCompaniesRaw(ctx context.Context, contactID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s/companies", url.PathEscape(contactID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddCompanyRaw attaches a company to a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/attachcontacttoacompany
func (s *ContactsService) AddCompanyRaw(ctx context.Context, contactID, companyID string) (*Result, error) {
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
func (s *ContactsService) RemoveCompanyRaw(ctx context.Context, contactID, companyID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("contacts/%s/companies/%s", url.PathEscape(contactID), url.PathEscape(companyID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListNotesRaw returns the notes for a contact with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listnotes
func (s *ContactsService) ListNotesRaw(ctx context.Context, contactID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s/notes", url.PathEscape(contactID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateNoteRaw creates a note on a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/createnote
func (s *ContactsService) CreateNoteRaw(ctx context.Context, contactID string, body *CreateNoteRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/notes", url.PathEscape(contactID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListSegmentsRaw returns the segments for a contact with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/listsegmentsforacontact
func (s *ContactsService) ListSegmentsRaw(ctx context.Context, contactID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s/segments", url.PathEscape(contactID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// ListSubscriptionsRaw returns the subscriptions for a contact with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/listsubscriptionsforacontact
func (s *ContactsService) ListSubscriptionsRaw(ctx context.Context, contactID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s/subscriptions", url.PathEscape(contactID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddSubscriptionRaw adds a subscription to a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/attachsubscriptiontypetocontact
func (s *ContactsService) AddSubscriptionRaw(ctx context.Context, contactID string, body *AddSubscriptionRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("contacts/%s/subscriptions", url.PathEscape(contactID)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// RemoveSubscriptionRaw removes a subscription from a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/subscription-types/detachsubscriptiontypetocontact
func (s *ContactsService) RemoveSubscriptionRaw(ctx context.Context, contactID, subscriptionID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("contacts/%s/subscriptions/%s", url.PathEscape(contactID), url.PathEscape(subscriptionID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// AddTagRaw adds a tag to a contact and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/tags/attachtagtocontact
func (s *ContactsService) AddTagRaw(ctx context.Context, contactID, tagID string) (*Result, error) {
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
func (s *ContactsService) RemoveTagRaw(ctx context.Context, contactID, tagID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("contacts/%s/tags/%s", url.PathEscape(contactID), url.PathEscape(tagID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
