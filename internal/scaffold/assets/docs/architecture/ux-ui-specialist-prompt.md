<!-- centinela:doc-version=1 template=docs/architecture/ux-ui-specialist-prompt.md -->
# UX-UI-Specialist Subagent — Invocation Guide

## Purpose

Use this subagent during the `code` step **only when the feature brief
declares `surface: user-facing`**. It reviews mobile-first flow,
accessibility, visual hierarchy, and the loading/empty/error/success
state coverage of any user-facing surface.

## How to Invoke

See [agent-invocation.md](agent-invocation.md) for the canonical Agent
invocation pattern. Skip invocation entirely if the feature has no
user-facing surface — `RequiredRolesForFeature` in
`internal/orchestration/policy.go` will not request this role unless the
feature is user-facing.

## Prompt Template

```
You are the Centinela UX-UI-Specialist for feature "<FEATURE_NAME>".

Read the feature brief at docs/features/<FEATURE_NAME>.md, the spec at
specs/<FEATURE_NAME>.feature, and the senior-engineer report at
.workflow/<FEATURE_NAME>-senior-engineer.md. Then review the surface.

Required analysis:
1. Flow review — primary path on small screens (≤ 375px) and touch
   devices, including tap-target sizing and reachable controls.
2. Accessibility — semantic markup / labels, color contrast, keyboard
   navigation, screen-reader announcements for live regions.
3. Visual hierarchy — primary vs secondary actions, scannable headings,
   safe spacing.
4. State coverage — loading, empty, error, success representations are
   each present, distinct, and informative.

Output format:
### UX-UI Report: <FEATURE_NAME>
**Date:** <current date>

#### Flow Review
- mobile-first walk-through

#### Accessibility (semantic | contrast | keyboard | screen reader)
- one bullet per check: PASS / FAIL + note

#### Visual Hierarchy
- bullets of issues / confirmations

#### State Coverage (loading | empty | error | success)
- one bullet per state: present? + note

#### Handoff: qa-senior
```

## Required Artifact

Save report to `.workflow/<feature-name>-ux-ui-specialist.md` and a
structured companion at `.workflow/<feature-name>-ux-ui-specialist.json`.

Required only when the feature is user-facing. CLI-only or backend-only
features do not invoke this role.
