---
feature: configurable-model-routing
summary: Route each subagent and tier to the concrete model you want, per runner, in centinela.toml—choose any provider's model without editing Go.
audience: end-user
status: done
---

## What it does

Centinela's role-to-model selection was hardcoded to Anthropic's models. If you ran Centinela under opencode or codex—or routed Claude Code through a vendor gateway—you were stuck using Claude even if you wanted Kimi for reasoning or DeepSeek for coding. 

Now you can point any tier (reasoning, balanced, fast) at any concrete model ID, per runner. You can even override a single role to use a different model than its tier. All overrides live in `centinela.toml`. Every unconfigured tier and role keeps the built-in Anthropic defaults, so zero-config still works.

## When you'd use it

You use this when you run Centinela on **opencode or codex** (or Claude Code via a gateway like OpenRouter) and want to choose the best model for each tier or role based on price, speed, or quality for your provider. Without this feature, you were locked to the built-in Anthropic defaults. With it, you can say "use Kimi for reasoning and DeepSeek for coding, but keep the defaults for everything else."

## How it behaves

- **Tier remapping.** Set `[orchestration.model_map.<tier>.<runner>]` in `centinela.toml` to override which concrete model backs a tier for a specific runner. For example, `[orchestration.model_map.reasoning] opencode = "moonshotai/kimi-k2"` means any role that uses the reasoning tier resolves to Kimi when running on opencode.
- **Per-role override.** Set `[orchestration.models.<role>.<runner>]` to pin a specific role to a concrete model for a runner. This beats the tier remap. For example, `[orchestration.models] senior-engineer = { opencode = "deepseek/deepseek-coder" }` makes senior-engineer always use DeepSeek, even if reasoning is remapped to Kimi.
- **Back-compat with tier strings.** A role value in `[orchestration.models]` can still be a plain tier name like `"balanced"`—the old behavior is not broken.
- **Unconfigured roles fall back.** If a role has no override and no tier remap for the active runner, it uses the built-in default (Claude models for the claude runner, Anthropic-prefixed models for opencode).
- **Zero-config is safe.** If you don't add any `[orchestration.model_map]` or `[orchestration.models]` sections, every role resolves to its built-in default. No changes needed for existing projects.
- **Missing runner mapping.** If the active runner (codex, for example) has no mapping for a tier, the directive emits the tier name instead of a wrong vendor's ID and warns you that the runner has no model for that tier.
- **Malformed config fails loudly.** Unknown runner keys, unknown role names, unknown tiers, or empty model strings all fail when config loads, with an error naming exactly which key is wrong. Mistakes are caught before a run starts.

## Examples

Override the reasoning tier for opencode to use Kimi, and pin senior-engineer to DeepSeek on opencode while keeping other roles on their defaults:

```toml
[orchestration.model_map.reasoning]
opencode = "moonshotai/kimi-k2"

[orchestration.model_map.balanced]
opencode = "deepseek/deepseek-chat"

[orchestration.models]
senior-engineer = { opencode = "deepseek/deepseek-coder" }
qa-senior = "balanced"  # still accepts a plain tier string
```

When you run Centinela on opencode:
- `big-thinker` (defaults to reasoning tier) resolves to `moonshotai/kimi-k2`.
- `senior-engineer` (overridden) resolves to `deepseek/deepseek-coder`.
- `feature-specialist` (defaults to balanced tier, no override) resolves to `deepseek/deepseek-chat`.
- `qa-senior` (overridden to balanced tier, no model_map entry for opencode) resolves to the built-in default for balanced on opencode.

When you run on claude:
- All roles resolve to their built-in defaults because no mappings are defined for the claude runner.
- `qa-senior` resolves to the balanced tier (back-compat string form), then to `claude-sonnet-4-6`.
