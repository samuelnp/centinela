#!/bin/sh
set -eu

DIRS="${FMT_DIRS:-cmd internal tests}"

# shellcheck disable=SC2086 # DIRS is a deliberate word-split list.
OFFENDERS="$(gofmt -l $DIRS)"

if [ -n "$OFFENDERS" ]; then
  {
    echo "format gate failed: files need gofmt -w:"
    echo "$OFFENDERS"
  } >&2
  exit 1
fi
