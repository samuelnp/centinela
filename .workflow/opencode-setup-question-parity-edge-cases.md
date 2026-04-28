# Edge Cases: opencode-setup-question-parity

- Missing `PROJECT.md` in OpenCode after greeting: setup directive must ask the exact six questions, not a compressed subset.
- Missing `PROJECT.md` in Claude after greeting: the same shared setup directive keeps the six-question checklist aligned.
- `PROJECT.md.template` must still be read before asking questions.
- Setup directive must not mention `centinela start <feature>` before project configuration exists.
