# DoRaw 3-Layer Refactor

## Overview
- Refactor `DoRaw` and `Result[T]` into a 3-layer architecture (Raw → Parse → Method) across all 33 service files
- Fixes inconsistent error returns: transport errors return `(nil, err)` while JSON unmarshal failures return `(partial result, err)`
- Enables raw body inspection before decoding — critical because Intercom docs are frequently wrong (~20+ times in 2 years)

## Context (from discovery)
- **Files/components involved**: `result.go`, `intercom.go`, `errors.go`, `calls.go`, 33 service files, 33 test files, `CLAUDE.md`
- **Current state**: All 34 service files already have `XxxRaw` companions using generic `DoRaw[T]`, but the core types and Parse layer don't exist yet
- **What needs to change**:
  - `Result[T]` (generic) → `Result` (non-generic, raw bytes)
  - `DoRaw[T]` (package-level generic) → `Client.DoRaw` (method, non-generic)
  - Add `Decode[T]` and 165 `ParseXxxResult` functions
  - Add `resultError`, remove `checkResponse`
  - Rewrite regular methods as thin wrappers: Raw + error check + Parse
- **Logger refactor**: handled separately, commit before starting this plan

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility for `Do`, `Response`, error predicates

## Testing Strategy
- **Unit tests**: required for every task
- Existing tests cover behavior — update signatures and assertions to match new types
- Raw tests: replace `result.Data.Field` with `Parse` or `Decode[T]` + access decoded value
- Error tests: verify `resultError` produces `*ErrorResponse` compatible with `IsNotFound`/`IsRateLimited`/`IsUnauthorized`

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Refactor `result.go` — non-generic `Result`, `Decode[T]`, `buildResult`
- [x] Replace `Result[T]` with non-generic `Result` struct: `StatusCode`, `Header`, `URL`, `Body []byte`, `Error *ErrorResult`
- [x] Remove `HTTPResult` struct (fields merged into `Result`)
- [x] Add `Decode[T any](r *Result) (*T, error)` generic helper for JSON unmarshaling
- [x] Replace `buildEmptyResult` with `buildResult(resp *http.Response, body []byte) *Result`
- [x] Remove `Empty` struct if no longer needed, or keep for type clarity
- [x] Write tests for `Decode[T]` — success, empty body, malformed JSON
- [x] Write tests for `buildResult` — status codes, headers, body capture
- [x] Run `go test ./...` — must pass before next task

### Task 2: Refactor `intercom.go` — `DoRaw` as Client method, refactor `Do`, update `Response`
- [x] Convert `DoRaw` from package-level generic function to `func (c *Client) DoRaw(ctx context.Context, req *http.Request) (*Result, error)`
- [x] Implement consistent error contract: `(nil, err)` for transport only, `(result, nil)` for any completed HTTP request
- [x] Populate `Result.Error` for 4xx/5xx responses inside `DoRaw`
- [x] Change `Response` struct to wrap `*Result` instead of `*http.Response`
- [x] Refactor `Do` to use `DoRaw` internally, keeping it public as escape hatch
- [x] Update `Do` to return `(*Response, error)` with `Response.Result` populated
- [x] Write tests for `DoRaw` — transport error, 200 response, 4xx response, 5xx response
- [x] Write tests for refactored `Do` — verify it delegates to `DoRaw` correctly
- [x] Update existing `intercom_test.go` tests for new signatures
- [x] Run `go test ./...` — must pass before next task

### Task 3: Refactor `errors.go` — add `resultError`, remove `checkResponse`
- [x] Add `resultError(r *Result) error` — converts `Result.Error` to `*ErrorResponse`
- [x] Ensure `IsNotFound`, `IsRateLimited`, `IsUnauthorized` still work with errors from `resultError`
- [x] Remove `checkResponse` function (error detection now lives in `DoRaw`)
- [x] Remove any remaining callers of `checkResponse`
- [x] Write tests for `resultError` — 404, 429, 401, generic 5xx
- [x] Write tests verifying error predicate compatibility
- [x] Run `go test ./...` — must pass before next task

