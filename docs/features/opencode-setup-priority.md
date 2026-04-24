# opencode-setup-priority

## Problem

OpenCode receives Centinela setup directives when `PROJECT.md` is missing, but a casual prompt like `Hi` can still be answered conversationally instead of starting setup.

## User Stories

- As an OpenCode user, I am guided into `PROJECT.md` setup even if my first message is only a greeting.
- As a maintainer, I keep Claude and OpenCode setup behavior aligned.

## Acceptance Criteria

- OpenCode prompt guidance makes setup directives higher priority than casual greetings.
- Generated OpenCode-facing instructions explicitly require honoring Centinela setup directives before normal conversation.
- Docs describe automatic setup in agent-neutral terms instead of Claude-only wording.

## Edge Cases

- Greeting-only prompts still trigger setup guidance when `PROJECT.md` is missing.
- Feature-intent autostart remains separate from setup-required behavior.

## Risks

- Over-prioritizing setup text could affect non-setup prompt handling if scoped too broadly.
- OpenCode prompt lifecycle behavior may vary across versions.

## Decomposition

- Tighten OpenCode instructions and prompt injection ordering.
- Add regression tests around missing `PROJECT.md` plus greeting prompts.
- Update generated and published setup docs to mention OpenCode explicitly.
