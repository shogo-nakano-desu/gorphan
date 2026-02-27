# AGENT.md

This file is the root map of Markdown documents in this repository.

## Current Docs
- [README](./README.md): Usage, flags, output behavior, and local quality checks.
- [Architecture](./docs/architecture.md): Package boundaries and data flow.
- [Testing Guide](./docs/testing.md): Unit/integration/golden/e2e test strategy.
- [Contributing](./CONTRIBUTING.md): Development workflow and quality checks.
- [Implementation Plan](./docs/plans/plan.md): Build plan for the Go CLI that detects orphan Markdown files.
- [Task Breakdown](./docs/plans/tasks.md): Detailed implementation checklist derived from `plan.md`.

## Planned Docs
- None.

## Notes
- Keep this file updated whenever a new Markdown document is added, moved, or removed.
- Treat this file as the navigation entry point for humans and tooling.

## Workflow Rule
- Always create a git commit immediately after completing each implementation phase.
- After making a change, commit the change.
- Always add or update tests when implementing a new feature or fixing a bug.
