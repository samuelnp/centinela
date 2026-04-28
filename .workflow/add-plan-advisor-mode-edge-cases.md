# Edge Cases — add-plan-advisor-mode

- Advisor mode must stay silent outside the `plan` step.
- Advisor mode must not repeat questions when the feature brief, plan, or spec already cover those topics.
- User-facing features should receive UX and mobile-first discovery prompts only when those details are still missing.
- Advisor mode must cap each round at 4 questions even when many planning gaps remain.
