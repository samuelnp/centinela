---
feature: enforce-actionable-orchestration-evidence
summary: Specialist evidence now must reference real project files, not free-text summaries.
audience: end-user
status: done
---

## What it does
When you run `centinela complete <feature>` to close a step, the orchestration validator now checks that every specialist evidence file (`.workflow/<feature>-<role>.json`) contains real, project-relative file paths in its `outputs` field. Each role has specific output rules: `big-thinker` and `feature-specialist` must point to the plan or spec they drove; `senior-engineer` must reference at least one implementation file; `qa-senior` must reference at least one test file and the edge-case report. If outputs contain invalid paths or are empty, the validator blocks completion with a clear error naming the broken path.

## When you'd use it
You feel this every time a subagent finishes a step and you run `centinela complete`. Before this change, agents could write prose summaries as outputs or skip linking to actual project artifacts entirely, forcing manual JSON rewrites. Now the validator enforces that findings are tied to real code—either the plan/spec they drove, the implementation they reviewed, or the tests they designed—so specialist work is always actionable and traceable.

## How it behaves
- `centinela complete` runs the orchestration validator as part of its gate checks.
- The validator reads each evidence JSON file for the feature's current step.
- For each role, it verifies outputs against role-specific rules:
  - `big-thinker` and `feature-specialist` outputs must include at least one path from `docs/features/` or `docs/plans/`.
  - `senior-engineer` outputs must include at least one non-evidence file (anything outside `.workflow/`).
  - `qa-senior` outputs must include at least one path from `tests/` and the `.workflow/<feature>-edge-cases.md` file.
  - `documentation-specialist` outputs are not checked (docs step outputs vary by project).
- If any output path does not exist in the repo or violates the role rule, completion fails with an error message naming the bad path and the role that violated it.
- Existing evidence with valid real file paths continues to pass without change.

## Examples
A typical evidence JSON after the plan step now looks like:

```json
{
  "feature": "my-auth-flow",
  "step": "plan",
  "role": "big-thinker",
  "status": "done",
  "generatedAt": "2026-05-28T14:00:00Z",
  "inputs": ["docs/features/my-auth-flow.md", "docs/features/user-signup.md"],
  "outputs": ["docs/features/my-auth-flow.md", "docs/plans/my-auth-flow.md"],
  "edgeCases": ["Token refresh edge case during session expiry"],
  "handoffTo": "feature-specialist"
}
```

The outputs point to real files that exist in the repo. If the plan file were missing, `centinela complete plan` would reject the evidence and report: `"big-thinker outputs reference docs/plans/my-auth-flow.md but it does not exist"`.
