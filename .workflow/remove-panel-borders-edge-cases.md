# Edge Cases: remove-panel-borders

## Covered

| # | Edge case | Handling | Test |
|---|-----------|----------|------|
| 1 | A CLI command panel renders border-free | header + body, no box | `acceptance:TestAccRoadmapPanelHasNoBorder` (`centinela roadmap`) |
| 2 | A hook directive panel renders border-free | same path (renderSystemPanel) | `unit:TestPanelDirectiveHasNoBorder` (RenderBlocked), `ui:TestRenderBlockedHasNoBorder` |
| 3 | Multi-line body panel border-free | RenderContext active-workflows panel | `integration:TestActiveWorkflowsPanelHasNoBorder` |
| 4 | Branding preserved | 🛡️👁️ + channel + title kept | all tests assert the persona + channel/title content |
| 5 | Header-then-body structure (blank separator, no framing) | `head + "\n\n" + body` | `ui:TestRenderSystemPanelKeepsHeaderThenBody` |
| 6 | No rounded border chars anywhere | assert none of `╭ ╮ ╰ ╯ │` | every test checks `ContainsAny` |

## Residual Risks

- Single-line `CLI` output (`RenderSuccess`/`RenderStep`) was already unboxed and
  is untouched — unchanged behavior.
- Plain concatenation (not `lipgloss.JoinVertical`) avoids the trailing-whitespace
  padding the box used to hide; verified clean on `centinela roadmap`.
- `boxStyle`/`errorBoxStyle`/`panelStyle` removed — no remaining consumer (the
  build + full suite guard against a missed reference).
