# completion-delivery-prompt

## Problem

When a feature's 5-step workflow reaches `done`, `centinela complete` prints
`Workflow complete for %q!` and stops. The completed work stalls in its
worktree: the orchestrating agent doesn't know the team's delivery convention,
so it either leaves the branch to rot or pushes/merges without asking. Centinela
already owns a local-merge flow (`centinela merge` + `--continue`) and the raw
ingredients for push + PR (`git push`, `gh pr create`), but nothing bridges
completion to either, and nothing detects which delivery options the repo even
supports.

## Who / Why

**Who.** The developer/operator (and the orchestrating agent) finishing a
feature in a Centinela-governed project, who must get completed work off the
worktree and into the team's delivery channel.

**Why.** Completion is the moment the delivery decision must be made, but it is
team- and repo-specific (open a PR vs. merge locally). Surfacing the valid
options and acting only on an explicit pick turns "stalled work / risky guess"
into "ask, confirm, deliver".

## In Scope

- A `CENTINELA DIRECTIVE: …` emitted at `complete.go`'s `done` branch telling
  the agent to **ask the user** how to deliver, listing only the **valid**
  options for this repo (from `origin`-remote presence × worktree mode), with
  the exact `centinela deliver` command per option. Mirrors the Merge Steward
  directive + a styled `ui.RenderDeliveryChoice` panel.
- `centinela deliver <feature> --via pr|merge` — performs the chosen delivery.
  `--via` is **required** (no default) so it can never act ambiguously.
  - `--via merge` composes the existing `centinela merge` flow (no
    reimplementation; `--continue` recovery inherited).
  - `--via pr` commits (if dirty), pushes the branch to `origin`, opens a PR
    via `gh` with a default body, prints the PR URL.
- New leaf `internal/gitutil`: `HasOriginRemote`, `GitHubCLIAvailable`, the
  `DeliveryOptions` matrix, and the directive-string builder.

## Out of Scope

- Rich PR description / `CHANGELOG` body (owned by
  `delivery-artifact-generation`, which this unblocks; default body only here).
- Auto-delivery without confirmation — never; the directive always asks.
- Native PR creation for non-GitHub remotes (`gh` is GitHub-specific) — those
  degrade to push + manual PR instructions.
- Merge conflict resolution (Merge Steward owns it).

## Acceptance Summary

- Completion at `done` emits a directive listing only the valid options and the
  matching `centinela deliver` commands; it never delivers on its own.
- With no `origin` remote, the PR option is not offered.
- `centinela deliver` with no `--via` (or an unsupported `--via` for the repo)
  refuses to act and exits non-zero.
- `--via merge` finalizes via the existing merge flow (incl. steward dispatch).
- `--via pr` pushes and opens a PR via `gh`; when `gh` is absent/unauth, it
  still pushes, prints honest manual-PR instructions, and exits non-zero
  without claiming a PR was opened.
