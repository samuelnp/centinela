# Edge Cases: claude-status-line

- No active workflow files should return `WF:none` and `BLOCK:NO_WORKFLOW`.
- Role-suffixed orchestration workflows should not be selected as primary workflow.
- Plan step should show blockers for missing brief, missing plan, or missing spec.
- Tests step should block on missing `.workflow/<feature>-edge-cases.md`.
- Validate step should block on missing gatekeeper and production-readiness artifacts.
- Empty feature or step fields in malformed workflow files should be ignored.
