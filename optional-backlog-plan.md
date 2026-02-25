# Optional Backlog Implementation Plan

## Scope
- Implement wikilink parsing support (`[[Page]]`, `[[path/to/file.md]]`, `[[Page|Alias]]`).
- Add unresolved-link reporting mode control.
- Add graph export support (DOT and Mermaid).
- Parallelize markdown graph construction for larger repositories.
- Add `.gorphan.yaml` config file support.

## Implementation Steps
1. Parser enhancements
   - Extend `internal/parser` to extract wikilinks.
   - Keep v1 rules (local links only, extension filtering) with wikilink fallback to `.md`.
   - Add unit tests for wikilink extraction variants.

2. Graph builder enhancements
   - Parallelize per-file graph extraction in `internal/graph`.
   - Keep deterministic output (sorted adjacency/warnings).
   - Add graph export helpers for DOT/Mermaid.
   - Add tests for export outputs and parallel behavior invariants.

3. CLI/report enhancements
   - Add `--unresolved` mode (`warn`, `report`, `none`).
   - Add `--graph` mode (`none`, `dot`, `mermaid`).
   - Extend report text/json payload to optionally include warning lists and graph text.
   - Add CLI tests for new flag behavior and outputs.

4. Config support
   - Add `.gorphan.yaml` loader with optional explicit `--config`.
   - Default behavior: auto-load `.gorphan.yaml` when present.
   - Explicit `--config` missing file should fail fast.
   - Add tests for config defaults and CLI override precedence.

5. Documentation/task updates
   - Update `README.md` with new optional features and examples.
   - Mark optional backlog items complete in `tasks.md`.

## Validation
- Run `gofmt -w .` on touched Go files.
- Run `go test ./...` and ensure all pass.
- Commit once optional backlog implementation is complete.
