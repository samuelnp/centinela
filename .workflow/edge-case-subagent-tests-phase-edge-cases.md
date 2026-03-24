### Edge-Case Report: edge-case-subagent-tests-phase
**Date:** 2026-03-23

#### Risk Matrix
- **Case:** Tests step bypass without hard-path review
- **Impact:** High
- **Likelihood:** Medium
- **Why:** Happy-path-only tests miss production failures.

- **Case:** Context reminder absent during tests step
- **Impact:** Medium
- **Likelihood:** Medium
- **Why:** Agent may complete tests without edge-case artifact.

#### Missing or Weak Scenarios
- tests-step validation without `.workflow/<feature>-edge-cases.md`
- tests-step context reminder when report is missing
- prompt doc missing required risk and residual sections

#### Proposed/Added Tests
- Unit: validate tests step requires edge-case report
- Integration: hook context reminder appears for tests step
- Acceptance: edge-case tester prompt includes mandatory sections

#### Residual Risks
- Subagent quality still depends on prompt adherence.
- Mitigation: keep prompt template strict and review report completeness in PR.
