# Consolidate Duplicate Types Across Sub-Packages

## Overview
- Multiple sub-packages define identical or near-identical structs (Author, LinkedObjectList, Part, ContactItem, Deleted, TagRefList) causing maintenance risk
- Consolidate these into `internal/api/types.go` as the single source of truth
- Sub-packages will import the shared types; minor breakage is acceptable (consumers update imports)
- Superset approach for near-duplicates: unified struct includes all fields with `omitempty`

## Context (from discovery)
- **Exact duplicates:** `Author` (conversations, tickets), `LinkedObjectList` (conversations, tickets)
- **Near-duplicates:** `Part` (conversations has extra `NotifiedAt`), `Contact`/`ContactItem` (conversations, tickets — identical but different names), `Deleted` (4 identical in tickets/articles/conversations/companies; contacts has unique variant)
- **TagRefList:** conversations and companies define TagList with `[]api.TagRef` — different from existing `api.TagList` which uses `[]Tag` with `json:"data"`
- **Kept separate:** `PartList` (different JSON tags: `conversation_parts` vs `ticket_parts`), `ReplyRequest` (too many field differences), `Message` (different contexts)
- **Existing shared types in `internal/api/types.go`:** TagRef, AdminRef, ContactRef, NoteAuthor, Note, SegmentRef, Tag, TagList, etc.

## Development Approach
- **Testing approach**: TDD — write tests that import the new shared types first, then move the types
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
  - tests are not optional - they are a required part of the checklist
  - write unit tests for new functions/methods
  - update existing test cases if behavior changes
  - tests cover both success and error scenarios
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility within each sub-package's public API where possible

## Testing Strategy
- **Unit tests**: existing tests in each sub-package already exercise JSON marshal/unmarshal of these types; they must continue to pass after consolidation
- **Compilation check**: `go vet ./...` after each task to catch import/type errors early
- **New tests**: add type-level tests in `internal/api/` verifying JSON round-trip for each new shared type

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Add Author to internal/api/types.go
- [x] Write tests in `internal/api/types_test.go` for `Author` JSON marshal/unmarshal (success + edge cases)
- [x] Add `Author` struct to `internal/api/types.go` with fields: `Type`, `ID`, `Name`, `Email`
- [x] Update `conversations/conversations.go`: remove local `Author`, import `api.Author`
- [x] Update `tickets/tickets.go`: remove local `Author`, import `api.Author`
- [x] Update all references to `Author` in both packages (Part, ReplyRequest fields, etc.)
- [x] Run `go test ./...` — must pass before next task

### Task 2: Add LinkedObjectList to internal/api/types.go
- [x] Write tests in `internal/api/types_test.go` for `LinkedObjectList` JSON marshal/unmarshal
- [x] Add `LinkedObjectList` struct to `internal/api/types.go` with fields: `Type`, `Data []any`, `TotalCount`, `HasMore`
- [x] Update `conversations/conversations.go`: remove local `LinkedObjectList`, use `api.LinkedObjectList`
- [x] Update `tickets/tickets.go`: remove local `LinkedObjectList`, use `api.LinkedObjectList`
- [x] Run `go test ./...` — must pass before next task

### Task 3: Add Part (superset) to internal/api/types.go
- [x] Write tests in `internal/api/types_test.go` for `Part` JSON marshal/unmarshal — include `NotifiedAt` present and absent cases
- [x] Add `Part` struct to `internal/api/types.go` — superset of conversations + tickets fields: `Type`, `ID`, `PartType`, `Body`, `CreatedAt`, `UpdatedAt`, `NotifiedAt`, `AssignedTo *Author`, `Author *Author`, `ExternalID`, `Redacted`
- [x] Update `conversations/conversations.go`: remove local `Part`, use `api.Part`
- [x] Update `tickets/tickets.go`: remove local `Part`, use `api.Part`
- [x] Update `PartList` structs in both packages to reference `[]api.Part` (keep PartList local — different JSON tags)
- [x] Run `go test ./...` — must pass before next task

### Task 4: Consolidate ContactItem into extended ContactRef
- [x] Write tests in `internal/api/types_test.go` for `ContactRef` with `ExternalID` field (backward compat: existing `ContactRef` usage without ExternalID must still work)
- [x] Add `ExternalID string` field to existing `api.ContactRef` in `internal/api/types.go`
- [x] Add `ContactRefList` struct to `internal/api/types.go` with fields: `Type`, `Contacts []ContactRef`
- [x] Update `conversations/conversations.go`: remove local `Contact` and `ContactList`, use `api.ContactRef` and `api.ContactRefList`
- [x] Update `tickets/tickets.go`: remove local `ContactItem` and `ContactList`, use `api.ContactRef` and `api.ContactRefList`
- [x] Verify existing `ContactRef` consumers (notes, contacts, companies) still compile and pass tests
- [x] Run `go test ./...` — must pass before next task

