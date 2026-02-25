# Tasks: Orphan Markdown Checker CLI (Go)

This task list is derived from [plan.md](./plan.md) and is organized for implementation order.

## Phase 0: Project Bootstrap
- [x] Initialize Go module (`go mod init`).
- [x] Create baseline project structure:
  - [x] `cmd/gorphan/main.go`
  - [x] `internal/scanner/scanner.go`
  - [x] `internal/parser/parser.go`
  - [x] `internal/graph/graph.go`
  - [x] `internal/report/report.go`
  - [x] `testdata/`
- [x] Add basic `README.md` with temporary usage placeholder.

## Phase 1: CLI Skeleton and Validation
- [x] Implement CLI command `gorphan`.
- [x] Add required flags:
  - [x] `--root`
  - [x] `--dir`
- [x] Add optional flags:
  - [x] `--ext` (default: `.md,.markdown`)
  - [x] `--ignore` (repeatable)
  - [x] `--format` (`text` default, plus `json`)
  - [x] `--verbose`
- [x] Validate inputs:
  - [x] Root file exists.
  - [x] Root path points to a file.
  - [x] Root is under scan directory.
  - [x] Scan directory exists.

## Phase 2: File Scanner
- [x] Recursively walk scan directory.
- [x] Filter files by configured markdown extensions.
- [x] Apply ignore rules for paths/patterns.
- [x] Normalize and store canonical paths for all markdown files in scope.
- [x] Return deterministic file ordering for stable output.

## Phase 3: Markdown Link Parser
- [x] Implement inline markdown link extraction (v1 mandatory).
- [x] If feasible, add reference-style link support.
- [x] Ignore non-local links:
  - [x] `http://`
  - [x] `https://`
  - [x] `mailto:`
- [x] Normalize parsed link targets:
  - [x] Strip fragment (`#...`)
  - [x] Strip query (`?...`)
  - [x] Decode basic escaped paths if needed
- [x] Define v1 behavior for no-extension targets (ignore by default).

## Phase 4: Graph Builder
- [x] Resolve each local link against the linking file directory.
- [x] Canonicalize resolved paths (`Clean`, absolute form).
- [x] Keep edges only when target:
  - [x] Is within scan directory.
  - [x] Matches markdown extension set.
  - [x] Exists in scanned markdown inventory.
- [x] Build adjacency map `source -> []target`.

## Phase 5: Reachability and Orphan Detection
- [x] Traverse graph from root (DFS or BFS).
- [x] Build reachable set.
- [x] Compute orphan list as: `all_markdown_files - reachable`.
- [x] Sort orphan paths lexicographically.
- [x] Convert output paths to relative paths from scan directory.

## Phase 6: Output, Exit Codes, and UX
- [x] Text formatter:
  - [x] Human-readable orphan list.
  - [x] Optional verbose summary (scanned, reachable, orphan counts).
- [x] JSON formatter:
  - [x] Structured output for automation/CI.
- [x] Exit code behavior:
  - [x] `0` when no orphans.
  - [x] `1` when orphans exist.
  - [x] `>1` on usage/runtime errors.
- [x] Warning strategy:
  - [x] Warn on unresolvable local links (do not fail v1).

## Phase 7: Tests
- [x] Unit tests:
  - [x] Link extraction edge cases.
  - [x] Path resolution/normalization.
  - [x] Ignore rule matching.
  - [x] Reachability traversal.
- [x] Integration tests using fixtures:
  - [x] No orphan scenario.
  - [x] Multiple orphan files.
  - [x] Cyclic links.
  - [x] Disconnected components.
  - [x] Invalid root validation.
- [x] Golden tests for text/json CLI outputs.
- [x] Ensure `go test ./...` passes.

## Phase 8: Documentation and Release Readiness
- [x] Finalize `README.md`:
  - [x] Problem definition and orphan semantics.
  - [x] Install/run instructions.
  - [x] Usage examples.
  - [x] Output format and exit codes.
  - [x] Known limitations (v1 scope).
- [x] Add sample command:
  - [x] `gorphan --root docs/index.md --dir docs`
- [x] Document future enhancements as non-goals for v1.

## Optional Backlog (Post-v1)
- [ ] Wikilink support (`[[Page]]`).
- [ ] Unresolved-link reporting mode.
- [ ] Graph export (dot/mermaid).
- [ ] Parallel parsing for large repositories.
- [ ] Config file support (`.gorphan.yaml`).

## Phase 9: CI Quality Gates
- [x] Add formatter check in GitHub Actions using `gofmt -l .`.
- [x] Fail CI when formatter output is non-empty.
- [x] Add linter check in GitHub Actions using `golangci-lint`.
- [x] Configure `golangci/golangci-lint-action` with a pinned version.
- [x] Keep/verify `go test ./...` after lint in CI.
- [x] Ensure CI runs on `push` to `main` and all `pull_request` events.
- [x] Enforce gate result: CI passes only if format, lint, and tests all pass.
- [x] Enforce failure behavior: any format/lint/test failure fails the workflow.
- [x] Document local pre-PR sequence: `gofmt`, `golangci-lint run`, `go test ./...`.
