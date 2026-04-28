# Feature Specialist — Plan Validation

- Feature: add-plan-advisor-mode
- Step: plan
- Role: feature-specialist
- Status: done

## Inputs
- docs/features/add-plan-advisor-mode.md
- specs/add-plan-advisor-mode.feature

## Outputs
- docs/plans/add-plan-advisor-mode.md
- specs/add-plan-advisor-mode.feature

## Edge Cases
- advisor mode may repeat generic questions if it does not inspect existing plan artifacts
- user-facing plan prompts should include UX and mobile-first guidance only when still missing
- advisor mode must not create a new orchestration evidence role or affect non-plan steps

## Handoff
- handoffTo: senior-engineer
