# Testing Guide

## Test Layers
- Unit tests: parser/scanner/graph/report behavior.
- Integration tests: CLI behavior over temporary markdown trees.
- Golden tests: stable text/json output snapshots.
- E2E tests: execute compiled CLI binary from tests.

## Commands
```bash
go test ./...
```

## CI Gates
- Format: `gofmt -l .` must be empty.
- Lint: `golangci-lint run`.
- Test: `go test ./...`.
