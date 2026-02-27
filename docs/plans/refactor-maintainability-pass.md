# Refactor Plan: `gorphan` Maintainability Pass (Moderate Scope)

## Summary
Refactor is warranted. Current behavior is solid (`go test ./...` passes), but orchestration and normalization logic are concentrated and duplicated, which raises maintenance cost and bug risk for future features.
This plan preserves CLI behavior and output contracts while improving structure, testability, and separation of concerns.

## Goals
- Preserve all current user-visible behavior (flags, exit codes, output shape, unresolved-link modes, graph export behavior).
- Reduce complexity in `run` flow by extracting pipeline steps and shared normalization/validation utilities.
- Remove duplicated extension/path normalization logic currently split across packages.
- Improve test clarity and reduce repeated fixture helpers.

## Non-Goals
- No flag/interface redesign.
- No config format changes.
- No dependency addition for YAML parsing.
- No new features.

## Public Interfaces and Type Changes
- Keep CLI flags and output unchanged in `cmd/gorphan/main.go`.
- Keep exported signatures in `internal/*` stable where possible:
  - `scanner.Scan`, `scanner.NormalizeExtensions`
  - `graph.Build`, `graph.Analyze`, `graph.ExportDOT`, `graph.ExportMermaid`
  - `report.RenderText`, `report.RenderJSON`
  - `config.FindConfigArg`, `config.Load`
- Add new internal-only helpers/types in `internal/pathutil`.

## Planned Refactor Work

### 1. Decompose CLI orchestration
- Extract pipeline phases from `run` into small private functions:
  - config/arg resolution
  - scan
  - graph build/analyze
  - post-processing (ignore-check filtering, relative conversion, sorting)
  - warning handling
  - rendering
  - exit-code decision
- Introduce a lightweight execution context/result struct to pass data between phases.
- Keep `main` and top-level `run(args, stdout, stderr)` signature unchanged.

### 2. Centralize normalization/path logic
- Create shared helpers for:
  - extension set construction
  - absolute/clean path normalization
  - relative slash conversion
  - path-within-dir checks
- Replace duplicated implementations in:
  - `internal/graph/graph.go`
  - `cmd/gorphan/main.go`
  - extension-set duplication in parser/scanner/graph

### 3. Simplify graph module internals
- Keep concurrency behavior intact, but isolate worker-pool setup and result reduction into dedicated helpers.
- Keep `pathIndex` approach, but separate traversal/index concerns from edge extraction for readability.
- Centralize warning formatting.

### 4. Improve config parsing structure (custom parser retained)
- Keep custom parser but split parsing into:
  - line tokenizer/scanner
  - key dispatch handlers
  - list accumulation
- Strengthen validation around malformed list/scalar mixes while preserving accepted current input patterns.
- Keep `FileConfig` shape and `Load` contract unchanged.

### 5. Test refactor and coverage hardening
- Consolidate repeated `mustWrite` helpers in tests under shared test utility file/package.
- Add focused tests for shared normalization helpers.
- Add parity tests ensuring refactor does not change:
  - unresolved modes behavior
  - graph export skip behavior with `--max-graph-nodes`
  - config + CLI override precedence
  - relative path output normalization
- Keep existing integration/golden/e2e tests intact.

## File-Level Target Map
- Primary refactor:
  - `cmd/gorphan/main.go`
  - `internal/graph/graph.go`
  - `internal/config/config.go`
- Secondary updates:
  - `internal/scanner/scanner.go`
  - tests under `cmd/gorphan`, `internal/*`, and `e2e`
- New helper package:
  - `internal/pathutil/*.go`

## Verification
- `go test ./...`

## Assumptions and Defaults
- Scope: moderate.
- Keep `.gorphan.yaml` custom parser (no new dependency).
- Behavior compatibility is mandatory; output drift is a regression unless approved.
- No dedicated perf rewrite in this pass.
