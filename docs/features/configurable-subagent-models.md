# Feature: Configurable Subagent Models

## Problem

Centinela fans a feature out to ~7 step subagents (big-thinker, feature-specialist,
senior-engineer, ux-ui-specialist, qa-senior, validation-specialist,
documentation-specialist) plus out-of-band report agents (gatekeeper,
production-readiness, edge-case-tester, merge-steward). Today every one of them
runs on whatever model the session happens to be using. The roles have very
different cognitive loads — architecture framing and hard implementation want a
reasoning-heavy model, while "read a report and summarize" or "run the validate
commands and tabulate" are well served by a fast, cheap model.

**Who is hurting:** developers/operators running Centinela who pay (in tokens,
latency, or rate-limit headroom) for an Opus-class model on every subagent even
when a cheaper tier would do the job identically. There is no knob to right-size
the model per role.

Centinela does **not** spawn subagents itself — the orchestrator (the host
session) does, by reading a role prompt and calling the Agent tool. Centinela's
only lever is the directive it injects from
`cmd/centinela/hook_orchestration.go:42`
(`CENTINELA DIRECTIVE: orchestrator only for …; delegate to [roles]`). So this
feature is **advisory**: it tells the orchestrator which model tier to use per
role; the orchestrator complies, exactly like the rest of the directive system.
It is not hard-enforced the way the file-size gate is, and the brief is honest
about that.

## Outcome

A project can declare a model **tier** per subagent role in `centinela.toml`.
Unconfigured roles fall back to sensible built-in defaults. When the
orchestration hook emits its delegate directive, each role is annotated with the
model it should run on, resolved for the active runner (Claude Code or opencode).

```toml
[orchestration.models]
big-thinker        = "reasoning"
feature-specialist = "balanced"
documentation-specialist = "fast"
```

Resulting directive (illustrative):

```
CENTINELA DIRECTIVE: orchestrator only for "x"/"plan"; delegate to
[big-thinker (model: reasoning), feature-specialist (model: balanced)].
```

## Design Decisions (locked with the user)

1. **Semantic tiers only** as the config interface — `reasoning` | `balanced` |
   `fast`. No raw model IDs in user config (they go stale on every model release
   and differ per runner). The tier→model mapping is an internal table.
2. **Smart per-role defaults** — unconfigured roles get a built-in tier (table
   below). Users override only what they care about.
3. **Both runners (parity)** — the tier→model table is defined per runner so
   `.claude` and `.opencode` both resolve to a valid model identifier.
4. **Advisory directive** — the resolved model is appended to the existing
   delegate directive. No evidence-schema change, no validation at
   `centinela complete`.

### Default role → tier

| Role | Default tier | Rationale |
|------|--------------|-----------|
| big-thinker | reasoning | architecture framing, scope, risk |
| senior-engineer | reasoning | hard implementation across layers |
| feature-specialist | balanced | structured spec/Gherkin authoring |
| qa-senior | balanced | test design, bounded |
| ux-ui-specialist | balanced | UX states, bounded |
| documentation-specialist | fast | summarize + render docs |
| validation-specialist | fast | run commands + tabulate gate report |
| gatekeeper | fast | read-and-report (out-of-band; see Decomposition) |
| edge-case-tester | fast | enumerate edge cases (out-of-band) |
| merge-steward | reasoning | conflict resolution is delicate (out-of-band) |

### Tier → model mapping (built-in, per runner)

| Tier | Claude Code | opencode |
|------|-------------|----------|
| reasoning | `claude-opus-4-7` | `anthropic/claude-opus-4-7` |
| balanced | `claude-sonnet-4-6` | `anthropic/claude-sonnet-4-6` |
| fast | `claude-haiku-4-5-20251001` | `anthropic/claude-haiku-4-5` |

These IDs live in one internal table so a model refresh touches one place.

## User Stories

- As a developer, I want to set a model tier per subagent role in
  `centinela.toml` so I can right-size cost/latency without editing prompts.
- As a developer, I want roles I don't configure to keep working on a sensible
  default so I never have to enumerate all of them.
- As an opencode user, I want the same tier config to resolve to a valid
  opencode model identifier so the feature is not Claude-Code-only.
- As an operator, I want a malformed tier or unknown role name to fail loudly at
  config-load time, not silently fall back, so misconfiguration is caught early.

## Acceptance Criteria (→ Gherkin scenarios in `specs/configurable-subagent-models.feature`)

1. Given a `[orchestration.models]` table mapping `big-thinker = "reasoning"`,
   when the orchestration hook emits the plan-step directive, then the
   `big-thinker` entry is annotated with the resolved `reasoning`-tier model.
