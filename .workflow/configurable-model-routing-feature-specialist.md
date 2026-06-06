### Feature-Specialist Report: configurable-model-routing
**Date:** 2026-06-05

#### Behavior Summary

`configurable-model-routing` opens Centinela's hardcoded `tier → concrete model` table to user configuration and adds a role-level concrete override, both keyed by runner (`claude` | `opencode` | `codex`). A project adds `[orchestration.model_map.<tier>]` in `centinela.toml` to remap which model backs a whole tier for a given runner (e.g., all `reasoning`-tier roles use `moonshotai/kimi-k2` under opencode). For finer control a role entry in `[orchestration.models]` can now be either the existing plain tier string (back-compatible) or a runner→model table that pins that role to a specific concrete model, winning over the tier layer. The resolver follows a strict 4-step precedence: (1) role override, (2) tier-map override, (3) built-in default, (4) emit the tier name with `ok=false` and warn when no mapping exists for the active runner. Config load validates all shape constraints (known runner keys, known tier keys, known role keys, non-empty model strings) with normalized (trimmed, lowercased) keys and fails loudly with the offending key named. Every unconfigured role/tier/runner keeps the built-in Anthropic defaults unchanged. The directive emission stays runner-agnostic: the hook renders all three runners' resolved IDs on the reference line; the orchestrator picks its own row. This is advisory — no evidence-schema change, no gate at `centinela complete`.

#### Gherkin Scenarios

Full spec: `specs/configurable-model-routing.feature`

| Scenario | Maps to AC |
|----------|-----------|
| Tier remap resolves the correct model for the active runner | AC#1 |
| Role override beats the role's tier for the active runner | AC#2 |
| Role with a tier override but no model_map entry for the runner uses the built-in default | AC#3 |
| Plain tier string value in orchestration.models is accepted and behaves as before | AC#4 |
| Unknown runner key in model_map is rejected at config load time | AC#5 |
| Unknown role key in orchestration.models is rejected at config load time | AC#5 |
| Unknown tier key in model_map is rejected at config load time | AC#5 |
| Empty model string in model_map is rejected at config load time | AC#5 |
| Absent model_map and models tables — all roles resolve to built-in defaults | AC#6 |
| Active runner with no mapping emits tier name and warning instead of another runner's concrete ID | AC#7 |
| Tier key with uppercase and surrounding whitespace is normalized and accepted | Edge: casing/whitespace normalization |
| Runner key with uppercase and surrounding whitespace is normalized and accepted | Edge: casing/whitespace normalization |
| Mixed role value forms in orchestration.models are both valid | Edge: mixed forms |
| Role-level concrete override wins over a model_map entry for the same runner | Edge: role-override-beats-tier |
| Codex runner before codex-support lands falls back to tier name with ok=false | Edge: codex rule-4 fallback |
| Empty model_map and models tables behave identically to absent tables | Edge: empty-tables-same-as-absent |

#### UX States

n/a — This is a CLI/config feature with no UI surface. Resolution happens inside the orchestration hook at directive emit time; the only observable output is text injected into the delegate directive written to the terminal. There are no loading spinners, empty states, or interactive UI components.

#### Out-of-Scope

- Model selection for out-of-band agents (gatekeeper, production-readiness, edge-case-tester, merge-steward) — different injection point, not emitted by the directive hook in v1.
- Recording the model actually used in role evidence and validating it at `centinela complete` — deferred to a follow-up feature.
- Provider-availability / model-existence checks — the runner's responsibility; Centinela validates shape only (non-empty, known keys); opaque model strings are advisory.
- Codex's concrete default model IDs — filled by the `codex-support` (Phase 8) feature; this feature ships the codex column and runner key only.
- Any `CENTINELA_RUNNER` / `--runner` runtime signal or `internal/setup/` runner-detection changes — locked out by design decision #1 (emit all-runners reference line; orchestrator picks its row).
- Splitting the feature into `model-routing-tier-remap` + `model-routing-role-override` — locked by design decision #2 (ship as one feature).

#### Handoff

- Next role: senior-engineer
- Open clarifications:
  1. Confirm the exact TOML decoder interface for the union field (`UnmarshalTOML` on a wrapper type vs. decode-to-`map[string]any` then narrow) — pick the form that stays ≤100 lines per file and preserves plain-string back-compat; verify against `config.Load`'s current decoder wiring before coding.
  2. Finalize the `ResolveModel` signature for the union `models` argument — a typed `RoleModels` struct carrying an optional tier string and an optional `map[string]string` (runner → model) vs. two parallel maps; the typed struct is cleaner but confirm the config package can construct it without importing the domain.
  3. Decide whether per-role annotations on the delegate line show resolved IDs per runner inline (one role → three columns) or rely solely on the factored `ModelReference` line — locked decision #1 favors the factored line; confirm directive readability with a sample before coding.
