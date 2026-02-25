# Architecture

## Packages
- `cmd/gorphan`: CLI parsing, execution flow, and exit codes.
- `internal/scanner`: Markdown file discovery and ignore handling.
- `internal/parser`: Markdown link extraction and normalization.
- `internal/graph`: Link graph construction and reachability/orphan analysis.
- `internal/report`: Text/JSON result rendering.

## Data Flow
1. Parse CLI flags and validate paths.
2. Scan markdown files under `--dir`.
3. Build link graph from scanned files.
4. Analyze reachability from `--root`.
5. Render text/json output and set exit code.
