# Package Restructure: Flat to Sub-Packages

## Overview
- Split 40 flat `.go` files (258 types) from single `package intercom` into 16 domain-scoped sub-packages
- Solves namespace pollution: users currently see all types in autocomplete/godoc
- Uses `internal/api` package with `Caller` interface to avoid circular imports
- Consumer API changes from `client.Contacts.Get()` to `client.Contacts().Get()`
- Core types (`Result`, `Iter`, `Filter`, `ErrorResponse`) stay in root as type aliases

Full design spec: `docs/superpowers/specs/2026-03-27-package-restructure-design.md`

## Context (from discovery)
- Files/components involved: all 40 `.go` files, 41 test files
- Key pattern: 3-layer method pattern (Raw → Parse → Method), `service` base type with pointer casting
- Dependencies: stdlib + `google/go-querystring` only
- Shared cross-package types: `TagRef` (used by contacts, conversations, tickets, companies)
- Greenfield — no backwards compatibility needed

## Development Approach
- **Testing approach**: Regular (code first, migrate existing tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Since this is a restructure of existing tested code, "tests" primarily means migrating and adapting existing test files

## Testing Strategy
- **Unit tests**: migrate existing `_test.go` files alongside their service code
- Adapt test imports to use `internal/api` types and sub-package types
- `testutil_test.go` helpers: evaluate whether to create `internal/testutil/` or duplicate per package
- Each sub-package must have its tests passing independently

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope
- Keep plan in sync with actual work done

## Implementation Steps

### Task 1: Create `internal/api` — Caller interface and Response type
- [x] create `internal/api/api.go` with `Caller` interface (`NewRequest`, `DoRaw`, `Do` methods)
- [x] define `Response` struct wrapping `*Result` in `api.go`
- [x] verify package compiles: `go build ./internal/api/...`

### Task 2: Move Result types to `internal/api`
- [x] create `internal/api/result.go` — move `Result`, `ErrorResult`, `ErrorDetail`, `Empty`, `Decode[T]`, `buildResult` from root `result.go`
- [x] export all types and functions needed by sub-packages
- [x] write tests for `Decode[T]` and `buildResult` in `internal/api/result_test.go`
- [x] run tests: `go test ./internal/api/...`

### Task 3: Move error types to `internal/api`
- [x] create `internal/api/errors.go` — move `ErrorResponse`, `IsNotFound`, `IsRateLimited`, `IsUnauthorized`, `hasStatusCode` from root `errors.go`
- [x] export `ResultError` (currently unexported `resultError`) for sub-package use
- [x] migrate error tests to `internal/api/errors_test.go`
- [x] run tests: `go test ./internal/api/...`

### Task 4: Move pagination types to `internal/api`
- [x] create `internal/api/pagination.go` — move `PagedResult[T]`, `Iter[T]`, `ListOptions`, `ScrollOptions`, `CursorPages`, `StartingAfterPage`, `NewIter`, `PageFetcher[T]` from root `pagination.go`
- [x] migrate pagination tests to `internal/api/pagination_test.go`
- [x] run tests: `go test ./internal/api/...`

### Task 5: Move search and query types to `internal/api`
- [x] create `internal/api/search.go` — move `Filter`, `SearchRequest`, `SearchPagination`, `SingleFilterOf`, `And`, `Or` from root `search.go`
- [x] create `internal/api/query.go` — move `addQueryOptions` as exported `AddQueryOptions` (wraps `google/go-querystring`)
- [x] create `internal/api/logger.go` — move `Logger` interface from root `logger.go`
- [x] run tests: `go test ./internal/api/...`

### Task 6: Create root type aliases and refactor Client
- [x] create `aliases.go` in root — type aliases for all `internal/api` public types (`Result`, `ErrorResponse`, `PagedResult`, `Iter`, `Filter`, `ListOptions`, etc.)
- [x] refactor `intercom.go` — import `internal/api`, ensure `Client` implements `api.Caller`
- [x] remove old root files (`result.go`, `errors.go`, `pagination.go`, `search.go`, `logger.go`) — their code now lives in `internal/api`
- [x] update `doc.go` if needed
- [x] verify all existing tests still pass: `go test ./...`

### Task 7: Create shared types in `internal/api`
- [x] create `internal/api/types.go` — extract `TagRef` and any other types used by 3+ services
- [x] identify all cross-package types by grepping for types referenced in multiple service files
- [x] run tests: `go test ./...`

### Task 8: Proof of concept — `tags` sub-package
- [x] create `tags/tags.go` — `Service` struct, `NewService(api.Caller)`, move all tag types and methods from root `tags.go`
- [x] rename types: `TagsService` → `Service`, `CreateTagRequest` → `CreateRequest`, etc.
- [x] add `Tags() *tags.Service` accessor method to `Client` in `intercom.go`
- [x] wire `tags.NewService(c)` in `NewClient`
- [x] migrate `tags_test.go` to `tags/tags_test.go`, adapt imports
- [x] remove root `tags.go`
- [x] run tests: `go test ./...`

### Task 9: Migrate `segments` sub-package
- [ ] create `segments/segments.go` — move types and methods from root `segments.go`
- [ ] add `Segments()` accessor to `Client`
- [ ] migrate tests to `segments/segments_test.go`
- [ ] remove root `segments.go`
- [ ] run tests: `go test ./...`

### Task 10: Migrate `export` sub-package
- [ ] create `export/reporting.go` — move from root `export_reporting.go`
- [ ] create `export/data.go` — move from root `data_export.go`
- [ ] add `ExportReporting()` and `DataExport()` accessors to `Client`
- [ ] migrate tests to `export/`
- [ ] remove root `export_reporting.go`, `data_export.go`
- [ ] run tests: `go test ./...`

### Task 11: Migrate `news` sub-package
- [ ] create `news/news.go` — move types and methods from root `news.go`
- [ ] add `News()` accessor to `Client`
- [ ] migrate tests to `news/news_test.go`
- [ ] remove root `news.go`
- [ ] run tests: `go test ./...`

### Task 12: Migrate `messaging` sub-package
- [ ] create `messaging/messages.go` — move from root `messages.go`
- [ ] create `messaging/emails.go` — move from root `emails.go`
- [ ] create `messaging/subscriptions.go` — move from root `subscription_types.go`
- [ ] add `Messages()`, `Emails()`, `SubscriptionTypes()` accessors to `Client`
- [ ] migrate tests to `messaging/`
- [ ] remove root `messages.go`, `emails.go`, `subscription_types.go`
- [ ] run tests: `go test ./...`

### Task 13: Migrate `settings` sub-package
- [ ] create `settings/brands.go` — move from root `brands.go`
- [ ] create `settings/ipallowlist.go` — move from root `ip_allowlist.go`
- [ ] create `settings/channelevents.go` — move from root `custom_channel_events.go`
- [ ] create `settings/jobs.go` — move from root `jobs.go`
- [ ] create `settings/notes.go` — move from root `notes.go`
- [ ] add `Brands()`, `IPAllowlist()`, `CustomChannelEvents()`, `Jobs()`, `Notes()` accessors to `Client`
- [ ] migrate tests to `settings/`
- [ ] remove root files
- [ ] run tests: `go test ./...`

### Task 14: Migrate `admins` sub-package
- [ ] create `admins/admins.go` — move from root `admins.go`
- [ ] create `admins/teams.go` — move from root `teams.go`
- [ ] create `admins/away.go` — move from root `away_status_reasons.go`
- [ ] add `Admins()`, `Teams()`, `AwayStatusReasons()` accessors to `Client`
- [ ] migrate tests to `admins/`
- [ ] remove root `admins.go`, `teams.go`, `away_status_reasons.go`
- [ ] run tests: `go test ./...`

### Task 15: Migrate `data` sub-package
- [ ] create `data/events.go` — move from root `data_events.go`
- [ ] create `data/attributes.go` — move from root `data_attributes.go`
- [ ] create `data/objects.go` — move from root `custom_objects.go`
- [ ] add `DataEvents()`, `DataAttributes()`, `CustomObjects()` accessors to `Client`
- [ ] migrate tests to `data/`
- [ ] remove root `data_events.go`, `data_attributes.go`, `custom_objects.go`
- [ ] run tests: `go test ./...`

### Task 16: Migrate `calls` sub-package
- [ ] create `calls/calls.go` — move from root `calls.go`
- [ ] create `calls/redirects.go` — move from root `phone_call_redirects.go`
- [ ] add `Calls()`, `PhoneCallRedirects()` accessors to `Client`
- [ ] migrate tests to `calls/`
- [ ] remove root `calls.go`, `phone_call_redirects.go`
- [ ] run tests: `go test ./...`

### Task 17: Migrate `helpcenter` sub-package
- [ ] create `helpcenter/helpcenter.go` — move from root `help_center.go`
- [ ] add `HelpCenter()` accessor to `Client`
- [ ] migrate tests to `helpcenter/helpcenter_test.go`
- [ ] remove root `help_center.go`
- [ ] run tests: `go test ./...`

### Task 18: Migrate `articles` sub-package
- [ ] create `articles/articles.go` — move from root `articles.go`
- [ ] create `articles/internal.go` — move from root `internal_articles.go`
- [ ] add `Articles()`, `InternalArticles()` accessors to `Client`
- [ ] migrate tests to `articles/`
- [ ] remove root `articles.go`, `internal_articles.go`
- [ ] run tests: `go test ./...`

### Task 19: Migrate `companies` sub-package
- [ ] create `companies/companies.go` — move from root `companies.go` (includes scroll pagination)
- [ ] add `Companies()` accessor to `Client`
- [ ] migrate tests to `companies/companies_test.go`
- [ ] remove root `companies.go`
- [ ] run tests: `go test ./...`

### Task 20: Migrate `tickets` sub-package
- [ ] create `tickets/tickets.go` — move from root `tickets.go`
- [ ] create `tickets/types.go` — move from root `ticket_types.go`
- [ ] create `tickets/states.go` — move from root `ticket_states.go`
- [ ] add `Tickets()`, `TicketTypes()`, `TicketStates()` accessors to `Client`
- [ ] migrate tests to `tickets/`
- [ ] remove root `tickets.go`, `ticket_types.go`, `ticket_states.go`
- [ ] run tests: `go test ./...`

### Task 21: Migrate `conversations` sub-package
- [ ] create `conversations/conversations.go` — move from root `conversations.go`
- [ ] add `Conversations()` accessor to `Client`
- [ ] migrate tests to `conversations/conversations_test.go`
- [ ] remove root `conversations.go`
- [ ] run tests: `go test ./...`

### Task 22: Migrate `contacts` sub-package
- [ ] create `contacts/contacts.go` — move from root `contacts.go` (largest service, ~909 lines)
- [ ] create `contacts/visitors.go` — move from root `visitors.go`
- [ ] add `Contacts()`, `Visitors()` accessors to `Client`
- [ ] migrate tests to `contacts/`
- [ ] remove root `contacts.go`, `visitors.go`
- [ ] run tests: `go test ./...`

### Task 23: Migrate `ai` sub-package
- [ ] create `ai/content.go` — move from root `ai_content.go`
- [ ] create `ai/voice.go` — move from root `fin_voice.go`
- [ ] add `AIContent()`, `FinVoice()` accessors to `Client`
- [ ] migrate tests to `ai/`
- [ ] remove root `ai_content.go`, `fin_voice.go`
- [ ] run tests: `go test ./...`

### Task 24: Clean up root package
- [ ] verify no old service files remain in root (only `intercom.go`, `aliases.go`, `doc.go`)
- [ ] remove `testutil_test.go` from root if all tests moved (or keep if shared)
- [ ] clean up any unused imports in `intercom.go`
- [ ] run tests: `go test ./...`

### Task 25: Verify acceptance criteria
- [ ] verify all 33 services accessible via accessor methods
- [ ] verify all types are in their correct sub-packages
- [ ] verify `internal/api` types are properly re-exported as aliases in root
- [ ] verify no cross-service imports between sub-packages
- [ ] run full test suite: `go test ./...`
- [ ] run race detector: `go test -race ./...`
- [ ] run linter: `go vet ./...`
- [ ] run format check: `gofmt -l .`
- [ ] verify test coverage: `go test -coverprofile=coverage.out ./...`

### Task 26: [Final] Update documentation
- [ ] update `README.md` with new import paths and usage examples
- [ ] update `CLAUDE.md` with new architecture, conventions, and "Adding a New Service" instructions
- [ ] update `doc.go` package documentation

## Technical Details

### Caller Interface
```go
type Caller interface {
    NewRequest(method, urlStr string, body any) (*http.Request, error)
    DoRaw(ctx context.Context, req *http.Request) (*Result, error)
    Do(ctx context.Context, req *http.Request, v any) (*Response, error)
}
```

### Sub-Package Service Pattern
```go
type Service struct {
    client api.Caller
}

func NewService(c api.Caller) *Service {
    return &Service{client: c}
}
```

### Type Renaming Convention
| Old (flat) | New (sub-package) |
|---|---|
| `ContactsService` | `contacts.Service` |
| `CreateContactRequest` | `contacts.CreateRequest` |
| `Contact` | `contacts.Contact` |
| `ListContactsOptions` | `contacts.ListOptions` |

### Package Grouping Reference
| Package | Services | Files |
|---|---|---|
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

## Post-Completion

**Manual verification:**
- Check godoc output for clean namespace per sub-package
- Verify IDE autocomplete shows only relevant types per import
- Test consumer experience with a sample program importing specific sub-packages
