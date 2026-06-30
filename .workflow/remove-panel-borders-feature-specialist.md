# remove-panel-borders — feature-specialist

## Behavior Summary

`renderSystemPanel` renders a branded header line (`🛡️👁️ <CHANNEL> <TITLE>`,
tone-colored) followed by the body, with NO rounded border. Applies to every
panel: CLI command output and hook directives alike.

## Acceptance Criteria (Gherkin)

See `specs/remove-panel-borders.feature`: CLI panel border-free; hook directive
border-free; branding preserved; single-line CLI output unchanged.

## UX States

- **Panel**: header line + blank + body, no box.
- **Branding**: 🛡️👁️ + channel tag + title retained, tone color retained.
- **Single-line CLI** (`RenderSuccess`): already unboxed → unchanged.

## Edge Cases

- Both CLI panels and hook directives lose the border (scope: everywhere).
- No rounded border chars `╭ ╮ ╰ ╯` remain; content preserved.
- `boxStyle` removed only if it has no other consumer (build guards it).

## Out-of-Scope

- Persona label / channel tag / colors / message text — all kept.

## Handoff

→ senior-engineer: edit `renderSystemPanel` to drop the border; remove dead
style; keep each file ≤100 lines.
