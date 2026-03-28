# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go SDK for the Intercom API v2.15. Single-package (`package intercom`), minimal dependencies (stdlib + `google/go-querystring`), 33 services.

## Commands

```bash
# Run all tests
go test ./...

# Run a single test
go test -run TestContactsService_Get ./...

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

**Single-package design** — all code lives in `package intercom` at the repo root. No sub-packages.

**Client core** (`intercom.go`):
- `Client` holds auth token, HTTP client, logger, and all service fields
- `service` is a base struct (holds `*Client`); every `XxxService` is a type alias for `service` (e.g., `type ContactsService service`)
- All services share the same `Client` via the `common` field pattern — `initialize()` type-casts `&c.common` to each service type
- `NewRequest` / `Do` handle JSON marshaling, auth headers (`Bearer` token), and `Intercom-Version: 2.15`
- `addQueryOptions` uses `google/go-querystring` to encode `url:"name,omitempty"` struct tags into query strings

**Service files** (one per API resource, e.g., `contacts.go`, `tickets.go`):
- Each file defines: the service type, request/response structs with JSON tags, and methods following the 3-layer pattern (Raw → Parse → Method)
- Every method takes `context.Context` as its first parameter
- URL path parameters use `url.PathEscape`
- **3-layer pattern**: Raw methods (`XxxRaw`) do `NewRequest` → `DoRaw` → return `(*Result, error)`; Parse functions (`ParseXxxResult`) decode `Result.Body` into typed structs; regular methods combine Raw + error check + Parse

**Pagination** (`pagination.go`):
- `PagedResult[T]` — generic paginated response
- `Iter[T]` — lazy auto-pagination iterator using cursor-based pagination (`starting_after`)
- `Iter[T].Collect()` — drains all pages into a `[]T` slice; returns items collected so far on error
- `Iter[T].ForEach(fn)` — callback-based iteration; stops on first error from `fn` or page fetch
- `ListAll` methods return `*Iter[T]`, created via `NewIter` with a `PageFetcher[T]` function
- Companies use separate scroll-based pagination (`ScrollOptions`)

**Search** (`search.go`):
- `Filter` struct serves dual purpose: single field filter or compound (AND/OR) filter
- Constructed via `SingleFilterOf`, `And`, `Or` helpers

**Error handling** (`errors.go`):
- `ErrorResponse` wraps API errors; `resultError` converts `Result.Error` into `*ErrorResponse`
- Helper predicates: `IsNotFound`, `IsRateLimited`, `IsUnauthorized`

**Raw results** (`result.go`):
- `Result` — non-generic struct: `StatusCode int`, `Header http.Header`, `URL string`, `Body []byte`, `Error *ErrorResult`
- `ErrorResult` — typed API error with Code, Message, Errors; implements `Error() string`
- `Decode[T any](r *Result) (*T, error)` — generic helper that JSON-unmarshals `Result.Body` into `*T`
- `buildResult(resp *http.Response, body []byte) *Result` — constructs a `Result` from an HTTP response
- `DoRaw` is a method on `Client`: `func (c *Client) DoRaw(ctx context.Context, req *http.Request) (*Result, error)`
- Error contract: `(nil, err)` for transport/IO failures only; `(result, nil)` for any completed HTTP request (including 4xx/5xx)
- API errors (4xx/5xx) populate `Result.Error`; use `resultError(r)` to convert to `*ErrorResponse` for predicate checks
- Excluded from Raw companions: `Download` methods (streaming body) and `ListAll` methods (iterators)

## Testing Patterns

Tests use `httptest.Server` via the `setup()` helper in `testutil_test.go`:
```go
client, mux, teardown := setup()
defer teardown()
mux.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) { ... })
```

Helper functions `testMethod` and `testHeader` assert request properties. Tests write raw JSON strings directly to the response writer. No external test frameworks — stdlib `testing` only.

## Conventions

- Service type naming: `XxxService` as a type alias of `service` (not embedding)
- Request structs: `CreateXxxRequest`, `UpdateXxxRequest`
- All JSON field tags use `omitempty` except for required/identity fields (`type`, `id`)
- Query param structs use `url:"name,omitempty"` tags (encoded via `google/go-querystring`)
- Timestamps are `int64` (Unix epoch); optional timestamps are `*int64`
- 3-layer pattern per method: see Architecture above
- Every public method (regular + Raw) includes a `// See:` comment linking to the Intercom API reference. Format: `// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/{tag}/{operationId}`. Skip: Parse functions, ListAll methods, model structs, service type declarations

## Adding a New Service

1. Create `xxx.go`: define `type XxxService service`, request/response structs, and methods following the 3-layer pattern
2. Register in `intercom.go`: add `Xxx *XxxService` field to `Client`, cast `&c.common` in `initialize()`
3. Create `xxx_test.go`: use `setup()` helper, `testMethod`/`testHeader`, raw JSON responses
4. Add `// See:` doc URL comments to every public method (regular + Raw) linking to the Intercom API reference
