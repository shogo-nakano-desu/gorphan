# Remaining Implementation Plan (Phases 6-8 + Pending Phase 9 Doc Item)

## Scope
- Complete Phase 6 (output formats, exit codes, warnings).
- Complete Phase 7 (unit/integration/golden tests and full test pass).
- Complete Phase 8 (README finalization and release-readiness docs).
- Complete pending Phase 9 doc item (local pre-PR quality command sequence).

## Phase 6 Plan
1. Add output rendering package in `internal/report`:
   - Text output for humans.
   - JSON output for automation.
2. Extend runtime flow in `cmd/gorphan/main.go`:
   - Use analysis result to render output.
   - Exit `0` when no orphans, `1` when orphans exist, `2` on usage/runtime errors.
3. Implement warning strategy:
   - Detect unresolved local markdown links during graph build.
   - Emit warnings without failing execution.
4. Add/adjust tests:
   - Verify text output, json output, and exit code behavior.
   - Verify warnings are non-fatal.

## Phase 7 Plan
1. Ensure unit tests cover required areas:
   - link extraction, path handling, ignore rules, reachability.
2. Add integration tests using temporary fixture trees for:
   - no orphan, multiple orphans, cyclic links, disconnected docs, invalid root.
3. Add golden tests for CLI text/json output:
   - Keep stable expected output snapshots in `testdata/golden/`.
4. Run and verify `go test ./...`.

## Phase 8 Plan
1. Rewrite `README.md` to production-ready state:
   - Problem definition and orphan semantics.
   - Installation/run instructions.
   - Concrete usage examples.
   - Output behavior and exit codes.
   - Known limitations for v1.
2. Ensure sample command is explicitly documented:
   - `gorphan --root docs/index.md --dir docs`.

## Pending Phase 9 Documentation Item
1. Document local pre-PR quality sequence in README:
   - `gofmt -w .`
   - `golangci-lint run`
   - `go test ./...`

## Execution Rules
- Add or update tests for every feature/bug change.
- Commit immediately after each phase completion (Phase 6, Phase 7, Phase 8/9-doc).
- Leave unrelated existing local changes untouched unless explicitly requested.
