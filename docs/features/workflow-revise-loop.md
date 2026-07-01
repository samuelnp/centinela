# Feature: workflow-revise-loop

## Problem

Centinela's workflow is forward-only: a feature advances `plan → code →
tests → validate → docs` exclusively through `centinela complete`. There is
no legitimate path *backward*. But real work loops back: the `validate` step
routinely discovers a defect that requires re-doing `code`, or a spec gap
that requires re-doing `plan`. When that happens the driving LLM has no
enforcer-sanctioned move — so it bypasses the enforcer entirely, editing
source via raw `bash` (heredoc / `sed`) outside the step it is gated into.
That silently defeats the architecture rules, the per-step validators, and
the certification evidence the whole framework exists to guarantee.

## Who hurts

- **The agent**, which is forced into an illegitimate bypass to do correct
  work, and the **prewrite hook**, which can only block or be evaded.
- **The reviewer / merger**, who trusts that a `validate`-passed feature was
  certified on the *final* tree — when in fact post-validate hand-edits were
  never re-gated.
- **The audit trail**: a backward jump today leaves no record of why work was
  redone, how often, or what triggered it.

## Goal

Add `centinela revise <feature> --to <step> --reason "<why>"`: a controlled,
auditable, gate-preserving backward transition. Rewinding to an earlier step
re-opens every step after it (status → pending) and invalidates only their
**certification evidence** — never the user's source or test code — so the
next `centinela complete` is forced to re-run those gates on the corrected
tree. The forward guarantee ("all gates pass on the final tree") then falls
out automatically from re-using `Complete()`.

The four properties that keep the guarantee intact:
1. **Explicit + logged** — `--reason` is REQUIRED (friction against
   thrashing); every revision is appended to an audit log on the state.
2. **Downstream evidence invalidation** — rewinding to step X deletes the
   `.workflow/<feature>-<role>.{json,md}` reports + `-edge-cases.md` for the
   steps after X; it MUST NOT touch source/test code.
3. **Forced re-entry through gates** — re-opened steps are pending, so
   `complete` re-validates them before re-advancing.
4. **Right target** — `--to` must be a real step in *this* feature's
   `OrderedSteps()` and strictly *before* the current step.

## Scope

**In (v1):**
- `revise` command: backward transition + required `--reason`.
- Pure domain `RewindTo` on `*Workflow` (mirror of `Complete`).
- Evidence `Invalidate(feature, role)` primitive (delete `.json`+`.md`).
- `Revisions []Revision` audit field on the workflow state; render the
  revision count/history in `centinela status`.
- Telemetry `RecordRevised` sibling of `RecordStepAdvanced`.

**Out (v1) — deliberate decisions, not omissions:**
- **Revising a completed (`done`) workflow** is out of scope. A shipped
  feature reopens via a new follow-up feature, not a rewind. (`revise` errors
  when `CurrentStep == "done"`.)
- **Marking evidence `.stale` vs deleting it:** v1 **deletes**. Git history
  plus the `Revisions` audit log already preserve the full record, so a
  parallel `.stale` lifecycle would be redundant complexity.
- Forward-skip / arbitrary jumps (only strictly-backward is allowed).
- Auto-detecting *which* step to rewind to (the human/agent supplies `--to`).

## Success criteria

- An agent that hits a `validate` defect runs `revise --to code` instead of a
  raw-bash bypass; the next `complete` re-runs the code→validate gates.
- Downstream role evidence is gone after a rewind; source/test code is intact.
- Every rewind is visible in `centinela status` with its reason.
