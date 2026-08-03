# Feature: docstring-gate

- surface: internal
- status: planned
- roadmap: Phase 13 — Lighter Centinela
- depends on: docs-step-markdown-first (merged)
- fixes: exported API ships with no doc comments; nothing mechanically requires them

## Problem

`docs-step-markdown-first` deleted the HTML pipeline and re-pointed the docs
step at markdown humans actually read (`docs/guides/`, README.md). It
deliberately left one hole open, named in its own out-of-scope list:

> Docstring enforcement (`centinela docs lint`, changed-files ratchet) —
> separate follow-up feature `docstring-gate`.

Nothing in Centinela requires a doc comment on an exported identifier. The
docs step gates on *prose about the feature*; the code step gates on
architecture and file size. An exported `func`, `type`, `const`, or interface
method can ship with zero explanation and every gate stays green. That is the
raw material a godoc or Docusaurus pipeline would consume, and today it is
mostly absent.

## Measured baseline (this repo, at plan time)

A throwaway `go/ast` walk over the whole worktree (excluding `.git`,
`vendor`, `node_modules`, `web`, `testdata`, `.worktrees`):

| Scope | Exported | Undocumented |
|-------|----------|--------------|
| All `.go` files, all kinds incl. struct fields | 5294 | **2410** |
| Non-test files only, all kinds incl. struct fields | 1534 | **812** |
| Non-test files, excluding struct fields (**the v1 check scope**) | 885 | **215** |

Breakdown of the 215 (non-test, no fields): const 89, method 44, func 38,
type 17, var 15, interface method 12. Separately: 572 of 599 non-test Go
files carry no package-clause doc comment, and 35 documented consts have a
doc comment that does not begin with the identifier name.

That 215 is the **legacy backlog**, not the gate's failure count: the gate
never scans an unchanged file. It is recorded so `centinela docs lint --full`
has a number to measure adoption against.

## What it verifies

Every **exported identifier in a changed file** carries a non-empty doc
comment. Scope is the changed-file set — the exact merge-base `*gitdiff.Set`
the other file-walking gates already receive, with no second notion of
"changed".

**Documentation means what `go/doc` means.** A doc comment above the
declaration counts; so does a **trailing line comment**
(`const X = 1 // X is the answer.`), because `godoc` renders it and this gate
exists to feed that pipeline. A doc on a grouped `const (` / `var (` / `type (`
block covers every spec inside it.

**One exclusion set, both surfaces.** `.git`, `node_modules`, `vendor`, `dist`,
`testdata`, `.worktrees`, `build` and `target` are never source. The gate and
`centinela docs lint --full` share the single set, so the report always
predicts the gate. This matters at `severity = "fail"`: a vendored tree under a
configured root would be a legacy backlog the ratchet was never asked to open,
and a deliberately-invalid `testdata/` parser fixture would be a hard failure
with no possible opt-out — an invalid file cannot carry `//centinela:nodoc` and
stay invalid.

**CI must resolve a merge base — and the failure is ref resolution, not
depth.** A ratcheted gate that cannot find a merge base Skips honestly and
enforces nothing, so "enforcing" would be a false claim. `actions/checkout`
leaves a **detached HEAD with no local branch ref for the default branch**, so
`git merge-base HEAD main` exits 128 with `Not a valid object name` even on a
full-history clone — `fetch-depth: 0` alone does not fix it. The real fix lives
in `internal/gitdiff`: when the configured diff base does not resolve as a
local ref, the resolver retries `origin/<base>` and reports the ref that
actually resolved. That repairs **every** diff-aware gate in CI, not just this
one. Full history (`fetch-depth: 0`) is still required so the retry has a
common ancestor to find. This is pinned by an acceptance test that reproduces
the ref state — clone, detach, delete the local branch — and asserts the gate
*enforces*, never by grepping the workflow YAML.

## The central constraint: enforcing at 215 legacy violations

A gate that permanently reports a warning is a permanently-red validator —
noise everyone learns to scroll past, and exactly the anti-pattern Phase 13
exists to kill. The three dishonest escapes are all refused:

- **Not** weakening the definition until the count reaches zero.
- **Not** shipping disabled, or shipping in `warn` forever, and calling it
  delivered.
- **Not** back-filling 215 placeholder comments to game the count.

The honest answer is the ratchet the roadmap entry already names. The gate
ships **enabled at `severity = "fail"`, scoped to changed files in every
mode** — including a run that asked for a full scan, where the gate resolves
the same merge-base set itself rather than opening the legacy backlog. Touch
a file, document its exports; never touch a file, and it is never inspected.
`main` therefore cannot go red for a pre-existing gap, and there is no
weakening anywhere: any new or modified exported identifier fails the build
until it is documented.

The whole-repo number stays measurable through `centinela docs lint --full`,
a human-invoked report that never exits non-zero. Paying that backlog down is
tracked separately as `docstring-full-scan-debt-paydown`.

When the changed-file scope is empty, contains no Go files, or cannot be
resolved at all, the gate reports **Skip** — a check that opened no file may
not report a confident pass.

## Goal

1. A built-in `[gates.docstring]` gate wired into `gates.RunWithFilter`,
   reporting Pass / Warn / Fail / **Skip** through the existing `Result`
   contract, honest about an empty or unresolvable scope.
2. `centinela docs lint [--changed|--full] [--json]` as the standalone
   surface, a thin `cmd/` wrapper over the same check — never a second
   implementation.
3. The senior-engineer prompt gains a write-doc-comments duty so comments are
   authored during the `code` step, with scaffold-mirror byte parity.

## Non-goals (v1)

- **Multi-language.** Go only. The language seam is named and shaped
  (`internal/docstring` provider, mirroring `internal/importgraph`), one
  provider registered.
- **Struct field docs** (597 of the 812 non-test violations) — behind
  `check_fields`, default off.
- **Package-clause docs** (572 files) — excluded entirely from v1.
- **Name-prefix strictness** (`// Foo does…`) — behind `require_name_prefix`,
  default off.
- **Hunk-level scope.** A changed file is scanned whole.
- **Quality judgement of comment text.** Presence, not usefulness.
- **A new baseline mechanism.** `audit-baseline` already exists.
- **Paying down the 215-item legacy backlog** so a full scan could be clean —
  deferred as `docstring-full-scan-debt-paydown`.

## Constraints

- Reuse `internal/gitdiff` scope; do not invent a second diff notion and do
  not add a gate-local `scope` knob.
- Empty, Go-free or unresolvable scope must report `Skip` — the
  `truthful-validators` rule that a gate which inspected nothing may not
  report a confident pass.
- `ui.RenderGateResult` expands `Result.Details` for `Fail` only, so on the
  enforcing surface the gate **message** must carry the exemption list and the
  violation count itself. A Pass may never call an exempted identifier
  documented, and a Warn may never end in a colon introducing a list the
  renderer will not print.
- `senior-engineer-prompt.md` is 109 lines against a 130-line budget
  (`promotedPromptLineBudget`) and is **not** on `mirrorParityAllowlist`:
  the duty must fit in the headroom and land byte-identically in
  `internal/scaffold/assets/docs/architecture/`.
- Every file ≤100 lines including `_test.go`; colocated tests; ≥97%
  coverage on touched packages.
