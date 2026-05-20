---
feature: merge-steward-auto-dispatch
summary: When a parallel-worktree merge needs human-quality judgement, Centinela now dispatches the Merge Steward for you and refuses to finalize the merge until valid evidence comes back.
audience: end-user
status: done
---

## What it does
When `centinela merge <feature>` hits a real text conflict, or when validation fails on the merged tree, Centinela no longer stops at a printed hint. It records a small pending-merge marker, emits a structured directive that tells the orchestrating Claude session to run the Merge Steward agent with the right prompt and inputs, and keeps re-surfacing that directive on every prompt until the agent writes back a verdict. You then run `centinela merge --continue <feature>`; only a valid APPLY verdict on a clean main checkout finalizes the merge, and only after the existing evidence validator has accepted the steward's JSON. Anything else — escalation, missing evidence, malformed evidence, or a dirty main — keeps the worktree and marker in place and exits non-zero, so a broken or unverified resolution can never silently land on `main`.

## When you'd use it
You'll feel this whenever you merge a feature back from its parallel worktree (see the [parallel-feature-worktrees](parallel-feature-worktrees.md) guide for that setup). On clean merges nothing changes — Centinela still removes the worktree and exits zero. But the moment a merge gets messy, you no longer need to notice a printed hint and copy-paste the Merge Steward prompt yourself: the dispatch happens automatically, the verdict gates the finalize, and your only job is to review the proposed diff before it lands.

## How it behaves
- A clean merge — no text conflict, validation passes on the merged tree — still removes the worktree and exits zero, exactly like before. No marker is written and no Merge Steward directive is emitted.
- When git reports a text conflict, Centinela writes a pending marker recording the reason (`git-text-conflict`), the conflicted file paths, and the worktree path. It prints a `CENTINELA DIRECTIVE` block naming the merge-steward prompt, the feature, and the `centinela merge --continue` resume command. The worktree is kept and the command exits non-zero.
- When git merges cleanly but `centinela validate` then fails on the merged tree, the same dispatch happens with the reason `post-merge-validate-failed`, so a "silent" breakage is treated as seriously as a textual conflict.
- While a pending marker exists and no valid Merge Steward evidence is on disk yet, a UserPromptSubmit hook re-emits the same dispatch directive on every prompt, so the request to run the agent never gets buried in scrollback.
- Once valid Merge Steward evidence is in place, the hook goes quiet — the next visible signal is what you get from `centinela merge --continue`.
- If no pending marker exists, the hook is silent. It only speaks up when a merge is actually waiting on the Steward.
- `centinela merge --continue <feature>` with valid `APPLY` evidence and a clean main tree finalizes the merge: it removes the worktree, clears the pending marker, and exits zero.
- `centinela merge --continue <feature>` with valid `ESCALATE` evidence prints the Steward's escalation note and the proposed diff to stderr, keeps the worktree and the marker, and exits non-zero — the resolution stays in your hands and main is never touched.
- `centinela merge --continue <feature>` with no Steward evidence yet refuses to finalize with an actionable "evidence required" error; nothing changes on disk and the command exits non-zero.
- `centinela merge --continue <feature>` with a Steward evidence file that fails the existing orchestration evidence validator (wrong feature, wrong role, malformed timestamp, missing fields) surfaces the exact validation error, leaves state unchanged, and exits non-zero.
- Even with a perfectly valid `APPLY` verdict, `centinela merge --continue <feature>` refuses to finalize if your main checkout has uncommitted changes — it re-checks the working tree right before finalize, keeps the worktree and marker, and exits non-zero.
- `centinela merge --continue <feature>` when there is no pending marker reports a clear "no pending merge to continue" error, changes no state, and exits non-zero.
- If you re-run `centinela merge <feature>` while a marker already exists — for example, the failure mode shifted from a text conflict to a post-merge validate failure — Centinela rewrites the single marker with the new reason rather than appending a second one, so the on-disk state never drifts out of sync with reality.

## Examples
Before this feature, a conflicted merge stopped at a printed hint and you had to hand-run the Merge Steward yourself. Now the dispatch is automatic.

A normal, clean merge looks the same as before:

    centinela merge checkout-redesign
    # merges, re-validates, removes .worktrees/checkout-redesign/, exit 0

When the merge needs the Steward, the dispatch directive shows up on stdout and the marker is written:

    centinela merge search-filters
    # ┌─ Merge Steward required ─────────────────────────────────────────┐
    # │ Conflict on search-filters needs Merge Steward review.            │
    # │ See .workflow/search-filters-merge-pending.json for the payload.  │
    # └───────────────────────────────────────────────────────────────────┘
    # CENTINELA DIRECTIVE: invoke merge-steward subagent
    #   prompt: docs/architecture/merge-steward-prompt.md
    #   feature: search-filters
    #   resume:  centinela merge --continue search-filters
    # exit 1

If you don't act on the directive in the same session, the next prompt re-surfaces it — you won't lose track of the pending merge.

Once the Steward has written `.workflow/search-filters-merge-steward.{md,json}`, resume the merge:

    centinela merge --continue search-filters
    # Steward verdict: APPLY (handoffTo: complete)
    # main tree clean — finalizing.
    # merged search-filters and removed its worktree.
    # exit 0

If the Steward escalates instead, the same command keeps you blocked and shows the proposed diff so you can decide what to do:

    centinela merge --continue search-filters
    # Steward verdict: ESCALATE (handoffTo: user)
    # ─── escalation note ───
    # Low-confidence resolution: conflict spans two specs that disagree
    # on the empty-state copy. Human review required.
    # ─── proposed diff ─────
    # … unified diff …
    # worktree .worktrees/search-filters/ kept; marker kept.
    # exit 1

Note: the parent `parallel-feature-worktrees` feature shipped only the printed hint and the Merge Steward prompt — the hand-off was manual. This feature wires the hand-off through, so what you actually experience is less manual juggling: the merge tells the orchestrator session what to do, the verdict comes back as evidence, and `--continue` is the single command that either lands the merge or surfaces an escalation. The escalation contract from the parent feature is preserved end-to-end: no diff ever lands on `main` without a deliberate APPLY verdict.
