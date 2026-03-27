# Plan: Enforce Plan Snapshot Inputs

1. Add plan-step evidence validation for required feature-doc snapshot inputs.
2. Require this rule only for `big-thinker` and `feature-specialist` roles.
3. Normalize and compare paths against all `docs/features/*.md` files.
4. Return actionable errors listing missing snapshot files.
5. Add orchestration and workflow tests for fail/pass branches.
6. Update workflow enforcement docs and scaffold template.