2. Given a role with **no** entry in `[orchestration.models]`, when the directive
   is emitted, then that role is annotated with its **default** tier's model.
3. Given an **absent** `[orchestration.models]` table, when the directive is
   emitted, then every role uses its default tier (feature is zero-config-safe).
4. Given an **invalid tier** value (e.g. `"genius"`), when config loads, then
   loading fails with an error naming the offending key and the allowed tiers.
5. Given an **unknown role** key (e.g. `"backend-wizard"`), when config loads,
   then loading fails with an error naming the offending key.
6. Given the active runner is opencode, when a `reasoning` tier resolves, then
   the emitted model identifier is the opencode form, not the Claude Code form.

## Edge Cases

- Empty `[orchestration.models]` table → all defaults apply (same as absent).
- Tier value with different casing/whitespace (`"Reasoning"`, `" fast "`) →
  normalized, then validated.
- Unknown role key → reject at load with a clear message (do not silently drop).
- Runner cannot be determined at hook-emit time (see Risks/Integration) → emit a
  runner-agnostic annotation (tier name and/or both model strings) rather than
  guessing wrong; never emit a Claude-only ID under opencode.
- Out-of-band roles (gatekeeper, production-readiness, edge-case-tester,
  merge-steward) are NOT emitted by the orchestration directive hook — their
  defaults are documented for the orchestrator but not injected in v1 (see
  Decomposition).
- A role configured to a tier whose model mapping is missing → fall back to the
  tier name in the directive and surface a warning; never crash the hook.

## Data Model

No persisted runtime entities. Pure configuration + resolution:

- `OrchestrationConfig.Models map[string]string` (`toml:"models"`) — role slug →
  tier name, parsed in `internal/config/` (leaf layer, no internal imports).
- Tier enum + role→default-tier table + per-runner tier→model table live in
  `internal/orchestration/` (domain). Resolution function:
  `role → tier (config or default) → model (runner)`.
- `cmd/centinela/hook_orchestration.go` stays thin: it calls the resolver and
  appends the annotation to the directive string. No decision logic in `cmd/`.

## Integration Points

- **`internal/config/`** — extend `OrchestrationConfig`; add validation
  mirroring the existing `file_size_exceptions` validator (reject unknown
  tier/role with a precise error).
- **`internal/orchestration/`** — new tier/model resolution (domain logic);
  `RequiredRolesForFeature` already enumerates the step roles to annotate.
- **`cmd/centinela/hook_orchestration.go`** — the single emission site.
- **Runner identity** — the same `centinela hook orchestration` binary runs
  under BOTH runners; there is currently no runtime runner signal at emit time
  (only an `--agent` flag on `init`/`migrate`/`setup`). Resolving how the hook
  learns its runner is the central design question (see Risks). Options for
  big-thinker: (a) emit tier name only + ship a tier→model reference for the
  orchestrator; (b) emit both runner model strings and let the orchestrator
  pick; (c) inject a `--runner`/`CENTINELA_RUNNER` signal at hook-wiring time in
  `internal/setup/` (both `.claude/settings.json` and the opencode plugin).
- **`.claude/settings.json` + `.opencode/plugins/centinela.js`** — touched only
  if option (c) is chosen.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Runner unknown at emit time → wrong/Claude-only ID under opencode | High | High | Pick a runner-agnostic emission strategy OR inject a runner signal at wiring time; never guess. Big-thinker decides. |
| Advisory only — orchestrator may ignore the model hint | Medium | Medium | Accept by design (consistent with directive model); document clearly; revisit evidence-check later if needed. |
| Hardcoded model IDs drift on model releases | Low | High | Single internal mapping table; tiers shield user config from churn. |
| Scope creep into out-of-band agents | Medium | Medium | v1 covers the 7 directive-injected step roles only; others documented, deferred. |
| opencode model identifier form is wrong/untested | Medium | Medium | Validate the opencode mapping against `opencode.json` conventions before shipping. |

## Decomposition

Single feature is appropriately sized for v1. Explicitly deferred to follow-ups
(not sub-features of this slug):

- Model selection for out-of-band agents (gatekeeper, production-readiness,
  edge-case-tester, merge-steward) — they are invoked from prompt files, not the
  orchestration directive hook, so they need a different injection point.
- Raw model-ID override escape hatch (rejected for v1 per design decision #1).
- Recording the model actually used in role evidence + validating it at
  `centinela complete` (the "advisory + evidence check" option, deferred).
