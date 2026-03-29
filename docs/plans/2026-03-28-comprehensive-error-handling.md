# Comprehensive Error Handling

## Overview
- Add all 22 Intercom API error code constants so users can match against specific error codes without string comparisons
- Extend `ErrorResponse` with a `*RateLimitInfo` field that parses `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, and `Retry-After` headers from 429 responses
- Add common status helpers: `IsBadRequest`, `IsForbidden`, `IsConflict`, `IsUnprocessableEntity`, `IsServerError`
- Ensure `ErrorResponse.Error()` includes the error code in its output for better debugging

## Context (from discovery)
- **Core error files:** `internal/api/errors.go`, `internal/api/result.go`
- **Type re-exports:** `aliases.go` — all new public types need aliases here
- **Error construction:** `BuildResult()` in `result.go` parses JSON errors; `ResultError()` in `errors.go` converts to `ErrorResponse`
- **Client methods:** `Do()`, `DoRaw()`, `DoDownload()` in `intercom.go` use the error pipeline
- **Existing helpers:** `IsNotFound` (404), `IsRateLimited` (429), `IsUnauthorized` (401) in `errors.go`
- **Test files:** `internal/api/errors_test.go`, `internal/api/result_test.go`, `errors_test.go`

## Development Approach
- **Testing approach**: TDD (tests first)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility — all existing public API remains unchanged

## Testing Strategy
- **Unit tests**: required for every task
- Tests follow existing patterns: `setup()`, `testMethod`, `testHeader`, raw JSON responses
- Use table-driven tests for error code constants and status helpers

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope
- Keep plan in sync with actual work done

## Implementation Steps

### Task 1: Add Intercom error code constants
- [x] Write tests in `internal/api/errors_test.go` asserting all 22 error code constants exist and have correct string values
- [x] Define error code constants in `internal/api/errors.go` as `const` block with type `ErrorCode string`
- [x] Add `ErrorCode` type alias in `aliases.go`
- [x] Run tests — must pass before next task

**Error codes to define (from Intercom docs):**
`server_error`, `client_error`, `type_mismatch`, `parameter_not_found`, `parameter_invalid`, `action_forbidden`, `conflict`, `api_plan_restricted`, `rate_limit_exceeded`, `unsupported`, `token_revoked`, `token_blocked`, `token_not_found`, `token_unauthorized`, `token_expired`, `missing_authorization`, `retry_after`, `job_closed`, `not_restorable`, `team_not_found`, `team_unavailable`, `admin_not_found`

### Task 2: Add RateLimitInfo type and populate on 429 responses
- [x] Write tests in `internal/api/errors_test.go` for `RateLimitInfo` parsing from response headers (success cases with all headers, partial headers, missing headers)
- [x] Write tests verifying `ErrorResponse.RateLimit` is populated when status is 429 and nil otherwise
- [x] Define `RateLimitInfo` struct in `internal/api/errors.go` with fields: `Limit int`, `Remaining int`, `Reset time.Time`, `RetryAfter time.Duration`
- [x] Add helper `parseRateLimitInfo(h http.Header) *RateLimitInfo` in `internal/api/errors.go`
- [x] Update `ResultError()` to call `parseRateLimitInfo` when status code is 429
- [x] Add `RateLimitInfo` type alias in `aliases.go`
- [x] Run tests — must pass before next task

**Rate limit headers to parse:**
- `X-RateLimit-Limit` → `Limit` (int)
- `X-RateLimit-Remaining` → `Remaining` (int)
- `X-RateLimit-Reset` → `Reset` (time.Time from Unix timestamp)
- `Retry-After` → `RetryAfter` (time.Duration in seconds)

### Task 3: Add common status code helpers
- [x] Write tests in `internal/api/errors_test.go` for each new helper: `IsBadRequest`, `IsForbidden`, `IsConflict`, `IsUnprocessableEntity`, `IsServerError`
- [x] Write tests for `IsServerError` covering 500, 502, 503, 504 (range check)
- [x] Write tests verifying helpers return false for nil errors and non-ErrorResponse errors
- [x] Implement `IsBadRequest` (400), `IsForbidden` (403), `IsConflict` (409), `IsUnprocessableEntity` (422) using existing `hasStatusCode` pattern
- [x] Implement `IsServerError` checking status code range 500-599
- [x] Add function aliases in `aliases.go` (or verify they're accessible via root package)
- [x] Run tests — must pass before next task

### Task 4: Add ErrorCode field to ErrorDetail and matching helper
- [ ] Write tests for `ErrorResponse.HasErrorCode(code ErrorCode) bool` — matches against any error in the Errors slice
- [ ] Write tests for `ErrorDetail.Code` being typed as `ErrorCode` (or a helper that accepts `ErrorCode`)
- [ ] Update `ErrorDetail.Code` field type from `string` to `ErrorCode` in `internal/api/result.go` (backward compatible since `ErrorCode` is `type ErrorCode string`)
- [ ] Add `HasErrorCode(code ErrorCode) bool` method to `ErrorResponse`
- [ ] Run tests — must pass before next task

### Task 5: Verify acceptance criteria
- [ ] Verify all 22 error code constants are defined and documented
- [ ] Verify `RateLimitInfo` is populated on 429 responses with correct header parsing
- [ ] Verify all 5 new status helpers work correctly
- [ ] Verify `HasErrorCode` method works for matching specific Intercom error codes
- [ ] Verify backward compatibility — no existing public API changed in breaking way
- [ ] Run full test suite (`go test ./...`)
- [ ] Run linter (`go vet ./...`)
- [ ] Run format check (`gofmt -l .`)

### Task 6: [Final] Update documentation
- [ ] Update `doc.go` with examples showing new error code constants and RateLimitInfo usage
- [ ] Update `example_test.go` with runnable examples for rate limit handling and error code matching

## Technical Details

### RateLimitInfo struct
```go
type RateLimitInfo struct {
    Limit      int           // X-RateLimit-Limit: total requests allowed
    Remaining  int           // X-RateLimit-Remaining: requests remaining in window
    Reset      time.Time     // X-RateLimit-Reset: when the rate limit window resets
    RetryAfter time.Duration // Retry-After: how long to wait before retrying
}
```

### ErrorCode constants
```go
type ErrorCode string

