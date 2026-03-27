# Raw Result Methods — Expose HTTP Data from Service Methods

## Overview
- Add `*Raw` companion methods to all service methods, returning a `Result[T]` that exposes HTTP metadata (status code, headers, raw body, URL) alongside typed response data
- Existing methods remain unchanged — zero breaking changes
- API errors (4xx/5xx) are represented as typed `ErrorResult` in `Result.Error` instead of Go errors; only transport/network failures return Go errors
- Motivation: users can't currently inspect HTTP details (rate-limit headers, status codes, raw bodies) — this limits debugging and advanced use cases like custom retry logic

## Context (from discovery)
- Files/components involved: `intercom.go` (Client.Do, Response), `errors.go` (ErrorResponse, checkResponse), all 33 service files
- Related patterns: `NewRequest → Do → decode` in every service method; `Response` wraps `*http.Response` but is discarded by all service methods
- Dependencies: no external deps (stdlib only constraint must be preserved)

## Development Approach
- **Testing approach**: TDD (tests first)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility — all existing methods and signatures unchanged

## Testing Strategy
- **Unit tests**: required for every task
- Tests use existing `httptest.Server` pattern via `setup()` helper in `testutil_test.go`
- Test both success and error paths for Raw methods
- Verify HTTP metadata is correctly populated in Result

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix
- Update plan if implementation deviates from original scope

## Special Cases

The following return types need specific treatment:

| Return Type | Count | Raw Treatment |
|---|---|---|
| `(*T, error)` | ~170 | Standard: `(*Result[T], error)` |
| `error` only | 5 | `(*Result[Empty], error)` where `Empty` is `struct{}` |
| `io.ReadCloser` | 2 (Download) | Skip — streaming body doesn't fit Result pattern |
| `*Iter[T]` | ~8 (ListAll) | Skip — wraps List, not direct HTTP; users can use ListRaw instead |
| `string` | 1 (GetRecordingURL) | Custom: extract string from raw body, still return Result |

## Implementation Steps

### Task 1: Define core types — Result[T], ErrorResult, HTTPResult
- [x] write tests for `Result[T]` struct field access (status code, headers, URL, body, data, error)
- [x] write tests for `ErrorResult` struct field access and Error() method
- [x] define `HTTPResult` struct in `result.go`: `StatusCode int`, `Header http.Header`, `URL string`, `Body []byte`
- [x] define `ErrorResult` struct in `result.go`: `Type string`, `RequestID string`, `Code string`, `Message string`, `Errors []ErrorDetail`; implement `Error() string`
- [x] define `Result[T any]` struct in `result.go`: embed `HTTPResult`, `Data *T`, `Error *ErrorResult`
- [x] run tests — must pass before next task

### Task 2: Add DoRaw method to Client
- [x] write tests for `Client.DoRaw` — success case (2xx, JSON body decoded into Result.Data)
- [x] write tests for `Client.DoRaw` — error case (4xx/5xx populates Result.Error, no Go error returned)
- [x] write tests for `Client.DoRaw` — transport error (network failure returns Go error)
- [x] write tests for `Client.DoRaw` — 204 No Content (Result.Data is zero value, no unmarshal)
- [x] implement `DoRaw[T any](ctx context.Context, req *http.Request) (*Result[T], error)` in `intercom.go`
  - Execute request, read full body into `[]byte`
  - Populate HTTPResult (StatusCode, Header, URL, Body)
  - On 4xx/5xx: unmarshal error body into ErrorResult, set Result.Error, return nil Go error
  - On 2xx: unmarshal body into T, set Result.Data
  - Only return Go error for transport/IO failures
- [x] run tests — must pass before next task

### Task 3: Add Raw methods to ContactsService (template service)
- [x] write tests for `ContactsService.GetRaw` — success returns Result with Data and HTTP metadata
- [x] write tests for `ContactsService.GetRaw` — 404 returns Result with Error populated, nil Go error
- [x] write tests for `ContactsService.ListRaw` — success returns Result with PagedResult data
- [x] write tests for `ContactsService.CreateRaw`, `UpdateRaw`, `DeleteRaw`, `SearchRaw`
- [x] write tests for `ContactsService.MergeRaw`, `ArchiveRaw`, `UnarchiveRaw`, `BlockRaw`
- [x] write tests for `ContactsService.FindByExternalIDRaw`
- [x] write tests for sub-resource Raw methods: `ListCompaniesRaw`, `AddCompanyRaw`, `RemoveCompanyRaw`, `ListNotesRaw`, `CreateNoteRaw`, `ListSegmentsRaw`, `ListSubscriptionsRaw`, `AddSubscriptionRaw`, `RemoveSubscriptionRaw`, `AddTagRaw`, `RemoveTagRaw`
- [x] implement all Raw companion methods in `contacts.go` following pattern: `NewRequest → DoRaw[T]`
- [x] verify existing (non-Raw) ContactsService tests still pass unchanged
- [x] run tests — must pass before next task

