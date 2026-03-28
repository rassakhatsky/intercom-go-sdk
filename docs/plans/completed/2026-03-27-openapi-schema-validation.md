# OpenAPI Schema Validation — SDK vs api.intercom.io.yaml

## Overview
- Validate every operation in `raw/api.intercom.io.yaml` (162 operations, Intercom API v2.15) against the Go SDK
- Identify missing operations, extra SDK methods, and struct field mismatches
- Fix all gaps: add missing methods, correct struct fields, add tests
- Ensure bidirectional completeness: schema → SDK and SDK → schema

## Context (from discovery)
- **OpenAPI schema**: `raw/api.intercom.io.yaml` — 23,719 lines, 162 operations across ~80 paths
- **SDK**: 33 services, ~165 public methods (excluding Raw/Parse variants), single `package intercom`
- **Known gaps identified in preliminary scan**:
  - `GET /contacts/{contact_id}/tags` (listTagsForAContact) — **missing from SDK**
  - `GET /export/workflows/{id}` (exportWorkflow) — **missing from SDK** (no Workflows service)
- **Pattern**: SDK follows 3-layer pattern (Raw → Parse → Method), all methods take `context.Context`
- **Testing**: `httptest.Server` via `setup()`, stdlib `testing` only

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change

## Testing Strategy
- **Unit tests**: required for every task (httptest patterns from `testutil_test.go`)
- No e2e tests in this project

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix

## Implementation Steps

### Task 1: Build validation script — endpoint coverage
Write a Python script (`scripts/validate_openapi.py`) that:
- Parses `raw/api.intercom.io.yaml` to extract all operations (method, path, operationId, request body schema, response schema, parameters)
- Scans all `*.go` files (non-test) to extract SDK methods with their HTTP method + URL path from `NewRequest(...)` calls
- Produces a mapping report: schema operation → SDK method
- Flags: missing SDK methods, extra SDK methods not in schema, HTTP method mismatches, path mismatches

- [x] Create `scripts/validate_openapi.py` with YAML parsing and Go source scanning
- [x] Run script and capture initial endpoint coverage report
- [x] Save report to `docs/reports/openapi-endpoint-coverage.md`
- [x] No tests needed (tooling script, not SDK code)

### Task 2: Extend validation script — request/response struct fields
Extend the script to:
- For each schema operation with a request body: resolve `$ref` to get property names/types
- For each SDK method: find the corresponding request struct and extract its JSON-tagged fields
- Compare: flag missing fields, extra fields, type mismatches
- Same for response schemas where possible

- [x] Add request body schema resolution (follow `$ref` chains)
- [x] Add Go struct field extraction (parse JSON tags from request/response structs)
- [x] Add field-level comparison (schema properties vs struct fields)
- [x] Run extended validation and save full report to `docs/reports/openapi-full-validation.md`
- [x] No tests needed (tooling script)

### Task 3: Fix gap — add `ContactsService.ListTags`
Add the missing `GET /contacts/{contact_id}/tags` operation.

- [x] Add `ListTags` method + `ListTagsRaw` to `contacts.go` following 3-layer pattern
- [x] Add `ParseListTagsResult` function
- [x] Add `// See:` doc URL comments
- [x] Write tests for `ListTags` in `contacts_test.go` (success case)
- [x] Write test for error/edge case
- [x] Run tests — must pass before next task

### Task 4: Fix gap — add `GET /export/workflows/{id}` support
Add the missing exportWorkflow operation. Decide service placement (new WorkflowsService or add to existing).

- [x] Evaluate: new `WorkflowsService` vs adding to `DataExportService` — choose based on Intercom API tag ("Workflows")
- [x] Implement the method + Raw variant + Parse function following 3-layer pattern
- [x] Register service in `intercom.go` if new service
- [x] Add `// See:` doc URL comments
- [x] Write tests for the new method (success + error cases)
- [x] Run tests — must pass before next task

### Task 5: Fix any additional gaps found by validation script
Address all issues found in Task 1-2 reports that weren't caught in preliminary scan.

- [x] Review validation reports from Tasks 1-2
- [x] For each missing operation: implement method + Raw + Parse + tests
- [x] For each missing struct field: add field with correct JSON tag + omitempty
- [x] For each type mismatch: fix Go type to match schema
- [x] Write tests for all new/modified code
- [x] Run tests — must pass before next task

