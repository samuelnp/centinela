# custom-gate-sdk — big-thinker

## Problem

Centinela's gates are a closed, hardcoded set (`gates.RunWithFilter`'s
`if cfg.Gates.X.Enabled` chain). A team with a project-specific mechanical rule
must fork Centinela or drop it into `[validate] commands` — where it runs as an
opaque pass/fail shell line with no severity, no structured violations, no
telemetry, and no participation in the gate ecosystem (enforcement profiles,
audit baseline/ratchet, the gate report).

## Scope

Add a `[[gates.custom]]` config surface: command-backed gates that produce the
same `gates.Result` as built-ins and so flow through the existing
render/severity/telemetry path AND become baseline-able by the audit ratchet.
The "SDK" is the config schema + the `Result` contract — NOT a Go-plugin API.
Built-ins are documented as reference implementations; none are rewritten.

## Dependencies & Assumptions

- Reuses the shell-exec pattern (`cmd/centinela/validate_runner.go`) + timeout
  pattern (`internal/gates/security_exec.go CommandContext`).
- Assumes config is checked-in code = trusted (same trust model as
  `validate.commands`); no allowlist/sandbox.
- Custom-gate `Details` fingerprint via the audit ratchet's GENERIC extractor;
  `output="lines"` yields per-line violations that fingerprint individually.

## Risks

- **Code-execution surface** — runs user shell commands. Mitigation: trusted
  checked-in config; per-gate timeout (`exec.CommandContext`, default 60s);
  document clearly; no privilege escalation.
- **Scope creep** — a full `Gate` interface/registry refactor of every built-in
  is large/risky and pressures the 100-line rule. v1 stays ADDITIVE: one new
  `customGates(cfg, filter) []Result` appended at the tail of `RunWithFilter`.
- **Determinism** — non-deterministic commands undermine ratchet fingerprint
  stability; documented as the user's responsibility.

## Rollout

Byte-identical no-op when no `[[gates.custom]]` configured (the append is gated
by `len(cfg.Gates.CustomGates) > 0`). Default severity **`fail`** (a deliberate
divergence from built-ins' `warn` — an explicitly-authored custom rule should
not silently no-op; soft rollout = set `warn` explicitly).

## Handoff

→ feature-specialist. Plan at `docs/plans/custom-gate-sdk.md`. Key calls: shell
exec + timeout; `output` modes blob (default, bounded) / lines (per-violation);
additive runner inside `internal/gates` (domain) importing only `config` (leaf)
+ stdlib — **no new import edge, no centinela.toml change, no scaffold mirror**
(cleaner than the audit gate). diff-aware kept opt-in via
`CENTINELA_CHANGED_FILES` env. Built-in name collision list specified for
`validateCustomGates`.
