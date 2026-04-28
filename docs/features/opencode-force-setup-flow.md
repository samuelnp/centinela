# opencode-force-setup-flow

## Problem

OpenCode now notices Centinela startup rules, but when `PROJECT.md` is missing it can still tell the user to run `centinela start <feature>` or ask what to work on. That is incorrect because a new project must complete project setup and roadmap bootstrap before feature workflows.

## User Stories

- As an OpenCode user, a greeting in an unconfigured Centinela project starts project setup, not feature discovery.
- As a maintainer, the generated OpenCode rules clearly separate bootstrap setup from feature work.

## Acceptance Criteria

- Generated `AGENTS.md` says missing `PROJECT.md` setup takes precedence over `centinela start <feature>`.
- Generated `AGENTS.md` forbids asking what feature to work on while project setup or roadmap bootstrap is incomplete.
- Existing setup panel guidance remains aligned with this behavior.

## Edge Cases

- Already-configured projects can still start feature workflows normally.
- Missing roadmap after `PROJECT.md` is created still blocks feature prompts.
- Unmanaged OpenCode files remain protected by existing sync behavior.

## Risks

- If the instruction is too broad, OpenCode could avoid feature work even after setup is complete.
- Static guidance must remain concise enough to be followed consistently.
