# gorphan

[![CI](https://github.com/shogo-nakano-desu/gorphan/actions/workflows/ci.yml/badge.svg)](https://github.com/shogo-nakano-desu/gorphan/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-00ADD8?logo=go)](https://go.dev/)

`gorphan` is a fast, focused CLI to find **orphan markdown files** in documentation trees.

An orphan markdown file is a markdown file that is not reachable from a given root file by following local links.

## Why gorphan?
- Keep docs navigation healthy as projects grow.
- Catch disconnected pages before publishing.
- Integrate in CI with clear exit codes.
- Export link graphs for visualization (DOT/Mermaid).

## Features
- Recursive markdown scan with extension filters.
- Link parsing:
  - inline markdown links
  - reference-style links
  - wikilinks (`[[Page]]`, `[[path/file.md]]`, `[[Page|Alias]]`)
- Reachability analysis from a root page.
- Orphan detection with human-readable or JSON output.
- Unresolved link handling modes (`warn`, `report`, `none`).
- Optional graph export (`dot`, `mermaid`).
- Optional `.gorphan.yaml` config with CLI override.

## Install

Build local binary:

```bash
go build -o gorphan ./cmd/gorphan
```

Install to `$GOBIN`:

```bash
go install ./cmd/gorphan
```

## Use as GitHub Action

This repository is also a Docker-based GitHub Action.

Example workflow:

```yaml
name: Docs Orphan Check

on:
  pull_request:
  push:
    branches: [main]

jobs:
  orphan-check:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Run gorphan
        uses: shogo-nakano-desu/gorphan@v1
        with:
          root: docs/architecture.md
          dir: docs
          format: text
          fail-on-orphans: true
```

Useful inputs:
- `root` (required)
- `dir` (default `.`)
- `ignore` (newline or comma separated patterns)
- `format` (`text` or `json`)
- `unresolved` (`warn`, `report`, `none`)
- `graph` (`none`, `dot`, `mermaid`)
- `fail-on-orphans` (`true` or `false`)

Action outputs:
- `exit-code`
- `has-orphans`

## Quick Start

```bash
./gorphan --root docs/index.md --dir docs
```

`--dir` is optional and defaults to the current directory.

## Usage

```bash
gorphan --root <root.md> [--dir <path>] [options]
```

### Flags
- `--root` (required): root markdown file (entry point).
- `--dir` (optional, default current directory): scan target.
- `--ext` (optional, default `.md,.markdown`): comma-separated markdown extensions.
- `--ignore` (optional, repeatable): ignore path prefix or glob.
- `--format` (optional, default `text`): `text` or `json`.
- `--verbose` (optional): include diagnostics summary.
- `--unresolved` (optional, default `warn`): unresolved-link handling (`warn`, `report`, `none`).
- `--graph` (optional, default `none`): graph export mode (`none`, `dot`, `mermaid`).
- `--config` (optional, default `.gorphan.yaml`): explicit config file path.

## Examples

Default text output:

```bash
gorphan --root docs/index.md --dir docs
```

JSON output for CI:

```bash
gorphan --root docs/index.md --dir docs --format json
```

Use current directory:

```bash
gorphan --root AGENT.md
```

Report unresolved links in output:

```bash
gorphan --root docs/index.md --dir docs --unresolved report
```

Suppress unresolved warnings:

```bash
gorphan --root docs/index.md --dir docs --unresolved none
```

Export graph:

```bash
gorphan --root docs/index.md --dir docs --graph dot
gorphan --root docs/index.md --dir docs --graph mermaid
```

## Output and Exit Codes

- Exit code `0`: no orphan files found.
- Exit code `1`: orphan files found.
- Exit code `2`: usage/runtime error.

Text output:
- Orphan list (relative to `--dir`).
- Optional summary when `--verbose`.
- Optional unresolved section when `--unresolved report`.

JSON output includes:
- `root`
- `dir`
- `orphans`
- `warnings`
- `graph`
- `summary` (`scanned`, `reachable`, `orphans`)

## Configuration (`.gorphan.yaml`)

If `.gorphan.yaml` exists in the current working directory, it is used as defaults.
CLI flags always win over config values.

```yaml
root: docs/index.md
dir: docs
ext: .md,.markdown
ignore:
  - drafts
  - archive/*
format: text
verbose: false
unresolved: warn
graph: none
```

## Development

Enable repository hooks once per clone:

```bash
git config core.hooksPath .githooks
```

Run local quality checks:

```bash
gofmt -w .
golangci-lint run
go test ./...
```

## Contributing

Contributions are welcome. Start with [CONTRIBUTING.md](./CONTRIBUTING.md).

## Release the Action

Publish and maintain stable Action tags:

```bash
# After merging to main
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Move major tag to latest v1 release
git tag -fa v1 -m "Release v1.0.0"
git push origin v1 --force
```

## Roadmap
- Richer markdown dialect support.
- Additional graph/report outputs.
- Performance tuning for very large documentation sets.

## License

No license file has been added yet.
