---
surface: internal
---

# Feature Brief: Refine UX Specialist Evidence

## Problem
The `ux-ui-specialist` role currently guarantees only that outputs point to UI files and that
`edgeCases` is non-empty. That is too loose to ensure consistent mobile-first, user-centered UI
review quality.

## Goal
Strengthen UX specialist evidence so it always declares mobile-first intent and covers a required
set of UX/UI review concerns through enforceable edge-case tags.

## Scope
- Add a dedicated `mobileFirst` evidence field.
- Require `mobileFirst: true` for `ux-ui-specialist` evidence.
- Replace arbitrary UX edge-case text with a required minimum tag set.
- Keep existing real UI output validation in place.
- Update docs and templates so future workflows emit the stricter UX evidence format.

## Acceptance Criteria
- UX evidence fails when `mobileFirst` is missing or false.
- UX evidence fails when required UX edge-case tags are missing.
- UX evidence still fails when outputs do not point to real UI files.
- UX evidence passes when mobile-first is true, required tags are present, and UI outputs are valid.
