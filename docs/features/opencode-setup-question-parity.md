# opencode-setup-question-parity

## Problem

Claude and OpenCode both enter Centinela setup when `PROJECT.md` is missing, but OpenCode compresses the setup questions and asks fewer details. Users should get the same setup checklist regardless of agent.

## User Stories

- As an OpenCode user, I get the same `PROJECT.md` setup questions Claude asks.
- As a maintainer, setup guidance is explicit enough that agents cannot collapse required questions into vague categories.

## Acceptance Criteria

- `RenderSetupNeeded` lists the exact six setup questions to ask.
- The questions cover project name, elevator pitch, tech stack, architecture archetype, locales, and folder layout.
- Tests assert the setup panel contains each exact question label.

## Edge Cases

- Agents must still read `PROJECT.md.template` first.
- Setup guidance must not mention `centinela start <feature>` before `PROJECT.md` exists.
- Roadmap handoff remains after `PROJECT.md` is written.

## Risks

- Overly specific wording could become stale if `PROJECT.md.template` changes.
- The prompt should stay concise enough not to overwhelm the first response.
