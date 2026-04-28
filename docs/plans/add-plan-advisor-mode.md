# Plan: Add Plan Advisor Mode

1. Add workflow config for `plan_advisor` and `plan_advisor_mode`, defaulting to enabled and
   `missing_info`.
2. Introduce an internal `planadvisor` package that inspects feature brief, plan, and spec files.
3. Detect missing planning coverage across strategic and feature-definition topics.
4. Generate at most 4 adaptive questions grouped into `big-thinker` and `feature-specialist`
   lenses.
5. Add a `centinela hook plan-advisor` command and wire it into Claude and OpenCode prompt hooks.
6. Keep `big-thinker` and `feature-specialist` as the evidence-producing roles; advisor mode only
   changes prompt behavior.
7. Add unit, integration, and acceptance tests for activation, suppression, adaptive questioning,
   and user-facing UX/mobile-first prompts.
