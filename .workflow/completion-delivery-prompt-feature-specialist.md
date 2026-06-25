### Feature-Specialist Report: completion-delivery-prompt
**Date:** 2026-06-24

#### Behavior Summary

When a feature's 5-step workflow reaches `done`, `centinela complete` emits a
2-line `CENTINELA DIRECTIVE` (mirroring the Merge Steward's shape) plus a styled
panel that tell the orchestrating agent to **ask the user how to deliver** the
completed work, listing **only the delivery options valid for this repo** — the
options come from the matrix `{origin-remote presence} × {worktree mode}` and
name the exact `centinela deliver <feature> --via pr|merge` command per option.
Completion itself never pushes or merges; it emits guidance text only. The new
`centinela deliver <feature>` command performs the chosen delivery: `--via` is
**required** (no default) so the command can never act on an unstated choice.
`--via merge` composes the existing `centinela merge` flow (clean merge removes
the worktree; a conflict reuses the merge-steward dispatch and `--continue`
recovery), and `--via pr` commits/pushes the branch and opens a PR via `gh` —
but if `gh` is absent or unauthenticated it still pushes, prints honest
manual-PR instructions, and exits non-zero, never claiming a PR was opened.
`--via pr` with no `origin` remote refuses to act entirely (no push).

#### Gherkin Scenarios   (reference specs/completion-delivery-prompt.feature)

The spec at `specs/completion-delivery-prompt.feature` defines 11 scenarios.
Each `Then` is concretely assertable against a fixture repo (with/without an
`origin` remote, with/without a worktree path, with `gh` stubbed present/absent)
— the assertions are exit code + emitted directive/message substrings + "no push
/ no merge happened", never a real GitHub PR.

1. **Origin + worktree → both options** — directive names the feature, lists
   both `--via pr` and `--via merge`, states "no push/merge without explicit
   choice"; no side effects; exit 0.
2. **No origin → merge only** — directive lists `--via merge`, lists no `--via
   pr`; exit 0.
3. **Single-checkout + origin → PR only** — directive lists `--via pr`, lists no
   `--via merge`; exit 0.
4. **Neither → no delivery target** — directive states none detected, lists no
   options; exit 0.
5. **Directive never delivers** — text only, no `git push`, no PR, worktree
   kept, branch not merged.
6. **deliver without `--via`** — reports "`--via pr|merge` must be chosen", no
   push/merge, exit non-zero.
7. **deliver `--via pr` with no origin** — reports "no origin remote — PR
   delivery unavailable", no push, no PR, exit non-zero.
8. **deliver `--via merge` clean** — finalizes through the existing merge flow,
   worktree removed, no pending marker, exit 0.
9. **deliver `--via merge` conflicted** — writes pending marker, emits the
   merge-steward CENTINELA DIRECTIVE, worktree kept, exit non-zero.
10. **deliver `--via pr` with origin + gh** — branch pushed to origin, PR opened
    via `gh`, PR URL reported, exit 0.
11. **deliver `--via pr` with gh absent** — branch still pushed, manual-PR
    instructions printed, does NOT claim a PR was opened, exit non-zero.

#### UX States  (table)

| State | Surface | Behavior |
|-------|---------|----------|
| Completion / delivery prompt | CLI (complete `done` branch) | Styled `RenderDeliveryChoice` panel + 2-line `CENTINELA DIRECTIVE` listing only valid options and their `centinela deliver` commands; exit 0; no side effects |
| Success — PR delivered | CLI (`deliver --via pr`) | Branch pushed to origin, PR opened via `gh`, PR URL printed; exit 0 |
| Success — merge delivered | CLI (`deliver --via merge`, clean) | Worktree merged + removed via existing flow; exit 0 |
| Error — no `--via` | CLI (`deliver`) | "choose --via pr|merge" message; no push/merge; exit non-zero |
| Error — PR without origin | CLI (`deliver --via pr`) | "no origin remote — PR delivery unavailable"; no push; exit non-zero |
| Degraded — gh missing on PR | CLI (`deliver --via pr`) | Push still happens; manual-PR instructions printed; no false "PR opened"; exit non-zero |
| Blocked — merge conflict | CLI (`deliver --via merge`) | Pending marker + merge-steward directive (inherited); worktree kept; exit non-zero |
| Loading | n/a | CLI; no async loading state |

#### Out-of-Scope

- Rich PR description / `CHANGELOG` body composition — owned by
  `delivery-artifact-generation`; this feature supplies the default/empty body
  only and unblocks it.
- Auto-delivery without confirmation — never; the directive always asks and
  `deliver` requires an explicit `--via`.
- Native PR creation for non-GitHub remotes (`gh` is GitHub-specific) — those
  degrade to push + manual instructions (same path as gh-absent).
- Re-implementing the merge flow — `--via merge` composes `centinela merge`
  (steward dispatch + `--continue`); conflict resolution stays with the Merge
  Steward.

#### Deferred Findings

none

#### Handoff — Next role: senior-engineer

Implement per the plan's file layout (each ≤100 lines): the leaf
`internal/gitutil/{remote.go,options.go,directive.go}`
(`HasOriginRemote`, `GitHubCLIAvailable`, `DeliveryOptions(hasOrigin,
worktreeMode)`, `DeliveryDirective`), `internal/ui/render_delivery.go`
(`RenderDeliveryChoice`), `cmd/centinela/deliver.go` (+ `deliver_pr.go` for the
push/gh path), and the `complete.go` `done`-branch wiring. Register
`internal/gitutil/**` in centinela.toml's `leaf` import-graph layer and add the
PROJECT.md G2 sentence. Keep `gitRun` overridable for unit tests (mirror
`worktree.gitRunner`) so detection and the `--via` guard are testable without a
real remote. Testable seams: `DeliveryOptions` (pure, all 4 matrix rows),
`DeliveryDirective` (string shape), `HasOriginRemote` (stubbed `gitRun`), and
`--via` gating (no remote needed). The directive wording must mirror
`MergeOutcome.StewardDirective()`'s 2-line imperative+details shape.
