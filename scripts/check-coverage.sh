#!/bin/sh
set -eu

THRESHOLD="${MIN_COVERAGE:-95.0}"
PROFILE="${COVERAGE_PROFILE:-coverage.out}"

if [ -n "${COVERAGE_VALUE:-}" ]; then
  TOTAL_PCT="$COVERAGE_VALUE"
else
  # Reuse branch: when COVERAGE_PROFILE is explicitly set AND the file exists,
  # the profile was written by this validate run's own suite execution — skip
  # the internal go test. Explicitly set but missing → fail-safe: run the
  # suite ourselves. Bare invocation → self-contained, exactly as before.
  if [ -z "${COVERAGE_PROFILE:-}" ] || [ ! -f "$PROFILE" ]; then
    go test ./... -coverprofile="$PROFILE" >/tmp/centinela-coverage.log
  fi

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
