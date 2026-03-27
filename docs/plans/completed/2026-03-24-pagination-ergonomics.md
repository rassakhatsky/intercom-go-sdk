# Pagination Ergonomics: Collect() and ForEach() for Iter[T]

## Overview
- Add `Collect()` and `ForEach()` methods to `Iter[T]` to eliminate repetitive iteration boilerplate
- Currently every consumer of `ListAll` must write 6-8 lines of `for iter.Next() { append } if iter.Err() { ... }` — this reduces to a single `Collect()` call
- Methods live on `Iter[T]` in `pagination.go`, keeping the API surface discoverable and cohesive

## Context (from discovery)
- Files/components involved: `pagination.go`, `pagination_test.go`
- Related patterns: 5 services have `ListAll` methods returning `*Iter[T]` (contacts, companies, articles, conversations, internal_articles)
- All 7 test files repeat the same `for iter.Next()` + `Err()` boilerplate
- No external dependencies — stdlib only

## Development Approach
- **Testing approach**: TDD (tests first)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility

## Testing Strategy
- **Unit tests**: required for every task (see Development Approach above)
- Tests follow existing pattern: `httptest.Server` via `setup()` helper, stdlib `testing` only
- Table-driven tests where multiple scenarios exist

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope
- Keep plan in sync with actual work done

## Implementation Steps

### Task 1: Add Collect() method — tests first
- [x] Write test `TestIter_Collect_MultiplePages` — verifies Collect() drains a 2-page iterator into a complete slice
- [x] Write test `TestIter_Collect_EmptyResult` — verifies Collect() returns empty slice (not nil) for 0 items
- [x] Write test `TestIter_Collect_Error` — verifies Collect() returns partial results + error when page 2 fails
- [x] Implement `Collect() ([]T, error)` on `Iter[T]` in `pagination.go`
- [x] Run tests — must pass before next task

### Task 2: Add ForEach() method — tests first
- [x] Write test `TestIter_ForEach_AllItems` — verifies callback is called for each item across pages
- [x] Write test `TestIter_ForEach_EarlyExit` — verifies returning an error from the callback stops iteration and propagates the error
- [x] Write test `TestIter_ForEach_FetchError` — verifies fetch errors are returned (not swallowed)
- [x] Implement `ForEach(fn func(T) error) error` on `Iter[T]` in `pagination.go`
- [x] Run tests — must pass before next task

### Task 3: Verify acceptance criteria
- [x] Verify Collect() works: drains iterator, returns `([]T, error)`, handles empty + error cases
- [x] Verify ForEach() works: calls fn per item, supports early exit via error, propagates fetch errors
- [x] Verify backward compatibility: existing `Next()`/`Current()`/`Err()` API unchanged
- [x] Run full test suite (`go test ./...`)
- [x] Run linter (`go vet ./...`)

### Task 4: [Final] Update documentation
- [x] Add doc comments on Collect() and ForEach() consistent with existing style
- [x] Update `doc.go` example to show Collect() usage alongside existing iterator pattern

## Technical Details

**Collect() signature:**
```go
func (it *Iter[T]) Collect() ([]T, error)
```
- Drains the iterator by calling `Next()` in a loop
- Returns all items collected + first error (if any)
- On error, returns items collected so far + the error
- Returns empty non-nil slice for 0 results

**ForEach() signature:**
```go
func (it *Iter[T]) ForEach(fn func(T) error) error
```
- Calls `fn` for each item yielded by `Next()`
- If `fn` returns non-nil error, stops iteration and returns that error
- If a fetch error occurs, returns `Iter.Err()`
- Distinction: callback errors vs fetch errors are both returned, but the caller can differentiate by checking `iter.Err()` after `ForEach`

## Post-Completion

**Manual verification:**
- Try using `Collect()` in a real Intercom API call to confirm ergonomics feel right
- Consider whether any existing `ListAll` test helpers should be updated to use `Collect()` (optional follow-up)
