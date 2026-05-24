### Big-Thinker Report: configurable-subagent-models
**Date:** 2026-05-24

#### Problem

Centinela fans every feature out to ~7 step subagents (big-thinker,
feature-specialist, senior-engineer, ux-ui-specialist, qa-senior,
validation-specialist, documentation-specialist) and a handful of out-of-band
report agents. Today they all run on whatever model the host session is using —
typically an Opus-class model. The roles have wildly different cognitive loads:
architecture framing and cross-layer implementation want a reasoning-heavy
model, while "read a report and tabulate" is served identically by a fast, cheap
tier.

**Who is hurting:** developers/operators running Centinela who pay (tokens,
latency, rate-limit headroom) for a top-tier model on every subagent, with no
knob to right-size cost per role.

Centinela does not spawn subagents itself — the orchestrator (host session) does,
by reading a role prompt and calling the Agent tool. Centinela's only lever is
the directive it injects from `cmd/centinela/hook_orchestration.go:42`. So this
feature is deliberately **advisory**: the directive tells the orchestrator which
model tier to use per role, and the orchestrator complies the same way it
complies with the rest of the directive system. It is not hard-enforced like the
file-size gate, and the brief is honest about that.

#### Scope

**In (v1):**
- A `[orchestration.models]` TOML table mapping role slug → semantic tier
  (`reasoning | balanced | fast`).
- Smart per-role default tiers (brief table) so unconfigured roles keep working.
- Load-time validation: reject unknown role keys and invalid tier values with a
  precise error naming the offending key (mirroring `file_size_exceptions`).
- Tier normalization (case/whitespace) before validation.
- Domain resolution `role → tier (config|default) → model (per runner)` living
  in `internal/orchestration/`.
- Per-runner tier→model table (Claude Code + opencode) in one internal place.
- The orchestration directive annotated with each step role's resolved model,
  for the **7 directive-injected step roles only**.

**Out (v1, deferred — not sub-features of this slug):**
- Model selection for out-of-band agents (gatekeeper, production-readiness,
  edge-case-tester, merge-steward) — invoked from prompt files, not the directive
  hook, so they need a different injection point.
