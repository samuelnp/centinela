### Edge-Case Report: add-personality-feedback
**Date:** 2026-04-06

#### Risk Matrix
- **Case:** Persona text overwhelms key guidance
- **Impact:** Medium
- **Likelihood:** Low
- **Why:** Added prefix could distract if too long.

- **Case:** ANSI colors not visible in hook bridges
- **Impact:** Low
- **Likelihood:** Medium
- **Why:** Non-TTY pipelines may strip color escapes.

#### Missing or Weak Scenarios
- Narrow terminals clipping long lines with persona prefix
- Environments with limited Unicode glyph rendering

#### Proposed/Added Tests
- Unit: persona primitives and key tokens exist
- Integration: success/info/error outputs include correct persona faces
- Acceptance: hook metadata and action hints remain intact with persona

#### Residual Risks
- Some users may prefer a compact plain style without expressions.