const (
    ErrServerError        ErrorCode = "server_error"
    ErrClientError        ErrorCode = "client_error"
    ErrTypeMismatch       ErrorCode = "type_mismatch"
    ErrParameterNotFound  ErrorCode = "parameter_not_found"
    ErrParameterInvalid   ErrorCode = "parameter_invalid"
    ErrActionForbidden    ErrorCode = "action_forbidden"
    ErrConflict           ErrorCode = "conflict"
    ErrAPIPlanRestricted  ErrorCode = "api_plan_restricted"
    ErrRateLimitExceeded  ErrorCode = "rate_limit_exceeded"
    ErrUnsupported        ErrorCode = "unsupported"
    ErrTokenRevoked       ErrorCode = "token_revoked"
    ErrTokenBlocked       ErrorCode = "token_blocked"
    ErrTokenNotFound      ErrorCode = "token_not_found"
    ErrTokenUnauthorized  ErrorCode = "token_unauthorized"
    ErrTokenExpired       ErrorCode = "token_expired"
    ErrMissingAuth        ErrorCode = "missing_authorization"
    ErrRetryAfter         ErrorCode = "retry_after"
    ErrJobClosed          ErrorCode = "job_closed"
    ErrNotRestorable      ErrorCode = "not_restorable"
    ErrTeamNotFound       ErrorCode = "team_not_found"
    ErrTeamUnavailable    ErrorCode = "team_unavailable"
    ErrAdminNotFound      ErrorCode = "admin_not_found"
)
```

### Usage examples
```go
// Check specific error code
_, err := client.Contacts().Get(ctx, "123")
var apiErr *intercom.ErrorResponse
if errors.As(err, &apiErr) {
    if apiErr.HasErrorCode(intercom.ErrParameterInvalid) {
        // handle invalid parameter
    }
}

// Handle rate limiting with parsed info
if intercom.IsRateLimited(err) {
    var apiErr *intercom.ErrorResponse
    errors.As(err, &apiErr)
    fmt.Printf("Rate limited. Retry after: %v\n", apiErr.RateLimit.RetryAfter)
    time.Sleep(apiErr.RateLimit.RetryAfter)
}

// Check for server errors (any 5xx)
if intercom.IsServerError(err) {
    // retry logic
}
```

## Post-Completion

**Manual verification:**
- Test against live Intercom API sandbox to verify rate limit header parsing
- Verify error codes match actual API responses (not just documentation)
