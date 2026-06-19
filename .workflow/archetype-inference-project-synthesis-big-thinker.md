# archetype-inference-project-synthesis — big-thinker

## Problem

Brownfield adopters must hand-author `PROJECT.md` and guess an architecture
archetype; a wrong guess makes every downstream gate (G2/G7/gatekeeper paths)
wrong from day one. The shipped `deep-codebase-analysis` already emits a
deterministic `.workflow/analysis.json` Inventory, but nothing consumes it to
bootstrap config.

## Scope

In: `centinela synthesize` — infer the best-fit archetype (Hexagonal /
Rails-native / N-Tier / ECS / Modular) from the Inventory with a confidence +
rationale, and draft a complete `PROJECT.md` (archetype-specific Architecture
Choice/G2/G7/layer-mapping/gatekeeper-paths + inventory-derived tech-stack/
folder/locales/naming). Output is a draft the user confirms/corrects; never
clobber an existing `PROJECT.md` (write `PROJECT.draft.md`). Out: LLM refinement,
monorepo per-module PROJECT files, validate-command synthesis, auto-promotion.

## Dependencies & Assumptions

Consumes the shipped `analyze.Inventory` contract (add `analyze.Load` to the
contract owner). **Layer decision:** `internal/synthesize` is an **aggregator**
(imports the `internal/analyze` domain read-only), mirroring doctor/insights/
calibration/audit — a domain→domain edge would be forbidden, so aggregator is
the correct placement; no cycle (analyze never imports synthesize). centinela.toml
adds `internal/synthesize/**` to the aggregator layer.

## Risks

**Deterministic, no-LLM** is the key call — justified: matches the analyze
substrate's philosophy, is byte-stable and unit-testable offline, and the
confirm/correct model means the bar is "best-fit starting point", not perfection.
A transparent scoring table the user can audit beats opaque generation. The
`Inferer` is an interface so an LLM backend can drop in later without touching
`cmd/`. Mis-classification risk is mitigated by surfacing confidence + rationale
and writing a draft, never the canonical file.

## Rollout

Additive: a new aggregator package + a new read-only command + `analyze.Load`.
No change to existing behavior; `PROJECT.md` is never overwritten. The draft
carries a doc-version marker + DRAFT/confidence banner; unmatched layer slots and
human-only sections emit `<!-- TODO: confirm -->` rather than fabrications.

## Handoff

→ feature-specialist: file-by-file plan under ≤100 lines, the inference scoring
table + confidence/ambiguity rules, the archetype-specific synthesis templates,
and the Gherkin spec. Edge cases enumerated in the JSON `edgeCases`.
