---
feature: add-personality-feedback
type: feat
---

# Feature: Add personality to Centinela feedback

Centinela output is clear but feels mechanical. We want messages to keep strict
guidance while adding a recognizable voice that makes feedback easier to scan
and less cold during day-to-day CLI use.

## Goals

- Add a consistent persona prefix to all rendered Centinela messages.
- Use tone-specific expressions for info, success, warning, and error output.
- Keep existing output structure (channel, title, actionability) unchanged.
- Preserve ANSI color support in terminal output with plain-text compatibility.

## Non-Goals

- Changing workflow policies, gate logic, or step validation behavior.
- Introducing config toggles for persona in this initial iteration.
