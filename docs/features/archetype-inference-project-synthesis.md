# Feature: archetype-inference-project-synthesis

**Phase:** Phase 9 — Brownfield Onboarding
**Archetype:** n-tier
**Depends on:** deep-codebase-analysis (shipped)
**Status:** plan

## Problem

A brownfield team adopting Centinela must hand-author `PROJECT.md` from a blank
template and *guess* an architecture archetype. A wrong guess makes every
downstream gate (G2 layer boundaries, G7 outer layer, gatekeeper paths) wrong
from day one. The shipped `deep-codebase-analysis` already emits a deterministic
`.workflow/analysis.json` Inventory (languages, manifests+frameworks, locales,
package layout, dependency graph) — but nothing consumes it to bootstrap config.

## User value

`centinela synthesize` reads that Inventory, **infers the best-fit archetype**
(Hexagonal / Rails-native / N-Tier / ECS / Modular) with a confidence level and
human-readable rationale, and **drafts a complete `PROJECT.md`** reflecting the
code as it actually is — Architecture Choice, G2/G7 rules, layer mapping
(derived from packages + graph), folder structure, locales, naming conventions,
gatekeeper paths. The user reviews and corrects a draft instead of authoring
from scratch.

## What ships

- A new **aggregator** package `internal/synthesize`: a swappable deterministic
  inference engine (per-archetype signal scoring over the Inventory) + an
  archetype-specific `PROJECT.md` synthesizer (template fill, no LLM).
- `analyze.Load(path)` on the contract owner (mirrors `analyze.Save`).
- A `centinela synthesize` command (mirrors `centinela analyze`) with
  `--in`/`--out`/`--json`.
- `internal/ui/render_synthesize.go` for the inference summary.

## Key decisions

- **Deterministic, no LLM** — matches the analyze substrate's philosophy,
  byte-stable, unit-testable offline; the bar is "best-fit starting point", the
  user confirms/corrects. The inferer is an interface so an LLM backend can drop
  in later without touching `cmd/`.
- **Never clobber** — when `PROJECT.md` exists, write `PROJECT.draft.md`.
- **Honest gaps** — sections that need human input (Elevator Pitch, Domain
  Language) or unmatched layer slots emit `<!-- TODO: confirm -->`, not guesses.

## Explicitly deferred

LLM refinement; monorepo per-module PROJECT files; `centinela.toml`
validate-command synthesis; auto-promoting the draft to `PROJECT.md`.
