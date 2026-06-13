# Feature: right-size-docs-step

- surface: internal
- status: planned
- roadmap: Phase 4 — Loop Velocity
- fixes: mandatory full docs ceremony runs on every feature including internal ones with no end-user story, burning tokens for zero reader value

## Problem

Every feature's `docs` step demands the same heavy output regardless of whether
anyone will ever read it: a plain-language KB guide (`kb/<feature>.md` + `.html`),
a full ~130 KB `index.html` portal regeneration, and a documentation-specialist
subagent evidence ceremony (`.md` + `.json`). That is right for a user-facing
feature with an end-user story. It is pure waste for an internal refactor,
bugfix, or chore — and this session proved it: every internal feature shipped
this session had its documentation-specialist flag that the portal regen was
heavier than its reader value. The docs step is the single biggest
ceremony-to-value mismatch left in the workflow.

## The idea: make the docs step surface-aware

The `code` step is already surface-aware — it requires the `ux-ui-specialist`
subagent ONLY for features whose brief declares `surface: user-facing`
(`RequiredRolesForFeature` gates it on `IsUserFacingFeature`). This feature
mirrors that exact mechanism for the `docs` step:

| Surface | KB guide (`kb/<f>.md`+`.html`) | Portal `index.html` per feature | documentation-specialist evidence | Required instead |
|---------|:---:|:---:|:---:|---|
| **user-facing** | required | required | required | — (unchanged) |
| **internal** | skipped | skipped | skipped | a one-line changelog entry |

- A **user-facing** feature's docs step is unchanged — full KB guide, portal,
  evidence. The KB guide is the one genuinely reader- and memory-useful artifact
  and it stays mandatory where it has a reader.
- An **internal** feature's docs step instead requires a single one-line
  changelog entry (`.workflow/<feature>-changelog.md`) and skips the KB guide,
  the per-feature portal regen, and the documentation-specialist ceremony.
- The portal stays current without per-feature regen by **regenerating at merge
  time** (a successful `centinela merge` refreshes `index.html` + `kb/*.html`),
  so the 130 KB rebuild happens once per delivery instead of once per feature.

## Surface detection (reuses existing parser)

`orchestration.IsUserFacingFeature(feature)` reads `docs/features/<feature>.md`
for `surface: user-facing`. Absence ⇒ internal — identical to how the code step
already treats it for ux-ui gating. So a feature opts into the heavy docs path by
declaring `surface: user-facing` (which user-facing features already declare for
the code step), and everything else gets the light path automatically.

## Goal

- `validateDocsOutput` and `RequiredRolesForFeature("docs")` become
  surface-conditional, mirroring the code step.
- An internal feature's docs step passes by producing a one-line changelog
  artifact instead of the KB/portal/evidence bundle.
- Portal `index.html` regeneration moves to merge time so it is not paid
  per-feature.
- Zero change for features that declare `surface: user-facing`.

## Non-goals (v1)

- **No full changelog automation.** The roadmap pairs this with
  `delivery-artifact-generation` (Phase 10, not built). v1 only requires a
  one-line `.workflow/<feature>-changelog.md`; assembling CHANGELOG.md from those
  entries is later work.
- **No new surface values.** Only the existing user-facing / internal split.
- **No change to gates or claim verification.** The docs step has no gate; this
  only changes which docs *artifacts* are required.
- **No removal of `centinela docs generate`.** It still exists and runs for
  user-facing features and at merge time; only its per-internal-feature
  obligation is dropped.
