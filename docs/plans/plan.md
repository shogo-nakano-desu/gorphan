# Plan: Orphan Markdown Checker CLI (Go)

## 1. Goal and Scope
- Build a lightweight CLI tool in Go to detect orphan Markdown files under a target directory.
- "Orphan" means a Markdown file is not reachable by links from other Markdown files, starting from one required root Markdown file.
- Keep scope focused on local file links only (no HTTP reachability checks).

## 2. Functional Requirements
- Accept a required root Markdown file path (entry point).
- Accept a target directory to scan recursively for Markdown files.
- Parse Markdown links from files and resolve local linked Markdown targets.
- Build a directed graph of Markdown files (source -> linked markdown target).
- Mark all files reachable from the root.
- Report Markdown files in scan scope that are not reachable (orphans).
- Exit codes:
  - `0`: no orphans.
  - `1`: orphans found.
  - `>1`: runtime/usage errors.

## 3. CLI Design
- Command name: `gorphan` (working name).
- Example usage:
  - `gorphan --root docs/index.md --dir docs`
- Flags:
  - `--root` (required): root markdown file.
  - `--dir` (required): directory to scan.
  - `--ext` (optional, default `.md,.markdown`): markdown extensions to include.
  - `--ignore` (optional, repeatable): glob patterns or path prefixes to exclude.
  - `--format` (optional, default `text`): output format (`text`, `json`).
  - `--verbose` (optional): include diagnostics (counts, skipped links).

## 4. Architecture
- `cmd/gorphan/main.go`
  - CLI arg parsing, input validation, invoking analysis, rendering output, exit codes.
- `internal/scanner`
  - Walk directory, collect markdown files, apply ignore rules.
- `internal/parser`
  - Extract markdown links from content.
  - Normalize link targets (strip anchors/query, decode simple escapes).
- `internal/graph`
  - Resolve relative links to filesystem paths.
  - Build adjacency map between markdown files in scope.
  - Traverse from root using DFS/BFS to compute reachable set.
- `internal/report`
  - Produce sorted orphan list and optional summary stats.

## 5. Link Handling Rules (Initial Version)
- Process inline and reference-style Markdown links if feasible; inline links are mandatory for v1.
- Only local links are considered for graph edges:
  - Relative paths (e.g. `./a.md`, `../guide/intro.md`).
  - Ignore external URLs (`http://`, `https://`, `mailto:`).
- Remove fragment identifiers (`#section`) and query strings before resolution.
- If link has no extension:
  - v1 default: ignore (or optionally attempt `.md` resolution behind a flag later).
- Resolve targets against the linking file's directory, then canonicalize (`filepath.Clean`, absolute path).
- Only create edges to files inside scan directory and matching markdown extensions.

## 6. Error Handling and UX
- Validate root exists, is a file, and is within scan directory.
- Warn (not fail) for unresolvable local links unless `--strict` is introduced later.
- Stable, deterministic output:
  - Sort orphan paths lexicographically.
  - Print paths relative to scan directory for readability.

## 7. Testing Strategy
- CI quality checks (GitHub Actions):
  - Format check: fail when `gofmt -l .` outputs any files.
  - Lint check: run `golangci-lint` via `golangci/golangci-lint-action`.
  - Test check: run `go test ./...`.
  - Execution order: format -> lint -> test (fast-fail).
- Unit tests:
  - Link extraction (anchors, external links, odd spacing, escaped chars).
  - Path resolution and normalization.
  - Graph traversal reachability.
  - Ignore pattern filtering.
- Integration tests (fixture directories):
  - No orphan case.
  - Multiple orphan files.
  - Cyclic links and disconnected components.
  - Root file validation failure.
- Golden tests for CLI output (`text` and `json`).

## 8. Implementation Milestones
- Milestone 1: Bootstrap Go module, CLI skeleton, argument parsing, and validation.
- Milestone 2: Directory scanner + markdown file inventory.
- Milestone 3: Link parser + path resolver + graph builder.
- Milestone 4: Reachability analysis + orphan detection output.
- Milestone 5: Tests (unit + integration) and CI quality gates (format, lint, test) via GitHub Actions workflow (`.github/workflows/ci.yml`).
  - CI commands: `test -z "$(gofmt -l .)"`, `golangci-lint run`, `go test ./...`.
- Milestone 6: README with usage examples and limitations.

## 9. Suggested Project Layout
- `go.mod`
- `cmd/gorphan/main.go`
- `internal/scanner/scanner.go`
- `internal/parser/parser.go`
- `internal/graph/graph.go`
- `internal/report/report.go`
- `testdata/` (fixtures)
- `README.md`

## 10. Future Enhancements (Out of Scope for v1)
- Support wikilinks (`[[Page]]`) and custom markdown dialects.
- Optional unresolved-link report.
- Dot/mermaid graph export.
- Parallel file parsing for large doc sets.
- Config file support (`.gorphan.yaml`).
