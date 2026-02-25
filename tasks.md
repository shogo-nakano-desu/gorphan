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
- [ ] Text formatter:
  - [ ] Human-readable orphan list.
  - [ ] Optional verbose summary (scanned, reachable, orphan counts).
- [ ] JSON formatter:
  - [ ] Structured output for automation/CI.
- [ ] Exit code behavior:
  - [ ] `0` when no orphans.
  - [ ] `1` when orphans exist.
  - [ ] `>1` on usage/runtime errors.
- [ ] Warning strategy:
  - [ ] Warn on unresolvable local links (do not fail v1).

## Phase 7: Tests
- [ ] Unit tests:
  - [ ] Link extraction edge cases.
  - [ ] Path resolution/normalization.
  - [ ] Ignore rule matching.
  - [ ] Reachability traversal.
- [ ] Integration tests using fixtures:
  - [ ] No orphan scenario.
  - [ ] Multiple orphan files.
  - [ ] Cyclic links.
  - [ ] Disconnected components.
  - [ ] Invalid root validation.
- [ ] Golden tests for text/json CLI outputs.
- [ ] Ensure `go test ./...` passes.

## Phase 8: Documentation and Release Readiness
- [ ] Finalize `README.md`:
  - [ ] Problem definition and orphan semantics.
  - [ ] Install/run instructions.
  - [ ] Usage examples.
  - [ ] Output format and exit codes.
  - [ ] Known limitations (v1 scope).
- [ ] Add sample command:
  - [ ] `gorphan --root docs/index.md --dir docs`
- [ ] Document future enhancements as non-goals for v1.

## Optional Backlog (Post-v1)
- [ ] Wikilink support (`[[Page]]`).
- [ ] Unresolved-link reporting mode.
- [ ] Graph export (dot/mermaid).
- [ ] Parallel parsing for large repositories.
- [ ] Config file support (`.gorphan.yaml`).
