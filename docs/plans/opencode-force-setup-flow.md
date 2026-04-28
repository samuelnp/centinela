# Plan: opencode-force-setup-flow

## Scope

Make OpenCode treat missing `PROJECT.md` and missing roadmap bootstrap as mandatory setup work, not as a reason to ask for a feature or suggest `centinela start <feature>`.

## Work Items

1. Update generated `AGENTS.md` OpenCode instructions to define bootstrap precedence before feature workflow rules.
2. Add regression tests that assert generated OpenCode guidance forbids feature prompts during setup.
3. Verify setup-render tests still ensure no `centinela start` appears in setup guidance.

## Validation

- `go test ./...`
- `centinela validate`

## Compatibility

- Preserve existing `centinela start <feature>` guidance for configured projects.
- Keep OpenCode plugin and config file locations unchanged.
- Do not alter existing unmanaged-file migration behavior.
