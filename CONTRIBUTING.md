# Contributing to gorphan

Thanks for contributing. This project aims to stay small, predictable, and high-signal.

## Ground Rules
- Keep changes focused. One concern per PR.
- Include tests for every feature or bug fix.
- Keep behavior deterministic (stable sort order, reproducible output).
- Do not silently change CLI semantics without documenting it.

## Development Setup

```bash
git clone https://github.com/shogo-nakano-desu/gorphan.git
cd gorphan
go test ./...
```

Enable hooks once per clone:

```bash
git config core.hooksPath .githooks
```

## Project Layout
- `cmd/gorphan`: CLI parsing and execution flow.
- `internal/scanner`: file discovery and ignore rules.
- `internal/parser`: markdown link extraction and normalization.
- `internal/graph`: graph build, analysis, and graph exports.
- `internal/report`: text/json rendering.
- `e2e/`: CLI end-to-end tests.
- `docs/`: architecture, testing, and planning docs.

## Making Changes
1. Create a focused branch.
2. Implement code and tests together.
3. Update docs when behavior, flags, or output changes.
4. Keep commit history clean and meaningful.

## Quality Checklist

Run before opening a PR:

```bash
gofmt -w .
golangci-lint run
go test ./...
```

Expected:
- No formatter diffs.
- No lint issues.
- All tests pass (unit, integration, golden, e2e).

## Pull Request Expectations
- Clear title and concise description.
- Explain user-visible behavior changes.
- Include test coverage for changed behavior.
- Include sample command/output when CLI behavior changes.

## Commit Style
Use direct, imperative messages:
- `Add mermaid graph export`
- `Fix parser escaped-space handling`
- `Make --dir optional with default current directory`

## Reporting Issues
When filing a bug, include:
- command used
- input markdown structure (minimal repro)
- expected behavior
- actual behavior
- environment (`go version`, OS)

## Communication
- Prefer concrete technical discussion over broad abstractions.
- If uncertain, open a draft PR early with assumptions listed.
