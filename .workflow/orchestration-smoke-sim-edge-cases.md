# Edge Cases: orchestration-smoke-sim

## Covered

- Step completion fails when required role evidence is missing.
- Step completion advances once required role evidence is present.
- Hook output changes required role list by step.

## Residual Risks

- This simulation validates evidence gating, not automatic subagent process spawning.
