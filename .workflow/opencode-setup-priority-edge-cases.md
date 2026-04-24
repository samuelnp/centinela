# Edge Cases: opencode-setup-priority

- Greeting-only prompts like `Hi` must still trigger setup-first guidance when `PROJECT.md` is missing.
- Feature-intent autostart must remain lower priority than setup and migration directives during bootstrap.
- Existing `opencode.json` instructions must keep unrelated custom entries while appending `AGENTS.md` and `CLAUDE.md` in that order.
- Plugin prompt handling must prepend setup text without breaking `output.prompt` or `output.context` variants.