### Task 6: Verify bidirectional completeness
Ensure no SDK methods exist without a schema counterpart (excluding legitimate convenience wrappers like TagCompany/UntagCompany).

- [x] Run validation script in "SDK → schema" direction
- [x] Review any SDK methods not mapped to schema operations
- [x] Document legitimate extras (convenience wrappers, ListAll iterators) vs actual mismatches
- [x] Fix or remove any truly orphaned SDK methods (none found — all extras are legitimate)
- [x] Run tests — must pass before next task

### Task 7: Verify acceptance criteria
- [x] Re-run full validation script — zero missing operations
- [x] Re-run struct field validation — zero missing required fields
- [x] Run full test suite (`go test ./...`)
- [x] Run linter (`go vet ./...`)
- [x] Run format check (`gofmt -l .`)
- [x] Verify all report files are up to date

### Task 8: [Final] Update documentation
- [x] Update CLAUDE.md if new services/patterns were added
- [x] Archive validation reports in `docs/reports/`

## Technical Details

### Operation-to-SDK Mapping (preliminary, 162 operations)

**Admins (5 ops)**: All present ✓
- `GET /admins` → AdminsService.List
- `GET /admins/activity_logs` → AdminsService.ListActivityLogs
- `GET /admins/{admin_id}` → AdminsService.Get
- `PUT /admins/{admin_id}/away` → AdminsService.SetAway
- `GET /me` → AdminsService.Me

**AI Content (10 ops)**: All present ✓
- 5 ContentImportSource operations → AIContentService.{List,Get,Create,Update,Delete}ContentImportSource
- 5 ExternalPage operations → AIContentService.{List,Get,Create,Update,Delete}ExternalPage

**Articles (6 ops)**: All present ✓

**Away Status Reasons (1 op)**: Present ✓

**Brands (2 ops)**: All present ✓

**Calls (5 ops)**: All present ✓

**Companies (11 ops)**: All present ✓
- GET /companies (retrieveCompany with query) → CompaniesService.CompanyList
- POST /companies/list → CompaniesService.List
- Plus Get, Create, Update, Delete, Scroll, ListContacts, ListSegments, ListNotes

**Contacts (23 ops)**: **1 MISSING** ❌
- ❌ `GET /contacts/{contact_id}/tags` (listTagsForAContact) — no ListTags method
- All other 22 operations present ✓

**Conversations (18 ops)**: All present ✓
- `POST /conversations/{id}/parts` (manageConversation) → split into Close, Open, Snooze, Assign

**Custom Channel Events (4 ops)**: All present ✓

**Custom Objects (5 ops)**: All present ✓

**Data Attributes (3 ops)**: All present ✓

**Data Events (3 ops)**: All present ✓

**Data Export (4 ops)**: All present ✓

**Emails (2 ops)**: All present ✓

**Export Reporting (4 ops)**: All present ✓

**Fin Voice (5 ops)**: All present ✓

**Help Center (7 ops)**: All present ✓

**Internal Articles (6 ops)**: All present ✓

**IP Allowlist (2 ops)**: All present ✓

**Jobs (1 op)**: Present ✓

**Messages (1 op)**: Present ✓

**News (8 ops)**: All present ✓

**Notes (1 op)**: Present ✓

**Phone Call Redirects (1 op)**: Present ✓

**Segments (2 ops)**: All present ✓

**Subscription Types (1 op)**: Present ✓

**Tags (4 ops)**: All present ✓

**Teams (2 ops)**: All present ✓

**Ticket States (1 op)**: Present ✓

**Ticket Types (6 ops)**: All present ✓

**Tickets (9 ops)**: All present ✓

**Visitors (3 ops)**: All present ✓

**Workflows (1 op)**: **MISSING** ❌
- ❌ `GET /export/workflows/{id}` (exportWorkflow) — no Workflows support

### SDK-only methods (not directly in schema but valid):
- `TagsService.TagCompany` / `UntagCompany` (+ Raw variants) — convenience wrappers over POST /tags (oneOf body)
- `*Service.ListAll` methods (5) — iterator wrappers: Articles, Companies, Contacts, Conversations, InternalArticles
- `ConversationsService.Close/Open/Snooze/Assign` (+ Raw variants) + `managePartsRaw` — SDK splits the single `POST /conversations/{id}/parts` (manageConversation) endpoint by action type for better ergonomics

## Post-Completion

**Manual verification**:
- Test against live Intercom API sandbox (if available) for new methods
- Verify new methods handle rate limiting correctly
