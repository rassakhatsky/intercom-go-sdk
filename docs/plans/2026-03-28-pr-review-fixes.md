# PR Review Fixes

## Overview
Address critical and important issues identified during comprehensive PR review of the Intercom Go SDK. Fixes cover nil-safety, error handling clarity, documentation accuracy, and test coverage gaps in the root package.

## Context (from discovery)
- **Critical**: `ResultError` panics on nil input; `Decode[T]` returns indistinguishable zero-value on empty body
- **Important**: Synthetic `*http.Response` in `ErrorResponse`; rate limit parse ambiguity; 0% test coverage on root accessors, `DoRawNoRedirect`, `DoDownload`, and new error aliases; doc comment typos
- **Files involved**: `internal/api/errors.go`, `internal/api/result.go`, `internal/api/pagination.go`, `intercom.go`, `aliases.go`, `export/data.go`, `data/events.go`, `contacts/contacts.go`

## Development Approach
- **Testing approach**: TDD — write failing tests first, then implement fixes
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility

## Testing Strategy
- **Unit tests**: required for every task (see Development Approach above)
- All fixes target `internal/api` or root package — use existing test patterns (`testCaller`, `setup`, stdlib `testing` only)

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope
- Keep plan in sync with actual work done

## Implementation Steps

### Task 1: Add nil guard to `ResultError`
- [x] write test: `TestResultError_NilResult` in `internal/api/errors_test.go` — call `ResultError(nil)`, assert returns `nil` (not panic)
- [x] add nil check at top of `ResultError` in `internal/api/errors.go:83`: `if r == nil || r.Error == nil { return nil }`
- [x] run tests — must pass before next task

### Task 2: Handle empty body in `Decode[T]` for non-204 responses
- [x] write test: `TestDecode_EmptyBodyNon204` in `internal/api/result_test.go` — create `Result{StatusCode: 200, Body: nil}`, call `Decode[SomeStruct]`, assert error is returned (not zero-value struct)
- [x] write test: `TestDecode_EmptyBody204` — confirm 204 still returns zero-value struct (no regression)
- [x] update `Decode[T]` in `internal/api/result.go:48` to return error on empty body when status is not 204
- [x] run tests — must pass before next task

### Task 3: Replace synthetic `*http.Response` in `ErrorResponse`
- [x] write test: `TestErrorResponse_ResponseMeta` in `internal/api/errors_test.go` — verify `ErrorResponse` exposes `StatusCode` and `Header` without nil-body `*http.Response`
- [x] add `StatusCode int` and `Headers http.Header` fields to `ErrorResponse` in `internal/api/errors.go`
- [x] remove `Response *http.Response` field from `ErrorResponse`
- [x] update `ResultError` to populate the new fields instead of creating synthetic `*http.Response`
- [x] update `Error()` method and status predicate functions (`IsNotFound`, `IsBadRequest`, etc.) to use new fields
- [x] update all tests referencing `ErrorResponse.Response` to use new fields
- [x] run tests — must pass before next task

### Task 4: Add test coverage for root package accessor methods
- [x] write test: `TestClient_AllAccessorsNonNil` in `intercom_test.go` — create `NewClient("tok")`, assert every accessor (`Contacts()`, `Tags()`, `Admins()`, etc.) returns non-nil
- [x] run tests — must pass before next task

### Task 5: Add test coverage for `DoRawNoRedirect` and `DoDownload`
- [x] write test: `TestClient_DoRawNoRedirect_302` in `do_raw_test.go` — verify 302 response is returned as result (redirect not followed)
- [x] write test: `TestClient_DoDownload_Success` — verify response body is streamed to writer
- [x] write test: `TestClient_DoDownload_Error` — verify 4xx returns proper error
- [x] run tests — must pass before next task

### Task 6: Add test coverage for new error alias wrappers
- [x] write test: `TestErrorAliases` in `errors_test.go` (root package) — verify `IsBadRequest`, `IsForbidden`, `IsConflict`, `IsUnprocessableEntity`, `IsServerError` delegate correctly
- [x] run tests — must pass before next task

### Task 7: Fix documentation issues
- [x] fix `export/data.go:46` — change comment from `ParseGetStatusResult` to `ParseDataGetStatusResult`
- [x] fix `data/events.go:161` — change `lisdataevents` to `listdataevents` in See URL (verify against OpenAPI spec first)
- [x] fix `contacts/contacts.go:301` — change "a api.TagList" to "an api.TagList"
- [x] run tests — must pass before next task

### Task 8: Add concurrency safety doc to `Iter[T]`
- [ ] add doc comment to `Iter[T]` type in `internal/api/pagination.go` noting it is not safe for concurrent use
- [ ] run tests — must pass before next task

### Task 9: Verify acceptance criteria
- [ ] verify all requirements from Overview are implemented
- [ ] verify edge cases are handled
- [ ] run full test suite (`go test ./...`)
- [ ] run linter (`go vet ./...` and `gofmt -l .`) — all issues must be fixed
- [ ] verify no regressions in existing tests

### Task 10: [Final] Update documentation
- [ ] update CLAUDE.md if any architectural patterns changed (e.g., `ErrorResponse` field rename)

## Technical Details

### Task 1: `ResultError` nil guard
Add `if r == nil` to existing nil check: `if r == nil || r.Error == nil { return nil }`

### Task 2: `Decode[T]` empty body handling
Split the current condition:
```go
if r.StatusCode == http.StatusNoContent {
    return new(T), nil
}
if len(r.Body) == 0 {
    return nil, fmt.Errorf("unexpected empty response body (HTTP %d)", r.StatusCode)
}
```

### Task 3: `ErrorResponse` field change
Replace `Response *http.Response` with direct fields. This is a breaking change to the internal error type but improves safety. Status predicates switch from `e.Response.StatusCode` to `e.StatusCode`.

## Post-Completion
**Manual verification:**
- Review that `ErrorResponse` field change doesn't break any downstream consumers (this is a new SDK, so no external consumers yet)
