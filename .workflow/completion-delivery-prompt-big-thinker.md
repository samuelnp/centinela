### Big-Thinker Report: completion-delivery-prompt
**Date:** 2026-06-24

#### Problem

When a feature's 5-step workflow reaches `done`, `centinela complete` prints
`Workflow complete for %q!` and stops. The completed work then **stalls in its
worktree**: the orchestrating agent does not know the team's delivery
convention, so it either does nothing (work rots on a branch) or — worse —
guesses and pushes / merges without asking, violating the project's intent.

There is real infrastructure to deliver work but no bridge to it at completion:

- **Local merge** already exists end-to-end: `centinela merge <feature>`
  (+ `--continue` recovery) performs the hybrid merge into `main` via the
  Merge Steward (`internal/worktree/merger.go`, `cmd/centinela/merge*.go`).
- **Push + PR** is a plain `git push` + `gh pr create`, but nothing tells the
  agent that this is the right path when an `origin` remote exists.

The completion message is the natural hook point, and the **Merge Steward's
directive pattern** (a 2-line `CENTINELA DIRECTIVE: …` printed at the `merge`
boundary, mirrored by a styled UI panel) is the precedent to follow: at
completion, emit a directive that tells the agent to **ask the user how to
deliver**, listing **only the valid options for this repo's setup**, then give
a single command that **acts on the chosen option after confirmation**.

The roadmap entry requires Centinela to "act on the choice" and to "never push
or merge without explicit confirmation" — so the surface is a directive (asks
first) **plus** a command that executes the picked option (merge → delegate to
the existing flow; pr → commit + push + `gh pr create`).

#### Scope (In / Out)

**In scope**

1. **Completion-time directive.** At `complete.go`'s `done` branch, after the
   "Workflow complete" line, emit a 2-line `CENTINELA DIRECTIVE: …` that:
   instructs the agent to **ask the user** which delivery path to take, lists
   **only the valid options** (computed from remote presence + worktree mode),
   and names the exact `centinela deliver` invocation per option. A styled UI
   panel (`ui.RenderDeliveryChoice`) mirrors `RenderMergeStewardNeeded`.
2. **New `centinela deliver <feature> --via pr|merge` command.** Performs the
   chosen delivery. **`--via` is required (no default)** so the command can
   never act ambiguously / without an explicit pick.
   - `--via merge`: delegates to the **existing** local-merge flow
     (`runMerge`/`worktree.Merge` + steward dispatch). Does NOT reimplement
     merge — composes it.
   - `--via pr`: commit (if dirty), `git push -u origin <branch>`, then
     `gh pr create` (default title/body — rich body is out of scope). On
     success prints the PR URL.
3. **Remote-detection leaf** (`internal/gitutil`): `HasOriginRemote(repo)` via
   `git remote get-url origin`, and a `GitHubCLIAvailable()` check (`gh`
   presence). Pure leaf (stdlib + `os/exec`), importable by any layer.
4. **Valid-option matrix** computed from `{has-origin} × {worktree-mode}` (see
   Dependencies). Drives both the directive's listed options and a guard in
   `deliver` that refuses an option the repo can't support.
5. **Honest `gh` degradation.** When `--via pr` is picked but `gh` is missing
   or unauthenticated, push still happens (that part is plain git) and the
   command prints clear manual "open a PR at <compare-url>" instructions and
   exits non-zero rather than claiming a PR was opened.

**Out of scope**

- Rich PR description / `CHANGELOG` body composition →
  `delivery-artifact-generation` (this feature UNBLOCKS it; it supplies the
  default/empty body only).
- **Auto-delivery without confirmation** — never. The directive always asks.
- Non-GitHub remotes — `gh` is GitHub-specific. GitLab/Bitbucket/etc. native PR
  creation is explicitly out; for those, `--via pr` degrades to push + manual
  instructions (same path as "gh absent").
- Merge conflict resolution — owned by the Merge Steward; `deliver --via merge`
  inherits its `--continue` flow unchanged.
- Multi-feature delivery trains / queued delivery.

#### Dependencies & Assumptions

**Reused infra (compose, do not reimplement):**

- `centinela merge <feature>` (+ `--continue`) — `--via merge` calls the same
  `runMerge` path (`worktree.Merge` + `dispatchSteward`).
- `wf.WorktreePath` (`internal/workflow/state.go`) — empty ⇒ single-checkout;
  non-empty ⇒ worktree mode (a branch exists to merge).
- Directive precedent: `MergeOutcome.StewardDirective()` (2-line shape) +
  `ui.RenderMergeStewardNeeded` + the `merge_dispatch.go` print pattern.
- `branchName(feature)` (worktree) for the push refspec.

**New remote-detection leaf — `internal/gitutil`:**

- `HasOriginRemote(repo string) (bool, error)` — `git remote get-url origin`.
- `GitHubCLIAvailable() bool` — `exec.LookPath("gh")` (and optionally
  `gh auth status` for the unauth case).
