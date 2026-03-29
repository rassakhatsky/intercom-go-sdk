# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go SDK for the Intercom API v2.15. 34 services organized into 16 domain-scoped sub-packages, minimal dependencies (stdlib + `google/go-querystring`).

## Commands

```bash
# Run all tests
go test ./...

# Run a single test (include sub-package path)
go test -run TestService_Get ./contacts/...

# Run tests with coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Vet
go vet ./...

# Format check
gofmt -l .

# Or use Make targets
make test        # go test -cover ./...
make test-race   # go test -race -cover ./...
make fix         # go vet + gofmt -l
```

## Architecture

**Sub-package design** — services live in 16 domain-scoped sub-packages (`contacts/`, `tickets/`, `tags/`, etc.). Core types and the Client live in the root `intercom` package.

**Internal API** (`internal/api/`):
- `Caller` interface: `NewRequest`, `DoRaw`, `DoRawNoRedirect`, `Do`, `DoDownload` — abstracts HTTP operations for sub-packages
- Shared types: `Result`, `ErrorResult`, `ErrorResponse`, `PagedResult[T]`, `Iter[T]`, `Filter`, `ListOptions`, etc.
- Cross-service domain types: `TagRef`, `AdminRef`, `ContactRef`, `Note`, `SegmentRef`, etc.
- Sub-packages import `internal/api` for the `Caller` interface and shared types; they never import each other

**Root package** (`intercom`):
- `Client` (`intercom.go`): holds auth token, HTTP client, logger, and all service instances; implements `api.Caller`
- `aliases.go`: type aliases re-exporting `internal/api` types so consumers use `intercom.Result`, `intercom.Iter[T]`, etc. without importing `internal/api`. Also re-exports error code constants (`ErrServerError`, `ErrParameterInvalid`, etc.) and status helper functions (`IsNotFound`, `IsBadRequest`, `IsServerError`, etc.)
- Accessor methods on `Client` (e.g., `Contacts()`, `Tags()`) return sub-package service pointers
- `NewRequest` / `Do` / `DoRaw` handle JSON marshaling, auth headers (`Bearer` token), and `Intercom-Version: 2.15`

**Sub-package services** (e.g., `contacts/contacts.go`, `tickets/tickets.go`):
- Each service struct holds an `api.Caller` field, constructed via `NewService(api.Caller)`
- Request/response structs with JSON tags; types are scoped to the package (e.g., `contacts.CreateRequest` instead of `CreateContactRequest`)
- Methods follow the 3-layer pattern (Raw → Parse → Method)
- Every method takes `context.Context` as its first parameter
- URL path parameters use `url.PathEscape`
- **3-layer pattern**: Raw methods (`XxxRaw`) do `NewRequest` → `DoRaw` → return `(*api.Result, error)`; Parse functions (`ParseXxxResult`) decode `Result.Body` into typed structs; regular methods combine Raw + error check + Parse

**Pagination** (`internal/api/pagination.go`):
- `PagedResult[T]` — generic paginated response
- `Iter[T]` — lazy auto-pagination iterator using cursor-based pagination (`starting_after`)
- `Iter[T].Collect()` — drains all pages into a `[]T` slice; returns items collected so far on error
- `Iter[T].ForEach(fn)` — callback-based iteration; stops on first error from `fn` or page fetch
- `ListAll` methods return `*Iter[T]`, created via `NewIter` with a `PageFetcher[T]` function
- Companies use separate scroll-based pagination (`ScrollOptions`)

**Search** (`internal/api/search.go`):
- `Filter` struct serves dual purpose: single field filter or compound (AND/OR) filter
- Constructed via `SingleFilterOf`, `And`, `Or` helpers

**Error handling** (`internal/api/errors.go`):
- `ErrorResponse` wraps API errors with `StatusCode int` and `Headers http.Header` fields (no synthetic `*http.Response`); `ResultError` converts `Result.Error` into `*ErrorResponse` (nil-safe: returns nil for nil input)
- Status predicates: `IsNotFound` (404), `IsRateLimited` (429), `IsUnauthorized` (401), `IsBadRequest` (400), `IsForbidden` (403), `IsConflict` (409), `IsUnprocessableEntity` (422), `IsServerError` (5xx)
- `ErrorCode` typed constants (e.g., `ErrParameterInvalid`, `ErrRateLimitExceeded`, `ErrTokenRevoked`); `ErrorResponse.HasErrorCode(code)` matches against the `Errors` slice
- `RateLimitInfo` (Limit, Remaining, Reset, RetryAfter) — auto-parsed from headers on 429 responses, available via `ErrorResponse.RateLimit`

