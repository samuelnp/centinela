#!/bin/sh
set -eu

THRESHOLD="${MIN_COVERAGE:-95.0}"
PROFILE="${COVERAGE_PROFILE:-coverage.out}"

if [ -n "${COVERAGE_VALUE:-}" ]; then
  TOTAL_PCT="$COVERAGE_VALUE"
else
  go test ./... -coverprofile="$PROFILE" >/tmp/centinela-coverage.log

  TOTAL_LINE="$(go tool cover -func="$PROFILE" | awk '/^total:/ {print $3}')"
  TOTAL_PCT="${TOTAL_LINE%%%}"
fi

python3 - "$TOTAL_PCT" "$THRESHOLD" <<'PY'
import sys

actual = float(sys.argv[1])
threshold = float(sys.argv[2])

if actual + 1e-9 < threshold:
    print(f"coverage gate failed: {actual:.1f}% < {threshold:.1f}%")
    sys.exit(1)

print(f"coverage gate passed: {actual:.1f}% >= {threshold:.1f}%")
PY
