# precommit-and-pr-gate — senior-engineer

## Files Touched

New (each ≤100 lines):
- `internal/gitdiff/staged.go` — `ChangedFilesStaged()` (`git diff --cached
  --name-only --diff-filter=ACMR`, `Base:"STAGED"`, degrade-safe).
- `internal/githooks/{install,splice}.go` — `package githooks`, **stdlib-only**
  (leaf layer, allow=[]): marker-delimited (`# >>> centinela >>>` / `<<<`)
  idempotent `Install`/`Uninstall`; preserves user hook content; `0o755`;
  deletes the file when only a bare shebang remains.
- `internal/ui/render_markdown.go` — `RenderGatesMarkdown([]gates.Result)`:
  plain markdown, `<!-- centinela:pr-gate -->` marker, count header, table,
  `<details>` per failing gate (Details capped). No Lipgloss; deterministic.
- `internal/config/{precommit,pr_gate}.go` — `PrecommitConfig{Enabled, SkipBuild}`
  (+ `RawSkipBuild *bool` so default-true is distinguishable from explicit-false,
  matching the repo's RawStepConfirmationMode pattern), `PrGateConfig{Enabled,
  FailOnWarning}`, with Normalize + validate.
- `cmd/centinela/precommit.go` (+ `precommit_skipbuild.go` `precommitCfg` shallow
  copy clearing `Gates.Build.Enabled`, + `precommit_install.go` install/uninstall),
  `cmd/centinela/pr_gate.go` (markdown to stdout, exit on Fail / Warn-if-config).

Edited: `internal/config/config.go` (+`Precommit`/`PrGate`; **extracted
`GatesConfig` to new `gates_config.go`** to keep config.go ≤100 — see below),
`defaults.go`, `file_size_exceptions.go`, `.github/workflows/validate.yml`
(pull_request-only pr-gate job: render → `gh pr comment --edit-last` w/ create
fallback; `permissions: pull-requests: write`; binary stays network-free),
`centinela.toml` (`internal/githooks/**` → leaf layer).

## Architecture Compliance

- **G1**: all files ≤100. NOTE: the senior-engineer's first pass pushed
  `config.go` to 102 (it was already at 100); fixed by extracting `GatesConfig`
  into `internal/config/gates_config.go` (config.go now 87) — caught by an
  independent `wc -l` sweep before the gate, not by `go test`.
- **G2 import-graph**: `internal/githooks` joins the **leaf** layer
  (allow=[]) — verified stdlib-only (no `samuelnp/centinela` imports). The
  markdown renderer in `internal/ui` adds NO new edge (`ui→gates` already
  exists). `import_graph` gate test passes. Scaffold-assets mirror is a **no-op**
  (template has no matrix — confirmed).
- **Network-free binary**: `pr-gate` renders + exits; CI does the `gh` posting.

## Type-Safety Notes

Strict Go, no `any`. `RawSkipBuild *bool` (toml `skip_build`) resolves
nil→true; `SkipBuild bool` is `toml:"-"`. `go vet`/`gofmt` clean.

## Trade-Offs

- precommit `skip_build` default true (cross-compile is the only slow,
  non-diff-aware gate) — keeps the hook fast; override per repo.
- render-in-Go / post-in-CI keeps the binary testable and dependency-free.

## Handoff

→ qa-senior. Verified by independent dogfood: staged-only gating (not-staged →
exit 0, staged oversized → exit 1 naming G1), installer fresh/idempotent/
preserve-user-hook/uninstall, and pr-gate marked Markdown verdict. Tests:
colocated `internal/gitdiff` (staged parse), `internal/githooks` (splice
idempotency / no-clobber / uninstall), `internal/ui` (markdown determinism),
`internal/config` (Normalize/validate, RawSkipBuild) — each ≤100 lines; tier
unit/integration over precommit/pr-gate; acceptance mapping the 15 scenarios.
