# Plan: Add Formatter + Linter to CI Documentation and Task Tracking

## Summary
Update `plan.md` and `tasks.md` so CI explicitly enforces:
- formatting check with `gofmt` in fail-on-diff mode
- linting with `golangci-lint`
- existing tests with `go test ./...`

This plan documents configuration decisions and adds actionable checklist items so implementation is unambiguous.

## Scope
In scope:
- Documentation/task updates in `plan.md` and `tasks.md` describing formatter/linter CI configuration.
- CI expectations for GitHub Actions job steps and pass/fail behavior.

Out of scope:
- Implementing workflow file changes in this step.
- Introducing additional linters beyond `golangci-lint`.

## Important Interface/Contract Changes
- CI quality gate contract is expanded from "tests only" to "format + lint + tests".
- Required CI checks become:
1. `gofmt` check: no formatting diffs allowed.
2. `golangci-lint` run: no lint errors.
3. `go test ./...`: all tests pass.

## Detailed File Update Plan

### 1) `plan.md`
Apply these content changes:

1. In **Section 7 (Testing Strategy)** add a CI quality subsection that states:
- formatter check uses `gofmt` and fails when `gofmt -l` returns files.
- linter check uses `golangci-lint` (GitHub Action-based run).
- tests remain `go test ./...`.
- CI order: format -> lint -> test (fast-fail preferred).

2. In **Section 8 (Implementation Milestones)** update Milestone 5 wording from test-only CI to:
- "Tests (unit + integration) and CI quality gates (format, lint, test) via GitHub Actions."

3. Add a short note in milestones (same section) specifying concrete commands expected in CI:
- format check: `test -z "$(gofmt -l .)"`
- lint: `golangci-lint run`
- test: `go test ./...`

### 2) `tasks.md`
Apply these checklist updates:

1. Add a dedicated **CI Quality Gates** phase (or extend existing Phase 7) with these unchecked items:
- Add formatter check in CI using `gofmt -l`.
- Fail CI when formatter output is non-empty.
- Add lint check in CI using `golangci-lint`.
- Configure/action-pin `golangci-lint` GitHub Action and version.
- Keep/verify `go test ./...` in CI after lint step.
- Ensure CI runs on `push` to `main` and all `pull_request`s.

2. In existing test/quality-related sections, keep current test items and add explicit completion criteria:
- CI is green only when all three gates pass.
- Any format/lint/test failure must fail the workflow.

3. Add a task for local developer parity:
- "Document local pre-PR command sequence: `gofmt`, `golangci-lint run`, `go test ./...`."

## Implementation Notes (for later execution)
- `gofmt` check should be repository-wide and deterministic.
- `golangci-lint` should be treated as authoritative lint gate.
- Avoid auto-format commits in CI; CI should only validate, not mutate.

## Test Cases and Scenarios
1. Formatting failure scenario:
- Intentionally misformat one `.go` file.
- Expected: formatter check fails before lint/test.

2. Lint failure scenario:
- Introduce a lint violation (e.g., unchecked error if rule enabled).
- Expected: lint check fails; tests may be skipped by fast-fail.

3. Test failure scenario:
- Break one unit test.
- Expected: format and lint pass, test step fails.

4. All-green scenario:
- Properly formatted code, clean lint, passing tests.
- Expected: full CI workflow passes.

5. Trigger coverage scenario:
- Open PR and push to `main`.
- Expected: same CI gates run in both events.

## Acceptance Criteria
- `plan.md` explicitly defines format/lint/test CI gates and commands.
- `tasks.md` contains actionable checkboxes for formatter and linter configuration.
- No ambiguity remains about chosen tools or strictness.
- CI success/failure behavior is explicitly documented.

## Assumptions and Defaults Chosen
- Linter: `golangci-lint`.
- Formatter policy: check-only (fail on unformatted files), no CI auto-fix.
- CI gate sequence: format -> lint -> test.
- Existing GitHub Actions trigger model (`push` to `main`, `pull_request`) remains in place.