- Pure **leaf**: imports stdlib + `os/exec` only; imports **no** internal
  package. Mirrors `internal/golist`/`internal/importgraph` (leaf git/exec
  wrappers).

**Layering / import-graph (precise edges):**

- `cmd/centinela/deliver.go` imports `internal/gitutil` (leaf), `internal/ui`,
  `internal/worktree`, `internal/workflow` — all already allowed for `cmd/`.
- `cmd/centinela/complete.go` already imports `internal/workflow`/`internal/ui`;
  it gains `internal/gitutil` to compute the option matrix for the directive.
- **The option-matrix decision is domain logic** but needs only the leaf
  (`HasOriginRemote`) + a bool (worktree mode). To keep `cmd/` thin (G7), place
  `DeliveryOptions(hasOrigin, worktreeMode bool) []Option` and the
  directive-string builder in `internal/gitutil` (a leaf that owns delivery
  concerns), so `cmd/` calls it directly. `internal/workflow` does **not**
  import it and does not decide delivery — this avoids the forbidden
  `workflow → worktree` edge entirely.
- **centinela.toml edit:** add `internal/gitutil/**` to the existing `leaf`
  layer `paths` (`[[gates.import_graph.layers]]` name=`leaf`). No new layer.
- **PROJECT.md G2 edit:** one sentence noting `internal/gitutil` is a leaf
  (stdlib + `os/exec`, GitHub/remote detection + delivery-option matrix),
  consumed by `cmd/` only, importing nothing internal — mirroring the
  `internal/golist`/`internal/importgraph` leaf wording.

**Assumptions:** delivery runs from the worktree (push its branch); `gh` when
present is GitHub-targeted; `origin` is the canonical remote name (non-`origin`
remotes treated as "no PR option").

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| `deliver` acts (push/merge) without explicit user confirmation | High — violates the core invariant | Medium | Directive **asks the user first**; `deliver` **requires explicit `--via`** (no default) so it can't act on an unstated choice; completion only emits the directive, never delivers. |
| `gh` absent or unauthenticated on `--via pr` | Medium — false "PR opened" claim, or hard crash | Medium | `GitHubCLIAvailable()` gate; on absence/unauth, still push, then print honest "open PR manually at <compare-url>", exit non-zero. Never claim a PR exists. |
| Directive offers an invalid option (e.g. PR with no remote) | Medium — agent runs a doomed command | Medium | Option matrix from `{has-origin}×{worktree-mode}`; directive lists only valid options; `deliver` re-guards and refuses an unsupported `--via`. |
| `deliver --via merge` duplicates / drifts from `centinela merge` | Medium — divergent merge behavior, double maintenance | Low | `--via merge` **composes** `runMerge` (same `worktree.Merge` + steward dispatch + `--continue`); no merge logic reimplemented. |
| `workflow → worktree` import added by mistake (G2 violation) | High — fails import_graph gate / breaks layering | Low | Delivery decision lives in the `internal/gitutil` **leaf** + `cmd/`; `internal/workflow` untouched. State.go only reads `WorktreePath`. |
| Push to a protected/diverged `origin` fails | Low–Medium — confusing error | Medium | Surface git's stderr verbatim with an actionable prefix; exit non-zero; worktree kept so the user can retry. |
| Non-GitHub remote present (GitLab etc.) | Low | Low | `--via pr` degrades to push + manual instructions (same as gh-absent); documented as out of scope for native PR creation. |

#### Rollout

Smallest correct slice first, then layer richness:

1. **Leaf + matrix (no side effects):** `internal/gitutil` with
   `HasOriginRemote`, `GitHubCLIAvailable`, `DeliveryOptions(matrix)`, and the
   directive-string builder. Unit-testable in isolation. Register in
   centinela.toml leaf layer + PROJECT.md G2.
2. **Completion directive:** wire the directive + `ui.RenderDeliveryChoice`
   into `complete.go`'s `done` branch (read-only; emits guidance only).
3. **`centinela deliver --via merge`:** compose the existing merge flow first
   (lowest new surface, reuses tested code).
4. **`centinela deliver --via pr`:** commit/push + `gh pr create` with honest
   degradation. Last, because it owns the only genuinely new side-effecting
   git/gh code.

Each step is independently shippable and gate-clean; the directive (1–2) gives
value even before `deliver` exists, since the listed commands are real.

#### Deferred Findings

none

#### Handoff — Next role: feature-specialist

Pin the exact directive wording (mirror `StewardDirective`'s 2-line shape),
the full option-matrix rows, the `deliver` command/flag surface and its
`--via` guard, the `gh`-degradation copy, and the per-file layout (each ≤100
lines): `internal/gitutil/{remote.go,options.go,directive.go}`,
`cmd/centinela/deliver.go` (+ `deliver_pr.go` if push/gh logic exceeds 100
lines), `internal/ui/render_delivery.go`, and the `complete.go` done-branch
wiring. Author `specs/completion-delivery-prompt.feature` mirroring
`merge-steward-auto-dispatch.feature` scenario style (option-matrix cases,
`--via` required, gh-absent degradation, no-origin → PR not offered).
