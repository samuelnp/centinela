# Feature Specialist — Plan Validation

- Feature: enrich-plan-advisor-context
- Step: plan
- Role: feature-specialist
- Status: done

## Inputs
- docs/features/enrich-plan-advisor-context.md
- specs/enrich-plan-advisor-context.feature

## Outputs
- docs/plans/enrich-plan-advisor-context.md
- specs/enrich-plan-advisor-context.feature

## Edge Cases
- roadmap dependencies should outrank same-phase siblings when both exist
- related edge-case reports should influence questions without causing prompt bloat
- missing roadmap artifacts should degrade gracefully to current-feature planning context

## Handoff
- handoffTo: senior-engineer
