# Edge-Case Review: auto-start-feature-intent

## Scenarios Reviewed

- Only `done` workflows exist: prewrite now requires `centinela start` for non-roadmap writes.
- Active workflow exists: autostart hook does not create a second workflow.
- Prompt is a review/advance confirmation: intent detector ignores it.
- Prompt payload is JSON or plain text: extraction supports both paths.
- Derived feature name already exists: autostart appends numeric suffix for uniqueness.

## Outcome

- Workflow enforcement remains strict after feature completion.
- New feature requests automatically open a tracked workflow instead of drifting.
