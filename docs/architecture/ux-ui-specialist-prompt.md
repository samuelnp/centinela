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

Save the Markdown report to
`.workflow/<feature-name>-ux-ui-specialist.md` and a structured JSON
companion at `.workflow/<feature-name>-ux-ui-specialist.json`.

The full schema and validator rules live in
[evidence-contract.md](evidence-contract.md). Read it before writing the
JSON — the orchestration validator rejects malformed evidence with no
auto-repair.

### ux-ui-specialist JSON skeleton

```json
{
  "feature": "<FEATURE_NAME>",
  "step": "code",
  "role": "ux-ui-specialist",
  "status": "done",
  "generatedAt": "<RFC 3339 timestamp>",
  "inputs": [
    "docs/features/<FEATURE_NAME>.md",
    "specs/<FEATURE_NAME>.feature",
    ".workflow/<FEATURE_NAME>-senior-engineer.md"
  ],
  "outputs": [
    "<real UI/asset paths declared for the feature surface>"
  ],
  "edgeCases": [
    "mobile-first",
    "visual-hierarchy",
    "typography-hierarchy",
    "responsive-layout",
    "loading-state",
    "empty-state",
    "error-state",
    "motion-and-reduced-motion"
  ],
  "mobileFirst": true,
  "handoffTo": "qa-senior"
}
```

### Rules that apply to this role (validator will check)

- `mobileFirst` MUST be present and set to `true`. Missing or `false`
  fails with `ux-ui-specialist evidence must declare mobileFirst: true`.
- `edgeCases` MUST contain all eight required UX tags (above). Match is
  case- and separator-insensitive (`Loading State`, `loading-state`,
  `loading_state` all count) — but you should write them in the exact
  hyphenated form shown.
- `outputs` MUST include real UI/asset paths declared for the feature
  surface (validator cross-references against the feature's `uiPaths`).
- `generatedAt` MUST be RFC 3339.
- `handoffTo` MUST be `qa-senior`.

Required only when the feature is user-facing. CLI-only or backend-only
features do not invoke this role.