### Task 4: Add Raw methods to ConversationsService
- [x] write tests for key Raw methods: `GetRaw`, `ListRaw`, `CreateRaw`, `UpdateRaw`, `DeleteRaw`, `SearchRaw`, `ReplyRaw`, `CloseRaw`, `OpenRaw`, `SnoozeRaw`, `AssignRaw`, `ConvertRaw`, `RedactRaw`
- [x] write tests for sub-resource Raw methods: `AddCustomerRaw`, `RemoveCustomerRaw`, `AddTagRaw`, `RemoveTagRaw`, `RunAssignmentRulesRaw`
- [x] implement all Raw companion methods in `conversations.go`
- [x] run tests — must pass before next task

### Task 5: Add Raw methods to TicketsService
- [x] write tests for Raw methods: `GetRaw`, `CreateRaw`, `UpdateRaw`, `DeleteRaw`, `SearchRaw`, `ReplyRaw`, `EnqueueRaw`
- [x] write tests for sub-resource Raw methods: `AddTagRaw`, `RemoveTagRaw`
- [x] implement all Raw companion methods in `tickets.go`
- [x] run tests — must pass before next task

### Task 6: Add Raw methods to CompaniesService
- [x] write tests for Raw methods: `GetRaw`, `ListRaw`, `CompanyListRaw`, `CreateRaw`, `UpdateRaw`, `DeleteRaw`, `ScrollRaw`
- [x] write tests for sub-resource Raw methods: `ListContactsRaw`, `ListSegmentsRaw`, `ListNotesRaw`
- [x] implement all Raw companion methods in `companies.go`
- [x] run tests — must pass before next task

### Task 7: Add Raw methods to ArticlesService
- [x] write tests for Raw methods: `GetRaw`, `ListRaw`, `CreateRaw`, `UpdateRaw`, `DeleteRaw`, `SearchRaw`
- [x] implement all Raw companion methods in `articles.go`
- [x] run tests — must pass before next task

### Task 8: Add Raw methods to AdminsService
- [x] write tests for Raw methods: `MeRaw`, `GetRaw`, `ListRaw`, `SetAwayRaw`, `ListActivityLogsRaw`
- [x] implement all Raw companion methods in `admins.go`
- [x] run tests — must pass before next task

### Task 9: Add Raw methods to AIContentService
- [x] write tests for Raw methods: `ListContentImportSourcesRaw`, `GetContentImportSourceRaw`, `CreateContentImportSourceRaw`, `UpdateContentImportSourceRaw`, `DeleteContentImportSourceRaw`
- [x] write tests for Raw methods: `ListExternalPagesRaw`, `GetExternalPageRaw`, `CreateExternalPageRaw`, `UpdateExternalPageRaw`, `DeleteExternalPageRaw`
- [x] note: `ListContentImportSourcesRaw` and `ListExternalPagesRaw` return `*Iter[T]` currently — these return iterators, add Raw to the underlying fetcher or skip
- [x] implement all applicable Raw companion methods in `ai_content.go`
- [x] run tests — must pass before next task

### Task 10: Add Raw methods to remaining services (batch 1 — simpler services)
Services: AwayStatusReasons, Brands, Calls, CustomChannelEvents, CustomObjects, DataAttributes

- [x] write tests for `AwayStatusReasonsService.ListRaw`
- [x] write tests for `BrandsService.GetRaw`, `ListRaw`
- [x] write tests for `CallsService.GetRaw`, `ListRaw`, `SearchRaw`, `GetRecordingURLRaw`, `GetTranscriptRaw`
- [x] write tests for `CustomChannelEventsService` Raw methods (4 methods)
- [x] write tests for `CustomObjectsService` Raw methods (5 methods)
- [x] write tests for `DataAttributesService` Raw methods (3 methods)
- [x] implement all Raw companion methods for these services
- [x] run tests — must pass before next task

### Task 11: Add Raw methods to remaining services (batch 2 — data/export services)
Services: DataEvents, DataExport, ExportReporting, Jobs, Messages

- [x] write tests for `DataEventsService.CreateRaw`, `ListRaw`, `CreateSummariesRaw`
- [x] write tests for `DataExportService.CreateRaw`, `GetStatusRaw`, `CancelRaw` (skip Download — streaming)
- [x] write tests for `ExportReportingService.EnqueueRaw`, `GetStatusRaw`, `GetDatasetsRaw` (skip Download — streaming)
- [x] write tests for `JobsService.GetStatusRaw`
- [x] write tests for `MessagesService.CreateRaw`
- [x] implement all Raw companion methods for these services
- [x] run tests — must pass before next task

