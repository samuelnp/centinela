# custom-gate-sdk — feature-specialist

## Behavior Summary

`[[gates.custom]]` entries (`name`, `command`, `severity`, optional `output`,
`enabled`) run during `centinela validate`: exit 0 → pass; non-zero → fail
(severity=fail blocks, exit 1) or warn (severity=warn, non-blocking, exit 0),
with command output in `Details`. Custom-gate failures emit telemetry and are
baseline-able/ratcheted by `centinela audit`. `output="lines"` makes each stdout
line a separate violation.

## Acceptance Criteria (Gherkin)

`specs/custom-gate-sdk.feature` — 20 scenarios, 1:1 mapped to Go acceptance
tests via `// Scenario: <name>`, modeled on `specs/audit-baseline-ratchet.feature`
(narrative Feature, Background, comment block, exit-code + determinism rigor).
Uses real commands (`true`, `false`, `sh -c 'exit 1'`, `sleep`, `printf`).

## UX States

CLI/text gate report. States: pass (named, ✓), fail (named, ✗ + details, blocks),
warn (⚠ details, non-blocking), config-error (validate/load fails with indexed
message), timeout/not-found (gate fails with clear message).

## Edge Cases

Exit 0 pass / non-zero fail; severity fail blocks vs warn non-blocking;
empty-output failure → generic Detail; multiple gates independent; enabled=false
skips; no entries → byte-identical; validation rejects empty command/name,
duplicate, built-in collision, bad severity, bad output (indexed); timeout fails
not hangs; command-not-found fails not crash; output=lines per-line violations;
baseline-able + ratcheted; gate-failure telemetry; determinism. (14 recorded in
evidence `edgeCases`.)

## Out-of-Scope

Go-plugin API; full `Gate` interface/registry refactor of built-ins;
deprecating `[validate] commands`; command allowlist/sandbox; Centinela-side
filtering of changed files (the command does its own with
`CENTINELA_CHANGED_FILES`).

## Handoff

→ senior-engineer. Implement per `docs/plans/custom-gate-sdk.md`; config schema,
exec/timeout, output modes, collision list, and the in-`gates` append seam are
fixed there.
