# Plan: Enrich Plan Advisor Context

1. Add a `ContextBundle` in `internal/planadvisor` that reads current feature artifacts plus roadmap,
   roadmap analysis, roadmap quality, selected related feature briefs/specs, and edge-case reports.
2. Resolve related features with this priority:
   - direct roadmap dependencies first
   - same-phase sibling features second
3. Add compact summary helpers so the advisor emits context findings instead of raw file dumps.
4. Extend question selection so related dependency risks, prior edge cases, and roadmap quality gaps
   influence `big-thinker` and `feature-specialist` questions.
5. Keep output capped and concise, with no new orchestration role or step changes.
6. Add unit, integration, and acceptance coverage for dependency-first context, sibling fallback,
   and related edge-case reuse.
