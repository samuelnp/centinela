---
feature: configurable-subagent-models
summary: Pick a model tier (reasoning/balanced/fast) per subagent role in centinela.toml to right-size cost and latency.
audience: end-user
status: done
---

## What it does
Centinela delegates each feature step to subagent roles (planning architect, code engineer, test designer, docs writer, and others) — all running on whatever model your session uses. Now you can declare a semantic tier (reasoning for hard thinking, balanced for standard work, fast for summaries) per role in `centinela.toml`, and the orchestration hook will annotate each role with the tier you picked. Unconfigured roles fall back to sensible built-in defaults (planning and code get reasoning; testing and docs get balanced or fast). The directive tells the orchestrator which tier to use for each role so it can pick the right model for its runtime.

## When you'd use it
You want to right-size cost and latency without editing prompts or code. For example: run planning and engineering on a strong model (reasoning tier: Opus) because architecture framing is hard, but run docs generation and test tabulation on a cheap fast model (fast tier: Haiku) because summarizing a report is straightforward. No configuration means all roles use their defaults, so the feature is zero-setup — pick only the roles where you want to override the built-in tier.

## How it behaves
- You declare a `[orchestration.models]` table in `centinela.toml` with role name → tier pairs: `big-thinker = "reasoning"`, `documentation-specialist = "fast"`, etc. Tiers are normalized (case + whitespace ignored), so `"Reasoning"` and `" fast "` work fine.
- If a role is in the config, it uses the tier you specified. If a role is not mentioned, it uses its default tier (planning and code: reasoning; testing and UX: balanced; docs and validation: fast). If the entire `[orchestration.models]` table is absent or empty, all roles use defaults (the feature is zero-config-safe).
- Invalid tier names (e.g., `"genius"`) or unknown role names (e.g., `"backend-wizard"`) are caught at config load time with a clear error message naming the problem and listing the allowed tiers and roles.
- When the orchestration hook emits the plan-step directive, each role is annotated with the tier you configured or defaulted — e.g., `delegate to [big-thinker (model: reasoning), documentation-specialist (model: fast)]`. This tells the orchestrator which tier each role should run on.
- The directive also includes a model reference line showing both runners' actual model IDs for each tier, e.g.: `model reference: reasoning: claude-opus-4-7 (Claude Code) / anthropic/claude-opus-4-7 (opencode); balanced: claude-sonnet-4-6 (Claude Code) / anthropic/claude-sonnet-4-6 (opencode); fast: claude-haiku-4-5-20251001 (Claude Code) / anthropic/claude-haiku-4-5 (opencode)`. The orchestrator uses this to pick the right model ID for its runner.
- Out-of-band roles invoked from prompt files (gatekeeper, production-readiness, edge-case-tester, merge-steward) are not yet annotated in the directive. Their defaults are documented for future use; that's a follow-up feature.

## Examples
A simple `centinela.toml` override:

```toml
[orchestration.models]
big-thinker = "reasoning"
senior-engineer = "reasoning"
feature-specialist = "balanced"
qa-senior = "balanced"
ux-ui-specialist = "balanced"
documentation-specialist = "fast"
validation-specialist = "fast"
```

When the hook emits the directive, you'll see:

```
CENTINELA DIRECTIVE: orchestrator only for "x"/"plan"; delegate to
[big-thinker (model: reasoning), feature-specialist (model: balanced), 
 senior-engineer (model: reasoning), qa-senior (model: balanced), 
 ux-ui-specialist (model: balanced), documentation-specialist (model: fast), 
 validation-specialist (model: fast)].

model reference: reasoning: claude-opus-4-7 (Claude Code) / anthropic/claude-opus-4-7 (opencode); 
                 balanced: claude-sonnet-4-6 (Claude Code) / anthropic/claude-sonnet-4-6 (opencode); 
                 fast: claude-haiku-4-5-20251001 (Claude Code) / anthropic/claude-haiku-4-5 (opencode)
```

A minimal override (only override roles you care about):

```toml
[orchestration.models]
documentation-specialist = "fast"
# other roles use their defaults
```

An invalid config that will fail at load time with a clear error:

```toml
[orchestration.models]
qa-senior = "genius"      # ❌ Error: invalid tier "genius"; allowed: reasoning, balanced, fast
unknown-role = "balanced" # ❌ Error: unknown role "unknown-role"
```

## Notes
This feature is advisory by design — the tier annotation goes into the directive, and the orchestrator honors it the same way it honors the rest of the directive system. If an orchestrator ignores the hint (unlikely in practice), Centinela has no hard enforcement hook to prevent it. The model actually used is not recorded in evidence or validated at `centinela complete`.

Future enhancements (not yet shipped):
- A `CENTINELA_RUNNER` signal injected at hook-wiring time so the hook knows its runner at emit time and can emit runner-specific model IDs directly instead of a both-runner reference.
- Model selection for out-of-band agents (gatekeeper, production-readiness, edge-case-tester, merge-steward) through a different injection point, since they're invoked from prompt files, not the orchestration directive hook.
- Evidence recording: capturing the model actually used in each role's evidence file and validating it at `centinela complete`.
