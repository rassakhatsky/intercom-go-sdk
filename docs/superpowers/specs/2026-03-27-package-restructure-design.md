# Package Restructure: Flat to Sub-Packages

## Context

The SDK has 40 `.go` files and 258 types in a single `package intercom`. This causes namespace pollution: users see every type from every service in autocomplete and godoc. The goal is to split services into domain-scoped sub-packages while keeping core types (`Result`, `Iter`, `Filter`, etc.) in the root `intercom` package.

**Constraints:** Greenfield (no consumers), Go 1.21+, single external dependency (`google/go-querystring`).

## Architecture

### Circular Import Avoidance

Sub-packages need `Client.NewRequest()`/`Client.DoRaw()`, but `Client` is in root `intercom`. To avoid a cycle:

```
internal/api    <- Caller interface + shared types (Result, Iter, Filter, etc.)
     ^                    ^
     |                    |
intercom/      ->    contacts/, tickets/, ...
(implements            (consume Caller interface,
 Caller)                import internal/api only)
```

- `internal/api` defines a `Caller` interface and all shared types
- Sub-packages import only `internal/api`, never `intercom`
- Root `intercom` imports sub-packages to wire accessor methods
- Root re-exports `internal/api` types as type aliases (`type Result = api.Result`)

### `internal/api` Package

```go
// internal/api/api.go
package api

type Caller interface {
    NewRequest(method, urlStr string, body any) (*http.Request, error)
    DoRaw(ctx context.Context, req *http.Request) (*Result, error)
    Do(ctx context.Context, req *http.Request, v any) (*Response, error)
}
```

Contains (moved from root):
- `api.go` — `Caller` interface, `Response` wrapper
- `result.go` — `Result`, `ErrorResult`, `Decode[T]`, `buildResult`
- `errors.go` — `ErrorResponse`, `ErrorDetail`, `IsNotFound`, `IsRateLimited`, `IsUnauthorized`, `ResultError`
- `pagination.go` — `PagedResult[T]`, `Iter[T]`, `ListOptions`, `ScrollOptions`, `CursorPages`, `NewIter`
- `search.go` — `Filter`, `SearchRequest`, `SearchPagination`, `SingleFilterOf`, `And`, `Or`
- `query.go` — `AddQueryOptions` (exported for sub-packages)
- `logger.go` — `Logger` interface
- `types.go` — Cross-package types: `TagRef` and any other types used by 3+ sub-packages

### Root `intercom` Package

After restructure, contains:
- `intercom.go` — `Client` struct (private service fields), `NewClient`, `NewRequest`, `DoRaw`, `Do`, accessor methods
- `aliases.go` — Type aliases re-exporting `internal/api` types for public API
- `doc.go` — Package documentation

### Sub-Package Pattern

Each sub-package defines a `Service` struct (not `ContactsService` — the package name provides context):

```go
// contacts/contacts.go
package contacts

import "github.com/rassakhatsky/intercom-go-sdk/internal/api"

type Service struct {
    client api.Caller
}

func NewService(c api.Caller) *Service {
    return &Service{client: c}
}

func (s *Service) Get(ctx context.Context, id string) (*Contact, error) { ... }
func (s *Service) GetRaw(ctx context.Context, id string) (*api.Result, error) { ... }
```

### Client Accessor Methods

```go
// intercom/intercom.go
type Client struct {
    token      string
    baseURL    string
    httpClient *http.Client
    logger     api.Logger

    contacts      *contacts.Service
    conversations *conversations.Service
    // ... 14 more
}

func (c *Client) Contacts() *contacts.Service           { return c.contacts }
func (c *Client) Conversations() *conversations.Service  { return c.conversations }
// ...
```

### Consumer API

```go
import (
    "github.com/rassakhatsky/intercom-go-sdk"
    "github.com/rassakhatsky/intercom-go-sdk/contacts"
)

c := intercom.NewClient("token")
contact, err := c.Contacts().Get(ctx, "123")
// contact is *contacts.Contact

// Core types still at root:
if intercom.IsNotFound(err) { ... }
var iter *intercom.Iter[contacts.Contact]
```

## Package Grouping (16 packages)