- Raw model-ID override escape hatch (rejected by design decision #1).
- Recording the model actually used in role evidence and validating it at
  `centinela complete` (the "advisory + evidence check" follow-up).
- Any evidence-schema change or new validation gate.

#### Dependencies & Assumptions

Respecting the n-tier layer rules (G2/G7): config is the leaf, resolution is
domain in `internal/orchestration/`, `cmd/` stays thin.

- **`internal/config/`** (leaf, no internal imports) — extend
  `OrchestrationConfig` with `Models map[string]string` (`toml:"models"`); add a
  validator hook in `validateConfig` that normalizes and rejects unknown
  roles/tiers. Config does NOT know the role enum or tier→model table; it
  validates against a flat list of allowed strings it owns or is handed, to avoid
  importing `internal/orchestration` (which would invert the dependency).
  **Decision:** keep the allowed-tier set (`reasoning|balanced|fast`) and the
  allowed-role set as local constants in config, because config is the leaf and
  may not import domain. The role/tier *semantics* (defaults, model mapping) live
  in `internal/orchestration`; only the *string allow-lists* are duplicated in
  config for load-time validation. This duplication is intentional and small;
  a unit test in both packages keeps them in sync.
- **`internal/orchestration/`** (domain) — new tier enum, role→default-tier
  table, per-runner tier→model table, and a `ResolveModel(role, cfgModels,
  runner)` resolver. `RequiredRolesForFeature` already enumerates the step roles
  to annotate; the resolver is called per role at emit time.
- **`cmd/centinela/hook_orchestration.go`** — the single emission site; it gains
  one call into the resolver and appends the annotation to the existing names
  list. No decision logic in `cmd/`.
- **Assumption:** the opencode model identifier forms in the brief
  (`anthropic/claude-*`) match `opencode.json` conventions. Flagged to
  feature-specialist/qa to verify against the opencode config before shipping.
- **Assumption:** `Config` is already passed (or trivially loadable) at the hook
  emit site; the hook currently does not load config, so the slice must thread
  `config.Load()` into `runHookOrchestration` (a thin wiring change, not logic).

#### Runner-Resolution Decision

The same `centinela hook orchestration` binary runs under BOTH runners and there
is **no runtime runner signal at emit time** — verified: the `--agent` flag
exists only on `init`/`migrate`/`setup` (wiring-time commands), and
`runHookOrchestration` reads only stdin (which it discards) and workflow state.
The opencode plugin and `.claude/settings.json` both invoke the identical
`centinela hook orchestration` command with no runner argument.

**Recommendation: (a) — emit the tier name in the directive plus a one-line
per-runner model reference, deferring the literal model-ID choice to the
orchestrator.**

Concretely, the directive annotates each role with its resolved **tier**
(`big-thinker (model: reasoning)`), and the hook emits a single compact reference
line mapping the tiers in play to their Claude Code and opencode model IDs (built
from the internal table). The orchestrator — which unambiguously knows whether it
is Claude Code or opencode — picks the correct ID from that reference.

**Why (a) over (b) and (c):**
- **(c) — inject a `--runner`/`CENTINELA_RUNNER` signal at wiring time** is the
  *correct long-term* answer but the *most work and the most blast radius*: it
  touches `internal/setup/hooks.go`, `settings_build.go`, the opencode plugin
  template (`opencode_plugin.go`), and every already-wired project would need a
  `centinela migrate` to gain the signal. Existing installs without the signal
  would still hit the no-runner case, so (a)'s fallback is needed regardless.
  This violates smallest-correct-slice-first.
- **(b) — emit both runner model strings inline per role** is noisier in the
  directive (two IDs × up to 7 roles) and couples the directive's verbosity to
  the runner table size. (a) achieves the same orchestrator outcome by factoring
  the runner table into one reference line, keyed by the small set of tiers
  actually in play, not per role.
- **(a)** keeps the hook **runner-agnostic by construction** — it can never emit
  a Claude-only ID under opencode because it emits the tier (always valid) and a
  both-runner reference. It is the smallest correct slice, requires no migrate,
  and leaves (c) available as a clean follow-up that would let the hook collapse
  the reference to a single ID once a runner signal exists.

The resolver is structured so the runner is a *parameter* (`ResolveModel(role,
models, runner)`), with `runner` defaulting to "unknown" → emit-both behavior.
Adding (c) later is then a one-line change at the call site (pass the detected
runner) with zero domain churn — the slice is forward-compatible with (c) without
paying for it now.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Runner unknown at emit time → a wrong/Claude-only ID leaks under opencode | High | High | Chosen strategy (a) is runner-agnostic by construction: emit tier + a both-runner reference line. The hook never resolves to a single runner-specific ID, so it cannot emit the wrong one. |
| Advisory only — orchestrator may ignore the model hint | Medium | Medium | Accepted by design, consistent with the entire directive model. Documented plainly; the "advisory + evidence check" enforcement is an explicit deferred follow-up. |
| Config↔domain allow-list drift (tier/role strings duplicated in config leaf and orchestration domain) | Medium | Medium | Single source of truth for *semantics* in `internal/orchestration`; config holds only the string allow-lists for load-time validation; a cross-package unit test asserts the two sets stay identical. |
| Hardcoded model IDs drift on model releases | Low | High | One internal per-runner table; tiers shield user config. A model refresh edits one file. |
| opencode model-ID form is wrong/untested | Medium | Medium | Verify the `anthropic/claude-*` forms against `opencode.json` conventions before shipping; cover with a resolver unit test asserting the exact opencode strings. |
| Scope creep into out-of-band agents | Medium | Medium | v1 covers only the 7 directive-injected step roles; the others' defaults are documented but not injected. |
| Missing tier→model mapping for a configured tier → hook crash | Low | Low | Resolver falls back to emitting the tier name and surfaces a warning; the hook never panics or aborts the directive. |
| Hook not loading config today → wiring regression | Low | Medium | Threading `config.Load()` into `runHookOrchestration` is a thin, well-covered change; on load error the hook degrades to defaults (zero-config-safe) rather than failing. |

#### Rollout

Smallest correct slice first. Every new/edited source file stays ≤100 lines (G1);
where a file would exceed budget, split into a sibling.

**Step 1 — Domain: tiers, defaults, and runner-agnostic resolution (no wiring).**
Create:
- `internal/orchestration/models.go` — `Tier` type + constants
  (`reasoning|balanced|fast`), `DefaultTierForRole(role Role) Tier` (brief
  table for the 7 step roles + documented out-of-band defaults), `NormalizeTier`,
  and `AllowedTiers()` / `AllowedRoleSlugs()` accessors. (≤100 lines; split the
  per-runner model table into `models_table.go` if needed.)
- `internal/orchestration/resolve.go` — `Runner` type (`claude|opencode|unknown`),
  the per-runner tier→model table, and
  `ResolveModel(role Role, models map[string]string, runner Runner) (modelOrTier string, ok bool)`
  plus a `ModelReference(tiers []Tier) string` helper that renders the compact
  both-runner reference line.
- `tests/unit/orchestration/models_test.go`, `resolve_test.go` — default tiers,
  config override, normalization, unknown-tier fallback, exact Claude/opencode
  IDs, unknown-runner → reference behavior.

**Step 2 — Config: parse + validate `[orchestration.models]`.**
Edit:
- `internal/config/orchestration.go` — add `Models map[string]string`
  (`toml:"models"`) to `OrchestrationConfig` and a small `OrchestrationModels(cfg)`
  accessor.
- Create `internal/config/orchestration_models.go` — `validateOrchestrationModels`
  (normalize tier, reject unknown role/tier with a precise message naming the
  key and listing allowed tiers), called from `validateConfig` in
  `file_size_exceptions.go`. Keep the allow-list constants local to config.
- `tests/unit/config/orchestration_models_test.go` — empty/absent table → no
  error; valid mapping; `"genius"` tier rejected with allowed tiers in the
  message; `"backend-wizard"` role rejected; `"Reasoning"`/`" fast "` normalized.

**Step 3 — Wire the resolver into the single emission site.**
Edit:
- `cmd/centinela/hook_orchestration.go` — load config (`config.Load()`; on error
  fall back to defaults), and for each required role append
  `<role> (model: <tier>)` to the names list; emit one
  `CENTINELA DIRECTIVE: model reference: <reference line>` after the delegate
  line. No decision logic added to `cmd/` — it only calls
  `orchestration.ResolveModel` / `orchestration.ModelReference`. Keep the file
  ≤100 lines; if the annotation loop grows, factor a tiny
  `cmd/centinela/orchestration_annotate.go` helper that still only delegates.
- `tests/integration/hook_orchestration_models_test.go` — drives the hook with a
  fixture config and asserts the annotated directive + reference line; asserts a
  config-less run uses defaults (zero-config-safe).

**Step 4 — Acceptance + docs handoff.**
- `specs/configurable-subagent-models.feature` (feature-specialist writes the
  Gherkin from the brief's 6 acceptance criteria).
- `tests/acceptance/configurable_subagent_models_test.go` — Gherkin-backed e2e.
- Document the out-of-band role defaults and the deferred (c) runner-signal path
  in the generated project docs at the docs step.

Rationale for the order: Step 1 is pure domain with zero wiring and is fully
unit-testable in isolation; it de-risks the runner-agnostic decision before any
config or hook change. Step 2 (leaf config) depends on the allow-list constants
conceptually but compiles independently. Step 3 is the only `cmd/` touch and is
trivially thin once Steps 1–2 exist. Step 4 closes acceptance + docs.

#### Handoff

**Next role:** feature-specialist.

Outstanding questions for feature-specialist:
1. Confirm the exact reference-line format (e.g.
   `reasoning → cc:claude-opus-4-7 / oc:anthropic/claude-opus-4-7`) so the
   Gherkin asserts a stable string.
2. Decide whether the directive annotates with the **tier** (`model: reasoning`)
   or the resolved Claude-Code ID under strategy (a). Recommendation: annotate
   with the tier and carry IDs in the reference line — keeps the per-role
   annotation runner-agnostic.
3. Verify the opencode model-ID forms against `opencode.json` conventions and
   lock them in the spec's tier→model scenario.
4. Confirm whether to also document (not inject) the out-of-band role defaults in
   v1 docs, per the brief's edge case.
