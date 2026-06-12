# Documentation-Specialist Report: enforcement-profiles

**Date:** 2026-06-12
**Step:** docs (5/5)

## KB entry

Authored `docs/project-docs/kb/enforcement-profiles.md` (audience: end-user =
Centinela operator). It explains enforcement profiles as a per-project /
per-feature strictness preset that scales how much *process* Centinela enforces
(step-gating, the stop-and-ask review prompt, mandatory subagent evidence) while
every validate gate and all claim verification keep running unchanged under every
preset.

- **What it does** — pick one of three presets; the preset sets the amount of
  enforced process, but the outcome checks (gates + claim verification) always run.
- **When you'd use it** — strict for a small/local model needing maximum rails,
  guided for a capable model driving the steps itself, outcome for a strong model
  working fast in any order; per-project default plus a per-feature override.
- **How it behaves** — one bullet per spec scenario as observable behavior: three
  presets; strict is the default and preserves today's behavior exactly; outcome
  allows out-of-order writes and suppresses the inter-step prompt; an explicit
  confirmation mode still beats the preset; strict requires the full subagent
  evidence while guided/outcome do not; profile chosen in `centinela.toml` or via
  `centinela start --profile <name>` (per-feature wins); an unknown profile is
  rejected at config load; and the key guarantee — gates and claim verification
  run identically under every profile.
- **Examples** — the `[workflow] enforcement_profile = "guided"` toml snippet, the
  `centinela start my-feature --profile outcome` command, and a note that
  `centinela status` shows the active profile.

## Generated outputs (confirmed to exist)

- `docs/project-docs/kb/enforcement-profiles.md` — 3.5K (source KB entry)
- `docs/project-docs/kb/enforcement-profiles.html` — 7.2K (rendered)
- `docs/project-docs/kb/index.html` — 19.1K (KB index, regenerated)
- `docs/project-docs/index.html` — 130.3K (portal index, regenerated)

`centinela docs validate` → exit 0.

## Scope note (right-size-docs-step)

A full HTML-portal regeneration is heavier than its reader value for this internal
governance feature, so beyond the standard KB entry + index regen I deliberately
skipped extra portal authoring and the feature/spec-relationship Mermaid diagram —
the right-sized output here is the operator-facing KB entry, nothing more.
