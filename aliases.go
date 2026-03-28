package intercom

import (
	"context"

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

// Shared domain types
type TagRef = api.TagRef
type AdminRef = api.AdminRef
type ContactRef = api.ContactRef
type NoteAuthor = api.NoteAuthor
type Note = api.Note
type NoteListResult = api.NoteListResult

// Tag sub-package aliases (for backward compat while other root services still reference these)
type Tag = tags.Tag
type TagList = tags.List
type CreateOrUpdateTagRequest = tags.CreateOrUpdateRequest
type TagCompanyItem = tags.TagCompanyItem
type TagCompanyRequest = tags.TagCompanyRequest
type UntagCompanyItem = tags.UntagCompanyItem
type UntagCompanyRequest = tags.UntagCompanyRequest

// Messaging sub-package aliases (for root services that reference these types)
type SubscriptionType = messaging.SubscriptionType
type Translation = messaging.Translation
type SubscriptionTypeList = messaging.SubscriptionTypeList
type Message = messaging.Message
type MessageSender = messaging.MessageSender
type MessageRecipient = messaging.MessageRecipient
type CreateMessageRequest = messaging.CreateMessageRequest
type EmailSetting = messaging.EmailSetting
type EmailSettingList = messaging.EmailSettingList

// Response wraps a Result to provide additional API-specific data.
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
type SearchRequest = api.SearchRequest
type SearchPagination = api.SearchPagination

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

// SingleFilterOf creates a filter that matches a single field.
func SingleFilterOf(field, operator string, value any) *Filter {
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
