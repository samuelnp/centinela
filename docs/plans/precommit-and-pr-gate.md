# Implementation Plan — precommit-and-pr-gate

> Feature brief: `docs/features/precommit-and-pr-gate.md`.
> Spec: `specs/precommit-and-pr-gate.feature`.

Phase 8 capstone (Continuous Governance). The mechanical gate suite already
exists and runs on-demand via `centinela validate` (`gates.RunWithFilter(cfg,
filter)` + `appendAuditGate`). This feature fires the **same** gate machinery at
the two moments developers want feedback — **before a commit lands** and **on a
PR** — without re-implementing any gate. Two new command surfaces, each reusing
`RunWithFilter` + `appendAuditGate`:

1. **`centinela precommit`** — runs gates scoped to the **staged index**
   (`git diff --cached`) and exits non-zero on a fail-severity gate, blocking the
   commit. Stays fast by skipping the cross-compile build gate by default.
   A marker-delimited installer wires `.git/hooks/pre-commit` to call it.
2. **`centinela pr-gate`** — runs gates scoped to the PR's changed-since-base
   files (reusing `Resolver.ChangedFiles`), renders a deterministic **Markdown**
   verdict to stdout, exits 0/1. The CI yaml posts the Markdown as an
   idempotent PR comment via `gh pr comment`. The Go binary stays network-free.

The load-bearing addition is **staged-diff support** in `internal/gitdiff`
(the resolver only does branch-diff today). Everything else reuses existing
machinery: gate execution, the audit gate seam, the `gates.Result` shape, and a
new **plain-Markdown renderer** alongside the terminal Lipgloss one.

## Decisions (DECIDED)

1. **Staged-diff = a new `staged bool` param is rejected; ship a dedicated
   `ChangedFilesStaged()` method (see Staged-diff CALL-OUT).** It has a different
   git invocation (`git diff --cached`, no `merge-base`/base ref, no
   `ls-files --others`), so overloading `ChangedFiles(base, includeUntracked)`
   with a third bool would muddy both call sites. A separate method keeps each
   resolver path single-purpose and independently testable, and degrades on the
   initial-commit / not-a-repo edges the brief calls out.

2. **Pre-commit speed = skip the build gate via a filtered cfg copy, not a
   gate allowlist.** The build gate (`checkBuild`) cross-compiles 6 release
   targets and is the **only** non-diff-aware heavy gate — it takes no filter, so
   it always runs whole-repo regardless of staged scope. A `[precommit]
   skip_build` knob (default **true**) produces a shallow cfg copy with
   `Gates.Build.Enabled = false`, then calls the normal
   `RunWithFilter(cfgCopy, stagedFilter)` + `appendAuditGate`. Every other gate
   is already diff-aware (it honors the filter), so scoping to the staged set
   makes them fast for free. A copy (vs. an allowlist) is minimal, needs no new
   gate-selection plumbing, and keeps `precommit` on the identical code path as
   `validate` (AC-7: custom + audit gates participate unchanged).

3. **Command surface — two cobra groups, no validate coupling.**
   - `centinela precommit` (default `RunE`) — staged-scoped gate run; render via
     the plain-Markdown renderer to **stderr**, exit 1 iff a fail-severity gate
     fails (`!gates.AllPassed(results)`), respecting `[precommit]`.
   - `centinela precommit install` / `centinela precommit uninstall` — write /
     remove the marker block in `.git/hooks/pre-commit` (see Installer CALL-OUT).
   - `centinela pr-gate` (default `RunE`) — changed-since-base gate run; render
     Markdown verdict to **stdout**; exit 1 on fail (or on warn when
     `[pr_gate] fail_on_warning`). `--post` deferred (Out-of-scope).
   We mount `install`/`uninstall` under `precommit` (not `hook`, which is
   reserved for the Claude/OpenCode editor-hook family in `main.go`) so the
   git-hook layer is clearly distinct from the editor-hook layer.

