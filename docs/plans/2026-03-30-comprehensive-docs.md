# Comprehensive Documentation

## Overview
- Add minimal, developer-focused documentation across the SDK
- Fill gaps: sub-package godoc, testable examples, contributing guide, README polish
- Keep everything concise — just what developers need to get productive

## Context (from discovery)
- Files/components involved: all 16 sub-packages + root + `api/`
- Existing docs: `README.md` (solid, 242 lines), `doc.go` (root only, 99 lines), `CLAUDE.md` (dev guide)
- Missing: sub-package `doc.go` files, testable `Example_*` functions, `CONTRIBUTING.md`
- No external test frameworks — stdlib only

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
  - testable examples (`Example_*` functions) serve as both documentation and tests
  - all examples must compile and produce expected output
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change

## Testing Strategy
- **Unit tests**: testable `Example_*` functions validate example code compiles and runs
- No E2E tests applicable (SDK library, not an application)

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope

## What Goes Where
- **Implementation Steps** (`[ ]` checkboxes): doc files, example tests, README updates
- **Post-Completion** (no checkboxes): publishing godoc, announcing changes

## Implementation Steps

### Task 1: Add sub-package doc.go files
- [x] add `api/doc.go` with package comment describing core types and Caller interface
- [x] add `contacts/doc.go` with package comment
- [x] add `conversations/doc.go` with package comment
- [x] add `companies/doc.go` with package comment
- [x] add `tickets/doc.go` with package comment
- [x] add `articles/doc.go` with package comment
- [x] add `helpcenter/doc.go` with package comment
- [x] add `news/doc.go` with package comment
- [x] add `ai/doc.go` with package comment
- [x] add `calls/doc.go` with package comment
- [x] add `tags/doc.go` with package comment
- [x] add `admins/doc.go` with package comment
- [x] add `segments/doc.go` with package comment
- [x] add `data/doc.go` with package comment
- [x] add `messaging/doc.go` with package comment
- [x] add `export/doc.go` with package comment
- [x] add `settings/doc.go` with package comment
- [x] add `workflows/doc.go` with package comment
- [x] run `go vet ./...` — must pass before next task

### Task 2: Add testable examples for core api/ package
- [ ] create `api/example_test.go` with examples for `Decode`, `SingleFilterOf`, `And`/`Or` filter builders
- [ ] add example for `Iter` usage pattern (Next/Current/Err loop)
- [ ] add example for `ErrorResponse` status predicates
- [ ] run tests `go test ./api/...` — must pass before next task

### Task 3: Add testable examples for key sub-packages
- [ ] create `contacts/example_test.go` with examples for Create, Get, ListAll, Search
- [ ] create `tags/example_test.go` with examples for Create, List, Tag/Untag
- [ ] create `conversations/example_test.go` with examples for Get, Reply, ListAll
- [ ] create `tickets/example_test.go` with examples for Create, Get, Update
- [ ] run tests `go test ./contacts/... ./tags/... ./conversations/... ./tickets/...` — must pass before next task

### Task 4: Add CONTRIBUTING.md
- [ ] create `CONTRIBUTING.md` covering: dev setup, running tests, coding conventions (3-layer pattern, naming), PR process
- [ ] keep it concise — reference CLAUDE.md for detailed architecture
- [ ] run `go vet ./...` — must pass (sanity check)

### Task 5: Polish README.md
- [ ] add link to CONTRIBUTING.md
- [ ] add link to pkg.go.dev for API reference (once published)
- [ ] review existing examples for accuracy against current API
- [ ] run full test suite `go test ./...` — must pass

### Task 6: Verify acceptance criteria
- [ ] verify all 18 sub-packages (16 domain + api + root) have package-level godoc
- [ ] verify example tests compile and pass
- [ ] verify CONTRIBUTING.md exists and is accurate
- [ ] run full test suite `go test ./...`
- [ ] run linter `go vet ./...` — all issues must be fixed
- [ ] run `gofmt -l .` — no formatting issues

### Task 7: [Final] Update project knowledge
- [ ] update CLAUDE.md if new patterns discovered
- [ ] update README.md table of contents if needed

*Note: ralphex automatically moves completed plans to `docs/plans/completed/`*

## Technical Details
- Sub-package `doc.go` files: single `// Package X ...` comment block, 2-5 lines max
- Testable examples: use `Example_*` naming convention with `// Output:` comments
- Examples need a mock server setup since they call API methods — use the existing `testCaller` pattern from test files
- CONTRIBUTING.md: ~50-80 lines, reference CLAUDE.md for architecture details

## Post-Completion
*Items requiring manual intervention or external systems — no checkboxes, informational only*

**Publishing:**
- Push to GitHub to trigger pkg.go.dev indexing
- Verify godoc renders correctly on pkg.go.dev
- Update README link to pkg.go.dev once available
