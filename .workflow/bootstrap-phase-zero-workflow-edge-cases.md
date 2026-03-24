### Edge-Case Report: bootstrap-phase-zero-workflow
**Date:** 2026-03-24

#### Risk Matrix
- **Case:** Legacy repositories missing `Project Stage` are interpreted incorrectly
- **Impact:** Medium
- **Likelihood:** Medium
- **Why:** Defaulting to greenfield can unexpectedly enforce bootstrap gating.

- **Case:** Roadmap has malformed bootstrap phase naming
- **Impact:** High
- **Likelihood:** Medium
- **Why:** Start command blocks non-bootstrap features without explicit Phase 0 shape.

#### Missing or Weak Scenarios
- Greenfield project with empty Phase 0 feature list
- Existing project with Phase 0 present but intentionally incomplete
- Mixed roadmap names (`Phase Zero`, localized variants)

#### Proposed/Added Tests
- Unit: project stage parsing + bootstrap helper behavior
- Integration: start guard gates by stage and bootstrap completion
- Acceptance: bootstrap order uses 3 steps and tests placeholders are rejected

#### Residual Risks
- Phase-name matching remains convention-based, not schema-validated.
- Mitigation: keep roadmap guidance strict and improve schema validation later.
