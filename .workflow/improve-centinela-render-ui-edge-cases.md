### Edge-Case Report: improve-centinela-render-ui
**Date:** 2026-03-24

#### Risk Matrix
- **Case:** Branded headers make output too noisy in tight loops
- **Impact:** Medium
- **Likelihood:** Medium
- **Why:** Hooks run frequently and could spam visual wrappers.

- **Case:** Color/style mismatch across terminals
- **Impact:** Low
- **Likelihood:** Medium
- **Why:** Different terminal themes render ANSI colors with varying contrast.

#### Missing or Weak Scenarios
- Narrow terminal widths clipping boxed metadata
- Non-UTF terminals rendering icon glyphs poorly
- Multiple active workflows causing dense context output

#### Proposed/Added Tests
- Unit: panel primitives and branding symbols exist
- Integration: blocked/context/tag outputs contain explicit system branding
- Acceptance: render modules use branded panel helpers

#### Residual Risks
- UX preference is subjective; some users may prefer minimal plain text.
- Future work: add config flag for compact/plain render modes.
