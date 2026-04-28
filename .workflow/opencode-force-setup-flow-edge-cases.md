# Edge Cases: opencode-force-setup-flow

- Missing `PROJECT.md` plus greeting-only prompt: generated OpenCode rules must start setup questions, not suggest `centinela start <feature>`.
- Missing `PROJECT.md` plus vague work prompt: generated OpenCode rules still require setup questions and `PROJECT.md` first.
- Existing `PROJECT.md` but missing roadmap: generated OpenCode rules require roadmap bootstrap before feature discovery.
- Setup and roadmap complete: generated OpenCode rules still allow normal `centinela start <feature>` for feature work.
