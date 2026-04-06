# Edge Cases: roadmap-quality-overall-threshold

## Covered

- Quality JSON omits a roadmap feature.
- Quality JSON references unknown feature names.
- Quality score values outside 1..10 are rejected.
- Feature with `overall` below 9 blocks roadmap validation and greenfield start.
- Low `effortEstimation` does not block when `overall` is at least 9.

## Residual Risks

- The evaluator role and threshold are currently fixed constants.
- Score narratives are validated for presence, not semantic depth.
