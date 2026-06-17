# Feature Brief — precommit-and-pr-gate

> Phase 8 capstone (Continuous Governance). Run the mechanical gates (1) as a
> fast **pre-commit hook** that blocks a commit when gates fail on the staged
> changes, and (2) as a **PR gate** that posts gate verdicts as PR review
> comments. Builds on the existing gate suite, the audit ratchet, and custom
> gates — the gates fire at commit-time and PR-time, not just on demand.

## Problem

`centinela validate` runs gates on demand (and in CI), but nothing fires the
gates at the two moments developers actually want feedback: **before a commit
lands** (so a violation never enters history) and **on a PR** (so reviewers see
gate verdicts inline without reading raw CI logs). Today centinela manages only
Claude/OpenCode editor hooks (`.claude/settings.json`) — it has no git-hook
install path and no GitHub PR integration.

## What this adds

1. **Pre-commit gate** — `centinela precommit` runs the gates scoped to the
   **staged** changes and exits non-zero (blocking the commit) on failure. An
   installer wires `.git/hooks/pre-commit` to call it. Must be **fast** (staged
   diff-aware; skip the slow whole-repo gates by default).
2. **PR gate** — `centinela pr-gate` runs the gates scoped to the PR's changed
   files, renders a Markdown verdict, and surfaces it as a PR review comment
   (posted from CI). Reuses the `internal/verdict` Packet shape.

## Key decisions to resolve in the plan

- **Staged-diff support (must add).** `internal/gitdiff/resolver.go.ChangedFiles`
  only does `merge-base HEAD <base>` (branch diff). Pre-commit needs the index:
  `git diff --cached --name-only --diff-filter=ACMR` (+ staged adds). Decide the
  API: a new `ChangedFilesStaged()` vs a `staged bool` param. This is the load-
  bearing addition.
- **Pre-commit speed / gate selection.** The default config enables the build
  gate (6-target cross-compile) — far too slow for every commit. Decide how
  pre-commit stays fast: a `[precommit]` config that runs only diff-aware gates
  and **skips build/cross-compile by default** (or an explicit gate allowlist).
  Recommend skipping non-diff-aware/heavy gates in pre-commit, configurable.
- **Git-hook installer (new layer).** Centinela only manages Claude hooks today.
  Decide the install command (`centinela precommit install` or
  `centinela hook install`), how it writes `.git/hooks/pre-commit` **without
  clobbering an existing hook** (idempotent, marker-delimited block, backup), and
  whether `centinela doctor` reports/repairs it.
- **PR-comment delivery: render vs post.** Two options: (a) centinela renders
  Markdown to stdout and the CI workflow posts it via `gh pr comment` (testable,
  no network/`gh`/auth dependency in the Go binary), or (b) centinela shells out
  to `gh` directly (`--post`). Recommend (a) as the v1 default — `centinela
  pr-gate` emits the Markdown verdict + exit code; the CI yaml posts it — with an
  optional `--post` convenience. There is ZERO existing GitHub integration, so
  keep the binary network-free and put the GitHub plumbing in `.github/workflows`.
- **Markdown renderer.** `internal/ui/render_gates.go` is terminal-only
  (Lipgloss). Add a plain-Markdown renderer for `[]gates.Result` / the verdict
  Packet (pass/fail icons, collapsed details) — reused by the PR comment.
- **Config sections.** `[precommit]` (enabled, gate selection/skip) and
  `[pr_gate]` (enabled, fail_on_warning) following the `roadmap_drift` pattern.
- **Idempotent PR comment.** When CI re-runs, the PR gate should update its
  existing comment rather than spam new ones (a stable marker the poster finds).
  Decide whether this lives in centinela's `--post` or the CI yaml.

## Acceptance Criteria

1. `centinela precommit` runs the gates against the **staged** files and exits 0
   when they pass, non-zero when a `fail`-severity gate fails (blocking a commit).
2. The installer writes a `.git/hooks/pre-commit` that calls `centinela
   precommit`; it is **idempotent** (re-install is a no-op) and **preserves** any
   pre-existing hook content (marker-delimited block, no clobber).
3. Pre-commit is fast: it does not run the cross-compile/build gate by default;
   only staged-scoped diff-aware gates run unless configured otherwise.
4. `centinela pr-gate` runs the gates against the PR's changed files and emits a
   Markdown verdict (gate names, pass/fail, details) plus a pass/fail exit code.
5. The Markdown verdict is deterministic and readable as a GitHub PR comment.
6. The CI workflow posts the verdict as a PR review comment, **updating** the
   prior comment on re-runs (no duplicate spam).
7. Custom gates and the audit-baseline gate participate in both precommit and
   pr-gate exactly as in `validate` (same `Result` path).
8. Config-driven (`[precommit]`, `[pr_gate]`); disabled/absent → no behavior
   change. All new source files ≤100 lines; no new cross-layer import violations.

## Edge Cases

- Not a git repo / no staged changes → precommit passes cleanly (nothing to
  gate), exit 0, no error.
- Staged-diff git command fails → graceful degrade (full scan or skip with a
  notice), never a crash or a false block.
- An existing `.git/hooks/pre-commit` is present → installer appends its block,
  preserves the rest; uninstall removes only its block.
- `pr-gate` run outside a PR / missing `GITHUB_*` context → renders the verdict
  to stdout without posting; clear message, no crash.
- A warn-severity gate failing → reported in both surfaces but does not block
  (unless `fail_on_warning`).
- Re-running pr-gate on the same PR → updates the single marker comment.
- Huge diff / many violations → Markdown bounded/collapsible, comment stays
  readable.
- Windows pre-commit hook (`sh` shebang) — note portability.

## Data Model

New `config.PrecommitConfig{Enabled bool; Gates []string / SkipBuild bool; ...}`
and `config.PrGateConfig{Enabled bool; FailOnWarning bool}` under `Config`,
normalized in `applyDefaults`, validated in `validateConfig`. New
`internal/gitdiff` staged method. Reuse `internal/verdict.Packet` /
`gates.Result`. New Markdown renderer (`internal/ui` or `internal/prgate`).

## Integration Points

- **Run**: `gates.RunWithFilter(cfg, stagedFilter)` (precommit) /
  `gates.RunWithFilter(cfg, changedFilter)` (pr-gate); reuse
  `appendAuditGate`.
- **Commands**: `cmd/centinela/precommit*.go`, `cmd/centinela/pr_gate.go`,
  installer.
- **CI**: extend `.github/workflows/validate.yml` with a PR-only step that runs
  `centinela pr-gate` and posts via `gh pr comment` (updating a marker comment).
- **Doctor**: optionally report pre-commit hook install status.

## Risks

- **Pre-commit speed** — the make-or-break UX risk; must skip heavy gates and
  scope to staged files. A slow pre-commit gets disabled by developers.
- **Hook clobbering** — overwriting a user's existing `.git/hooks/pre-commit`
  would be destructive; the installer must be marker-delimited and idempotent.
- **GitHub coupling** — keep the Go binary network-free (render, don't post by
  default); GitHub plumbing lives in CI yaml to stay testable and portable.
- **Scope** — two surfaces in one feature; keep each minimal and reuse the
  existing gate/verdict machinery rather than re-implementing.
