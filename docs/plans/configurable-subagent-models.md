# Plan: Configurable Subagent Models

> Feature brief (locked design): [docs/features/configurable-subagent-models.md](../features/configurable-subagent-models.md)

## Problem

Centinela delegates every feature step to ~7 subagents, all running on whatever
model the host session uses. Architecture framing and cross-layer implementation
want a reasoning-heavy model; "read a report and tabulate" is served identically
by a fast, cheap tier. There is no knob to right-size cost/latency per role.

Centinela does not spawn subagents — the orchestrator does. Centinela's only
lever is the delegate directive emitted from
`cmd/centinela/hook_orchestration.go:42`. So this feature is **advisory**: it
annotates each role in the directive with a model tier; the orchestrator complies
the same way it complies with the rest of the directive system. No hard
enforcement, no evidence-schema change (per locked design decision #4).

## Solution

Three semantic tiers in user config (`reasoning | balanced | fast`), smart
per-role defaults, and a per-runner tier→model table living in one internal
place. The orchestration hook annotates each step role with its resolved tier and
emits a compact both-runner model reference, so the orchestrator picks the right
ID for its runner.

### Runner-resolution decision (the central design question)

The same `centinela hook orchestration` binary runs under both runners and has
**no runtime runner signal at emit time** — the `--agent` flag exists only on
`init`/`migrate`/`setup`. Chosen strategy: **(a) emit the tier name per role plus
a one-line both-runner model reference**, deferring the literal ID choice to the
orchestrator (which knows its runner).

Rejected: (c) inject a `--runner`/`CENTINELA_RUNNER` signal at wiring time — the
correct long-term fix but the most blast radius (touches `internal/setup` and
both runner configs) and would require every install to `migrate`; (b) emit both
IDs per role — noisier than factoring the runner table into one reference line.
The resolver takes `runner` as a parameter defaulting to "unknown" → emit-both,
so adding (c) later is a one-line call-site change with zero domain churn.

### Slice 1 — Domain: tiers, defaults, runner-agnostic resolution

Domain logic in `internal/orchestration`; no wiring yet. Fully unit-testable in
isolation, de-risking the runner-agnostic decision before any config/hook change.

- **`internal/orchestration/models.go` — new.** `Tier` type + constants
  (`reasoning`/`balanced`/`fast`); `DefaultTierForRole(role Role) Tier` encoding
  the brief's table for the 7 step roles (plus documented out-of-band defaults);
  `NormalizeTier(s string) (Tier, bool)` (trim + lowercase, then validate);
  `AllowedTiers()` and `AllowedRoleSlugs()` accessors. Keep ≤100 lines; if the
  per-runner table pushes the budget, split it into `models_table.go`.
- **`internal/orchestration/resolve.go` — new.** `Runner` type
  (`claude`/`opencode`/`unknown`); the per-runner tier→model table (Claude Code:
  `claude-opus-4-7` / `claude-sonnet-4-6` / `claude-haiku-4-5-20251001`;
  opencode: `anthropic/claude-opus-4-7` / `anthropic/claude-sonnet-4-6` /
  `anthropic/claude-haiku-4-5`);
  `ResolveModel(role Role, models map[string]string, runner Runner) (string, bool)`
  resolving role → tier (config override or default) → model (or tier name on a
  missing mapping); and `ModelReference(tiers []Tier) string` rendering the
  compact both-runner reference line. The resolver never panics — a missing
  mapping returns the tier name + `ok=false` so the caller can warn.

### Slice 2 — Config: parse + validate `[orchestration.models]`

Leaf layer (`internal/config`), no internal imports. Validation mirrors the
existing `file_size_exceptions` validator.

- **`internal/config/orchestration.go` — modify.** Add
  `Models map[string]string` (`toml:"models"`) to `OrchestrationConfig`; add an
  `OrchestrationModels(cfg *Config) map[string]string` accessor returning the
  normalized map (nil-safe).
- **`internal/config/orchestration_models.go` — new.**
  `validateOrchestrationModels(cfg *Config) error` — for each entry: normalize
  the tier (trim/lowercase), reject an unknown role key (message names the key),
  reject an invalid tier (message names the key and lists allowed tiers). Called
  from `validateConfig` in `file_size_exceptions.go`. The allowed-tier and
  allowed-role string sets are **local constants in config** — the leaf may not
  import `internal/orchestration`; a cross-package unit test keeps the two sets
  in sync with the domain's `AllowedTiers()`/`AllowedRoleSlugs()`.

### Slice 3 — Wire the resolver into the single emission site

- **`cmd/centinela/hook_orchestration.go` — modify.** Load config
  (`config.Load()`; on error fall back to defaults — zero-config-safe). For each
  required role append `<role> (model: <tier>)` to the names list, and after the
  delegate line emit one
  `CENTINELA DIRECTIVE: model reference: <reference line>` built from the tiers in
  play. The command only calls `orchestration.ResolveModel` /
  `orchestration.ModelReference` — no decision logic in `cmd/` (G7). If the file
  exceeds 100 lines, factor a thin `cmd/centinela/orchestration_annotate.go`
  helper that still only delegates.

### Slice 4 — Acceptance + docs

- **`specs/configurable-subagent-models.feature` — new** (feature-specialist):
  the brief's 6 acceptance criteria as Gherkin scenarios.
- **`tests/acceptance/configurable_subagent_models_test.go` — new**: Gherkin-backed
  e2e driving the hook end-to-end.
- Document the out-of-band role defaults and the deferred (c) runner-signal path
  in the generated project docs at the docs step.

## Files Changed / Added

Added:
- `internal/orchestration/models.go` — `Tier`, defaults, normalization, accessors.
- `internal/orchestration/resolve.go` — `Runner`, per-runner table, `ResolveModel`,
  `ModelReference`.
- `internal/config/orchestration_models.go` — `validateOrchestrationModels`.
- `tests/unit/orchestration/models_test.go` — defaults, normalization, accessors.
- `tests/unit/orchestration/resolve_test.go` — override, exact IDs per runner,
  unknown-runner reference, missing-mapping fallback.
- `tests/unit/config/orchestration_models_test.go` — absent/empty → no error,
  valid mapping, unknown tier/role rejected, casing/whitespace normalized,
  allow-list parity with the domain.
- `tests/integration/hook_orchestration_models_test.go` — annotated directive +
  reference line; config-less run uses defaults.
- `tests/acceptance/configurable_subagent_models_test.go` — Gherkin e2e.
- `specs/configurable-subagent-models.feature` — written by feature-specialist.

Modified:
- `internal/config/orchestration.go` — `Models` field + accessor.
- `internal/config/file_size_exceptions.go` — call `validateOrchestrationModels`
  from `validateConfig`.
- `cmd/centinela/hook_orchestration.go` — load config, annotate roles, emit the
  model reference line (thin).

## Rollout Sequence

1. **Slice 1 — domain first.** Pure, no wiring, fully unit-testable; locks the
   runner-agnostic resolution and tier/default tables. Zero dependency on the
   rest.
2. **Slice 2 — leaf config.** Adds parsing + load-time validation; compiles
   independently of Slice 1 (uses local allow-list constants), with a parity test
   binding it to the domain sets.
3. **Slice 3 — single `cmd/` touch.** Trivially thin once Slices 1–2 exist; wires
   the resolver into the one emission site and threads `config.Load()`.
4. **Slice 4 — acceptance + docs.** Closes the 6 acceptance criteria and documents
   defaults + the deferred runner-signal follow-up.

## Risks & Mitigations

- **Runner unknown at emit time → wrong/Claude-only ID under opencode** (High).
  Strategy (a) is runner-agnostic by construction: per-role annotation carries the
  *tier*, IDs live in a both-runner reference line. The hook never resolves to a
  single runner-specific ID, so it cannot emit the wrong one.
- **Advisory only — orchestrator may ignore the hint** (Medium). Accepted by
  design, consistent with the directive model; the evidence-check enforcement is
  an explicit deferred follow-up.
- **Config↔domain allow-list drift** (Medium). Semantics live only in
  `internal/orchestration`; config holds the string allow-lists for leaf-safe
  validation; a cross-package test asserts the sets are identical.
- **Model-ID drift on releases** (Low/High likelihood). One internal per-runner
  table; tiers shield user config. A refresh edits one file.
- **opencode ID form wrong/untested** (Medium). Verify `anthropic/claude-*`
  against `opencode.json` conventions; a resolver test asserts the exact strings.
- **Missing tier→model mapping → hook crash** (Low). Resolver returns the tier
  name + `ok=false`; the hook warns and never aborts the directive.
- **Cross-layer leak (G2/G7)** (Low). All resolution in `internal/orchestration`;
  config stays leaf; `cmd/centinela/hook_orchestration.go` remains a thin
  orchestrator ≤100 lines.
