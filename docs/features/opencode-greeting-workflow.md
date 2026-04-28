# opencode-greeting-workflow

## Problem

OpenCode can answer a first greeting conversationally in a Centinela project instead of first surfacing the required setup or workflow guidance. Claude Code already treats the Centinela bootstrap rules as startup-critical, so users see inconsistent behavior across supported agents.

## User Stories

- As an OpenCode user, I am told about Centinela setup and workflow requirements at the beginning of a new project session.
- As a maintainer, I want generated OpenCode assets to keep Centinela rules visible even for casual first prompts.

## Acceptance Criteria

- Generated OpenCode instructions explicitly require mentioning Centinela setup or workflow requirements before answering greetings when required.
- OpenCode prompt injection continues to put setup and migration directives ahead of other context.
- Regression tests cover greeting-first OpenCode behavior through generated assets.

## Edge Cases

- No setup warning is emitted for directories unrelated to Centinela.
- Existing unmanaged OpenCode files are not overwritten without review.
- Feature autostart remains separate from setup-required guidance.

## Risks

- Overly broad instructions could make normal conversation noisy after setup is complete.
- OpenCode plugin event behavior can change by version, so static instructions must carry the critical rule.

## Decomposition

- Strengthen generated `AGENTS.md` OpenCode startup rules.
- Add tests asserting greeting-first behavior is described in generated OpenCode assets.
- Preserve existing plugin ordering tests for setup and migration directives.
