---
id: bacb676c33af8e1a
feature: remove-panel-borders
step: tests
type: lesson
title: - Single-line `CLI` output (`RenderSuccess`/`RenderStep`) was already unboxed and
tags: edge-cases, lesson
sourceArtifact: .workflow/remove-panel-borders-edge-cases.md
createdAt: 2026-06-29T20:44:23Z
---

- Single-line `CLI` output (`RenderSuccess`/`RenderStep`) was already unboxed and
- Plain concatenation (not `lipgloss.JoinVertical`) avoids the trailing-whitespace
- `boxStyle`/`errorBoxStyle`/`panelStyle` removed — no remaining consumer (the