### Task 5: Consolidate Deleted (standard variant)
- [x] Write tests in `internal/api/types_test.go` for `Deleted` JSON marshal/unmarshal
- [x] Add `Deleted` struct to `internal/api/types.go` with fields: `ID`, `Object`, `Deleted` (the 4-way identical variant)
- [x] Update `tickets/tickets.go`: remove local `Deleted`, use `api.Deleted`
- [x] Update `articles/articles.go`: remove local `Deleted`, use `api.Deleted`
- [x] Update `conversations/conversations.go`: remove local `Deleted`, use `api.Deleted`
- [x] Update `companies/companies.go`: remove local `Deleted`, use `api.Deleted`
- [x] Keep `contacts/contacts.go` `Deleted` as-is (different structure with `Type` and `ExternalID`)
- [x] Run `go test ./...` — must pass before next task

### Task 6: Add TagRefList to internal/api/types.go
- [x] Write tests in `internal/api/types_test.go` for `TagRefList` JSON marshal/unmarshal (with `json:"tags"` key)
- [x] Add `TagRefList` struct to `internal/api/types.go` with fields: `Type string`, `Tags []TagRef` — distinct from existing `TagList` which uses `Data []Tag` with `json:"data"`
- [x] Update `conversations/conversations.go`: remove local `TagList`, use `api.TagRefList`
- [x] Update `companies/companies.go`: remove local `TagList`, use `api.TagRefList`
- [x] Run `go test ./...` — must pass before next task

### Task 7: Update aliases.go (root package re-exports)
- [x] Add type aliases in `aliases.go` for new shared types: `Author`, `LinkedObjectList`, `Part`, `ContactRefList`, `Deleted`, `TagRefList`
- [x] Verify the root package compiles: `go build ./...`
- [x] Run `go test ./...` — must pass before next task

### Task 8: Verify acceptance criteria
- [x] Verify all 7 types consolidated: Author, LinkedObjectList, Part, ContactRef (extended), Deleted, TagRefList, ContactRefList
- [x] Verify no duplicate type definitions remain (grep for each type name across sub-packages)
- [x] Verify edge cases: contacts.Deleted still has its unique variant, PartList still local with correct JSON tags
- [x] Run full test suite: `go test ./...`
- [x] Run linter: `go vet ./...`
- [x] Run format check: `gofmt -l .`

### Task 9: [Final] Update documentation
- [ ] Update CLAUDE.md to reflect new shared types in internal/api/types.go
- [ ] Add comments to internal/api/types.go documenting which services use each new type

## Technical Details

### Type definitions to add to `internal/api/types.go`

**Author** (exact duplicate from conversations + tickets):
```go
type Author struct {
    Type  string `json:"type"`
    ID    string `json:"id"`
    Name  string `json:"name,omitempty"`
    Email string `json:"email,omitempty"`
}
```

**LinkedObjectList** (exact duplicate from conversations + tickets):
```go
type LinkedObjectList struct {
    Type       string `json:"type"`
    Data       []any  `json:"data"`
    TotalCount int    `json:"total_count"`
    HasMore    bool   `json:"has_more"`
}
```

**Part** (superset — conversations adds `NotifiedAt`):
```go
type Part struct {
    Type       string  `json:"type"`
    ID         string  `json:"id"`
    PartType   string  `json:"part_type,omitempty"`
    Body       string  `json:"body,omitempty"`
    CreatedAt  int64   `json:"created_at,omitempty"`
    UpdatedAt  int64   `json:"updated_at,omitempty"`
    NotifiedAt int64   `json:"notified_at,omitempty"`
    AssignedTo *Author `json:"assigned_to,omitempty"`
    Author     *Author `json:"author,omitempty"`
    ExternalID string  `json:"external_id,omitempty"`
    Redacted   bool    `json:"redacted,omitempty"`
}
```

**ContactRef** (extend existing — add `ExternalID`):
```go
type ContactRef struct {
    Type       string `json:"type"`
    ID         string `json:"id"`
    ExternalID string `json:"external_id,omitempty"`
}
```

**ContactRefList** (new — from conversations.ContactList + tickets.ContactList):
```go
type ContactRefList struct {
    Type     string       `json:"type"`
    Contacts []ContactRef `json:"contacts"`
}
```

**Deleted** (standard 4-way duplicate):
```go
type Deleted struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Deleted bool   `json:"deleted"`
}
```

**TagRefList** (new — distinct from existing TagList):
```go
type TagRefList struct {
    Type string   `json:"type"`
    Tags []TagRef `json:"tags"`
}
```

### Types kept separate (with rationale)
- **PartList**: different JSON keys (`conversation_parts` vs `ticket_parts`) — consolidation requires custom marshaling
- **ReplyRequest**: conversations has `AttachmentFiles` and `SkipNotifications` that tickets doesn't; different `ReplyOption` types
- **Message**: different contexts (conversation message vs outbound message), `Subject` field difference
- **contacts.Deleted**: unique structure with `Type` and `ExternalID` fields instead of `Object`

## Post-Completion

**Manual verification:**
- Review that godoc output for `internal/api` is clean and well-organized
- Verify that any downstream consumers of this SDK (if any) can update imports cleanly
