# Plan: remove-panel-borders

## Summary

Drop the rounded border from every `renderSystemPanel` rendering. Keep the
branded header line + body. One-function change in `internal/ui/panel.go` plus
dead-style cleanup; existing content assertions are unaffected.

## Change

`internal/ui/panel.go`:
```go
func renderSystemPanel(channel, title string, t tone, body string) string {
	head := renderSystemLine(channel, title, t)
	return lipgloss.JoinVertical(lipgloss.Left, head, "", body)
}
```
(was: `return panelStyle(t).Render(content)`). Then remove `panelStyle` (only
used here). Check `boxStyle` consumers; delete it too if `renderSystemPanel` was
the only one, else leave it.

## Verification of blast radius

- Existing UI tests (`render_core_test`, `improve_centinela_render_ui_integration`)
  assert content (`🛡️👁️`, channel, title, body) — all preserved → still green.
- `renderSystemLine` (header) and tone colors unchanged → no color regression.

## Test strategy

- **unit** (`internal/ui`): assert `renderSystemPanel(...)` output contains the
  channel/title/body but **none** of `╭ ╮ ╰ ╯` (no rounded border); same for
  `RenderBlocked` and `RenderContext`.
- **integration**: a representative CLI panel (e.g. `RenderRoadmap` /
  `RenderDeliveryChoice`) renders border-free while keeping its content.
- **acceptance**: binary-driven — run a command that prints a panel (e.g.
  `centinela roadmap`) in a temp repo and assert the output has the title but no
  border chars.

## Rollout

Cosmetic-only; additive-safe. No config, no migration. Affects every panel
uniformly (CLI + hook directives), per the chosen scope.

## Risks

Low. Only risk is deleting a still-referenced `boxStyle` — guarded by building
the whole module before commit.
