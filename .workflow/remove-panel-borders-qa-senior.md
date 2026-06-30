# remove-panel-borders — qa-senior

## Test Inventory

- **Colocated** `internal/ui/panel_test.go`: `renderSystemPanel` + `RenderBlocked`
  have no border chars; header-then-body structure; content preserved.
- **unit** `tests/unit/remove_panel_borders_unit_test.go`: `ui.RenderBlocked`
  directive border-free with branding.
- **integration** `tests/integration/...`: `ui.RenderContext` active-workflows
  panel border-free with content.
- **acceptance** `tests/acceptance/...`: binary `centinela roadmap` prints the
  PHASE OVERVIEW header with zero `╭ ╮ ╰ ╯ │`. All files ≤100 lines.

## Coverage Gaps

None. Total 95.0% ≥ gate; the change removed dead styles (fewer statements) and
`renderSystemPanel` is exercised by both colocated and existing UI tests.

## Acceptance Wiring

`specs/remove-panel-borders.feature` → `TestAccRoadmapPanelHasNoBorder`
(CLI panel), plus the unit/integration tiers for hook + multi-line panels.
`centinela.toml` already runs `go test ./tests/acceptance/...`.

## Handoff

→ validation-specialist: full suite + gates green; produce the gatekeeper report.
