# Plan: Roadmap Quality Overall Threshold

1. Add roadmap quality model and validator in `internal/roadmap`.
2. Validate role, per-feature coverage, score range, and `overall >= 9`.
3. Require quality artifacts in `centinela roadmap validate`.
4. Block greenfield start when roadmap quality validation fails.
5. Update setup and roadmap UI guidance for evaluator iteration loop.
6. Add unit, integration, and acceptance tests for pass/fail paths.
