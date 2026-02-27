#!/usr/bin/env sh

set -eu

root="${INPUT_ROOT:-}"
dir="${INPUT_DIR:-.}"
ext="${INPUT_EXT:-.md,.markdown}"
ignore="${INPUT_IGNORE:-}"
ignore_check_files="${INPUT_IGNORE_CHECK_FILES:-}"
format="${INPUT_FORMAT:-text}"
verbose="${INPUT_VERBOSE:-false}"
unresolved="${INPUT_UNRESOLVED:-fail}"
graph="${INPUT_GRAPH:-none}"
config="${INPUT_CONFIG:-}"
fail_on_orphans="${INPUT_FAIL_ON_ORPHANS:-}"
if [ -z "$fail_on_orphans" ]; then
  fail_on_orphans="$(printenv "INPUT_FAIL-ON-ORPHANS" 2>/dev/null || true)"
fi
if [ -z "$fail_on_orphans" ]; then
  fail_on_orphans="true"
fi

if [ -z "$root" ]; then
  echo "error: input 'root' is required" >&2
  exit 2
fi

set -- --root "$root" --dir "$dir" --ext "$ext" --format "$format" --unresolved "$unresolved" --graph "$graph"

if [ "$verbose" = "true" ]; then
  set -- "$@" --verbose
fi

if [ -n "$config" ]; then
  set -- "$@" --config "$config"
fi

# Support newline or comma separated ignore patterns.
if [ -n "$ignore" ]; then
  ignore_lines=$(printf "%s\n" "$ignore" | tr ',' '\n')
  OLDIFS=$IFS
  IFS='
'
  for pattern in $ignore_lines; do
    if [ -n "$pattern" ]; then
      set -- "$@" --ignore "$pattern"
    fi
  done
  IFS=$OLDIFS
fi

# Support newline or comma separated ignore-check file entries.
if [ -n "$ignore_check_files" ]; then
  ignore_check_lines=$(printf "%s\n" "$ignore_check_files" | tr ',' '\n')
  OLDIFS=$IFS
  IFS='
'
  for rule in $ignore_check_lines; do
    if [ -n "$rule" ]; then
      set -- "$@" --ignore-check-file "$rule"
    fi
  done
  IFS=$OLDIFS
fi

set +e
/usr/local/bin/gorphan "$@"
exit_code=$?
set -e

has_orphans=false
if [ "$exit_code" -eq 1 ]; then
  has_orphans=true
fi

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  {
    printf "exit-code=%s\n" "$exit_code"
    printf "has-orphans=%s\n" "$has_orphans"
  } >>"$GITHUB_OUTPUT"
fi

if [ "$exit_code" -eq 1 ] && [ "$fail_on_orphans" != "true" ]; then
  echo "orphan markdown files found, but fail-on-orphans=false so exiting successfully."
  exit 0
fi

exit "$exit_code"