### Task 4: Refactor `calls.go` — special-case methods using `buildResult`
- [x] Update `GetRecordingURLRaw` and `GetTranscriptRaw` to use new `buildResult` signature
- [x] Update remaining Raw methods to use `s.client.DoRaw` and return `(*Result, error)`
- [x] Add Parse functions: `ParseCallGetResult`, `ParseCallListResult`, `ParseCallSearchResult`, `ParseCallGetRecordingURLResult`, `ParseCallGetTranscriptResult`
- [x] Rewrite regular methods as Raw + error check + Parse wrappers
- [x] Update `calls_test.go` — Raw test assertions, Parse function tests
- [x] Run `go test ./...` — must pass before next task

### Task 5: Refactor `admins.go` — pilot service file (5 Raw methods)
- [x] Update Raw methods (`MeRaw`, `GetRaw`, `ListRaw`, `SetAwayRaw`, `ListActivityLogsRaw`) to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add Parse functions: `ParseAdminMeResult`, `ParseAdminGetResult`, `ParseAdminListResult`, `ParseAdminSetAwayResult`, `ParseAdminListActivityLogsResult`
- [x] Rewrite regular methods as thin wrappers: Raw + `resultError` check + Parse
- [x] Update `admins_test.go` — Raw test assertions use Parse/Decode
- [x] Run `go test ./...` — must pass before next task

### Task 6: Refactor `contacts.go` (22 Raw methods)
- [x] Update all 22 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 22 Parse functions following `Parse{Service}{Method}Result` naming
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Keep `ListAll` method unchanged (iterator, excluded from refactor)
- [x] Update `contacts_test.go` — all Raw test assertions
- [x] Run `go test ./...` — must pass before next task

### Task 7: Refactor `conversations.go` (19 Raw methods)
- [x] Update all 19 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 19 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Keep `ListAll` method unchanged
- [x] Update `conversations_test.go`
- [x] Run `go test ./...` — must pass before next task

### Task 8: Refactor `companies.go` (10 Raw methods)
- [x] Update all 10 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 10 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Keep `ListAll` method unchanged
- [x] Update `companies_test.go`
- [x] Run `go test ./...` — must pass before next task

### Task 9: Refactor `ai_content.go` (10 Raw methods)
- [x] Update all 10 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 10 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Update `ai_content_test.go`
- [x] Run `go test ./...` — must pass before next task

### Task 10: Refactor `tickets.go` (9 Raw methods)
- [x] Update all 9 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 9 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Update `tickets_test.go`
- [x] Run `go test ./...` — must pass before next task

### Task 11: Refactor `news.go` (8) and `help_center.go` (7)
- [x] Update all 15 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 15 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Update `news_test.go` and `help_center_test.go`
- [x] Run `go test ./...` — must pass before next task

### Task 12: Refactor `articles.go` (6) and `internal_articles.go` (6)
- [x] Update all 12 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 12 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Keep `ListAll` methods unchanged
- [x] Update `articles_test.go` and `internal_articles_test.go`
- [x] Run `go test ./...` — must pass before next task

### Task 13: Refactor `ticket_types.go` (6) and `tags.go` (6)
- [x] Update all 12 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 12 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Update `ticket_types_test.go` and `tags_test.go`
- [x] Run `go test ./...` — must pass before next task

### Task 14: Refactor `custom_objects.go` (5) and `fin_voice.go` (5)
- [x] Update all 10 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 10 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Update `custom_objects_test.go` and `fin_voice_test.go`
- [x] Run `go test ./...` — must pass before next task

### Task 15: Refactor `custom_channel_events.go` (4), `data_attributes.go` (3), `data_events.go` (3)
- [x] Update all 10 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 10 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Update corresponding test files
- [x] Run `go test ./...` — must pass before next task

### Task 16: Refactor `data_export.go` (3), `export_reporting.go` (3), `visitors.go` (3)
- [x] Update all 9 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 9 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Keep `Download` methods unchanged (streaming binary, excluded)
- [x] Update `data_export_test.go`, `export_reporting_test.go`, `visitors_test.go`
- [x] Run `go test ./...` — must pass before next task

