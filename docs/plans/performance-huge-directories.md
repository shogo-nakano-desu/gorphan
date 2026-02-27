# Performance Plan for Huge Directories

## Goal
Improve runtime and memory behavior when scanning and analyzing very large documentation trees (tens of thousands to hundreds of thousands of files) while preserving current CLI behavior.

## Baseline and Success Metrics
- Capture baseline wall-clock time for `scan`, `build`, `analyze`, and report generation.
- Capture peak memory (RSS) on representative large datasets.
- Define regression gates for CI-sized fixture datasets.

## Task List
- [x] Task 1: Reduce scanner path normalization overhead.
  - Remove redundant per-file `filepath.Abs` work in `scanner.Scan`.
  - Keep deterministic sorted output and existing ignore semantics.
- [x] Task 2: Pre-compile ignore rules for fast matching.
  - Parse ignore rules once into prefix and glob buckets.
  - Reuse compiled rules during directory walk.
- [x] Task 3: Remove repeated scan-dir resolution in graph export.
  - Resolve and clean `scanDir` once per export call.
  - Reuse cached value for all node label conversions.
- [x] Task 4: Add benchmark scaffolding for large trees.
  - Add focused benchmarks for scanner and graph build.
  - Include fixture generators sized for local performance testing.
- [x] Task 5: Tune graph build concurrency strategy.
  - Introduce configurable worker cap to reduce I/O contention.
  - Measure throughput changes with benchmark fixtures.
- [x] Task 6: Evaluate memory-focused graph representation.
  - Prototype ID-based node indexing to reduce duplicate path strings.
  - Keep output compatibility by converting IDs to paths at render time.
- [x] Task 7: Add guardrails for very large graph exports.
  - Add optional threshold warning/skip behavior for huge graph outputs.
  - Keep default behavior backward compatible unless explicitly configured.

## Execution Notes
- Implement tasks in order.
- After each task: run relevant tests and record outcome.
- Prefer behavior-preserving refactors first, then optional feature flags for larger changes.
