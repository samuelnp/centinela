# Feature Brief: Enforce Actionable Orchestration Evidence

## Problem
Strict orchestration currently proves that specialist evidence files exist, but it does not prove
that discoveries from `big-thinker`, `feature-specialist`, `senior-engineer`, or `qa-senior`
were applied to the project. Evidence can pass with free-text outputs instead of real file changes.

## Goal
Require strict-orchestration evidence to reference actionable project outputs so specialist findings
must be reflected in concrete repo updates before a step can complete.

## Scope
- Validate evidence `outputs` as real project-relative paths instead of free-text summaries.
- Enforce role-specific actionable outputs for `big-thinker`, `feature-specialist`, `senior-engineer`, and `qa-senior`.
- Keep docs-step `documentation-specialist` behavior unchanged.
- Return clear validation errors naming the missing or invalid output paths.

## Acceptance Criteria
- `big-thinker` evidence fails unless outputs include the feature plan or spec artifact it drove.
- `feature-specialist` evidence fails unless outputs include a real spec or feature-plan artifact.
- `senior-engineer` evidence fails unless outputs include at least one real non-evidence implementation file.
- `qa-senior` evidence fails unless outputs include at least one real test file and the edge-case report.
- Existing valid evidence with real file paths continues to pass.
