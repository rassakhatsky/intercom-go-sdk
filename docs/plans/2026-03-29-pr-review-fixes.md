# Fix PR Review Findings

## Overview
- Address all important issues and selected suggestions identified during comprehensive PR review of the `init` branch
- Fixes span 5 categories: comment accuracy, nil safety, URL consistency, parse function naming, and type design suggestions
- All changes are backward-compatible — no public API signature changes

## Context (from PR review)
- Files/components involved: `data/attributes.go`, `data/objects.go`, `export/reporting.go`, `workflows/workflows.go`, `ai/voice.go`, `internal/api/result.go`, `calls/calls.go`, `messaging/subscriptions.go`, `search.go`
- 5 review agents analyzed ~34K lines across 94 files
- 7 important issues + selected suggestions from type design and error handling reviews
- All tests currently pass (19 packages)

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility

## Testing Strategy
- **Unit tests**: required for every task with code changes
- Existing test infrastructure: `setup()` / `testMethod` / `testHeader` pattern per sub-package
- stdlib `testing` only — no external frameworks

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Fix `// See:` URL typos and inconsistencies
- [x] Fix `data/attributes.go:99,143` — change `lisdataattributes` to `listdataattributes` in both `List` and `ListRaw` URLs
- [x] Fix `export/reporting.go:82-163` — replace all 7 OpenAPI path-encoded URLs with `{tag}/{operationId}` format
- [x] Fix `workflows/workflows.go:60,77` — normalize URL casing to match convention (lowercase tag/operationId)
- [x] run tests (`go test ./data/... ./export/... ./workflows/...`) - must pass before next task

### Task 2: Fix comment-code mismatches
- [x] Fix `ai/voice.go:88-90` — change "external reference ID" to "Intercom ID" in `Collect` comment
- [x] Fix `data/attributes.go:80,85,90` — align Parse function comment prefixes with actual function names (`ParseAttributeListResult`, `ParseAttributeCreateResult`, `ParseAttributeUpdateResult`)
- [x] Fix `data/objects.go:52-73` — align Parse function comment prefixes with actual function names (`ParseObjectGetResult`, `ParseObjectGetByExternalIDResult`, etc.)
- [x] Fix `messaging/subscriptions.go:36` — change `ParseListResult` to `ParseSubscriptionListResult` in comment
- [x] run tests (`go test ./ai/... ./data/... ./messaging/...`) - must pass before next task

### Task 3: Add nil guard in `BuildResult`
- [x] Add nil check for `resp.Request` and `resp.Request.URL` in `internal/api/result.go:70` before accessing `resp.Request.URL.String()`
- [x] Write test for `BuildResult` with nil `resp.Request` — verify no panic and `URL` is empty string
- [x] run tests (`go test ./internal/api/...`) - must pass before next task

### Task 4: Add validation to `ParseGetTranscriptResult`
- [x] Add nil guard for `r` parameter in `calls/calls.go:104-106`
- [x] Add empty body check — return error for empty transcript
- [x] Write test for `ParseGetTranscriptResult` with nil `*Result` — verify error returned
- [x] Write test for `ParseGetTranscriptResult` with empty body — verify error returned
- [x] Write test for `ParseGetTranscriptResult` with valid body — verify transcript returned
- [x] run tests (`go test ./calls/...`) - must pass before next task

### Task 5: Add nil guard to `Decode[T]`
- [x] Add nil check for `r` parameter in `internal/api/result.go:50` — return error for nil input
- [x] Write test for `Decode[T]` with nil `*Result` — verify error returned (not panic)
- [x] run tests (`go test ./internal/api/...`) - must pass before next task

### Task 6: Improve `ParseGetRecordingURLResult` error message
- [x] Include response body and status code in error message for unexpected status codes in `calls/calls.go:92-101`
- [x] Write test for `ParseGetRecordingURLResult` with 200 status — verify error includes body context
- [x] run tests (`go test ./calls/...`) - must pass before next task

### Task 7: Add `Filter` operator constants
- [x] Define typed operator constants in `internal/api/search.go` (e.g., `OpEquals = "="`, `OpNotEquals = "!="`, `OpContains = "~"`, `OpAND = "AND"`, `OpOR = "OR"`, etc.)
- [x] Update `SingleFilterOf`, `And`, `Or` to use the constants internally
- [x] Update existing tests to use constants where applicable
- [x] run tests (`go test ./internal/api/...`) - must pass before next task

### Task 8: Verify acceptance criteria
- [x] verify all 7 important issues from PR review are addressed
- [x] verify suggestions (nil guards, operator constants, error messages) are implemented
- [x] run full test suite (`go test ./...`)
- [x] run linter (`go vet ./...` && `gofmt -l .`)
- [x] verify no new test failures introduced

### Task 9: [Final] Update documentation
- [ ] update CLAUDE.md if operator constants or new conventions were added
- [ ] update `aliases.go` if new public types were added (operator constants)

*Note: ralphex automatically moves completed plans to `docs/plans/completed/`*

## Technical Details
- URL format convention: `// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/{tag}/{operationId}`
- All tags and operationIds use lowercase
- Parse function comments must start with the exact function name per godoc convention
- Nil guards follow pattern: early return with `fmt.Errorf(...)` for public functions

## Post-Completion
**Manual verification:**
- Spot-check `// See:` URLs by opening a few in a browser to confirm they resolve
- Review `go doc` output for fixed Parse functions to confirm godoc renders correctly