### Task 12: Add Raw methods to remaining services (batch 3 — content/config services)
Services: Emails, HelpCenter, InternalArticles, IPAllowlist, News, Notes, PhoneCallRedirects

- [x] write tests for `EmailsService.GetRaw`, `ListRaw`
- [x] write tests for `HelpCenterService` Raw methods (7 methods)
- [x] write tests for `InternalArticlesService` Raw methods (6 methods, skip ListAll)
- [x] write tests for `IPAllowlistService.GetRaw`, `UpdateRaw`
- [x] write tests for `NewsService` Raw methods (8 methods)
- [x] write tests for `NotesService.GetRaw`
- [x] write tests for `PhoneCallRedirectsService.CreateRaw`
- [x] implement all Raw companion methods for these services
- [x] run tests — must pass before next task

### Task 13: Add Raw methods to remaining services (batch 4 — reference/config services)
Services: Segments, SubscriptionTypes, Tags, Teams, TicketStates, TicketTypes, Visitors, FinVoice

- [x] write tests for `SegmentsService.GetRaw`, `ListRaw`
- [x] write tests for `SubscriptionTypesService.ListRaw`
- [x] write tests for `TagsService` Raw methods (5 methods, skip Delete if error-only or use Result[Empty])
- [x] write tests for `TeamsService.GetRaw`, `ListRaw`
- [x] write tests for `TicketStatesService.ListRaw`
- [x] write tests for `TicketTypesService` Raw methods (6 methods)
- [x] write tests for `VisitorsService` Raw methods (3 methods)
- [x] write tests for `FinVoiceService` Raw methods (5 methods)
- [x] implement all Raw companion methods for these services
- [x] run tests — must pass before next task

### Task 14: Verify acceptance criteria
- [x] verify all services have Raw companion methods (except Download/ListAll exclusions)
- [x] verify existing tests pass unchanged (backward compatibility)
- [x] verify Result.Error is populated on 4xx/5xx with correct ErrorResult fields
- [x] verify Result.Data is populated on 2xx success
- [x] verify HTTPResult fields (StatusCode, Header, URL, Body) are always populated
- [x] run full test suite (`go test ./...`)
- [x] run linter (`go vet ./...`)

### Task 15: [Final] Update documentation
- [x] add package-level example showing Raw method usage in doc.go or example_test.go
- [x] update README.md with Raw methods section

## Technical Details

### New Types (in `result.go`)

```go
// HTTPResult contains raw HTTP response metadata.
type HTTPResult struct {
    StatusCode int
    Header     http.Header
    URL        string
    Body       []byte
}

// ErrorResult represents a typed API error response.
type ErrorResult struct {
    Type      string        `json:"type"`
    RequestID string        `json:"request_id,omitempty"`
    Code      string        `json:"-"` // populated from Errors[0].Code
    Message   string        `json:"-"` // populated from Errors[0].Message
    Errors    []ErrorDetail `json:"errors"`
}

func (e *ErrorResult) Error() string { ... }

// Result wraps an HTTP response with typed data and error.
type Result[T any] struct {
    HTTPResult
    Data  *T
    Error *ErrorResult
}
```

### DoRaw Flow

```
Client.DoRaw[T](ctx, req)
  ├─ httpClient.Do(req)
  ├─ read full body → []byte
  ├─ populate HTTPResult{StatusCode, Header, URL, Body}
  ├─ if StatusCode >= 400:
  │    ├─ json.Unmarshal(body, &ErrorResult)
  │    ├─ set Result.Error
  │    └─ return &Result, nil  ← no Go error
  ├─ if StatusCode == 204:
  │    └─ return &Result{Data: &T{}}, nil
  └─ else:
       ├─ json.Unmarshal(body, &T)
       ├─ set Result.Data
       └─ return &Result, nil
```

### Raw Method Pattern

```go
func (s *ContactsService) GetRaw(ctx context.Context, id string) (*Result[Contact], error) {
    req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("contacts/%s", url.PathEscape(id)), nil)
    if err != nil {
        return nil, err
    }
    return DoRaw[Contact](ctx, s.client, req)
}
```

Note: `DoRaw` is a package-level generic function (not a method) because Go doesn't support generic methods on non-generic types. Signature: `func DoRaw[T any](ctx context.Context, c *Client, req *http.Request) (*Result[T], error)`

## Post-Completion

**Manual verification:**
- Test with real Intercom API to verify HTTP metadata is correctly populated
- Verify rate-limit headers (X-RateLimit-Limit, X-RateLimit-Remaining) are accessible via Result.Header
- Check that Result.Body contains the exact raw JSON for debugging purposes
