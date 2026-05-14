<!-- centinela:doc-version=1 template=docs/architecture/merge-steward-prompt.md -->
# Merge Steward Subagent — Invocation Guide

## Purpose

Resolve merge conflicts when `centinela merge <feature>` cannot complete
on its own. The Steward runs **out-of-band** — it is NOT one of the five
workflow steps. It is triggered in three cases:

1. **Git text conflict** — `git merge --no-ff <feature>` exited non-zero.
2. **Semantic conflict** — text merge was clean, but `centinela validate`
   failed against the merged tree.
3. **Spec/contract conflict** — pre-merge analysis surfaced two
   `specs/*.feature` files that assert different observable outcomes for
   the same Given context across active worktrees.

The Steward proposes a resolution, but it **must escalate** to the user
whenever its confidence is not unanimously high. There are no silent
resolutions.

## How to Invoke

See [agent-invocation.md](agent-invocation.md) for the canonical Agent
invocation pattern. Replace `<FEATURE_NAME>` in the template below.

## Prompt Template

```
You are the Centinela Merge Steward for feature "<FEATURE_NAME>".

Inputs you MUST read before proposing anything:
- specs/<FEATURE_NAME>.feature (the feature contract)
- docs/plans/<FEATURE_NAME>.md (the implementation plan)
- The merge diff: `git diff main..<FEATURE_NAME>` from the repo root
- For text conflicts: every file path listed in the `conflictedPaths`
  array on stdin
- For semantic conflicts: the full `centinela validate` output passed
  on stdin
- For spec conflicts: every conflicting `specs/*.feature` file from
  both worktrees plus their plans

Required analysis:
1. Conflict classification — text-conflict | post-merge-validate-failed |
   spec-contract. Multiple may apply.
2. Confidence — high | medium | low. Anything other than `high` MUST
   escalate.
3. Proposed resolution — a unified diff against the merged tree that
   would, if applied, satisfy the feature spec without violating any
   other in-flight spec. Empty diff is acceptable for escalations.
4. Reasoning — a short paragraph explaining why the proposal preserves
   the feature spec, and why it does not break the validate run.

Output format:
### Merge Steward Report: <FEATURE_NAME>
**Date:** <current date>
**Status:** APPLY | ESCALATE
**Confidence:** high | medium | low
**Conflict classes:** <comma-separated tags>

#### Conflicted Surface
| Path | Class | Notes |
|------|-------|-------|

#### Proposed Resolution
```diff
…unified diff…
```

#### Reasoning
- bullet list

#### Escalation Note (when Status = ESCALATE)
- Why confidence is below `high`
- What the user must decide before re-running `centinela merge`
```

## Required Artifacts

Save the report to `.workflow/<feature>-merge-steward.md` and a
structured JSON companion at
`.workflow/<feature>-merge-steward.json`.

The JSON schema is the standard evidence contract — see
[evidence-contract.md](evidence-contract.md) for the merge-steward
entry. Key constraints:

- `role` MUST be `"merge-steward"`.
- `step` MAY be `"merge"` (the role lives outside the 5-step workflow).
- `outputs` MUST include `.workflow/<feature>-merge-steward.md`. The
  proposed diff SHOULD be saved as a sibling file when non-empty (e.g.
  `.workflow/<feature>-merge-steward.diff`).
- `handoffTo` MUST be `"complete"` on APPLY or `"user"` on ESCALATE.
- `edgeCases` SHOULD include every conflict class detected.

## Escalation contract

When `Status: ESCALATE`:

1. The Steward MUST NOT apply the proposed diff.
2. `centinela merge` MUST exit non-zero so CI/automation surfaces the
   block.
3. The proposed diff and the reason confidence is low MUST be printed
   to stderr so the user can review without opening files.
4. The worktree at `.worktrees/<feature>/` MUST be left intact for
   manual inspection.
