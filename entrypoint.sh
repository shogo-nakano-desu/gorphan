#!/usr/bin/env sh

set -eu

root="${INPUT_ROOT:-}"
dir="${INPUT_DIR:-.}"
ext="${INPUT_EXT:-.md,.markdown}"
ignore="${INPUT_IGNORE:-}"
format="${INPUT_FORMAT:-text}"
verbose="${INPUT_VERBOSE:-false}"
unresolved="${INPUT_UNRESOLVED:-warn}"
graph="${INPUT_GRAPH:-none}"
config="${INPUT_CONFIG:-}"
fail_on_orphans="${INPUT_FAIL_ON_ORPHANS:-true}"

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
