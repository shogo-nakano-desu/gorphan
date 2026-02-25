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
- [ ] Implement CLI command `gorphan`.
- [ ] Add required flags:
  - [ ] `--root`
  - [ ] `--dir`
- [ ] Add optional flags:
  - [ ] `--ext` (default: `.md,.markdown`)
  - [ ] `--ignore` (repeatable)
  - [ ] `--format` (`text` default, plus `json`)
  - [ ] `--verbose`
- [ ] Validate inputs:
  - [ ] Root file exists.
  - [ ] Root path points to a file.
  - [ ] Root is under scan directory.
  - [ ] Scan directory exists.

## Phase 2: File Scanner
- [ ] Recursively walk scan directory.
- [ ] Filter files by configured markdown extensions.
- [ ] Apply ignore rules for paths/patterns.
- [ ] Normalize and store canonical paths for all markdown files in scope.
- [ ] Return deterministic file ordering for stable output.

## Phase 3: Markdown Link Parser
- [ ] Implement inline markdown link extraction (v1 mandatory).
- [ ] If feasible, add reference-style link support.
- [ ] Ignore non-local links:
  - [ ] `http://`
  - [ ] `https://`
  - [ ] `mailto:`
- [ ] Normalize parsed link targets:
  - [ ] Strip fragment (`#...`)
  - [ ] Strip query (`?...`)
  - [ ] Decode basic escaped paths if needed
- [ ] Define v1 behavior for no-extension targets (ignore by default).

## Phase 4: Graph Builder
- [ ] Resolve each local link against the linking file directory.
- [ ] Canonicalize resolved paths (`Clean`, absolute form).
- [ ] Keep edges only when target:
  - [ ] Is within scan directory.
  - [ ] Matches markdown extension set.
  - [ ] Exists in scanned markdown inventory.
- [ ] Build adjacency map `source -> []target`.

## Phase 5: Reachability and Orphan Detection
- [ ] Traverse graph from root (DFS or BFS).
- [ ] Build reachable set.
- [ ] Compute orphan list as: `all_markdown_files - reachable`.
- [ ] Sort orphan paths lexicographically.
- [ ] Convert output paths to relative paths from scan directory.

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
