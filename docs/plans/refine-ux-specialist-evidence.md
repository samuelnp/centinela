# Plan: Refine UX Specialist Evidence

1. Extend orchestration evidence with a `mobileFirst` field.
2. Add a UX-specific evidence validator for `ux-ui-specialist`.
3. Require `mobileFirst: true` for UX evidence.
4. Define and validate a minimum UX edge-case tag set:
   - `mobile-first`
   - `visual-hierarchy`
   - `typography-hierarchy`
   - `responsive-layout`
   - `loading-state`
   - `empty-state`
   - `error-state`
   - `motion-and-reduced-motion`
5. Preserve existing UI output path validation and role selection behavior.
6. Add unit, integration, and acceptance coverage for missing mobile-first and missing-tag failures.
7. Update runtime and scaffold docs to reflect the stricter UX evidence contract.
