# completion-delivery-prompt — senior-engineer

## Files Touched

| File | Lines | Change |
|------|-------|--------|
| `internal/gitutil/remote.go` | 35 | NEW leaf — `gitRun` seam + `HasOriginRemote` |
| `internal/gitutil/options.go` | 40 | NEW — `Option` type + `DeliveryOptions` matrix + `Supports` |
| `internal/gitutil/directive.go` | 37 | NEW — `GitHubCLIAvailable` + `DeliveryDirective` |
| `internal/ui/render_delivery.go` | 32 | NEW — `RenderDeliveryChoice` panel (read-only) |
| `cmd/centinela/deliver.go` | 58 | NEW — `deliver` cmd, required `--via`, matrix guard, merge dispatch |
| `cmd/centinela/deliver_pr.go` | 47 | NEW — PR path: push + `gh` with honest degradation |
| `cmd/centinela/complete.go` | 98 | EDIT — done-branch directive wiring (no side effects) |
| `centinela.toml` | — | EDIT — `internal/gitutil/**` added to the `leaf` import-graph layer |
| `PROJECT.md` | — | EDIT — G2 sentence registering `internal/gitutil` as a leaf |

All Go files are ≤100 lines.

## Architecture Compliance

- **G2 leaf placement.** `internal/gitutil` imports only stdlib + `os/exec` —
  no internal package. It is added to the `leaf` import-graph layer in
  `centinela.toml` and documented in PROJECT.md G2, mirroring
  `internal/golist`/`internal/importgraph`. `internal/ui` importing the leaf
  (read-only, for the `Option` type) is the same allowance already granted to
  `ui` for domain rendering types; build confirms no cycle. No new
  `workflow → worktree` edge is introduced; `internal/workflow` is read-only.
- **G1 sizes.** Every new/edited source file ≤100 lines (max is complete.go at 98).
- **G7 thin cmd.** `deliver.go` makes no matrix decisions of its own — the
  offered set comes from `gitutil.DeliveryOptions` and the `--via` validity
  from `gitutil.Supports`. `--via merge` composes `runMerge` unchanged (no merge
  logic reimplemented); `--via pr` delegates to `runDeliverPR`.
- **No side effects at completion.** The `complete` done-branch only calls
  read-only `HasOriginRemote(".")` and prints; it never pushes or merges.

## Type-Safety Notes

`Option` is a distinct string type (not bare `string`); `--via` is parsed into
it and validated against the two consts before any dispatch. The
`HasOriginRemote` error path distinguishes a real exec failure (`*exec.ExitError`
⇒ "no", everything else ⇒ error) so a missing remote is never confused with a
broken git. No `interface{}`/`any`. `go vet ./...` clean.

## Trade-Offs

- **Commit-if-dirty (rejected).** The plan flagged auto-commit as risky. I push
  the already-committed branch and, if the worktree is dirty, error clearly
  rather than silently committing unreviewed work. Auto-staging the user's tree
  during delivery is the more dangerous default; an explicit error is safer.
- **`--via pr` degradation.** With no origin it refuses entirely (no push). With
  origin but `gh` absent/unauthenticated it still pushes the branch, prints
  honest manual-PR instructions, and returns a non-nil error (non-zero exit) —
  it never claims a PR was opened.
- **Required-flag mechanics.** Cobra's `MarkFlagRequired("via")` rejects an
  *absent* flag with `required flag(s) "via" not set` (non-zero, no side
  effects); an *empty or unknown* value reaches RunE and yields
  `choose --via pr|merge`. Both paths refuse and exit non-zero, satisfying the
  spec's "refuses to act" assertions through two complementary guards.
- **PR title/body.** `gh pr create --fill` uses commit-derived defaults; a rich
  PR/CHANGELOG body is deliberately out of scope (`delivery-artifact-generation`).

## Deferred Findings

none

## Handoff

Next role: **qa-senior**. Testable seams: `DeliveryOptions` (pure, all 4 matrix
rows), `DeliveryDirective` (string shape, incl. empty-opts single-line), and
`HasOriginRemote` via the overridable `gitRun` var (stub a `*exec.ExitError` for
the no-remote case, a generic error for the broken-git case). The `--via` guard
is testable without a real remote. `cmd/centinela/deliver_pr.go` exposes an
overridable `gitDeliver` var for push-path tests. `complete`'s done-branch emits
the directive with only valid options — assert directive substrings + "no push /
no merge happened" against fixture state files, never a real GitHub PR.