4. **PR delivery = render-in-Go, post-in-CI (see PR delivery CALL-OUT).**
   `centinela pr-gate` emits Markdown to stdout + an exit code and does **zero**
   network I/O. The CI workflow captures stdout and posts it with `gh pr comment`,
   updating a single marker comment on re-runs. There is no existing GitHub
   integration in the repo (no `gh`, no API client), so keeping the binary
   network-free keeps it unit-testable and portable; the GitHub plumbing lives in
   `.github/workflows/validate.yml` where `gh` and `GITHUB_TOKEN` already exist.

5. **Markdown renderer lives in `internal/ui` (`render_markdown.go`), not a new
   package.** It is pure formatting over `[]gates.Result` (the same input as the
   existing `RenderGateResult`), reused by both `precommit` (stderr) and `pr-gate`
   (stdout). Putting it next to `render_gates.go` avoids a new package + new
   import-graph edge. It is Lipgloss-free (plain `###`/`|`/`<details>` text) so
   output is byte-deterministic (AC-5).

6. **Git-hook installer = `internal/githooks/` new leaf package.** Pure file I/O
   over `.git/hooks/pre-commit` with marker-delimited idempotent splice (see
   Installer CALL-OUT). It imports nothing from the project (stdlib only), so it
   is a **leaf** — added to the leaf layer's `paths`. `cmd/` wires it.

7. **Config — `[precommit]` + `[pr_gate]` as top-level `Config` sections**
   (siblings of `[validate]`, not under `[gates]`), mirroring the
   `RoadmapDriftConfig` Normalize+validate pattern. Both default to safe values:
   `precommit.skip_build = true`; `pr_gate.fail_on_warning = false`. `Enabled`
   governs only the installer/CI advisory surface — the commands themselves run
   whenever invoked (a hook that calls `precommit` is itself the opt-in).

8. **Doctor integration deferred** (Out-of-scope) — keeps v1 to the two surfaces.

## Staged-diff (CALL-OUT — the load-bearing addition)

`Resolver.ChangedFiles` does `merge-base HEAD <base>` then
`git diff --name-only --diff-filter=ACMR <merge-base>` + `git ls-files --others`.
Pre-commit needs the **index**, not the branch delta. Add to
`internal/gitdiff/resolver.go`:

```go
// ChangedFilesStaged returns the set of files staged in the index
// (git diff --cached), the input to the pre-commit gate run. It takes no base
// ref (the index is compared against HEAD, or against the empty tree on the
// initial commit). On any git failure it returns
// (nil, Summary{Degrade: reason}, nil) so the caller degrades — never crashes,
// never false-blocks (brief edge cases).
func (r *Resolver) ChangedFilesStaged() (*Set, Summary, error)
```

