### Feature-Specialist Report: workflow-revise-loop
**Date:** 2026-06-30

#### Behavior Summary

`centinela revise <feature> --to <step> --reason "<why>"` provides a
controlled, auditable backward transition in the otherwise forward-only
workflow. When an agent hits a defect at `validate` (or a spec gap at `plan`),
it runs `revise --to code` (or `revise --to plan`) rather than bypassing the
enforcer via raw bash. The command validates four pre-conditions — the target
is a real step in the feature's own `OrderedSteps()`, it is strictly before
the current step, the workflow is not `done`, and `--reason` is non-empty —
then re-opens every downstream step to `pending` (clearing their
`CompletedAt`), sets the target step to `in-progress`, appends a `Revision`
entry to the audit log, and deletes the certification evidence
(`.workflow/<feature>-<role>.{json,md}` and `-edge-cases.md`) for all
re-opened steps. Source code, test files, and documentation files are never
touched. The next `centinela complete` is forced to re-run the gates on the
corrected tree because the deleted evidence is again reported missing by
`validateOrchestration`. Every rewind is visible in `centinela status` as a
revision count plus the most recent reason — the designed friction against
thrashing. Because this feature is a pure CLI domain feature (no UI surface),
`ux-ui-specialist` is excluded from the `code`-step invalidation role set;
re-opening `code` invalidates only `senior-engineer`.

#### Gherkin Scenarios

See `specs/workflow-revise-loop.feature` for the full executable spec.
Scenarios covered:

1. **Happy path** — `validate → code` rewind: current step becomes `code`,
   downstream evidence (gatekeeper, validation-specialist, edge-cases) is
   deleted, senior-engineer evidence and source files survive, a Revision
   entry is appended.
2. **Re-gating (blocked)** — after rewind, `centinela complete` at the
   re-opened step exits non-zero when evidence is missing.
3. **Re-gating (unblocked)** — `centinela complete` advances once evidence
   is regenerated and gates pass.
4. **Missing `--reason` flag** — command exits non-zero with no state change.
5. **Empty/whitespace `--reason`** — rejected after trim, no state change.
6. **Forward target** — rejected with a clear error naming the direction
   violation.
7. **Equal target** — rejected as not strictly backward.
8. **Unknown step name** — rejected, error names the bad value.
9. **Done workflow** — rejected, no state change.
10. **Audit visibility** — `centinela status` shows revision count and latest
    reason inline.
11. **Archetype-awareness** — hotfix order (`code,tests,validate`) is
    respected; no plan or docs step appears in the re-opened set.
12. **Safety** — source, test, and docs files survive the rewind unchanged;
    only `.workflow/<feature>-*` files are removed.
13. **Idempotency** — already-absent evidence produces no error.
14. **Accumulation** — multiple rewinds produce multiple `Revision` entries
    (append-only, not overwritten).
15. **Internal feature code invalidation** — `ux-ui-specialist` is not
    referenced; only `senior-engineer` is invalidated.
16. **CompletedAt cleared** — re-opened steps have their `CompletedAt`
    field set to null.

#### UX States

| State   | Trigger                                              | Surface |
|---------|------------------------------------------------------|---------|
| n/a     | This feature has no UI surface (pure CLI domain)     | n/a     |

All UX states are n/a — `centinela revise` is an internal CLI command with
no web or TUI interface. Status render (`centinela status`) is already tested
in the audit scenario; it is not a new UI surface.

#### Edge Cases

- Empty or whitespace-only `--reason` is rejected before any state mutation.
- Target equal to current step is rejected (backward-only).
- Target after current step is rejected (backward-only).
- Unknown step name is rejected with a clear error naming the value.
- Done workflow is rejected before any state mutation.
- Idempotent invalidation: already-absent evidence files produce no error.
- Non-canonical archetype step order (hotfix: code,tests,validate) is
  respected.
- Source, test, and docs files are never deleted by invalidation.
- Multiple rewinds accumulate as separate Revision entries (not overwritten).
- Re-opened steps have CompletedAt cleared to null.
- Internal feature code-step invalidation excludes ux-ui-specialist; only
  senior-engineer is invalidated.
- Re-opened steps are set to pending and complete blocks re-advance until
  evidence is regenerated.

#### Out-of-Scope

- **Revising a `done` workflow.** A shipped feature reopens via a new
  follow-up feature. `RewindTo` errors when `CurrentStep == "done"`.
- **Marking evidence `.stale` vs. deleting it.** v1 deletes. Git history
  and the `Revisions` audit log already preserve the full record; a parallel
  `.stale` lifecycle is redundant complexity.
- **Forward skip or arbitrary jump.** Only strictly-backward transitions
  are allowed.
- **Auto-detecting the target step.** The human/agent supplies `--to`
  explicitly.
- **`ux-ui-specialist` invalidation on `code` step.** This is an internal
  CLI feature; `RequiredRolesForFeature` does not include `ux-ui-specialist`.
  Re-opening `code` invalidates `senior-engineer` only. (Decision handed to
  feature-specialist by big-thinker; pinned here and in spec scenarios.)
- **Full revision list in `status` output.** v1 renders the count and the
  latest reason inline; a full `--verbose` list is a deliberate future slice.
- **A `--dry-run` flag** showing what would be deleted without mutating
  state.
- **Interactive confirmation prompt** before deletion.

#### Deferred Findings

none — all boundary calls (no-rewind-when-done, delete-not-stale,
no-ux-on-internal-code-step) are deliberate v1 decisions pre-agreed in the
big-thinker report, not new gaps discovered during feature-specialist work.

#### Handoff

- Next role: senior-engineer
- Open clarifications: none — the two big-thinker questions are resolved:
  (1) `code`-step invalidation for internal features = `senior-engineer`
  only (no `ux-ui-specialist`); (2) `status` shows count + latest reason
  inline, full list deferred to a future `--verbose` slice.

