# gorphan

`gorphan` is a lightweight Go CLI to detect orphan Markdown files under a target directory.

An orphan Markdown file is a file that is not reachable from a specified root Markdown file by following local Markdown links.

## Install

```bash
go build -o gorphan ./cmd/gorphan
```

## Usage

```bash
gorphan --root docs/index.md --dir docs
```

### Flags
- `--root` (required): root markdown file (entry point).
- `--dir` (required): directory to scan recursively.
- `--ext` (optional, default `.md,.markdown`): comma-separated markdown extensions.
- `--ignore` (optional, repeatable): ignore path prefix or glob.
- `--format` (optional, default `text`): output format (`text` or `json`).
- `--verbose` (optional): include summary diagnostics.

### Examples

Text output:

```bash
gorphan --root docs/index.md --dir docs
```

JSON output:

```bash
gorphan --root docs/index.md --dir docs --format json
```

Ignore directories/files:

```bash
gorphan --root docs/index.md --dir docs --ignore drafts --ignore archive/*
```

Verbose summary:

```bash
gorphan --root docs/index.md --dir docs --verbose
```

## Output and Exit Codes

- Exit code `0`: no orphan files found.
- Exit code `1`: orphan files found.
- Exit code `2`: usage/runtime error (invalid args, missing root, scan/build/analyze failure).

Text format prints a human-readable list of orphan files (relative to `--dir`).
JSON format prints structured output with:
- `root`
- `dir`
- `orphans`
- `warnings`
- `summary` (`scanned`, `reachable`, `orphans`)

Warnings for unresolved local markdown links are emitted to stderr and do not fail execution by themselves.

## Local Pre-PR Quality Checks

Run these before opening a pull request:

```bash
gofmt -w .
golangci-lint run
go test ./...
```

## Limitations (v1)

- The root Markdown file is the reachability entry point.
- Local markdown links only; external links are ignored.
- Links without file extensions are ignored by default.
- Absolute-path links are ignored.
- Unresolved local links are warned but not treated as fatal errors.