Behavior:
- Runs `git diff --cached --name-only --diff-filter=ACMR` (A=added, C=copied,
  M=modified, R=renamed — same filter as `ChangedFiles`; D=deleted excluded
  because a deleted file can't violate a file-scoped gate). `--cached` compares
  the index against `HEAD`, and on a repo with **no commits yet** git diffs the
  index against the empty tree automatically — so the initial-commit case needs
  no special revision, it just lists every staged file.
- **No `ls-files --others`** — untracked-but-unstaged files are not part of a
  commit, so they must not be gated (AC-1 scope = staged only).
- `Summary{Base: "STAGED", Files: set.Len()}` for the header; on failure
  `Summary.Degrade` is set via the existing `degradeReason`-style mapping
  (`"not a git repository"` when not a repo).
- **Not-a-repo / empty staged set** both yield a usable result: not-a-repo →
  `Degrade` set, caller treats as "nothing to gate, exit 0"; empty staged set →
  `NewSet(nil)` (`Len()==0`), gates run over an empty filter and pass (AC edge:
  no staged changes → exit 0 cleanly).

The precommit command consumes it:
```go
set, summary, _ := gitDiffResolver.ChangedFilesStaged()
if summary.Degrade != "" { /* notice + exit 0, never block */ }
results := appendAuditGate(cfg, gates.RunWithFilter(precommitCfg(cfg), set))
```
where `precommitCfg(cfg)` is the skip-build copy (Decision #2).

## Git-hook installer safety (CALL-OUT — destructive if wrong)

Overwriting a user's `.git/hooks/pre-commit` is data loss. The installer in
`internal/githooks/install.go` performs a **marker-delimited splice**, never a
clobber:

```go
const (
    BeginMarker = "# >>> centinela >>>"
    EndMarker   = "# <<< centinela <<<"
)

// Block is the managed hook body, fenced by the markers.
const Block = BeginMarker + "\n" +
    "#!/bin/sh\n" +
    "# Managed by centinela — do not edit between the markers.\n" +
    "centinela precommit\n" +
    EndMarker + "\n"

// Install writes/refreshes the centinela block in the pre-commit hook at
// hooksDir/pre-commit, preserving any content outside the markers, creating the
// file (and hooksDir) when absent, and setting the executable bit. Idempotent:
// re-install replaces only the fenced block.
func Install(hooksDir string) (changed bool, err error)

// Uninstall removes the centinela block (and its surrounding blank lines),
// leaving any user content intact; removes the file only if it becomes empty.
func Uninstall(hooksDir string) (changed bool, err error)

// splice replaces the text between BeginMarker and EndMarker (inclusive) with
// newBlock; appends newBlock when no markers are present. Pure string fn.
func splice(existing, newBlock string) string
```

Rules (each an AC / edge case):
- **No pre-existing hook** → write `#!/bin/sh\n` shebang (once) + the block,
  `chmod 0755`. Missing `.git/hooks` dir → `os.MkdirAll(hooksDir, 0o755)` first.
- **Pre-existing hook present** → if it lacks a shebang, the user's content is
  preserved verbatim and the block is **appended**; if markers already exist,
  only the fenced region is replaced (idempotent — re-install of an unchanged
  block returns `changed=false`).
- **Uninstall** removes exactly the fenced region (and trims orphaned blank
  lines), never touching surrounding user lines; if the file is left with only a
  bare shebang / whitespace it is deleted.
- **Executable bit** always re-asserted on install (`os.Chmod(path, 0o755)`).
- **Portability** — `#!/bin/sh` (POSIX), single `centinela precommit` line; works
  under Git-for-Windows' bundled `sh`. Noted in docs.
- `cmd/` resolves `hooksDir` = `.git/hooks` of the repo root (via `git
  rev-parse --git-path hooks`, falling back to `.git/hooks`); `internal/githooks`
  takes the dir as a param so it is filesystem-pure and `t.TempDir()`-testable.

## PR delivery: render-vs-post (CALL-OUT)

**Verdict: render in Go, post in CI (Decision #4).** `centinela pr-gate` writes a
deterministic Markdown verdict to stdout and exits 0/1. The CI step pipes that to
`gh pr comment`. Rationale:
- **Testable** — the renderer is a pure `[]gates.Result → string` fn; no network,
  no `gh`, no token; covered by golden-string unit tests (AC-5 determinism).
- **Portable / no new deps** — zero GitHub client code enters the Go module; the
  binary runs identically on a laptop (prints to stdout) and in CI (piped).
- **Single responsibility** — Go decides *what* the verdict is + the exit code;
  CI decides *where* it goes and handles auth + comment de-duplication.
- `--post` (shell out to `gh pr comment` from Go) is **deferred** — it would pull
  `gh`/token assumptions into the binary for no v1 benefit; the CI yaml already
  does this cleanly.

**Markdown format** (rendered by `ui.RenderGatesMarkdown(results)`), bounded &
collapsible for huge diffs (AC: comment stays readable):
```
<!-- centinela:pr-gate -->        ← stable marker line, first line, for CI find/update
### Centinela PR Gate — ❌ 1 failed, 2 passed

| Gate | Status | Message |
| --- | --- | --- |
| G1: File Size | ❌ fail | 1 file over 100 lines |
| import_graph | ✅ pass | no forbidden edges |
| G11: i18n | ✅ pass | all locales present |

<details><summary>Failing details (G1: File Size)</summary>

- internal/x.go (142 lines)
</details>
```
- The first line is an HTML-comment **marker** (`<!-- centinela:pr-gate -->`) the
  CI poster greps to find-and-update its single comment (idempotency, AC-6).
- Header summarizes counts; one table row per gate; `<details>` per failing gate,
  with Details capped (e.g. first 50 lines + "… N more") so a huge diff can't
  blow the comment size limit (edge case).
- Status icons fixed (`✅`/`❌`/`⚠️`/`➖`) and gates emitted in `results` order →
  byte-deterministic for fixed input (AC-5). No `time.Now`, no Lipgloss.

## Import-graph / layering (CALL-OUT)

**Verdict: one new leaf package + one renderer file; no new failing edge, no
cycle, no `centinela.toml` matrix change beyond adding the leaf path.**

- `internal/githooks` → **leaf** (stdlib-only file I/O, imported only by `cmd/`).
  Add `internal/githooks/**` to the existing leaf layer `paths`
  (line 61, next to `internal/config/**`, `internal/gitdiff/**`). A leaf
  `allow = []` is fine — githooks imports no project package.
- `internal/ui/render_markdown.go` → stays in `internal/ui` (already imports
  `internal/gates`; the new file adds **no** new import edge — `ui → gates`
  already exists via `render_gates.go`). `ui` is unmapped today (rendering
  helper); adding a file changes nothing in the matrix.
- `internal/gitdiff` gains a method, no new import.
- `internal/config` gains two config types, no new import.
- `cmd/centinela` (the `cmd` layer, `allow = [domain, leaf, aggregator]`) imports
  `githooks` (leaf), `gitdiff` (leaf), `gates` (domain), `ui`, `config` (leaf) —
  all allowed. No edge into `cmd` from any internal package.
- **`centinela.toml` change:** add `"internal/githooks/**"` to the leaf layer
  `paths` (one entry). **Mirror to `internal/scaffold/assets/centinela.toml`**
  (known trap — the scaffold parity test covers only the 8 arch docs, NOT the
  toml, so a drift is silent). Verify both files during the code step. No other
  matrix edits — `pr-gate`/`precommit` logic produces no aggregator package.

## v1 scope

**In:**
- `internal/gitdiff.ChangedFilesStaged()` (staged index → `*Set`, graceful
  degrade on initial-commit / not-a-repo).
- `internal/githooks` leaf package: `Install` / `Uninstall` / `splice`,
  marker-delimited, idempotent, no-clobber, exec-bit, MkdirAll.
- `internal/ui/render_markdown.go`: `RenderGatesMarkdown([]gates.Result) string`
  — deterministic, bounded/collapsible, marker-prefixed.
- `centinela precommit` (staged run, skip-build, exit 0/1, stderr render) +
  `precommit install` / `precommit uninstall`.
- `centinela pr-gate` (changed-since-base run, Markdown to stdout, exit 0/1,
  `fail_on_warning`).
- `[precommit]` (`enabled`, `skip_build`) + `[pr_gate]` (`enabled`,
  `fail_on_warning`) config — Normalize + validate, wired into defaults +
  validateConfig.
- `.github/workflows/validate.yml`: a `pull_request`-only job posting the
  `pr-gate` Markdown as an updating marker comment via `gh pr comment`.

**Out (deferred):**
- Direct GitHub API / GraphQL from Go — none; CI owns posting.
- `centinela pr-gate --post` (shelling to `gh` from the binary).
- `centinela doctor` reporting/repairing the git-hook install status.
- Multi-host beyond GitHub (GitLab/Bitbucket comment posting).
- A pre-push or commit-msg hook; only `pre-commit` is wired.
- Auto-installing the git hook during `centinela start` / setup — install is
  always explicit (`centinela precommit install`).

## Step 2 — code

New / edited source files (each ≤100 lines):

| File | Change | Budget |
|------|--------|--------|
| `internal/gitdiff/resolver.go` | add `ChangedFilesStaged()` (`git diff --cached --name-only --diff-filter=ACMR`, degrade-safe, `Base:"STAGED"`); reuse `splitNonEmpty`/`degradeReason` | +~25 (split a helper out if >100) |
| `internal/githooks/install.go` | NEW. `BeginMarker`/`EndMarker`/`Block` consts; `Install(hooksDir) (bool, error)`; `Uninstall(hooksDir) (bool, error)`; exec-bit + MkdirAll | ~85 |
| `internal/githooks/splice.go` | NEW. pure `splice(existing, newBlock) string` + marker find/replace/append helpers | ~55 |
| `internal/ui/render_markdown.go` | NEW. `RenderGatesMarkdown([]gates.Result) string` — marker line, header counts, table, `<details>` per fail, Details cap | ~95 |
| `internal/config/precommit.go` | NEW. `PrecommitConfig{Enabled bool; SkipBuild bool}`; `NormalizePrecommit` (SkipBuild defaults true); `validatePrecommit` (no-op, reserved) — mirrors `roadmap_drift.go` | ~40 |
| `internal/config/pr_gate.go` | NEW. `PrGateConfig{Enabled bool; FailOnWarning bool}`; `NormalizePrGate`; `validatePrGate` | ~35 |
| `internal/config/config.go` | add `Precommit PrecommitConfig \`toml:"precommit"\`` + `PrGate PrGateConfig \`toml:"pr_gate"\`` to `Config` | +2 lines |
| `internal/config/defaults.go` | `cfg.Precommit = NormalizePrecommit(cfg.Precommit)` + `cfg.PrGate = NormalizePrGate(cfg.PrGate)` in `applyDefaults` | +2 lines |
| `internal/config/file_size_exceptions.go` | call `validatePrecommit` + `validatePrGate` in `validateConfig` | +6 lines |
| `cmd/centinela/precommit.go` | NEW. `precommitCmd` group + default `RunE`: `ChangedFilesStaged` → degrade-or-run `RunWithFilter(precommitCfg(cfg), set)` + `appendAuditGate` → `RenderGatesMarkdown` to stderr → exit 1 on `!AllPassed` | ~85 |
| `cmd/centinela/precommit_skipbuild.go` | NEW. `precommitCfg(cfg) *config.Config` — shallow copy with `Gates.Build.Enabled=false` when `cfg.Precommit.SkipBuild` | ~30 |
| `cmd/centinela/precommit_install.go` | NEW. `install`/`uninstall` subcommands → resolve hooks dir (`git rev-parse --git-path hooks`) → `githooks.Install`/`Uninstall` → confirmation line | ~80 |
| `cmd/centinela/pr_gate.go` | NEW. `prGateCmd` default `RunE`: `ChangedFiles(diffBase, true)` → degrade-or-run `RunWithFilter(cfg, set)` + `appendAuditGate` → `RenderGatesMarkdown` to stdout → exit 1 on fail (or warn if `FailOnWarning`) | ~90 |
| `.github/workflows/validate.yml` | add a `pull_request`-only job/step (see CI snippet) | +~20 lines |
| `centinela.toml` | add `internal/githooks/**` to leaf `paths`; (optional) `[precommit]`/`[pr_gate]` example blocks | +~3 |
| `internal/scaffold/assets/centinela.toml` | **mirror** the leaf `paths` edit (known trap) | +1 |

**Reuse seam (load-bearing):** both new commands call the *identical*
`appendAuditGate(cfg, gates.RunWithFilter(cfgX, set))` pipeline as `validate.go`,
so custom gates + the audit-baseline gate participate with no new code path
(AC-7). The only difference is the **filter source** (staged vs. changed-since-
base) and, for precommit, the skip-build cfg copy.

### CI snippet intent (`.github/workflows/validate.yml`)

A second job (or step) gated on `if: github.event_name == 'pull_request'`, with
`permissions: pull-requests: write`:
```yaml
  pr-gate:
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    permissions:
      pull-requests: write
    steps:
      - uses: actions/checkout@v6
        with: { fetch-depth: 0 }          # need base ref for merge-base
      - uses: actions/setup-go@v6
        with: { go-version-file: go.mod }
      - name: Render gate verdict
        run: go run ./cmd/centinela pr-gate > verdict.md || true   # capture even on fail
      - name: Post/update PR comment
        env: { GH_TOKEN: ${{ github.token }} }
        run: |
          # find prior comment by the <!-- centinela:pr-gate --> marker; edit it or create one
          gh pr comment ${{ github.event.pull_request.number }} --body-file verdict.md --edit-last \
            || gh pr comment ${{ github.event.pull_request.number }} --body-file verdict.md
      - name: Enforce gate exit code
        run: go run ./cmd/centinela pr-gate    # re-run for the real exit code (fails the job)
```
Intent notes: `fetch-depth: 0` so `merge-base HEAD <base>` resolves; the marker
comment line makes `--edit-last`/find-by-marker idempotent (AC-6); the verdict is
rendered once for the comment and the exit code re-asserted to fail the check.
(Implementation may collapse the two `pr-gate` invocations by capturing `$?`.)

## Step 3 — tests

Colocated per-package `_test.go` (the 95% per-package coverage gate is NOT moved
by `tests/` tier files — add coverage next to the code). Each ≤100 lines (G1
applies to `_test.go` too):

- `internal/gitdiff/resolver_staged_test.go` — **staged-diff parsing.** Stub
  `Resolver.Run`: `git diff --cached --name-only --diff-filter=ACMR` output →
  expected `Set`; not-a-repo error → `Summary.Degrade=="not a git repository"`,
  nil set, **nil error** (never crash); empty stdout → empty set, `Files==0`,
  no degrade (AC edge: no staged changes); initial-commit (no HEAD) still lists
  staged files via the stub.
- `internal/githooks/install_test.go` — **installer idempotency / no-clobber.**
  In `t.TempDir()`: install into an empty dir creates an executable hook with the
  fenced block; install **twice** is a no-op (`changed=false`, byte-identical);
  install over a pre-existing user hook **preserves** the user lines and appends
  the block; uninstall removes only the fenced region and keeps user lines;
  uninstall of a centinela-only hook deletes the file; missing `.git/hooks` dir
  is created. Assert exec bit (`0o755`).
- `internal/githooks/splice_test.go` — pure `splice`: append when no markers;
  replace-in-place when markers present; preserve surrounding text; multiple
  install/uninstall round-trips converge (idempotent).
- `internal/ui/render_markdown_test.go` — **markdown render determinism.** Golden
  string for a mixed pass/fail/warn/skip result set: marker line first, header
  counts correct, one row per gate in input order, `<details>` only for fails,
  Details capped with "… N more" past the cap; **two renders of the same input
  are byte-identical** (AC-5); empty results → "all passed" header, no table rows.
- `internal/config/precommit_test.go` — `NormalizePrecommit` defaults
  `SkipBuild=true` (and leaves an explicit `false` intact); round-trips Enabled.
- `internal/config/pr_gate_test.go` — `NormalizePrGate` defaults
  `FailOnWarning=false`; `validatePrGate` no-op; disabled → no behavior change.
- `cmd/centinela/precommit_skipbuild_test.go` — **skip-build.** `precommitCfg`
  with `SkipBuild=true` returns a copy whose `Gates.Build.Enabled==false` while
  the original cfg is **unmutated**; with `SkipBuild=false` the build gate stays
  enabled.
- `cmd/centinela/precommit_test.go` — with a stubbed `gitDiffResolver`: degrade
  (`Summary.Degrade` set) → exit 0 + a notice, **never blocks** (AC edge: not a
  repo / staged cmd fails → no false block); a fail-severity result → non-zero
  exit; all-pass → exit 0.
- `cmd/centinela/pr_gate_test.go` — fail → exit 1; warn with `FailOnWarning=false`
  → exit 0; warn with `FailOnWarning=true` → exit 1; degrade → render-to-stdout +
  clear message, no crash (AC edge: run outside a PR / missing base).

**Integration:** `tests/integration/precommit_test.go` — in a `t.TempDir()` git
repo (init, configure user, write `centinela.toml` enabling the file-size gate):
`git add` an **oversized** staged file → `centinela precommit` (built `/tmp`
binary or package API) exits **non-zero** and names the file; with no staged
changes → exit 0; stage a clean file → exit 0. Plus a `pr_gate` integration: two
commits on a branch with one oversized changed file → `pr-gate` emits Markdown
containing the marker + the failing gate row and exits 1. Drive the installer
end-to-end: `precommit install` writes an executable `.git/hooks/pre-commit`;
re-install is a no-op; a real `git commit` of an oversized file is blocked by the
hook (the make-or-break UX assertion).

**Acceptance:** `tests/acceptance/precommit_*` + `tests/acceptance/pr_gate_*`
(executable, one per Gherkin scenario) — run the real built binary against a
fixture repo and assert exit codes + summary/marker lines for: staged-pass,
staged-fail-blocks, no-staged-changes-passes, install-idempotent, install-no-
clobber (pre-existing hook preserved), uninstall-removes-only-block, pr-gate-
markdown-deterministic, pr-gate-fail-exit, pr-gate-warn-non-blocking,
not-a-repo-degrade-no-block. Register the acceptance runner in `validate.commands`
in `centinela.toml`.

`.workflow/precommit-and-pr-gate-edge-cases.md` — map every brief edge case
(not-a-git-repo, no staged changes, staged-diff git failure → degrade, existing
hook preserved, uninstall removes only block, pr-gate outside a PR, warn-severity
non-blocking, re-run updates single comment, huge-diff bounded markdown, Windows
`sh` shebang portability) to the test covering it.

Note: `go test ./...` runs ~75s; `[verify] verify_timeout = 240` gives margin.

## Step 4 — validate

Gatekeeper report `.workflow/precommit-and-pr-gate-gatekeeper.md`; `centinela
validate` green (lint + types + full suite). **Confirm the G2 import-graph gate
output: the new `internal/githooks/**` leaf mapping must produce zero new
*failing* edges** (githooks imports stdlib only; `cmd → githooks` is leaf,
allowed); confirm `internal/ui/render_markdown.go` adds no new edge (`ui → gates`
already exists). Confirm `internal/gitdiff` / `internal/config` gain no new
import. Confirm every touched source file ≤100 lines (including `_test.go`; split
the staged resolver into a helper file if `resolver.go` would exceed 100).
**Mirror the `centinela.toml` leaf-path edit into
`internal/scaffold/assets/centinela.toml` and verify both by hand** (parity test
doesn't cover the toml). Dogfood `centinela precommit` / `precommit install` /
`pr-gate` from a `/tmp` binary built from `./cmd/centinela` before relying on the
installed binary (known trap: installed binary lags the worktree). Production-
readiness subagent if the gate is enabled.

## Step 5 — docs

Documentation-specialist `.md` + `.json`; regenerate
`docs/project-docs/index.html`; changelog artifact
`.workflow/precommit-and-pr-gate-changelog.md` (create early via `evidence
artifact new` so completion doesn't fail). Document: `centinela precommit` and
its `install`/`uninstall` subcommands (exit codes, what gets gated = staged
files, skip-build default); the `.git/hooks/pre-commit` marker-block contract +
idempotency + no-clobber guarantee + `#!/bin/sh` portability note; `centinela
pr-gate` (Markdown output, exit codes, `fail_on_warning`); the Markdown verdict
format + the `<!-- centinela:pr-gate -->` marker the CI poster uses to update one
comment; the `[precommit]`/`[pr_gate]` config knobs + safe defaults
(`skip_build=true`, `fail_on_warning=false`); the CI yaml integration (PR-only
step, `permissions: pull-requests: write`, `fetch-depth: 0` for `merge-base`).
Add a PROJECT.md G2 one-line note that `internal/githooks` is a stdlib-only leaf
imported only by `cmd/`, and confirm the `centinela.toml`/scaffold mirror.
