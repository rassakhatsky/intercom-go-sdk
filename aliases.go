package intercom

import (
	"context"

	"github.com/rassakhatsky/intercom-go-sdk/contacts"
	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
	"github.com/rassakhatsky/intercom-go-sdk/messaging"
	"github.com/rassakhatsky/intercom-go-sdk/tags"
)

// Type aliases re-export internal/api types as part of the public intercom API.
// These allow consumers to use intercom.Result, intercom.Iter[T], etc. without
// importing the internal package directly.

// Result types
type Result = api.Result
type ErrorResult = api.ErrorResult
type ErrorDetail = api.ErrorDetail
type Empty = api.Empty

// Error types
type ErrorResponse = api.ErrorResponse
type ErrorCode = api.ErrorCode
type RateLimitInfo = api.RateLimitInfo

// Error code constants — re-exported from internal/api for consumer convenience.
const (
	ErrServerError       = api.ErrServerError
	ErrClientError       = api.ErrClientError
	ErrTypeMismatch      = api.ErrTypeMismatch
	ErrParameterNotFound = api.ErrParameterNotFound
	ErrParameterInvalid  = api.ErrParameterInvalid
	ErrActionForbidden   = api.ErrActionForbidden
	ErrConflict          = api.ErrConflict
	ErrAPIPlanRestricted = api.ErrAPIPlanRestricted
	ErrRateLimitExceeded = api.ErrRateLimitExceeded
	ErrUnsupported       = api.ErrUnsupported
	ErrTokenRevoked      = api.ErrTokenRevoked
	ErrTokenBlocked      = api.ErrTokenBlocked
	ErrTokenNotFound     = api.ErrTokenNotFound
	ErrTokenUnauthorized = api.ErrTokenUnauthorized
	ErrTokenExpired      = api.ErrTokenExpired
	ErrMissingAuth       = api.ErrMissingAuth
	ErrRetryAfter        = api.ErrRetryAfter
	ErrJobClosed         = api.ErrJobClosed
	ErrNotRestorable     = api.ErrNotRestorable
	ErrTeamNotFound      = api.ErrTeamNotFound
	ErrTeamUnavailable   = api.ErrTeamUnavailable
	ErrAdminNotFound     = api.ErrAdminNotFound
)

// Shared domain types
type TagRef = api.TagRef
type AdminRef = api.AdminRef
type ContactRef = api.ContactRef
type ContactRefList = api.ContactRefList
type NoteAuthor = api.NoteAuthor
type Note = api.Note
type NoteListResult = api.NoteListResult
type SegmentRef = api.SegmentRef
type SegmentListResult = api.SegmentListResult
type ArticleContent = api.ArticleContent
type ArticleTranslatedContent = api.ArticleTranslatedContent
type Author = api.Author
type LinkedObjectList = api.LinkedObjectList
type Part = api.Part
type Deleted = api.Deleted
type TagRefList = api.TagRefList

// Tag sub-package aliases re-export tags types so consumers can use
// intercom.Tag, intercom.TagList, etc. without importing the tags package.
type Tag = tags.Tag
type TagList = tags.List
type CreateOrUpdateTagRequest = tags.CreateOrUpdateRequest
type TagCompanyItem = tags.TagCompanyItem
type TagCompanyRequest = tags.TagCompanyRequest
type UntagCompanyItem = tags.UntagCompanyItem
type UntagCompanyRequest = tags.UntagCompanyRequest

// Contacts sub-package aliases
type Contact = contacts.Contact
type ContactLocation = contacts.Location
type ContactListRef = contacts.ListRef
type ContactDeleted = contacts.Deleted
type ContactArchived = contacts.Archived
type ContactUnarchived = contacts.Unarchived
type ContactBlocked = contacts.Blocked
type CreateContactRequest = contacts.CreateRequest
type UpdateContactRequest = contacts.UpdateRequest
type MergeContactsRequest = contacts.MergeRequest
type CompanyRef = contacts.CompanyRef
type CompanyListResult = contacts.CompanyListResult
type CreateNoteRequest = contacts.CreateNoteRequest
type SubscriptionListResult = contacts.SubscriptionListResult
type AddSubscriptionRequest = contacts.AddSubscriptionRequest
type SocialProfileList = contacts.SocialProfileList
type SocialProfile = contacts.SocialProfile
type ContactsService = contacts.Service
type VisitorsService = contacts.VisitorsService
type Visitor = contacts.Visitor
type VisitorAvatar = contacts.VisitorAvatar
type VisitorCompanies = contacts.VisitorCompanies
type VisitorLocation = contacts.VisitorLocation
type VisitorTags = contacts.VisitorTags
type VisitorSegments = contacts.VisitorSegments
type UpdateVisitorRequest = contacts.UpdateVisitorRequest
type ConvertVisitorRequest = contacts.ConvertVisitorRequest
type ConvertVisitorIdentifier = contacts.ConvertVisitorIdentifier
type ConvertVisitorUser = contacts.ConvertVisitorUser
type VisitorSocialProfiles = contacts.VisitorSocialProfiles

