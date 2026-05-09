# agent-performance-audit

## Problem

Centinela now creates native OpenCode specialist agents, but their prompts and step coverage need an audit for token efficiency, duplicated context, and missing workflow roles.

## User Stories

- As a maintainer, I want to know which agents are configured and whether each workflow step has specialist coverage.
- As a user paying LLM API costs, I want unnecessary prompt text identified so agent context stays small.
- As a product owner, I want a prioritized recommendation list for improving agent performance without losing workflow quality.

## Acceptance Criteria

- An audit document lists every configured Centinela OpenCode agent and its workflow step coverage.
- The audit identifies missing agent coverage for plan, code, tests, validate, and docs.
- Missing validate-step native OpenCode agent coverage is implemented as `validation-specialist`.
- The audit recommends prompt text reductions and performance improvements.
- Tests verify the audit mentions all workflow steps and all configured agents.

## Edge Cases

- `ux-ui-specialist` is conditional and should not be counted as required for every feature.
- Gatekeeper and production-readiness prompts are validation specialists but not currently configured as native OpenCode agents.
- Existing generated subagent prompts must be reviewed without changing behavior in this audit.

## Risks

- Over-optimizing prompts can reduce specialist quality.
- Adding more agents can improve coverage but increase orchestration cost.
