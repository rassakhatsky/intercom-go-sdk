# Remove internal/api — Make API Package Public

## Overview
- Move `internal/api/` → `api/` so consumers can import shared types directly
- Delete `aliases.go` — no longer needed since `api/` is public
- Replace unexported var wrappers with direct `api.` calls
- Update `doc.go` and `CLAUDE.md` to reflect new structure
- **Motivation**: `internal/` adds indirection without value — the SDK re-exports nearly everything via aliases anyway. A public `api/` package is simpler and lets consumers use `api.Result`, `api.Filter` directly

## Context (from discovery)
- Files/components involved: `internal/api/` (8 source + 6 test files), `aliases.go`, `intercom.go`, `doc.go`, 34 sub-package source files, 17 sub-package test files, root test files (`result_test.go`, `errors_test.go`, `intercom_test.go`, `pagination_test.go`)
- No other packages under `internal/` — directory can be removed entirely
- 3 unexported vars (`buildResult`, `resultError`, `addQueryOptions`) used in `intercom.go` and root test files need replacing with direct `api.` calls

## Development Approach
- **Testing approach**: TDD — write import-verification tests before moving files, ensure all existing tests pass after each task
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run `go test ./...` after each change
- Maintain backward compatibility within sub-packages (import path change only)

## Testing Strategy
- **Unit tests**: existing test suite covers all functionality — refactor must not break any tests
- **Verification**: `go vet ./...` and `gofmt -l .` must pass after refactor
- **Import verification**: confirm `api/` package is importable and types resolve correctly

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Write TDD verification test for public api/ package
- [x] Create `api_public_test.go` in root package that imports `api/` (will fail initially since path doesn't exist yet)
- [x] Test that key types are accessible: `api.Result`, `api.Filter`, `api.Iter`, `api.ErrorResponse`
- [x] Test that key functions are accessible: `api.BuildResult`, `api.ResultError`, `api.SingleFilterOf`
- [x] Verify test fails with expected import error (confirms TDD red phase)

### Task 2: Move `internal/api/` → `api/`
- [x] Move directory: `internal/api/` → `api/`
- [x] Remove empty `internal/` directory
- [x] Verify `api/` package compiles: `go build ./api/...`
- [x] Run `go test ./api/...` — existing api tests must pass

### Task 3: Update sub-package imports (34 files)
- [x] Replace `github.com/rassakhatsky/intercom-go-sdk/internal/api` → `github.com/rassakhatsky/intercom-go-sdk/api` in all sub-package source files
- [x] Replace same import in all sub-package test files (17 files with testutil helpers)
- [x] Run `go build ./...` — all packages must compile
- [x] Run `go test ./...` — all tests must pass

### Task 4: Replace unexported vars with direct `api.` calls
- [x] In `intercom.go`: replace `buildResult(...)` → `api.BuildResult(...)`, `resultError(...)` → `api.ResultError(...)` (add `api` import)
- [x] In root test files (`result_test.go`, `errors_test.go`, `intercom_test.go`, `pagination_test.go`): replace `buildResult` → `api.BuildResult`, `resultError` → `api.ResultError`, `addQueryOptions` → `api.AddQueryOptions`
- [x] Run `go test ./...` — all tests must pass

### Task 5: Delete `aliases.go`
- [ ] Delete `aliases.go`
- [ ] Fix any compilation errors from removed aliases (check if `intercom.go` or root test files reference alias types like `Result`, `Logger`)
- [ ] For types used in `intercom.go` (e.g., `Result`, `Response`, `Logger`), add direct `api.` qualification
- [ ] Run `go build ./...` — must compile
- [ ] Run `go test ./...` — all tests must pass

### Task 6: Update documentation
- [ ] Update `doc.go`: remove references to `internal/api` and aliases pattern
- [ ] Update `CLAUDE.md`: reflect new `api/` package structure, remove aliases section
- [ ] Run `go vet ./...` — must pass
- [ ] Run `gofmt -l .` — no formatting issues

### Task 7: Verify acceptance criteria
- [ ] Verify `internal/` directory no longer exists
- [ ] Verify `aliases.go` no longer exists
- [ ] Verify no Go files reference `internal/api`
- [ ] Run full test suite: `go test ./...`
- [ ] Run `go vet ./...`
- [ ] Verify TDD test from Task 1 passes (green phase)
- [ ] Clean up TDD test if it duplicates existing coverage

## Technical Details
- Package declaration in `api/*.go` is already `package api` — no changes needed
- Go module path: `github.com/rassakhatsky/intercom-go-sdk`
- Import path changes: `github.com/rassakhatsky/intercom-go-sdk/internal/api` → `github.com/rassakhatsky/intercom-go-sdk/api`
- Root package (`package intercom`) will import `api` directly instead of through aliases
- Sub-packages continue to import `api` for the `Caller` interface — only the path changes

## Post-Completion

**Manual verification:**
- Confirm consumers can `import "github.com/rassakhatsky/intercom-go-sdk/api"` and access all types
- Review godoc output to ensure `api/` package documentation renders correctly
