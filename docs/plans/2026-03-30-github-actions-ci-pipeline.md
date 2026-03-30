# GitHub Actions CI Pipeline

## Overview
- Add a GitHub Actions CI workflow for the intercom-go-sdk Go project
- Validates code quality (vet, format), dependency integrity, and correctness (tests with race detector)
- Triggers on pushes to `main` and pull requests targeting `main`

## Context (from discovery)
- **No existing CI** — `.github/workflows/` does not exist
- **Makefile** has `test`, `test-race`, and `fix` targets (can reuse concepts)
- **Go 1.26.1**, single dependency (`google/go-querystring`)
- **No linter config** — linter will be added separately later
- **17 sub-packages** with tests using stdlib `testing` only

## Development Approach
- **Testing approach**: Regular (verify workflow YAML is valid, then test by pushing)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: all checks must pass before marking task complete**

## CI Checks (in order)
1. `go vet ./...` — static analysis
2. `gofmt -l .` — format check (fail if any files need formatting)
3. `go mod verify` — dependency integrity
4. `go mod tidy` + `git diff --exit-code go.mod go.sum` — catch stale module files
5. `go build ./...` — compilation check
6. `go test -race -cover ./...` — tests with race detector and coverage

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with warning prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Create the GitHub Actions workflow file
- [x] create `.github/workflows/` directory
- [x] create `.github/workflows/ci.yml` with workflow name, triggers (push to main, PR to main)
- [x] add single job `ci` running on `ubuntu-latest` with Go 1.26.x
- [x] add step: checkout code (`actions/checkout@v4`)
- [x] add step: setup Go (`actions/setup-go@v5` with `go-version: '1.26'`)
- [x] add step: `go vet ./...`
- [x] add step: format check — `gofmt -l .` with failure on non-empty output
- [x] add step: `go mod verify`
- [x] add step: `go mod tidy` then `git diff --exit-code go.mod go.sum`
- [x] add step: `go build ./...`
- [x] add step: `go test -race -cover ./...`
- [x] verify YAML syntax is valid

### Task 2: Verify acceptance criteria
- [x] workflow triggers on push to main and PRs to main
- [x] all 6 checks are present and ordered correctly
- [x] Go version matches go.mod (1.26.x)
- [x] format check fails the build on unformatted code (not just warns)
- [x] mod tidy check fails on stale go.mod/go.sum
- [x] review final YAML for correctness

### Task 3: [Final] Update documentation
- [ ] update README.md with CI badge if desired
- [ ] note in CONTRIBUTING.md that CI runs these checks (if applicable)

## Technical Details
- **Runner**: `ubuntu-latest`
- **Go version**: `1.26` (patch version resolved automatically by `actions/setup-go`)
- **Caching**: `actions/setup-go@v5` handles module caching automatically
- **Format check**: `gofmt -l .` lists files needing formatting; pipe to check for non-empty output
- **Mod tidy check**: run `go mod tidy` then `git diff --exit-code go.mod go.sum` to detect drift

## Post-Completion
**Manual verification:**
- Push a branch or open a PR to confirm the workflow runs successfully
- Intentionally break formatting or a test to verify failure behavior
- Confirm GitHub shows the check status on PRs

**Future enhancements (not in scope):**
- Add golangci-lint step (user will add later)
- Add code coverage reporting/thresholds
- Add release automation
