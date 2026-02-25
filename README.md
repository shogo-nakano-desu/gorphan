# gorphan

`gorphan` is a lightweight Go CLI to detect orphan Markdown files under a target directory.

## Status

Project bootstrap is complete. Core CLI functionality is under implementation.

## Planned Usage

```bash
gorphan --root docs/index.md --dir docs
```

## Notes

- The root Markdown file is the reachability entry point.
- Files not reachable from the root via Markdown links are reported as orphans.
