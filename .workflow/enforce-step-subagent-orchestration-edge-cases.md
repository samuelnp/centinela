# Edge-Case Review: enforce-step-subagent-orchestration

## Scenarios Reviewed

- Legacy workflows without `orchestrationMode` metadata are not blocked.
- Missing markdown evidence with valid JSON still blocks step completion.
- Missing JSON evidence with markdown present still blocks step completion.
- JSON with wrong `step` or `role` is rejected as invalid evidence.
- Unknown extra JSON fields are accepted.
- `checksum` field is optional and does not block validation when omitted.
- `feature-specialist` and `qa-senior` evidence require non-empty `edgeCases`.

## Outcome

- Strict orchestration is enforced only for new workflows.
- Evidence validation gives deterministic blocking for incomplete delegation.
