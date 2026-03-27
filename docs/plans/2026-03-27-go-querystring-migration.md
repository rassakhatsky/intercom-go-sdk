# Replace custom query encoding with google/go-querystring

## Overview
- Replace the custom reflection-based `addQueryOptions`, `parseTag`, and `isZero` functions with `google/go-querystring`
- Simplifies ~60 lines of custom reflection code to a ~5-line wrapper
- All existing `url:"name,omitempty"` struct tags are already compatible — zero tag changes needed
- Refactor callers in service files to use the simplified encoding

## Context (from discovery)
- **Core functions to remove**: `addQueryOptions`, `parseTag`, `isZero` in `intercom.go` (lines 239–326)
- **Callers**: 15 call sites across 11 service files + `pagination_test.go`
- **Struct tags**: 10 files use `url:"..."` tags — all compatible with go-querystring
- **Tests**: `TestAddQueryOptions` and `TestAddQueryOptions_NilOpts` in `intercom_test.go`
- **Dependency**: First external dependency — google/go-querystring has zero transitive deps, stdlib only

## Development Approach
- **Testing approach**: Regular (code first, verify existing tests pass)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility (all service method signatures unchanged)

## Testing Strategy
- **Unit tests**: existing `TestAddQueryOptions*` tests validate behavior; update to cover go-querystring wrapper
- **Integration**: all service `_test.go` files exercise query params end-to-end via `httptest`
- **Regression**: full `go test ./...` must pass after each task

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope
- Keep plan in sync with actual work done

## Implementation Steps

### Task 1: Add google/go-querystring dependency
- [ ] run `go get github.com/google/go-querystring`
- [ ] verify `go.mod` and `go.sum` are updated
- [ ] run `go vet ./...` — must pass

### Task 2: Replace addQueryOptions internals
- [ ] import `github.com/google/go-querystring/query` in `intercom.go`
- [ ] rewrite `addQueryOptions` body to use `query.Values(opts)` instead of custom reflection
- [ ] remove `parseTag` function (no longer needed)
- [ ] remove `isZero` function (no longer needed)
- [ ] run existing `TestAddQueryOptions` and `TestAddQueryOptions_NilOpts` — must pass
- [ ] write additional test cases: pointer fields (*bool), url:"-" skip, non-struct input
- [ ] run `go test ./...` — all tests must pass

### Task 3: Simplify caller pattern across service files
- [ ] audit all 15 call sites to confirm no caller-specific changes are needed
- [ ] if any callers use `addQueryOptions` in a non-standard way, refactor them
- [ ] run `go test ./...` — all tests must pass

### Task 4: Verify acceptance criteria
- [ ] verify `addQueryOptions` works identically to before (same query string output)
- [ ] verify `parseTag` and `isZero` are fully removed
- [ ] verify no other code references removed functions
- [ ] run full test suite: `go test ./...`
- [ ] run `go vet ./...` — all issues must be fixed
- [ ] run `gofmt -l .` — no formatting issues

### Task 5: [Final] Update documentation
- [ ] update CLAUDE.md if architecture section needs changes (mention go-querystring)
- [ ] update README.md if dependency section needs changes

## Technical Details
- **Import**: `github.com/google/go-querystring/query`
- **Key function**: `query.Values(v interface{}) (url.Values, error)` — encodes struct to `url.Values`
- **Tag format**: `url:"name,omitempty"` — identical to current convention
- **Nil handling**: `query.Values(nil)` returns empty `url.Values`, no error
- **Pointer support**: automatically dereferences; nil pointers produce no output (same as current)
- **Bool encoding**: "true"/"false" strings (same as current `fmt.Sprintf("%v", ...)`)

## Post-Completion
- This is the first external dependency — consuming projects that vendor dependencies will need to update
- Consider documenting the dependency policy decision (stdlib + google/go-querystring)
