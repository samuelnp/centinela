# workflow-archetypes — documentation-specialist

**Date:** 2026-06-12
**Handoff →** complete

## KB entry

Authored `docs/project-docs/kb/workflow-archetypes.md` — an operator-facing
knowledge-base entry for an operator choosing a workflow track for a piece of
work. Plain language throughout: no Go package/function names, no Given/When/Then
prose. Follows the enforcement-profiles house style (front matter + What it does /
When you'd use it / How it behaves / Examples).

**Summary line:** Named lightweight workflow tracks let you pick a right-sized
path for work that isn't a full feature — a hotfix, a refactor, or a throwaway
spike — instead of always running the full plan-to-docs workflow, while
verification still applies wherever the chosen track includes the ship step.

Coverage:
- **What it does** — pick a lightweight track for non-feature work; tracks reuse
  the same canonical steps, so verification still applies on the steps a track keeps.
- **When you'd use it** — hotfix (urgent bug: fix, test, ship — no design doc);
  refactor (restructure without behavior change: plan, change, prove, ship — no
  user docs); spike (timeboxed throwaway experiment: plan, code — no ship gate).
- **How it behaves** — one bullet per spec scenario: the four tracks and their
  step lists; canonical is the default (nothing changes unless you choose a track);
  selection via `centinela start --archetype <name>` or a roadmap entry, flag wins;
  the active track shows in `centinela status` (spike marked "no ship gate");
  validate-bearing tracks are ship-gated identically (gates + claim verification);
  spike has no validate step so it isn't ship-gated, but isn't a verification hole
  (gate is step-keyed, not name-keyed; promoted spike work is validated at merge);
  unknown track names are rejected; a track is independent of the strictness profile.
- **Examples** — `centinela start fix-login --archetype hotfix`,
  `centinela start probe-idea --archetype spike`, and the `centinela status` note.

## Generated outputs (all confirmed present)

| File | Status |
|------|--------|
| `docs/project-docs/kb/workflow-archetypes.md` | written (source KB entry) |
| `docs/project-docs/kb/workflow-archetypes.html` | generated |
| `docs/project-docs/kb/index.html` | regenerated (KB index) |
| `docs/project-docs/index.html` | regenerated (docs portal landing) |

`centinela docs validate` passes; `centinela docs generate --out
docs/project-docs/index.html` rendered the KB page, the KB index, and the portal
landing in one pass.

## Right-sizing the docs step

workflow-archetypes is an internal, operator-facing feature. The KB entry is the
load-bearing deliverable — it's what an operator reads to choose a track. The full
HTML portal regeneration is heavier than its reader value for an internal feature,
but it's a single command and the standard step output, so it was run as required
rather than expanded further. Mermaid diagram: skipped (no diagram adds reader
value over the four-row track table already in prose).

## Handoff → complete

Docs step artifacts complete: KB markdown + generated HTML (page, KB index,
portal), this report, and the evidence JSON. Ready for `centinela complete`.
