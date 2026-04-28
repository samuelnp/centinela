---
surface: internal
---

# Feature Brief: Add Plan Advisor Mode

## Problem
Strict orchestration already requires `big-thinker` and `feature-specialist` plan evidence, but the
runtime prompt behavior does not consistently make them act like adaptive advisors. The model still
risks jumping into implementation or asking the same generic planning questions even when the
feature brief, plan, or spec already answer them.

## Goal
Add a default-on `plan-advisor` mode that activates during the `plan` step and orchestrates between
 `big-thinker` and `feature-specialist` so the model asks only the missing high-value questions that
 improve the feature brief, plan, and spec.

## Scope
- Add a plan-step prompt hook for advisor mode.
- Enable advisor mode by default through workflow config.
- Limit each advisor round to at most 4 questions.
- Use existing feature brief, plan, and spec artifacts to detect missing planning coverage.
- Split advisor output into `big-thinker` and `feature-specialist` lenses without adding a new
  evidence role.

## Acceptance Criteria
- Advisor mode only activates during the `plan` step.
- Advisor mode stays silent outside `plan`.
- When planning coverage is missing, the hook asks up to 4 targeted questions.
- When relevant areas are already covered, the hook avoids repeating those questions.
- User-facing features receive UX and mobile-first planning questions only when those topics are
  still missing.
