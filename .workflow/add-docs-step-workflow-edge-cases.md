# Edge Cases: add-docs-step-workflow

## Covered

- Default workflows advance from `validate` to `docs` instead of `done`.
- Bootstrap workflows include `docs` as final step without reintroducing tests step.
- Docs step fails when `docs/project-docs/index.html` is missing.
- Strict orchestration requires `documentation-specialist` evidence during docs step.
- Statusline and context hooks surface actionable docs-step reminders.

## Residual Risks

- Existing external integrations that assume `[x/4]` progress may need adjustment.
- Projects with custom docs output paths are not configurable in docs-step gate yet.
