# merge-steward-auto-dispatch

## Problem

`parallel-feature-worktrees` added `centinela merge <feature>` and a
`merge-steward` role (prompt, evidence contract, validator), but the
wiring stops short: on a git text conflict or a post-merge
`centinela validate` failure, `runMerge` only prints "Merge Steward
required" and errors with a path hint. Nothing dispatches the Steward,
nothing consumes its evidence, and merge can never be finalized after a
conflict. The operator must notice the hint, hand-run the agent, and
re-drive the merge by hand.

## Goal

Make the Merge Steward engage automatically when a merge needs it.
`centinela merge` records a pending-merge marker and emits a structured
CENTINELA DIRECTIVE instructing the orchestrator session to invoke the
merge-steward subagent with the right prompt and inputs. A
UserPromptSubmit hook keeps re-surfacing the directive until valid
`.workflow/<feature>-merge-steward.json` evidence exists. `centinela
merge --continue` then reads that evidence and gates finalization on it.

## Scope

- `centinela merge` on text-conflict OR post-validate-failure: write
  `.workflow/<feature>-merge-pending.json` (reason, conflicted paths,
  worktree path), print a CENTINELA DIRECTIVE naming the merge-steward
  prompt + inputs, exit non-zero.
- New `cmd/centinela/hook_merge.go` (UserPromptSubmit) that, while a
  pending marker exists without valid steward evidence, re-emits the
  dispatch directive every prompt.
- `centinela merge --continue <feature>`: validate steward evidence via
  the existing orchestration validator. `handoffTo: complete` + APPLY →
  finalize (remove worktree, clear marker). `handoffTo: user` /
  ESCALATE / invalid → stay blocked, print escalation note + diff to
  stderr, exit non-zero, keep worktree and marker.
- Pending/finalize state logic lives in `internal/worktree/`; the cmd
  layer stays a thin orchestrator.

## Edge Cases

- Steward escalates (low confidence) — never auto-apply; surface to
  user, exit non-zero, keep worktree (parent escalation contract).
- `--continue` with no marker → clear error, no-op.
- `--continue` before evidence written → blocked, directive re-emitted.
- Invalid/mismatched steward JSON → treated as not-yet-resolved, not a
  silent pass.
- Conflict reason changes between runs (text → validate) — marker is
  rewritten, not appended.
- Main working tree dirtied during steward work → re-checked at
  `--continue` before finalize.

## Out of Scope

- Perfectly resolving arbitrary semantic conflicts — the Steward
  proposes, humans approve.
- Auto-applying the proposed diff without an APPLY/complete handoff.
- Multi-feature merge trains, queued merges.
- Any GUI or non-CLI surface.
