# Plan: opencode-setup-priority

## Scope

Make OpenCode treat Centinela setup guidance as the first instruction to follow when project bootstrap artifacts are missing.

## Work Items

1. Update generated OpenCode instructions so setup directives override casual conversation.
2. Adjust OpenCode plugin prompt injection to surface setup and migration directives before other context.
3. Add regression tests covering greeting-only prompts with missing `PROJECT.md`.
4. Replace Claude-only README setup wording with agent-neutral guidance.

## Validation

- `go test ./...`
- Generated OpenCode assets contain the stronger setup instructions and prompt ordering.

## Compatibility

- Keep existing `centinela hook` commands unchanged.
- Preserve feature-intent autostart behavior when setup is already complete.
