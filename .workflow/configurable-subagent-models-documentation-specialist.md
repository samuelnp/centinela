# Documentation Specialist Report — configurable-subagent-models

**Feature:** configurable-subagent-models  
**Step:** docs  
**Role:** documentation-specialist  
**Status:** done  
**Date:** 2026-05-24

## Summary

Knowledge base entry written for end-users covering the new ability to configure model tiers per subagent role in centinela.toml. Entry follows the house format (feature YAML frontmatter + four body sections: What it does, When you'd use it, How it behaves, Examples, Notes). HTML generation completed successfully.

## Inputs Processed

- `docs/features/configurable-subagent-models.md` — feature brief (locked design, tier/role tables, user stories, 6 acceptance criteria, edge cases)
- `docs/plans/configurable-subagent-models.md` — implementation strategy (runner-agnostic directive emission, Slice 1–4 breakdown, risks/mitigations)
- `specs/configurable-subagent-models.feature` — 12 Gherkin scenarios (AC1–6 + 6 edge cases covering tier normalization, unknown roles, missing mappings, out-of-band roles, etc.)
- `docs/project-docs/kb/merge-steward-auto-dispatch.md` — house style reference (frontmatter, section structure, tone for end-users)

## Outputs Generated

1. **`docs/project-docs/kb/configurable-subagent-models.md`** — 5.4 KiB
   - Frontmatter: feature, summary (one sentence), audience (end-user), status (done)
   - **What it does** — semantic tiers (reasoning/balanced/fast) per role, unsset roles use defaults, directive annotates each role with the tier, both runners get model reference line
   - **When you'd use it** — right-sizing cost/latency without editing code
   - **How it behaves** — config table in centinela.toml, tier normalization, default fallback, zero-config-safe, validation errors on invalid tier/unknown role, out-of-band roles deferred
   - **Examples** — sample TOML block, annotated directive output, both-runner model reference line, invalid configs showing error format
   - **Notes** — advisory by design, future enhancements (CENTINELA_RUNNER signal, out-of-band agent model selection, evidence recording)

2. **`docs/project-docs/kb/configurable-subagent-models.html`** — 8.4 KiB (auto-rendered from markdown)

3. **`docs/project-docs/index.html`** — 89 KiB (main docs index, includes KB entry in navigation)

## Scenarios Documented

All 12 feature scenarios rewritten as user-visible behavior:

1. Configured tier is honored in the directive annotation
2. Unconfigured role falls back to its default tier
3. Absent `[orchestration.models]` table → all defaults (zero-config-safe)
4. Invalid tier value rejected at config load with precise error
5. Unknown role key rejected at config load with precise error
6. Directive is runner-agnostic (tier name + both-runner model reference line)
7. Empty `[orchestration.models]` table behaves like absent
8. Tier value normalized (uppercase/whitespace) before validation
9. Whitespace-only tier values normalized and accepted
10. Invalid value after normalization rejected with error
11. Missing tier→model mapping falls back to tier name with warning (no crash)
12. Out-of-band roles (gatekeeper, production-readiness, edge-case-tester, merge-steward) not annotated in directive

## Verification

- ✅ KB markdown entry written at `/Users/samuelnp/projects/personal/centinela/docs/project-docs/kb/configurable-subagent-models.md`
- ✅ HTML rendered at `/Users/samuelnp/projects/personal/centinela/docs/project-docs/kb/configurable-subagent-models.html`
- ✅ Main docs index regenerated at `/Users/samuelnp/projects/personal/centinela/docs/project-docs/index.html`
- ✅ `./centinela docs generate` completed without errors
- ✅ House format matched (YAML frontmatter, plain-language sections, TOML examples, model reference line shown)
- ✅ End-user audience (no Gherkin Given/When/Then, no internal jargon, clear use-case triggers)
- ✅ Future deferred work documented (CENTINELA_RUNNER signal, out-of-band agent model selection, evidence recording)

## Handoff

Ready for `centinela complete configurable-subagent-models`.
