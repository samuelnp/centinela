# Plan: opencode-setup-question-parity

## Scope

Make the missing-`PROJECT.md` setup directive require the same six-question checklist across Claude and OpenCode.

## Work Items

1. Update `internal/ui/render_setup.go` to replace category-only guidance with exact setup question labels.
2. Extend setup rendering tests to assert each question is present.
3. Keep existing constraints: read `PROJECT.md.template`, avoid `centinela start`, and hand off to roadmap setup after `PROJECT.md` is written.

## Validation

- `go test ./...`
- `centinela validate`

## Compatibility

- No changes to OpenCode plugin event names or config paths.
- No changes to the workflow state model.
- Existing setup behavior remains a prompt directive, not an automatic file write.
