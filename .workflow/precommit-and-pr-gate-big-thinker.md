# precommit-and-pr-gate — big-thinker

## Problem

The mechanical gates only run on demand / in CI. Nothing fires them at the two
moments feedback matters most: before a commit lands (so a violation never
enters history) and on a PR (so reviewers see verdicts inline). Centinela today
manages only Claude/OpenCode editor hooks — no git-hook path, no PR integration.

## Scope

`centinela precommit` (gates on staged changes, blocks on fail) + an idempotent
`.git/hooks/pre-commit` installer; `centinela pr-gate` (gates on PR changed
files → deterministic Markdown verdict + exit code) posted from CI. Reuses the
gate suite, `appendAuditGate`, and the verdict shape. Built-ins, custom gates,
and the audit gate participate via the same `Result` path.

## Dependencies & Assumptions

- Builds on custom-gate-sdk + audit-baseline-ratchet (same `RunWithFilter`).
- MUST add staged-diff support to `internal/gitdiff` (`git diff --cached`).
- Config = checked-in/trusted; the Go binary stays network-free (render, don't
  post); GitHub plumbing lives in `.github/workflows`.

## Risks

- **Pre-commit speed** (make-or-break): skip the cross-compile/build gate by
  default (`[precommit] skip_build=true` → shallow cfg copy with
  `Gates.Build.Enabled=false`) and scope to staged files. A slow hook gets
  disabled.
- **Hook clobbering**: the installer must be marker-delimited (`# >>> centinela
  >>>` … `# <<< centinela <<<`), idempotent, preserve pre-existing hook content,
  set `0o755`, MkdirAll the hooks dir.
- **GitHub coupling**: zero existing integration; keep `pr-gate` to stdout +
  exit code, CI posts/updates one marker comment via `gh pr comment`.

## Rollout

Config-gated; absent/disabled → byte-identical behaviour. Safe defaults
(`skip_build=true`, `fail_on_warning=false`).

## Handoff

→ feature-specialist. Plan at `docs/plans/precommit-and-pr-gate.md`. Verified:
`internal/githooks` joins the **leaf** layer (allow=[] → stdlib-only, which the
installer is) — one `centinela.toml` line; the markdown renderer in `internal/ui`
adds NO new edge (`ui→gates` already exists). NOTE for code step: the
`internal/scaffold/assets/centinela.toml` mirror is a **no-op** — that template
has no import-graph matrix (confirmed 0 entries), so skip it. `--post`, direct
GitHub API, and doctor integration are out of v1 scope.
