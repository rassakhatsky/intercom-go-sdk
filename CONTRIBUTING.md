# Contributing

## Dev Setup

1. Clone the repo and `cd` into it
2. Ensure Go 1.26.1+ is installed
3. Run `go mod download` to fetch dependencies

## Running Tests

```bash
make test        # go test -cover ./...
make test-race   # go test -race -cover ./...
make fix         # go vet + gofmt -l

# Single sub-package
go test ./contacts/...

# Single test
go test -run TestService_Get ./contacts/...
```

All tests use the standard `testing` package — no external frameworks.

## Project Structure

Services live in 17 domain-scoped sub-packages (`contacts/`, `tickets/`, `tags/`, etc.). Core types and the HTTP client live in the root `intercom` package. Shared types used across sub-packages live in `api/`.

Sub-packages never import each other. If a type is needed by multiple packages, it belongs in `api/`.

See [CLAUDE.md](CLAUDE.md) for full architecture details, package grouping, and the complete list of conventions.

## Coding Conventions

**3-layer method pattern** — every API method follows Raw -> Parse -> Method:

1. `XxxRaw` — calls `NewRequest` -> `DoRaw`, returns `(*api.Result, error)`
2. `ParseXxxResult` — decodes `Result.Body` into a typed struct
3. `Xxx` — combines Raw + error check + Parse into one call

**Naming:**

- Service types: `Service` (or `XxxService` for multi-service packages)
- Request structs: `CreateRequest`, `UpdateRequest` (scoped to the sub-package)
- JSON tags use `omitempty` except for required fields
- Every public method includes a `// See:` comment with the Intercom API reference URL

**Other:**

- Every method takes `context.Context` as its first parameter
- URL path parameters use `url.PathEscape`
- Timestamps are `int64` (Unix epoch); optional timestamps are `*int64`

## Adding a New Service

1. Create the service file in the appropriate sub-package
2. Register it in `intercom.go` (field, init, accessor)
3. Add tests using the sub-package's `setup()` helper
4. Add `// See:` doc URL comments to every public method

For a new sub-package, also create `testutil_test.go` with `setup()`, `testMethod`, and `testHeader` helpers.

See [CLAUDE.md](CLAUDE.md) for the detailed step-by-step guide.

## Pull Requests

- Keep changes focused — one feature or fix per PR
- All tests must pass (`make test`)
- Code must be formatted (`gofmt`) and pass `go vet`
- Include tests for new functionality