// Contacts Parse function aliases
var (
	ParseContactGetResult                = contacts.ParseGetResult
	ParseContactListResult               = contacts.ParseListResult
	ParseContactCreateResult             = contacts.ParseCreateResult
	ParseContactUpdateResult             = contacts.ParseUpdateResult
	ParseContactDeleteResult             = contacts.ParseDeleteResult
	ParseContactSearchResult             = contacts.ParseSearchResult
	ParseContactMergeResult              = contacts.ParseMergeResult
	ParseContactArchiveResult            = contacts.ParseArchiveResult
	ParseContactUnarchiveResult          = contacts.ParseUnarchiveResult
	ParseContactBlockResult              = contacts.ParseBlockResult
	ParseContactFindByExternalIDResult   = contacts.ParseFindByExternalIDResult
	ParseContactListCompaniesResult      = contacts.ParseListCompaniesResult
	ParseContactAddCompanyResult         = contacts.ParseAddCompanyResult
	ParseContactRemoveCompanyResult      = contacts.ParseRemoveCompanyResult
	ParseContactListNotesResult          = contacts.ParseListNotesResult
	ParseContactCreateNoteResult         = contacts.ParseCreateNoteResult
	ParseContactListSegmentsResult       = contacts.ParseListSegmentsResult
	ParseContactListSubscriptionsResult  = contacts.ParseListSubscriptionsResult
	ParseContactAddSubscriptionResult    = contacts.ParseAddSubscriptionResult
	ParseContactRemoveSubscriptionResult = contacts.ParseRemoveSubscriptionResult
	ParseContactAddTagResult             = contacts.ParseAddTagResult
	ParseContactRemoveTagResult          = contacts.ParseRemoveTagResult
	ParseContactListTagsResult           = contacts.ParseListTagsResult
	ParseVisitorGetResult                = contacts.ParseVisitorGetResult
	ParseVisitorUpdateResult             = contacts.ParseVisitorUpdateResult
	ParseVisitorConvertResult            = contacts.ParseVisitorConvertResult
)

// Messaging sub-package aliases re-export messaging types so consumers can use
// intercom.Message, intercom.SubscriptionType, etc. without importing the messaging package.
type SubscriptionType = api.SubscriptionType
type Translation = api.Translation
type SubscriptionTypeList = messaging.SubscriptionTypeList
type Message = messaging.Message
type MessageSender = messaging.MessageSender
type MessageRecipient = messaging.MessageRecipient
type CreateMessageRequest = messaging.CreateMessageRequest
type EmailSetting = messaging.EmailSetting
type EmailSettingList = messaging.EmailSettingList

// Response wraps a Result returned by the Do method.
type Response = api.Response

// Pagination types
type ListOptions = api.ListOptions
type ScrollOptions = api.ScrollOptions
type CursorPages = api.CursorPages
type StartingAfterPage = api.StartingAfterPage
type PagedResult[T any] = api.PagedResult[T]
type PageFetcher[T any] = api.PageFetcher[T]
type Iter[T any] = api.Iter[T]

// Search types
type Filter = api.Filter
type Operator = api.Operator
type SearchRequest = api.SearchRequest
type SearchPagination = api.SearchPagination

// Operator constants — re-exported from internal/api for consumer convenience.
const (
	OpEquals      = api.OpEquals
	OpNotEquals   = api.OpNotEquals
	OpGreaterThan = api.OpGreaterThan
	OpLessThan    = api.OpLessThan
	OpContains    = api.OpContains
	OpNotContains = api.OpNotContains
	OpIn          = api.OpIn
	OpNotIn       = api.OpNotIn
	OpStarts      = api.OpStarts
	OpEnds        = api.OpEnds
	OpAND         = api.OpAND
	OpOR          = api.OpOR
)

// Logger types
type Logger = api.Logger

// Decode unmarshals the JSON body of a Result into a value of type T.
func Decode[T any](r *Result) (*T, error) {
	return api.Decode[T](r)
}

// NewIter creates a new paginated iterator.
func NewIter[T any](ctx context.Context, opts *ListOptions, fetcher PageFetcher[T]) *Iter[T] {
	return api.NewIter[T](ctx, opts, fetcher)
}

// IsNotFound returns true if the error is an Intercom 404 response.
func IsNotFound(err error) bool { return api.IsNotFound(err) }

// IsRateLimited returns true if the error is an Intercom 429 response.
func IsRateLimited(err error) bool { return api.IsRateLimited(err) }

// IsUnauthorized returns true if the error is an Intercom 401 response.
func IsUnauthorized(err error) bool { return api.IsUnauthorized(err) }

// IsBadRequest returns true if the error is an Intercom 400 response.
func IsBadRequest(err error) bool { return api.IsBadRequest(err) }

// IsForbidden returns true if the error is an Intercom 403 response.
func IsForbidden(err error) bool { return api.IsForbidden(err) }

// IsConflict returns true if the error is an Intercom 409 response.
func IsConflict(err error) bool { return api.IsConflict(err) }

// IsUnprocessableEntity returns true if the error is an Intercom 422 response.
func IsUnprocessableEntity(err error) bool { return api.IsUnprocessableEntity(err) }

// IsServerError returns true if the error is an Intercom 5xx response.
func IsServerError(err error) bool { return api.IsServerError(err) }

// SingleFilterOf creates a filter that matches a single field.
func SingleFilterOf(field string, operator Operator, value any) *Filter {
	return api.SingleFilterOf(field, operator, value)
}

// And creates a compound filter that requires all sub-filters to match.
func And(filters ...*Filter) *Filter { return api.And(filters...) }

// Or creates a compound filter that requires any sub-filter to match.
func Or(filters ...*Filter) *Filter { return api.Or(filters...) }

// Unexported function wrappers used by service files and intercom.go.
var (
	buildResult     = api.BuildResult
	resultError     = api.ResultError
	addQueryOptions = api.AddQueryOptions
)
