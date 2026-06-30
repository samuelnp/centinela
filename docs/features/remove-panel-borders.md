# Feature: remove-panel-borders

## Problem

Centinela wraps its system messages in a rounded `╭──╮` border box
(`renderSystemPanel` → `boxStyle`). The user finds the border boxes noisy and
wants them gone from **all** centinela output — both CLI command panels
(ROADMAP PHASE OVERVIEW, DELIVER, MIGRATE, …) and hook directives (BLOCKED
WRITE, MIGRATION REQUIRED, REVIEW REQUIRED, …).

## Goal

Remove the rounded border from every `renderSystemPanel` rendering. Keep the
existing **content**: the branded header line (`🛡️👁️  <CHANNEL>  <TITLE>` with
its tone color) and the body below it. Output becomes a header line + body with
no box.

## Scope

- `internal/ui/panel.go`: `renderSystemPanel` returns the joined header + body
  directly instead of `panelStyle(t).Render(content)`.
- Remove now-dead `panelStyle` (and `boxStyle` if unused elsewhere — verify
  against the dashboard renderers before deleting).
- `renderSystemLine` (the header) and the tone coloring are unchanged.

## Out of scope

- The `🛡️👁️` persona label and the `<CHANNEL>` tag stay (the user chose to keep
  branding; only the border box is removed).
- Single-line `CLI` output (`RenderSuccess`/`RenderStep`) is already unboxed.
- No change to colors, icons, or message text.

## Acceptance

- `renderSystemPanel` / `RenderBlocked` / `RenderContext` / roadmap / deliver /
  migrate panels render **without** the rounded border characters
  (`╭ ╮ ╰ ╯ │ ─`) while still containing `🛡️👁️`, the channel, the title, and the
  body.

## Risks

- **Low.** Existing UI tests assert content (channel/title/persona/body), not
  border chars, so they keep passing. A new test pins the no-border behavior.
- Verify `boxStyle` has no other consumer before deleting it.
