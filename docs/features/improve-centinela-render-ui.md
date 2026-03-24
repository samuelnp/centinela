---
feature: improve-centinela-render-ui
type: feat
---

# Feature: Explicit and Attractive Centinela Render UI

Centinela output should be visually distinct from LLM conversational output so
users can immediately identify system messages. The current rendering is useful
but ambiguous in mixed CLI streams.

## Goals

- Make every Centinela message clearly branded (`CENTINELA` + channel/type).
- Keep output compact while improving readability and visual hierarchy.
- Use color, icons, and boxed layout consistently across hooks and commands.
- Preserve existing behavior and enforcement logic (presentation-only changes).

## Non-Goals

- Changing workflow rules, gate semantics, or hook policy behavior.
- Introducing heavy/verbose UI that obscures actionable information.
