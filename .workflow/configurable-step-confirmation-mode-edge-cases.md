# Edge Cases: configurable-step-confirmation-mode

## Covered

- Unknown `step_confirmation_mode` values normalize to `every_step`.
- `after_plan` shows review prompt at plan only, not code/tests/validate/docs.
- `auto` suppresses review prompt while keeping explicit `centinela complete` flow.
- Strict orchestration can hide review prompts unless evidence requirements are met.

## Residual Risks

- Prompt policy is advisory in hook context and depends on agent compliance.
- Existing custom CLAUDE rules may still enforce stricter behavior than config.