### Task 17: Refactor small service files — `brands.go` (2), `emails.go` (2), `ip_allowlist.go` (2), `segments.go` (2), `teams.go` (2)
- [x] Update all 10 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 10 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Update corresponding 5 test files
- [x] Run `go test ./...` — must pass before next task

### Task 18: Refactor smallest service files — `away_status_reasons.go` (1), `jobs.go` (1), `messages.go` (1), `notes.go` (1), `phone_call_redirects.go` (1), `subscription_types.go` (1), `ticket_states.go` (1)
- [x] Update all 7 Raw methods to use `s.client.DoRaw`, return `(*Result, error)`
- [x] Add 7 Parse functions
- [x] Rewrite all regular methods as Raw + error check + Parse wrappers
- [x] Update corresponding 7 test files
- [x] Run `go test ./...` — must pass before next task

### Task 19: Verify acceptance criteria
- [x] Verify all 165 Raw methods return `(*Result, error)` with non-generic `Result`
- [x] Verify all 165 Parse functions exist and follow `Parse{Service}{Method}Result` naming
- [x] Verify all regular methods use Raw + error check + Parse pattern
- [x] Verify `Do` still works as public escape hatch
- [x] Verify error predicates (`IsNotFound`, `IsRateLimited`, `IsUnauthorized`) work end-to-end
- [x] Run `go vet ./...` — all issues must be fixed
- [x] Run `go test ./...` — full suite must pass
- [x] Verify no unused code remains (`HTTPResult`, `checkResponse`, old `DoRaw[T]`)

### Task 20: Update documentation
- [x] Update `CLAUDE.md` architecture section to reflect 3-layer pattern
- [x] Update `CLAUDE.md` conventions section with Parse function naming
- [x] Update `CLAUDE.md` Raw results section for non-generic `Result`

## Technical Details

### New `Result` struct (non-generic)
```go
type Result struct {
    StatusCode int
    Header     http.Header
    URL        string
    Body       []byte
    Error      *ErrorResult  // non-nil for 4xx/5xx
}
```

### `Decode[T]` generic helper
```go
func Decode[T any](r *Result) (*T, error) {
    if r.StatusCode == http.StatusNoContent || len(r.Body) == 0 {
        return new(T), nil
    }
    var data T
    if err := json.Unmarshal(r.Body, &data); err != nil {
        return nil, err
    }
    return &data, nil
}
```

### `DoRaw` error contract
- `(nil, err)` — transport/IO failure only
- `(result, nil)` — always, for any completed HTTP request (including 4xx/5xx)

### Parse function pattern
```go
func Parse{Service}{Method}Result(r *Result) (*T, error) { return Decode[T](r) }
```

### Regular method pattern
```go
func (s *XxxService) Method(ctx context.Context, ...) (*T, error) {
    result, err := s.MethodRaw(ctx, ...)
    if err != nil { return nil, err }
    if result.Error != nil { return nil, resultError(result) }
    return Parse{Service}{Method}Result(result)
}
```

### `resultError` helper
Converts `Result.Error` → `*ErrorResponse`, preserving compatibility with `IsNotFound`/`IsRateLimited`/`IsUnauthorized`.

### Files excluded from 3-layer refactor
- `Download` methods in `data_export.go`, `export_reporting.go` (streaming binary)
- `ListAll` methods in `articles.go`, `companies.go`, `contacts.go`, `conversations.go`, `internal_articles.go` (iterators)

## Post-Completion

**Manual verification:**
- Test with real Intercom API to verify Raw methods return usable `Result.Body`
- Verify that `Decode[T]` handles unexpected response shapes gracefully
- Confirm error predicate compatibility in consuming code

**Consuming projects:**
- Any code using `Result[T]` must migrate to non-generic `Result` + `Decode[T]` or Parse functions
- Code using `DoRaw[T]` must switch to `client.DoRaw` + `Decode[T]`
