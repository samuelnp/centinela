# custom-gate-sdk — senior-engineer

## Files Touched

New (each ≤100 lines):
- `internal/config/custom_gate.go` (84) — `CustomGate` struct, `NormalizeCustomGates`
  (default severity `fail`, output `blob`, timeout 60), `validateCustomGates`
  (indexed: name non-empty/unique/no built-in collision; command non-empty;
  severity∈{fail,warn}; output∈{blob,lines}) + `builtinGateNames` set.
- `internal/gates/custom_command_exec.go` (51) — `runCustom` (shell exec via
  `exec.CommandContext` timeout, `CENTINELA_CHANGED_FILES` env injection, reuses
  the existing `exitCode` helper).
- `internal/gates/custom_command.go` (96) — `customGates`, `customResult`,
  blob (4 KiB cap) / lines (≤200, blanks dropped) Details, empty→generic.

Edited: `internal/config/config.go` (+`CustomGates` field, 100 lines),
`defaults.go` (+normalize), `file_size_exceptions.go` (+validate),
`internal/gates/gates.go` (append `customGates` at tail of `RunWithFilter`,
gated by `len>0` for a byte-identical no-op), `internal/gitdiff/set.go`
(+`Paths()` accessor for the changed-files env), `internal/audit/participation.go`
(fold custom-gate Names into ratchet participants — bug fix below).

## Architecture Compliance

- **G1**: every file ≤100 (`config.go` exactly 100).
- **G2 import-graph**: NO change needed. The runner lives in `internal/gates`
  (domain), importing only `internal/config` (leaf) + `internal/gitdiff` (leaf)
  + stdlib — all already reachable. No new package, no `centinela.toml` change,
  no scaffold mirror. (Cleaner than the audit gate.)
- **Additive, no registry refactor**: built-ins unchanged; the shared contract
  is `gates.Result`. Custom-gate results flow through the existing
  render/telemetry/severity path with no new code there.

## Type-Safety Notes

Strict Go, no `any`. Shell exec is the trusted-config model (matches
`validate.commands`); per-gate timeout prevents hangs. `go vet`/`gofmt` clean;
full suite (2204) green.

## Trade-Offs

- Shell (`sh -c`/`cmd /C`) over argv — custom rules need pipes/globs; documented
  trust model.
- Default severity `fail` (vs built-ins' `warn`) so an explicit rule can't
  silently no-op.

## Real bug found + fixed (cross-feature)

`internal/audit/participation.go` used a hardcoded `defaultParticipants` allowlist
that excluded dynamic custom-gate Names, so `record.go` silently dropped custom
violations — `audit baseline` reported "0 baselined" and the headline
`output="lines"` per-violation ratcheting (AC-4) did nothing. Fixed:
`participatingGates` now also includes every configured custom-gate Name (still
honoring `target_gates`). Verified end-to-end: a failing `output="lines"` custom
gate baselines its 2 violations, `audit` tolerates them, and a new 3rd line
blocks as "1 new". No new import edge (audit already imports config).

## Handoff

→ qa-senior. Not user-facing (CLI). Tests: colocated `internal/config`
(Normalize/validate, collisions) + `internal/gates` (customResult mapping, blob/
lines, exec via `true`/`false`/`sleep`) ≤100 lines each for coverage; tier
unit/integration over `RunWithFilter`; acceptance mapping the 20
`specs/custom-gate-sdk.feature` scenarios. Add an `internal/audit` colocated test
for custom-gate participation.