**Raw results** (`internal/api/result.go`):
- `Result` — non-generic struct: `StatusCode int`, `Header http.Header`, `URL string`, `Body []byte`, `Error *ErrorResult`
- `ErrorResult` — typed API error with Code, Message, Errors; implements `Error() string`
- `Decode[T any](r *Result) (*T, error)` — generic helper that JSON-unmarshals `Result.Body` into `*T`; returns zero-value `*T` for any 2xx response with empty body; returns error on empty body for non-2xx responses
- `BuildResult(resp *http.Response, body []byte) *Result` — constructs a `Result` from an HTTP response
- Error contract: `(nil, err)` for transport/IO failures only; `(result, nil)` for any completed HTTP request (including 4xx/5xx)
- API errors (4xx/5xx) populate `Result.Error`; use `ResultError(r)` to convert to `*ErrorResponse` for predicate checks
- Excluded from Raw companions: `Download` methods (streaming body) and `ListAll` methods (iterators)

## Testing Patterns

Each sub-package embeds test helpers (`testCaller`, `setup`, `testMethod`, `testHeader`) inline in its `*_test.go` files. The `ai/` package is the exception, using a dedicated `testutil_test.go`. The root package also has a `testutil_test.go`.

```go
svc, mux, teardown := setup()
defer teardown()
mux.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) { ... })
```

Helper functions `testMethod` and `testHeader` assert request properties. Tests write raw JSON strings directly to the response writer. No external test frameworks — stdlib `testing` only.

## Conventions

- Service type naming: `Service` (or `XxxService` for multi-service packages like `contacts.VisitorsService`)
- Request structs: `CreateRequest`, `UpdateRequest` (scoped to the sub-package)
- All JSON field tags use `omitempty` except for required/identity fields (`type`, `id`)
- Query param structs use `url:"name,omitempty"` tags (encoded via `google/go-querystring`)
- Timestamps are `int64` (Unix epoch); optional timestamps are `*int64`
- 3-layer pattern per method: see Architecture above
- Every public method (regular + Raw) includes a `// See:` comment linking to the Intercom API reference. Format: `// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/{tag}/{operationId}`. Skip: Parse functions, ListAll methods, model structs, service type declarations
- Sub-packages never import each other; shared types live in `internal/api`

## Package Grouping

| Package | Services | Files |
|---------|----------|-------|
| `contacts` | contacts, visitors | contacts.go, visitors.go |
| `conversations` | conversations | conversations.go |
| `companies` | companies | companies.go |
| `tickets` | tickets, ticket_types, ticket_states | tickets.go, types.go, states.go |
| `articles` | articles, internal_articles | articles.go, internal.go |
| `helpcenter` | help_center | helpcenter.go |
| `news` | news | news.go |
| `ai` | ai_content, fin_voice | content.go, voice.go |
| `calls` | calls, phone_call_redirects | calls.go, redirects.go |
| `tags` | tags | tags.go |
| `admins` | admins, teams, away_status_reasons | admins.go, teams.go, away.go |
| `segments` | segments | segments.go |
| `data` | data_events, data_attributes, custom_objects | events.go, attributes.go, objects.go |
| `messaging` | messages, emails, subscription_types | messages.go, emails.go, subscriptions.go |
| `export` | data_export, export_reporting | data.go, reporting.go |
| `settings` | brands, ip_allowlist, custom_channel_events, jobs, notes | brands.go, ipallowlist.go, channelevents.go, jobs.go, notes.go |
| `workflows` | workflows | workflows.go |

## Adding a New Service

1. **Create the service file** in the appropriate sub-package (e.g., `contacts/newservice.go`):
   - Define `type NewService struct { client api.Caller }`
   - Add `func NewNewService(c api.Caller) *NewService { return &NewService{client: c} }`
   - Define request/response structs with JSON tags
   - Implement methods following the 3-layer pattern (Raw → Parse → Method)

2. **Register in `intercom.go`**:
   - Add a field to `Client`: `newService *contacts.NewService`
   - Initialize in `initialize()`: `c.newService = contacts.NewNewService(c)`
   - Add accessor method: `func (c *Client) NewService() *contacts.NewService { return c.newService }`

3. **Create tests** in the sub-package (e.g., `contacts/newservice_test.go`):
   - Use the sub-package's `setup()` helper, `testMethod`/`testHeader`, raw JSON responses

4. **Add `// See:` doc URL comments** to every public method (regular + Raw) linking to the Intercom API reference

5. **If adding a new sub-package**:
   - Create the directory and add a `testutil_test.go` with `setup()`, `testMethod`, `testHeader` helpers
   - Import `internal/api` for the `Caller` interface and shared types
   - If the new service needs types used by other sub-packages, add them to `internal/api/types.go`
   - Add type aliases in `aliases.go` if consumers need them from the root package
