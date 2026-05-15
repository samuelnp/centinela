### Feature-Specialist Report: merge-steward-auto-dispatch

**Date:** 2026-05-15

#### Behavior Summary

`centinela merge <feature>` activates the dormant Merge Steward contract.
On a clean text merge whose post-merge `centinela validate` passes, behavior
is unchanged from the parent feature: no marker, no directive, worktree
removed, exit zero. On a git text conflict OR a post-merge validate failure,
the command writes a single `.workflow/<feature>-merge-pending.json` marker
(reason, conflicted paths, worktree path, RFC 3339 timestamp), prints a
structured `CENTINELA DIRECTIVE:` naming the merge-steward prompt, the
feature, and the `centinela merge --continue <feature>` resume command, keeps
the worktree, and exits non-zero. A new UserPromptSubmit hook re-emits that
same directive on every subsequent prompt for as long as the marker exists
and no valid `.workflow/<feature>-merge-steward.json` evidence is present;
the moment valid evidence appears (or the marker is gone) the hook goes
silent. `centinela merge --continue <feature>` then re-runs the existing
orchestration evidence validator and re-checks the clean tree, and gates
finalization on exactly three states: valid + APPLY + `handoffTo:complete`
finalizes (remove worktree, clear marker, exit zero); valid + ESCALATE /
`handoffTo:user` stays blocked (note + proposed diff to stderr, worktree and
marker kept, non-zero); missing or schema-invalid evidence refuses with an
actionable error from the validator (state unchanged, non-zero). The parent
feature's no-silent-resolution escalation contract is preserved end to end.

#### Gherkin Scenarios

Specified in `specs/merge-steward-auto-dispatch.feature`:

- **Clean merge does not dispatch the Steward (regression guard)** — no
  marker, no directive, worktree removed, exit zero (parent behavior intact).
- **Text conflict writes the pending marker and dispatches the Steward** —
  marker records `git-text-conflict` + conflicted paths + worktree path;
  directive names prompt, feature, and `--continue` resume; worktree kept;
  non-zero exit.
- **Post-merge validate failure dispatches the Steward like a text
  conflict** — same dispatch path, marker reason `post-merge-validate-failed`.
- **The hook re-emits the directive while the marker exists without valid
  evidence** — directive re-surfaces on a subsequent prompt.
- **The hook stops re-emitting once valid steward evidence is present** —
  hook silent when valid evidence exists.
- **The hook is silent when no pending marker exists** — no false directive.
- **Continue with APPLY evidence finalizes the merge** — worktree removed,
  marker cleared, exit zero.
- **Continue with ESCALATE evidence keeps the merge blocked** — not
  finalized; worktree + marker kept; escalation note + proposed diff to
  stderr; non-zero exit.
- **Continue with missing steward evidence refuses to finalize** —
  actionable "evidence required" error; state unchanged; non-zero exit.
- **Continue with schema-invalid steward evidence refuses to finalize** —
  orchestration validator error surfaced; state unchanged; non-zero exit.
- **Continue with APPLY evidence but a dirty main tree refuses to
  finalize** — clean-tree re-check blocks finalize even on valid APPLY.
- **Continue with no pending marker reports a clear error** — nothing to
  continue; no state change; non-zero exit.
- **Re-running merge while a pending marker exists does not lose the
  marker** — exactly one marker; rewritten (not appended) with the new
  reason; idempotent and safe.

#### UX States

CLI/terminal only — no graphical surface; "empty" is n/a.

| State    | Trigger | Surface |
|----------|---------|---------|
| loading  | n/a (synchronous CLI command; no spinner) | n/a |
| empty    | n/a (no list/collection surface) | n/a |
| error    | `--continue` with no marker / missing / schema-invalid evidence; ESCALATE; dirty tree at finalize | Actionable stderr message; non-zero exit; for ESCALATE the steward note + proposed diff are printed to stderr |
| dispatch | Text conflict or post-merge validate failure | `ui.RenderStep`-style "Merge Steward required" block + `CENTINELA DIRECTIVE:` line; re-emitted each prompt by the hook while pending |
| success  | `centinela merge` clean, or `--continue` with valid APPLY + clean tree | `ui.RenderSuccess` "Merged … and removed its worktree"; exit zero; marker cleared |

#### Out-of-Scope

- Perfectly resolving arbitrary semantic conflicts — the Steward proposes a
  diff; humans approve via an explicit APPLY + `complete` handoff.
- Auto-applying the proposed diff without an APPLY/`complete` handoff.
- Multi-feature merge trains or queued/pending-merge queues.
- A `centinela merge --status` convenience view (deferrable per the plan).
- Spec/contract conflicts — blocked pre-merge by `DetectSpecConflicts`
  before `Merge` runs, so they never reach the auto-dispatch path.
- Any GUI or non-CLI surface; the Go binary never calls an LLM.
- Changing the merge-steward prompt's analysis contract — only its
  invocation path changes.

#### Handoff

- Next role: senior-engineer
- Open clarifications:
  - **`--continue` vs `merge finalize`**: the spec assumes the `--continue`
    flag (single command, `git rebase --continue` mental model). Confirmed
    direction; no separate subcommand specified.
  - **Pending-marker JSON shape**: `{feature, reason, conflictedPaths,
    worktreePath, generatedAt}`. The spec asserts reason + conflicted paths
    + worktree path are recorded; senior-engineer to confirm this payload
    fully satisfies the merge-steward prompt's stdin expectations or whether
    conflicted paths must also ride on the directive line.
  - **Directive line content**: the spec pins that the directive names the
    merge-steward prompt, the feature, and the `--continue` resume command;
    exact wording is left to the senior-engineer to align with the existing
    `CENTINELA DIRECTIVE:` hook style.
