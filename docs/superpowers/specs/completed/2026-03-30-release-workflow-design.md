# Release Workflow Design

## Summary

Add a GitHub Actions workflow that creates versioned releases via `workflow_dispatch`. Only CODEOWNERS can trigger it. Supports `patch` and `minor` bumps. First release will be `v1.0.0`.

## Trigger

- `workflow_dispatch` with a required `bump` input: `patch` or `minor`
- No major bump option — Go module v2+ requires module path changes and is always a manual process

## Authorization

- Workflow reads `.github/CODEOWNERS`, extracts GitHub usernames (strips `@` prefix)
- Compares against `github.actor`
- Fails early with a clear error message if the actor is not a CODEOWNER

## Version Calculation

1. Fetch the latest semver tag matching `v*.*.*` (sorted by version, not date)
2. If no tags exist, default to `v0.0.0`
3. Apply bump:
   - `patch`: `v1.0.0` → `v1.0.1`
   - `minor`: `v1.0.0` → `v1.1.0`
4. First release: since baseline is `v0.0.0`, a `minor` bump → `v1.0.0`

## Workflow Steps

1. Checkout repository (with full history for tag discovery)
2. Validate actor is a CODEOWNER
3. Determine latest tag
4. Calculate next version
5. Create and push git tag
6. Create GitHub Release using `gh release create` with `--generate-notes`

## GitHub Release

- Title: the tag name (e.g., `v1.0.0`)
- Body: auto-generated release notes (GitHub lists PRs/commits since previous tag)
- No binary attachments (this is a Go library, not a binary)

## Constraints

- Workflow runs against `main` branch only (default for workflow_dispatch)
- No goreleaser, no CHANGELOG.md, no build artifacts
- No `v0.x.x` releases — first release is `v1.0.0`

## Files Changed

- `.github/workflows/release.yml` — new workflow file
