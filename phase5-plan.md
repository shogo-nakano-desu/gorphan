# Phase 5 Plan: Reachability and Orphan Detection

## Summary
Implement Phase 5 by extending `internal/graph` with deterministic reachability/orphan analysis, then wiring it into `cmd/gorphan/main.go` for computation (not final formatting yet).

Chosen defaults:
- If `--root` is not in scanned inventory, fail fast with an error.
- Keep Phase 5 analysis logic in `internal/graph`.

## Scope (Phase 5 only)
- Traverse graph from root (DFS or BFS).
- Build reachable set.
- Compute orphan list (`all files - reachable`).
- Sort orphans lexicographically.
- Convert orphan paths to scan-dir-relative paths (slash-separated for stable output).

Out of scope for this phase:
- Final user-facing text/json rendering design (Phase 6).
- Exit code `1` for orphan presence (Phase 6).
- Unresolved-link warnings (Phase 6).

## API / Interface Changes
Add to `internal/graph`:

1. `type Analysis struct {`
   - `Reachable []string` (absolute, sorted)
   - `Orphans []string` (absolute, sorted)
   - `OrphansRelative []string` (scan-dir relative, slash-separated, sorted)
   - `ReachableSet map[string]struct{}` (for internal/advanced use)
   `}`

2. `func Analyze(g *Graph, scanDir string, allFiles []string) (*Analysis, error)`
   - Validates `g != nil`.
   - Validates root exists in `allFiles`; otherwise returns error (`root not in scanned markdown inventory`).
   - Traverses from `g.Root` over `g.Adjacency`.
   - Computes and sorts reachable + orphan absolute paths.
   - Produces sorted relative orphan paths from `scanDir`.

3. Internal helpers:
   - `func traverseReachable(root string, adjacency map[string][]string) map[string]struct{}`
   - `func toSortedSlice(set map[string]struct{}) []string`
   - `func toRelativeSlash(scanDir string, absPaths []string) ([]string, error)`

No CLI flags added in Phase 5.

## Implementation Details
1. Graph traversal
   - Use iterative DFS with stack (deterministic behavior by pushing neighbors in reverse sorted order if needed).
   - Include root in reachable even if no outgoing edges.
   - Ignore adjacency references to unknown nodes (defensive; should not occur with current builder).

2. Root inventory validation
   - Build a set from `allFiles`.
   - If `g.Root` absent, return error immediately.
   - Error message format: `root markdown file is not in scan result: <abs-path>`.

3. Orphan computation
   - For each file in `allFiles`, if not in reachable set => orphan.
   - Sort absolute orphan list.
   - Convert to relative paths via `filepath.Rel(scanDir, absPath)`.
   - Normalize to slash format with `filepath.ToSlash`.

4. CLI integration (`cmd/gorphan/main.go`)
   - After `graph.Build`, call `graph.Analyze(linkGraph, cfg.Dir, files)`.
   - On error, print and return exit code `2`.
   - In `--verbose`, add:
     - reachable count
     - orphan count
   - Keep return code `0` for now (Phase 6 will introduce `1` when orphans exist).

## Tests
Add/update tests with feature coverage and bug prevention:

1. `internal/graph/graph_test.go`
   - `TestAnalyze_ReachableAndOrphans`
     - fixture graph with connected + disconnected nodes
     - assert reachable absolute paths sorted
     - assert orphan absolute paths sorted
     - assert orphan relative paths sorted/slash-normalized
   - `TestAnalyze_RootNotInInventory_ReturnsError`
   - `TestAnalyze_HandlesRootWithNoEdges`

2. `cmd/gorphan/main_test.go`
   - `TestRun_Phase5AnalysisVerboseCounts`
     - docs with one orphan
     - assert verbose includes reachable/orphan counts
   - `TestRun_RootExcludedByIgnore_Fails`
     - use `--ignore` that excludes root
     - expect exit code `2` and root-inventory error message

3. Regression guard
   - Ensure existing parser/scanner/graph tests continue passing.

## Acceptance Criteria
- `go test ./...` passes.
- `run()` computes reachability/orphans without panic on empty-edge graphs.
- Root excluded from scan produces deterministic error and exit code `2`.
- Orphan relative paths are stable and deterministic across platforms (slash-separated).

## Assumptions and Defaults
- Analysis consumes graph built from scanned markdown files only.
- Paths inside analysis are absolute except explicit `OrphansRelative`.
- No behavior changes to `--format` output yet (deferred to Phase 6).
- Existing unrelated uncommitted repo changes (e.g., `plan.md`, `.github/`) are untouched.
