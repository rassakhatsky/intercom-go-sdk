# PR Review: Critical + Important Fixes

## Overview
Address 2 critical and 8 important findings from the comprehensive PR review of the Intercom Go SDK. These fixes improve type safety, error handling robustness, and documentation accuracy without changing the SDK's public API behavior.

## Context (from PR review)
- **Critical**: `Operator` type alias provides zero type safety (the `Decode[T]` empty body finding was dismissed — callers can use Raw methods to inspect bodies directly)
- **Important**: `ErrorResponse.Error()` nil receiver panic; incomplete status predicate lists in comments; misleading "backward compat" comments; `NewClient` missing token validation; `SingleFilterOf` accepts `string` not `Operator`; `parseRateLimitInfo` zero-value ambiguity; `doc.go` uses raw `"="` instead of `OpEquals`
- Files involved: `internal/api/result.go`, `internal/api/search.go`, `internal/api/errors.go`, `intercom.go`, `aliases.go`, `doc.go`

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
- Existing test files: `internal/api/result_test.go`, `internal/api/search_test.go`, `internal/api/errors_test.go`, `intercom_test.go`, `search_test.go`

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Change `Operator` from type alias to defined type (Critical #2)
- [x] write test in `internal/api/search_test.go`: verify `Operator` constants have expected string values and that `string(OpEquals)` round-trips correctly
- [x] change `type Operator = string` to `type Operator string` in `internal/api/search.go:4`
- [x] change `Filter.Operator` field from `string` to `Operator` in `internal/api/search.go:43`
- [x] change `SingleFilterOf` parameter from `string` to `Operator` in `internal/api/search.go:49`
- [x] update the `Operator` alias in `aliases.go` if needed (it re-exports `api.Operator`)
- [x] run tests — must pass before next task

### Task 2: Add nil receiver guard to `ErrorResponse.Error()` (Important #4)
- [x] write test in `internal/api/errors_test.go`: call `Error()` on nil `*ErrorResponse`, expect non-panic return
- [x] add nil receiver check at top of `ErrorResponse.Error()` in `internal/api/errors.go:59`
- [x] run tests — must pass before next task

### Task 3: Validate non-empty token in `NewClient` (Important #6)
- [x] write test in `intercom_test.go`: `NewClient("")` should panic with descriptive message
- [x] write test in `intercom_test.go`: `NewClient("   ")` (whitespace-only) should panic
- [x] write test in `intercom_test.go`: `NewClient("valid-token")` should not panic
- [x] add `strings.TrimSpace(token) == ""` panic check at top of `NewClient` in `intercom.go:110`
- [x] run tests — must pass before next task

### Task 4: Fix incomplete status predicate lists in comments (Important #10)
- [x] update `ResultError` doc comment in `internal/api/errors.go:82-83` to reference all predicates or use general phrase
- [x] update `Do` method doc comment in `intercom.go:442-445` to reference all predicates or use general phrase
- [x] run tests — must pass before next task (no code changes, but verify nothing broke)

### Task 5: Fix misleading comments in `aliases.go` and `doc.go` (Important #9, #16)
- [x] update `doc.go:33`: change `"="` to `intercom.OpEquals` in the search example
- [x] update `aliases.go:65`: rewrite "backward compat" comment to explain actual purpose
- [x] update `aliases.go:136`: rewrite "for root services" comment to explain actual purpose
- [x] run tests — must pass before next task

### Task 6: Verify acceptance criteria
- [x] verify all critical issues addressed: `Operator` type safety, `Decode[T]` empty body
- [x] verify all important issues addressed: nil guard, token validation, comment fixes, `SingleFilterOf` signature
- [x] run full test suite (`go test ./...`)
- [x] run linter (`go vet ./...`)
- [x] run format check (`gofmt -l .`)

### Task 7: Update documentation
- [ ] update CLAUDE.md if any conventions changed (e.g., `Operator` is now a defined type)
- [ ] update `internal/api/search.go` Filter doc comment to explain `Value` union type (scalar vs `[]*Filter`)
