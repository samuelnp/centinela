### Feature-Specialist Report: configurable-subagent-models
**Date:** 2026-05-24

#### Behavior Summary

The `configurable-subagent-models` feature adds an optional `[orchestration.models]` TOML table to `centinela.toml` that maps any of the seven directive-injected step-role slugs (`big-thinker`, `feature-specialist`, `senior-engineer`, `ux-ui-specialist`, `qa-senior`, `validation-specialist`, `documentation-specialist`) to one of three semantic tier strings (`reasoning`, `balanced`, `fast`). At config-load time the validator trims and lowercases each tier value, then rejects unknown role keys or invalid tier strings with a precise error that names the offending key and lists the allowed values — mirroring the existing `file_size_exceptions` validator. Unconfigured roles fall back to built-in default tiers (e.g. `big-thinker` → `reasoning`, `documentation-specialist` → `fast`); a project with no `[orchestration.models]` section at all is therefore zero-config-safe. When the orchestration hook emits its delegate directive it annotates each step role with `<role> (model: <tier>)` — where `<tier>` is always the tier name, never a raw runner-specific model ID — and appends one compact `CENTINELA DIRECTIVE: model reference:` line that maps every tier in play to both its Claude Code model ID and its opencode model ID (e.g. `reasoning → cc:claude-opus-4-7 / oc:anthropic/claude-opus-4-7`). This makes the emission runner-agnostic by construction: the orchestrator, which unambiguously knows its own runner, reads the tier annotation and picks the correct ID from the reference line; the hook never emits a Claude-only bare ID that could be invalid under opencode. Out-of-band agents (gatekeeper, production-readiness, edge-case-tester, merge-steward) are deliberately excluded from the directive annotation in v1 because they are invoked from prompt files, not the orchestration hook.

#### Gherkin Scenarios

All scenarios are defined in `specs/configurable-subagent-models.feature`.

| # | Scenario | Maps to |
|---|----------|---------|
| 1 | Configured tier is reflected in the orchestration directive | AC1 |
| 2 | Unconfigured role falls back to its default tier | AC2 |
| 3 | Absent orchestration.models table — all defaults apply | AC3 |
| 4 | Invalid tier value is rejected at config load time | AC4 |
| 5 | Unknown role key is rejected at config load time | AC5 |
| 6 | Directive is runner-agnostic — emits tier name and both-runner reference | AC6 |
| 7 | Empty orchestration.models table — all defaults apply | Edge: empty table |
| 8 | Tier value with uppercase is normalized and accepted | Edge: "Reasoning" |
| 9 | Tier value with surrounding whitespace is normalized and accepted | Edge: " fast " |
| 10 | Tier value that is still invalid after normalization is rejected | Edge: " Genius " |
| 11 | Missing internal tier-to-model mapping falls back to tier name without crashing | Edge: missing mapping |
| 12 | Out-of-band roles are not annotated in the orchestration directive | Edge: out-of-band roles |

#### UX States

| State | Status |
|-------|--------|
| All UX states | n/a — this feature has no UI surface; it is pure CLI config + hook output |

#### Out-of-Scope

- Model selection for out-of-band agents (gatekeeper, production-readiness, edge-case-tester, merge-steward) — they are invoked from prompt files, not the orchestration directive hook, and need a different injection point (deferred per brief Decomposition).
- Raw model-ID override escape hatch — rejected by design decision #1 (semantic tiers only in user config).
- Recording the model actually used in role evidence and validating it at `centinela complete` — the "advisory + evidence check" follow-up is explicitly deferred.
- Any evidence-schema change or new validation gate tied to model compliance.
- Injecting a `--runner` / `CENTINELA_RUNNER` signal at hook-wiring time (strategy (c)) — deferred as the correct long-term follow-up once strategy (a) ships.

#### Handoff

**Next role:** senior-engineer

**Open clarifications resolved by this spec:**

1. The directive annotates with the **tier name** (`model: reasoning`), not a raw runner-specific ID. IDs live exclusively in the one-line both-runner model reference (`CENTINELA DIRECTIVE: model reference: reasoning → cc:claude-opus-4-7 / oc:anthropic/claude-opus-4-7`). This is locked per strategy (a) and is reflected in AC6 / Scenario 6.
2. The reference-line format is: `reasoning → cc:claude-opus-4-7 / oc:anthropic/claude-opus-4-7` (for each tier in play, space-separated). Scenario 6 asserts both the Claude Code ID and the opencode ID for the `reasoning` tier.
3. The opencode model-ID forms (`anthropic/claude-opus-4-7`, `anthropic/claude-sonnet-4-6`, `anthropic/claude-haiku-4-5`) follow the `anthropic/<slug>` convention from `opencode.json`. The resolver unit tests (Slice 1) must assert the exact strings to catch any drift before shipping.
4. Out-of-band role defaults are documented in the brief's default-tier table and will be documented (not injected) at the docs step, per the brief's edge-case note.
