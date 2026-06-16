# Feature Brief — custom-gate-sdk

> Phase 8: Continuous Governance. Let teams define their own mechanical gates
> (project-specific rules) via config — run inside `centinela validate` without
> forking Centinela. The built-in gates become reference implementations of one
> `Result` contract, not the ceiling. This is the pivot from opinionated
> workflow tool to **policy engine**.

## Problem

Centinela's gates are a closed set: every gate is a hardcoded
`if cfg.Gates.X.Enabled { checkX(...) }` call in `gates.RunWithFilter`. A team
with a project-specific rule ("no `console.log` in `src/`", "every migration has
a rollback", "OpenAPI spec matches handlers") has two bad options: fork
Centinela, or drop the rule into `[validate] commands` — where it runs as an
opaque pass/fail shell line with **no severity, no structured violations, no
telemetry, and no participation in the gate ecosystem** (enforcement profiles,
the audit baseline/ratchet, the gate report). Custom rules are second-class.

## What this adds

A `[[gates.custom]]` config surface: teams declare command-backed gates that
become **first-class** — they produce the same `gates.Result{Name, Status,
Message, Details}` as built-ins and therefore flow through the existing
rendering, `warn` vs `fail` severity, telemetry (`gate-failure` events), and —
crucially — are **baseline-able by `audit-baseline-ratchet`** (their `Details`
fingerprint like any gate's). The "SDK" is the documented config schema + the
`Result` contract, NOT a Go-plugin API.

## Key decisions to resolve in the plan

- **Exec model: shell vs argv.** The build gate uses `strings.Fields` (no
  shell); `[validate] commands` uses `sh -c`/`cmd /C`. Custom rules realistically
  need shell features (pipes, globs, `&&`). Decide — likely shell-based for
  ergonomics — and document the trust model (config is checked-in code = trusted;
  no allowlist/sandbox, matching `validate.commands` today). Add a per-gate
  timeout.
- **Pass/fail mapping + structured violations.** At minimum: exit 0 = pass,
  non-zero = fail/warn per `severity`, with captured stdout/stderr in `Details`.
  Decide whether to support **`output = "lines"`** so each stdout line becomes a
  separate `Details` entry (a real violation) — this is what makes a custom gate
  meaningfully baseline-able by the ratchet (per-line fingerprints) rather than
  one opaque blob. Recommend supporting it; keep the default simple.
- **Differentiation from `[validate] commands`.** Be explicit: custom gates are
  the structured, governed path (severity/telemetry/baseline); `validate.commands`
  remains the quick opaque path. Decide whether to deprecate/redirect or leave
  both.
- **Interface generalization scope.** The roadmap says built-ins become
  "reference implementations." Decide v1 scope: an **additive custom-gate runner**
  (built-ins keep their hardcoded chain; custom gates loop through one new
  runner, all emitting the shared `Result`) vs a full `Gate` interface + registry
  refactor of every built-in (large, risky, 100-line-file pressure). Recommend
  additive — the shared contract is already `Result`; document built-ins AS the
  reference, don't rewrite them.
- **diff-aware.** Decide whether a custom gate can opt into changed-files scope
  (e.g. `CENTINELA_CHANGED_FILES` env var) or is always full-scan in v1
  (recommend: expose changed files via env, opt-in `diff_aware = true`; defer if
  it bloats scope).
- **Wiring seam.** Append custom-gate results in `gates.RunWithFilter` (domain
  layer) or from `cmd/` like the audit gate. Recommend in-`gates` (it's a
  built-in-style domain gate reading config + exec, no new cross-layer edge).

## Acceptance Criteria

1. A `[[gates.custom]]` entry with `name`, `command`, `severity` runs during
   `centinela validate` and appears in the gate report with its name.
2. Command exit 0 → the gate passes; non-zero → it fails (`severity=fail`,
   blocks, exit 1) or warns (`severity=warn`, non-blocking), with command output
   surfaced in `Details`.
3. Custom-gate failures are recorded in telemetry as `gate-failure` events (same
   path as built-ins).
4. A failing custom gate's violations are baseline-able by `centinela audit
   baseline` and tolerated/ratcheted exactly like built-in gate violations.
5. Multiple custom gates run independently; one failing doesn't stop others.
6. Config is validated: required fields present, `severity ∈ {fail,warn}`,
   duplicate/empty names rejected with indexed errors; `enabled=false` skips.
7. A custom gate with `output = "lines"` turns each stdout line into a separate
   `Details` violation (so the ratchet fingerprints them individually).
8. Per-gate timeout; a hung command fails the gate rather than hanging validate.
9. All new source files ≤100 lines; no new cross-layer import violations.

## Edge Cases

- Empty/whitespace `command` → validation error (not a runtime panic).
- Command not found / non-executable → gate fails with a clear message, not a
  crash.
- Command times out → gate fails with a timeout message.
- `name` collides with a built-in gate name or another custom gate → rejected.
- Command prints nothing on failure → `Details` falls back to a generic message.
- `enabled=false` or empty `[[gates.custom]]` list → no behavior change
  (byte-identical validate output).
- Huge command output → captured but bounded (truncate Details to keep the
  report readable).
- Non-zero exit but `severity=warn` → reported, non-blocking.
- `output = "lines"` with thousands of lines → bounded.

## Data Model

New `config.CustomGate` (`internal/config/custom_gate.go`):
`{Name, Command, Severity string; Output string; TimeoutSeconds int; DiffAware
bool; Enabled bool}` with `toml` tags; `CustomGates []CustomGate \`toml:"custom"\``
added to `GatesConfig`. Normalized (default severity, timeout) in `applyDefaults`,
validated in `validateConfig` with indexed errors. New gate runner
`internal/gates/custom_command.go` producing `[]gates.Result`.

## Integration Points

- **Run**: append custom-gate results in `gates.RunWithFilter` (loop over
  `cfg.Gates.CustomGates`).
- **Exec**: shell exec with timeout (model on `internal/gates/build_runner.go` +
  `security_exec.go` timeout), changed-files via env when diff-aware.
- **Telemetry/render**: unchanged — results flow through `emitGateFailures` +
  `ui.RenderGateResult`.
- **Audit ratchet**: custom gate `Details` fingerprint via the existing generic
  extractor; confirm the generic fallback handles arbitrary custom output.

## Risks

- **Code execution surface** — running user-defined shell commands. Mitigation:
  config is checked-in code (trusted), same model as `validate.commands`;
  document clearly; per-gate timeout; no privilege escalation.
- **Scope creep** — a full `Gate` interface/registry refactor of every built-in
  is tempting but large and risky; v1 should stay additive.
- **Determinism** — custom commands can be non-deterministic; that's the user's
  responsibility, but the ratchet's fingerprint stability depends on stable
  output. Document the expectation.
