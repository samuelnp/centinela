# remove-panel-borders — senior-engineer

## Files Touched

- `internal/ui/panel.go`: `renderSystemPanel` now returns
  `renderSystemLine(...) + "\n\n" + body` — branded header line + body, no border.
  Removed the dead `panelStyle` helper.
- `internal/ui/styles.go`: removed `boxStyle` (only consumer was `panelStyle`)
  and `errorBoxStyle` (already had no consumer).

## Architecture Compliance

Pure `internal/ui` cosmetic change; no layer/import changes. `renderSystemLine`
(the persona+channel+title header) and `styleForTone` (tone colors) are
unchanged, so branding and color are preserved.

## Type-Safety Notes

Plain string concatenation instead of `lipgloss.JoinVertical` — the latter pads
every line to the block width, leaving trailing whitespace once the box (which
hid it) is gone. Concatenation yields clean lines. `lipgloss` import stays
(it's the return type of `styleForTone`). Verified: `centinela roadmap` renders
0 of `╭ ╮ ╰ ╯ │`; existing `internal/ui` tests remain green.

## Trade-Offs

Removed both box styles outright (the user chose "border box everywhere") rather
than gating behind a flag — simplest, and there is no remaining consumer.

## Handoff

→ qa-senior: unit (panel output has channel/title/body but no border chars),
integration (a real CLI panel renders border-free), acceptance (a binary command
that prints a panel shows the title with no border).
