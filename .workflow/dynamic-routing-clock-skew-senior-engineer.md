# dynamic-routing-clock-skew — senior-engineer

## Files Touched

| Path | Reason |
|------|--------|
| `internal/workflow/model_routes_underway.go` | Compare evidence mtime against `StartedAt - clockSkewGrace` instead of `StartedAt` exactly |

## Problem and root cause

PR #90 (`dynamic-model-routing`) passed every local gate — three adversarial
verifier rounds, a full `centinela validate`, and the machine-side suite run at
`complete` — then failed CI on two tests, both exercising
`roleEvidenceFromThisRun`:

- `internal/workflow` — `TestRoleStepUnderway_CurrentStepNeedsEvidenceOnDisk`
- `tests/unit` — `TestRoleStepUnderway_JSONStubFromThisRunCounts`

`New()` stamps `StartedAt` from `time.Now().UTC()` — Go's precise clock. A file's
mtime is stamped by the kernel from a COARSE clock updated once per timer tick.
On Linux that value can read several milliseconds EARLIER than a `time.Now()`
taken microseconds before the write, so a stub written immediately after `start`
compares as older than `StartedAt` and is misjudged as "left over from an earlier
run". macOS stamps mtimes finely enough that the ordering held, which is why
every local run was green and only CI was red.

This is a real defect, not a flaky test: on Linux the same skew would let a
downgrade slip past the "role's step is underway" refusal for a short window
after `start`.

## Architecture Compliance

- Boundary checks passed: change confined to `internal/workflow`; only stdlib
  `time` added; no cmd/ or config-leaf edits; no new package edges.
- G1 file size: `model_routes_underway.go` is 69 lines (≤100).
- G7 outer-layer rule: no business logic moved into an outer layer.

## Type-Safety Notes

- `clockSkewGrace` is a typed `time.Duration` constant, not a bare number.
- Comparison stays in `time.Time`/`time.Duration` arithmetic; no unix-int
  conversions, no float seconds.

## Trade-Offs

- **Grace window vs. exact comparison:** exactness is unachievable across two
  different clock sources. Stamping `StartedAt` from the filesystem instead
  would add I/O to every `start` and still race.
- **2 seconds:** far above any plausible tick granularity, far below the gap
  that defines a genuinely stale stub (an earlier run predates the current one
  by minutes at minimum), so the check keeps everything it was meant to catch.

## Deferred Findings

none

## Handoff

- Next role: qa-senior.
- Outstanding: the regression test must reproduce the skew DETERMINISTICALLY —
  set the artifact's mtime just before `StartedAt` with `os.Chtimes` rather than
  relying on the host clock's granularity. The original tests could not fail on
  macOS, which is exactly how this reached CI.