| # | Package | Services | ~Lines | Rationale |
|---|---------|----------|--------|-----------|
| 1 | `contacts` | contacts, visitors | 1128 | Visitors are pre-identification contacts |
| 2 | `conversations` | conversations | 744 | Large, cohesive, standalone |
| 3 | `companies` | companies | 414 | Unique scroll pagination |
| 4 | `tickets` | tickets, ticket_types, ticket_states | 820 | Types/states are ticket metadata |
| 5 | `articles` | articles, internal_articles | 628 | External vs internal articles |
| 6 | `helpcenter` | help_center | 294 | Collections + centers |
| 7 | `news` | news | 334 | News items + newsfeeds |
| 8 | `ai` | ai_content, fin_voice | 586 | AI-powered features |
| 9 | `calls` | calls, phone_call_redirects | 310 | Phone/call domain |
| 10 | `tags` | tags | 246 | Standalone, widely referenced |
| 11 | `admins` | admins, teams, away_status_reasons | 394 | Teammates domain |
| 12 | `segments` | segments | 104 | Distinct resource |
| 13 | `data` | data_events, data_attributes, custom_objects | 541 | Data modeling domain |
| 14 | `messaging` | messages, emails, subscription_types | 249 | Outbound messaging |
| 15 | `export` | data_export, export_reporting | 337 | Async export jobs |
| 16 | `settings` | brands, ip_allowlist, custom_channel_events, jobs, notes | 485 | Workspace config + small utilities |

## Directory Layout

```
intercom-go-sdk/
  intercom.go              # Client, accessor methods, NewClient, NewRequest, DoRaw, Do
  aliases.go               # type Result = api.Result, type Filter = api.Filter, etc.
  doc.go                   # package documentation

  internal/api/
    api.go                 # Caller interface, Response
    result.go              # Result, ErrorResult, Decode[T], buildResult
    errors.go              # ErrorResponse, IsNotFound, IsRateLimited, IsUnauthorized
    pagination.go          # PagedResult[T], Iter[T], ListOptions, ScrollOptions, NewIter
    search.go              # Filter, SearchRequest, SingleFilterOf, And, Or
    query.go               # AddQueryOptions (uses google/go-querystring)
    logger.go              # Logger interface
    types.go               # TagRef, other cross-package types

  contacts/                # contacts.go, visitors.go
  conversations/           # conversations.go
  companies/               # companies.go
  tickets/                 # tickets.go, types.go, states.go
  articles/                # articles.go, internal.go
  helpcenter/              # helpcenter.go
  news/                    # news.go
  ai/                      # content.go, voice.go
  calls/                   # calls.go, redirects.go
  tags/                    # tags.go
  admins/                  # admins.go, teams.go, away.go
  segments/                # segments.go
  data/                    # events.go, attributes.go, objects.go
  messaging/               # messages.go, emails.go, subscriptions.go
  export/                  # reporting.go, data.go
  settings/                # brands.go, ipallowlist.go, channelevents.go, jobs.go, notes.go
```

## Shared Types Strategy

- Types used by **3+ sub-packages** move to `internal/api/types.go` (e.g., `TagRef`)
- Types used by **2 packages** — each defines its own lightweight ref struct
- Types used by **1 package** — stay in that package

## Naming Conventions

- Sub-package service type: `Service` (not `ContactsService` — package name provides context)
- Sub-package request types: `CreateRequest`, `UpdateRequest` (not `CreateContactRequest`)
- Sub-package main entity: `Contact`, `Conversation`, etc. (no prefix needed)
- Sub-package list options: `ListOptions` (scoped to package)
- Root re-exports: `type Result = api.Result` (type aliases preserve identity)

## Migration Strategy

### Phase 1: Internal Foundation
1. Create `internal/api/` with `Caller` interface
2. Move shared types: Result, ErrorResponse, PagedResult, Iter, Filter, Logger
3. Move `addQueryOptions` as `AddQueryOptions` (exported)
4. Add `aliases.go` in root with type aliases
5. Verify: `go build ./...`

### Phase 2: Proof of Concept — `tags` package
6. Create `tags/` with `Service`, `Tag`, `CreateRequest`, etc.
7. Add `tags` accessor to `Client`
8. Move tests to `tags/`
9. Verify: `go test ./...`

### Phase 3: Migrate Remaining (dependency order)
10. Independent packages first: `segments`, `export`, `news`, `messaging`, `settings`
11. Then packages with shared type refs: `admins`, `data`, `calls`, `helpcenter`, `articles`
12. Then heavy packages: `companies`, `tickets`, `conversations`, `contacts`, `ai`
13. Each migration: move types + methods, add NewService, add accessor, update tests

### Phase 4: Cleanup
14. Remove old flat service files from root
15. Update `doc.go`, `README.md`, `CLAUDE.md`
16. Final `go vet ./...`, `gofmt -l .`, `go test -race ./...`

## Testing

- Each sub-package gets its own `_test.go` files
- The `setup()` helper moves to `internal/testutil/` (test-only internal package) or is duplicated per package
- Test pattern stays the same: `httptest.Server`, `testMethod`, `testHeader`, raw JSON
- `testutil_test.go` helpers could become an `internal/testutil` package for shared test setup

## Verification

After full migration:
```bash
go build ./...          # all packages compile
go test ./...           # all tests pass
go test -race ./...     # no race conditions
go vet ./...            # no vet issues
gofmt -l .              # no formatting issues
```

Also manually verify:
- `godoc` shows clean namespace per package
- Autocomplete in IDE shows only relevant types per import
