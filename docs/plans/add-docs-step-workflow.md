# Plan: Add Docs Step to Workflow

1. Update workflow step order and related progress helpers to include `docs`.
2. Add `docs` artifact validation in `internal/workflow`.
3. Add strict orchestration role `documentation-specialist` for `docs` step.
4. Update CLI/hook messaging and progress output to support 5 steps.
5. Update managed docs/templates that describe workflow step count/order.
6. Add tests for workflow transitions, orchestration, hook UX, and acceptance.
