# remove-panel-borders — big-thinker

## Problem

Centinela wraps system messages in a rounded ╭──╮ border box
(`renderSystemPanel` → `boxStyle`). The user wants the border boxes gone from
all output (CLI command panels + hook directives), keeping the content.

## Scope

`renderSystemPanel` returns the branded header line + body without the border.
Remove dead `panelStyle` (and `boxStyle` if unused). Header (`renderSystemLine`)
and tone colors unchanged.

## Dependencies & Assumptions

- No dependency features. Pure `internal/ui` cosmetic change.
- Existing UI tests assert content (🛡️👁️, channel, title, body), not border
  chars, so they stay green; a new test pins the no-border behavior.

## Risks

- Low. Only risk: deleting a still-referenced `boxStyle` — guarded by building
  the whole module before commit.

## Rollout

Cosmetic-only, no config/migration. Affects every panel uniformly per the
user's chosen "everywhere" scope.

## Handoff

→ feature-specialist: encode no-border + content-preserved + branding-kept as
acceptance scenarios.
