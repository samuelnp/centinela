# cost-governance — senior-engineer

## Files Touched

- **config**: `internal/config/cost.go` (`[cost]` schema + Normalize + IsActive),
  wired into `config.go` struct and `defaults.go`.
- **telemetry leaf**: `event.go` (`TypeCostSample` + InputTokens/OutputTokens),
  `constructors.go` (`RecordCostSample`, no-op on zero).
- **new `internal/cost`** (aggregator layer): `transcript.go` (tolerant JSONL
  token summer with offset), `cursor.go` (per-transcript read position),
  `aggregate.go` (fold samples → feature/step/model), `budget.go` (Status math),
  `report.go` (Build + ActiveStatus + AnyOver/Empty).
- **ui**: `render_cost.go` (report + soft-gate ⚠ line).
- **cmd**: `cost.go` (`centinela cost [--json]`), `hook_cost.go` (capture),
  `validate_cost.go` (non-failing validate ⚠), `active_feature.go` (worktree or
  root active-feature resolver, shared by hook + validate).
- **setup**: `hooks.go` + `settings_build.go` wire `centinela hook cost` to the
  harness **Stop** event (which provides `transcript_path`); `hooks_test.go`
  updated for the new param.
- **config/docs**: `centinela.toml` adds `internal/cost/**` to the import_graph
  aggregator layer and a commented `[cost]` example.

## Architecture Compliance

`internal/cost` is a read-only aggregator importing the `internal/telemetry` and
`internal/config` leaves + stdlib only; imported solely by `cmd/` (its Report
type by `internal/ui`). Added to the aggregator layer `paths`, mirroring
`internal/calibration`/`insights`. No new cross-layer edge. Telemetry stays a
config-only leaf (token fields are plain ints).

## Type-Safety Notes

Strict typing throughout: budgets are `int` (0 = off), token usage is summed via
typed `Usage`. The transcript reader is tolerant by design (unknown shapes →
zero) but typed — no `any`. End-to-end smoke confirmed: a 2-message transcript
(one garbage line skipped) summed to 1500 in / 1000 out, a repeated hook fire
added nothing (cursor at EOF), and `validate` exited 0 with the ⚠.

## Trade-Offs

- Spend is attributed to the active feature/step at hook time (transcript is
  per-session); a byte cursor prevents double-counting. Cross-feature precision
  within one Stop is inherently approximate — acceptable for a visibility gate.
- Tier budgets are keyed by model id (tier-name resolution deferred); v1's
  soft gate evaluates step + feature budgets.
- Lock/transcript/config failures are all silent no-ops (telemetry is
  non-blocking) — cost never fails the host command.

## Handoff

→ qa-senior: unit (config Normalize/IsActive, budget math, transcript summing +
tolerant parse + cursor), integration (telemetry round-trip → aggregate), and
acceptance (binary-driven capture → `cost` over-budget → `validate` exit 0 ⚠).
