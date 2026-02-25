# Contributing

## Workflow
- Create focused changes and keep commits scoped.
- Add or update tests for new features and bug fixes.
- Enable repository hooks once per clone:
  - `git config core.hooksPath .githooks`
- Run local quality checks before opening a PR:
  - `gofmt -w .`
  - `golangci-lint run`
  - `go test ./...`

## Development Notes
- The CLI entrypoint is `cmd/gorphan/main.go`.
- Core logic lives under `internal/` packages.
- Keep markdown links valid so orphan analysis can run without warnings.
