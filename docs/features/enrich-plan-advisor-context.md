---
surface: internal
---

# Feature Brief: Enrich Plan Advisor Context

## Problem
`plan-advisor` currently inspects only the current feature brief, plan, and spec. That helps avoid
generic repeated questions, but it misses high-value planning context that already exists elsewhere
in the project, such as roadmap dependencies, same-phase sibling features, prior edge-case reports,
and roadmap analysis or quality concerns.

## Goal
Enrich plan-advisor mode with a compact context bundle so `big-thinker` and `feature-specialist`
ask better planning questions using what is already known across the roadmap and related feature
artifacts.

## Scope
- Read roadmap phase and dependency context from `.workflow/roadmap.json` and analysis artifacts.
- Prefer dependency context first, then same-phase sibling features.
- Read current feature artifacts plus selected related feature briefs, specs, and edge-case reports.
- Summarize related context into a small planning bundle instead of dumping full files into prompts.
- Use roadmap quality and dependency concerns to influence question selection.

## Acceptance Criteria
- Plan-advisor uses current feature brief, plan, spec, and local edge-case report when present.
- If roadmap analysis exists, dependencies are considered before same-phase siblings.
- If related edge-case reports exist, advisor questions can reuse those lessons.
- Advisor output remains concise and capped at the existing question limit.
- Advisor does not dump large raw file contents into the prompt.
