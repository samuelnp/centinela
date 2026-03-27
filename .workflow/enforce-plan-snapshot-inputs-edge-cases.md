# Edge Cases: enforce-plan-snapshot-inputs

## Covered

- Plan snapshot rule applies only to `plan` step evidence.
- Rule applies only to `big-thinker` and `feature-specialist` roles.
- Missing any `docs/features/*.md` path fails with explicit missing-file output.
- Current feature brief path is required even when glob results are sparse.
- Mixed input path styles (`./`, absolute, slash variants) normalize correctly.

## Residual Risks

- Very large feature-doc sets can make evidence `inputs` arrays large.
- Enforcement assumes feature briefs are markdown files under `docs/features/`.
