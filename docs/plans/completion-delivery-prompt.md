# Plan: completion-delivery-prompt

## Mechanism (decided)

A **directive at completion** + a **`centinela deliver` command that acts on the
choice**. This mirrors the Merge Steward pattern exactly: the Go binary never
delivers on its own at completion — it emits a `CENTINELA DIRECTIVE: …` telling
the orchestrator to **ask the user** which delivery path to take (listing only
the valid options for this repo), and the agent runs `centinela deliver
<feature> --via pr|merge` only after the user confirms. `--via` is required
(no default), so the command cannot act on an unstated choice.

## Resolved surface

```
centinela deliver <feature> --via pr      # commit (if dirty) → push origin → gh pr create
centinela deliver <feature> --via merge   # delegate to the existing local-merge flow
```

- `--via` is a **required** string flag; an empty or unsupported value errors,
  no-op, exits non-zero.
- `--via merge` composes `runMerge(feature)` (same `worktree.Merge` + steward
  dispatch + `--continue` recovery). No merge logic is reimplemented.
- `--via pr` runs the PR path (commit/push/`gh`) with honest degradation.
- `centinela complete` at the `done` branch prints a styled panel + a 2-line
  directive; it **never** delivers.

## Option matrix

`worktree mode` = `wf.WorktreePath != ""`. `has-origin` = `gitutil.HasOriginRemote`.

| has-origin | worktree mode | Options offered (directive + `deliver` guard) |
|-----------|---------------|-----------------------------------------------|
| yes | yes | `--via pr`  **and**  `--via merge` |
| yes | no  | `--via pr` only (no worktree branch to local-merge) |
| no  | yes | `--via merge` only (no remote to push/PR) |
| no  | no  | none — directive states "no delivery target detected; configure an `origin` remote or use worktree mode" |

`deliver` re-checks the same matrix and refuses a `--via` the repo can't
support (e.g. `--via pr` with no origin), so the directive and the command
agree.

## Remote-detection leaf — `internal/gitutil`

A new **leaf** package (stdlib + `os/exec` only; imports no internal package),
mirroring `internal/golist`/`internal/importgraph`. Split for ≤100 lines/file:

- `internal/gitutil/remote.go` — `HasOriginRemote(repo string) (bool, error)`
  via `git remote get-url origin` (exit-nonzero ⇒ false, nil). Overridable
  `gitRun` var for tests, mirroring `worktree.gitRunner`.
- `internal/gitutil/options.go` — `Option` type + `DeliveryOptions(hasOrigin,
  worktreeMode bool) []Option`. Encodes the matrix above.
- `internal/gitutil/directive.go` — `GitHubCLIAvailable() bool`
  (`exec.LookPath("gh")`); `DeliveryDirective(feature string, opts []Option)
  string` building the 2-line `CENTINELA DIRECTIVE` (see below).

### Import edges

- `cmd/centinela/complete.go` → `internal/gitutil` (leaf): allowed for `cmd/`.
- `cmd/centinela/deliver.go` → `internal/gitutil`, `internal/ui`,
  `internal/worktree`, `internal/workflow`: all already allowed for `cmd/`.
- `internal/gitutil` imports nothing internal — **no** `workflow → worktree`
  edge is introduced. `internal/workflow` is untouched (state.go is read only).

### centinela.toml + PROJECT.md edits (exact)

- **centinela.toml** `[[gates.import_graph.layers]]` name=`leaf`: append
  `"internal/gitutil/**"` to `paths`. Add a comment noting it is a stdlib +
  `os/exec` git/remote + delivery-option leaf consumed by `cmd/` only.
- **PROJECT.md** G2 paragraph: add one sentence — "`internal/gitutil` is a leaf
  (stdlib + `os/exec`): `origin`-remote / `gh` detection and the delivery-option
  matrix for `centinela deliver`; it imports nothing internal and is consumed by
  `cmd/` only, mirroring `internal/golist`/`internal/importgraph`."

## Directive format (mirror StewardDirective's 2-line shape)

```
CENTINELA DIRECTIVE: %q is complete — ask the user how to deliver it; do NOT push or merge without their explicit choice.
Valid options: <pr: `centinela deliver %s --via pr`> <merge: `centinela deliver %s --via merge`>. Run only the one the user picks.
```

The second line lists only the options from the matrix (one, two, or — when
neither applies — a "no delivery target detected" sentence). Imperative line +
details line, exactly like `MergeOutcome.StewardDirective()`.

## UI renderer

`internal/ui/render_delivery.go` — `RenderDeliveryChoice(feature string, opts
[]Option) string` using `renderSystemPanel("DELIVER", "CHOOSE DELIVERY", …)`,
mirroring `RenderMergeStewardNeeded`. Read-only; lists the valid options with
their commands. ≤100 lines.

## cmd wiring

- `cmd/centinela/complete.go` `done` branch (after line 81): compute
  `hasOrigin, _ := gitutil.HasOriginRemote(".")`; `worktreeMode := wf.WorktreePath
  != ""`; `opts := gitutil.DeliveryOptions(hasOrigin, worktreeMode)`; print
  `ui.RenderDeliveryChoice(feature, opts)` then
  `gitutil.DeliveryDirective(feature, opts)`. Guidance only — no side effects.
- `cmd/centinela/deliver.go` — new Cobra command. Required `--via` flag;
  validate against the live matrix; `--via merge` → `runMerge(feature)`;
  `--via pr` → `runDeliverPR(feature, wf)`. Thin orchestrator (G7).
- `cmd/centinela/deliver_pr.go` — the PR path: commit-if-dirty, `git push -u
  origin <branchName(feature)>`, then if `GitHubCLIAvailable()` →
  `gh pr create` (default title/body, print URL); else push-only + honest
  "open a PR manually at <compare URL>" + exit non-zero. Split here to keep
  both cmd files ≤100 lines.

## File layout (each ≤100 lines)

| File | Role |
|------|------|
| `internal/gitutil/remote.go` | `HasOriginRemote` + overridable `gitRun` |
| `internal/gitutil/options.go` | `Option` + `DeliveryOptions` matrix |
| `internal/gitutil/directive.go` | `GitHubCLIAvailable` + `DeliveryDirective` |
| `internal/ui/render_delivery.go` | `RenderDeliveryChoice` panel |
| `cmd/centinela/deliver.go` | `deliver` command, `--via` guard, merge dispatch |
| `cmd/centinela/deliver_pr.go` | PR path (commit/push/gh + degradation) |
| `cmd/centinela/complete.go` | done-branch directive wiring (edit) |
| `centinela.toml` / `PROJECT.md` | register `internal/gitutil` leaf (edit) |

## Tests (tests step)

- **Unit:** `DeliveryOptions` for all four matrix rows; `HasOriginRemote` with
  a stubbed `gitRun`; `DeliveryDirective` string shape; `--via` required /
  unsupported guard.
- **Integration:** `complete` done-branch emits the directive with only valid
  options (origin present vs. absent); `deliver --via merge` composes the merge
  flow.
- **Acceptance** (`tests/acceptance/`, mirroring merge-steward style): no
  origin → PR option not offered; `deliver` without `--via` errors non-zero;
  `--via pr` with `gh` absent pushes + prints manual instructions + exits
  non-zero. Plus `.workflow/completion-delivery-prompt-edge-cases.md`.

## Out of scope

Rich PR/CHANGELOG body (→ `delivery-artifact-generation`); auto-delivery without
confirmation; native PR creation for non-GitHub remotes; merge conflict
resolution (Merge Steward).
