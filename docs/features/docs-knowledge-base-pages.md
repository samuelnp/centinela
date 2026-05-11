# Feature Brief: Per-Feature Knowledge Base in the Docs Step

## Problem

The docs step (5/5) currently produces only a workflow inventory at
`docs/project-docs/index.html` — feature counts, plan/spec lists, evidence
tables, state matrix. A non-technical reader cannot learn what any single
feature actually does. Specs (`.feature` files) hold the behavioral truth but
in Gherkin form, which is unfriendly for end-users.

## Goal

Extend the docs step so it also produces a plain-language knowledge base:
one HTML page per feature, written for Centinela end-users (non-tech), backed
by LLM-authored markdown narratives that the deterministic generator wraps
into the docs shell.

## Scope

- New artifact: `docs/project-docs/kb/<feature>.md` — LLM-authored narrative.
- New deterministic output: `docs/project-docs/kb/<feature>.html` per page,
  plus `docs/project-docs/kb/index.html` listing every feature.
- Extend `internal/docgen` with KB loader, types, and renderer.
- Add a "Knowledge Base" link to the main `index.html` nav.
- `centinela docs validate` requires the KB md + html for the feature
  currently in the docs step. Pre-existing features without KB md show a
  "guide not yet written" placeholder card in the KB index.
- Update `documentation-generator-prompt.md` so the agent writes the KB md as
  part of its contract.

## Non-Goals

- Backfilling KB md for every existing feature in this change. Only the
  feature passing through the docs step is required to ship its KB md.
- Changing the technical/workflow report. The existing `index.html`
  continues to serve maintainers.
- Multi-language output (use PROJECT.md locales only when applicable).
