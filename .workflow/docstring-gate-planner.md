### Planner Report: docstring-gate
**Date:** 2026-07-31

Full plan: `docs/plans/docstring-gate.md` · Brief: `docs/features/docstring-gate.md`
· Spec: `specs/docstring-gate.feature`

#### Problem

`docs-step-markdown-first` (merged) deleted the HTML pipeline — `internal/docgen`,
`centinela docs generate`, the `kb/*.md → .html` contract and the
`docs/project-docs/index.html` completion gate are gone; `centinela docs` is now a
bare cobra parent with one subcommand, `docs context <feature>`, and the docs step
gates on documentation-specialist `outputs` naming a real file under `docs/` or
`README.md`. Its own out-of-scope list names this feature and states that the
redesigned step produces "the markdown + (later) **docstrings** such a pipeline would
consume". Nothing today produces or requires those docstrings: an exported `func`,
`type`, `const`, or interface method can ship with zero explanation and every gate
stays green. Measured now, 215 exported identifiers in this repo's non-test Go files
have no doc comment at all, and the code-step agent has no instruction to write one.

#### Measured baseline (decides the shipping posture)

`go/ast` walk over the worktree, excluding `.git`, `vendor`, `node_modules`, `web`,
`testdata`, `.worktrees`:

| Scope | Exported | Undocumented |
|-------|----------|--------------|
| All `.go`, all kinds incl. struct fields | 5294 | **2410** |
| Non-test files, all kinds incl. struct fields | 1534 | **812** |
| Non-test files, excluding struct fields (**v1 gate scope**) | 885 | **215** |

The 215: const 89, method 44, func 38, type 17, var 15, interface method 12.
Separately: 572/599 non-test files have no package-clause doc; 35 documented consts
are not name-prefixed. **No strictness setting makes a full-scan `fail` gate green
today**, so it ships `enabled = true, severity = "warn"` — the same honest adoption
posture `[gates.spec_traceability]` and `[gates.roadmap_drift]` already carry in
`centinela.toml`, with the measured count written into the comment. Not weakened,
not disabled, not back-filled.

#### Scope

- **In:** `internal/docstring` (stdlib-only leaf AST scanner + named language seam);
  `DocstringConfig` + Normalize/validate in `internal/config`;
  `internal/gates/docstring.go` registered in `RunWithFilter`; `centinela docs lint`
  as a thin `cmd/` wrapper; `[gates.docstring]` adoption + leaf-layer entry in
  `centinela.toml`; the senior-engineer prompt duty + byte-identical scaffold mirror.
- **Out:** multi-language (seam named, Go-only built), struct fields
  (`check_fields=false`), package-clause docs, name-prefix strictness
  (`require_name_prefix=false`), hunk-level scope, comment-quality judgement, a
  second baseline mechanism, doc-site generation, ratcheting this repo to `fail`.

#### Dependencies & Assumptions

- `internal/gitdiff` — the gate consumes the `*gitdiff.Set` only, never shells to git.
- `internal/gates` — `Result{Name,Status,Message,Details}`, `Status ∈ {Pass,Fail,Warn,Skip}`.
- `internal/config` — the `Normalize*`/`validate*` pair pattern.
- `cmd/centinela/validate_mode.go` — `resolveDiffFilter`, `currentEnv`, `ResolveMode`.
- Assumption: an unparseable `.go` file is a reported finding, never a silent skip.
- Assumption: 215 must be re-measured with the real gate at code time; the acceptance
  test asserts the configured **posture** (`warn`), not the count.

#### Key decisions

1. **Ratchet scope = the existing one, unchanged.** `checkDocstring(cfg, filter)`
   takes the same `*gitdiff.Set` as `checkFileSize`/`checkSpecTraceability`
   (`git diff --diff-filter=ACMR <merge-base HEAD main>` ∪ untracked). Local runs are
   diff-aware, CI full-scans via `ResolveMode`, `precommit` gets staged scope free.
   **No `scope` knob** — a second notion of "changed" is the thing being prevented.
   Hunk-level scope rejected (needs line-range mapping, defeated by gofmt, a rename
   dodges it). File granularity: *touch a file, document its exports.*
2. **Documented = any non-empty doc comment.** Name-prefixing is a deprecated golint
   rule worth 35 extra findings and no information; it lives behind
   `require_name_prefix=false`. Grouped `const`/`var`/`type` blocks inherit the
   GenDecl doc; `Deprecated:` counts. `_test.go`, generated files
   (`// Code generated … DO NOT EDIT.`) and unexported identifiers are out;
   `internal/` is **in** (knob `include_internal=true`) or the gate is vacuous here.
   Opt-out is `//centinela:nodoc` — deliberately not `//nolint`, which is
   golangci-lint's namespace — and every exemption is listed in the passing report.
3. **Skip is first-class.** Two branches: empty filter, and a non-empty filter with
   zero in-scope Go files (a docs-only branch). Neither may report a confident pass —
   the `truthful-validators` rule.
4. **Go only for v1**, with the seam named and shaped after `internal/importgraph`'s
   provider registry.
5. **Prompt duty** in `senior-engineer-prompt.md` — 109 lines against a 130-line
   `promotedPromptLineBudget`, and the file is **not** on `mirrorParityAllowlist`, so
   ≤8 lines and a byte-identical mirror.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Enabling at `fail` turns main red (215 violations, CI full-scans) | High | High if unmeasured | Measured; ships `warn`; `fail` path acceptance-tested on a fixture |
| A second "changed" notion drifts from the other gates | High | Medium | Gate only calls `filter.Contains`; no git access, no scope knob |
| Unparseable `.go` file silently vanishes (false green) | High | Low | Parse error is its own reported finding |
| Empty/Go-free scope reports a confident Pass | Medium | Medium | Two explicit Skip branches, one scenario each |
| Prompt edit breaks the 130-line budget or mirror parity | Medium | Medium | 21 lines headroom; duty ≤8; both tests catch it |
| G1: AST walker grows past 100 lines | Medium | High | Pre-split scan/decl/filter/directive/report |
| Coverage dips below the 95% gate | Medium | Medium | Colocated `_test.go` per file (coverage is per-package) |
| `docs lint` duplicates gate logic | Medium | Low | Wrapper over the same entry point; test asserts agreement |

#### Rollout

Slice 1 `internal/docstring` leaf (no wiring) → Slice 2 config (`enabled=false`
default, no behavior change) → Slice 3 the gate + `RunWithFilter` registration →
Slice 4 `centinela docs lint` → Slice 5 adoption (`centinela.toml` warn + leaf layer)
and the senior-engineer prompt + mirror.

#### Behavior Summary

When enabled, `validate` / `precommit` / `pr-gate` / `docs lint` walk the Go AST of
every in-scope file and report each exported func, type, method, interface method,
const and var with no doc comment. Scope is whatever `*gitdiff.Set` the run already
resolved. `_test.go`, generated and out-of-root files are never inspected;
`//centinela:nodoc` identifiers are exempt and listed in the passing report. Zero
violations → `Pass` naming the inspected count; violations → `Warn` or `Fail` per
severity with one `path:line: kind Name` line each; nothing in scope → `Skip`.

#### Gherkin Scenarios

14 scenarios in `specs/docstring-gate.feature`: undocumented exported func fails at
`fail`; documented identifiers pass; `warn` reports without failing; empty diff scope
→ Skip; Go-free diff scope → Skip; unchanged legacy file not reported (the ratchet);
grouped const block doc covers its members; `_test.go` + generated never reported;
`//centinela:nodoc` exempt and reported; struct fields ignored at `check_fields=false`;
unparseable file reported not skipped; unknown severity is a config error;
`docs lint` exits 1 on Fail / 0 on Warn; the prompt duty present and byte-identical
in both copies within the 130-line budget.

#### UX States

| State | Trigger | Surface |
|-------|---------|---------|
| loading | n/a — synchronous sub-second AST walk | — |
| empty (Skip) | no changed files, or no Go files in scope | validate gate line: "no Go files in scope; nothing inspected" |
| success (Pass) | zero violations | "N exported identifiers documented" + exemption lines |
| warning (Warn) | violations at `severity="warn"` | violation list; validate still exits 0 |
| error (Fail) | violations at `severity="fail"` | same list; `AllPassed` false; exit 1 |
| config error | unknown severity while enabled | `gates.docstring.severity must be fail or warn, got "nope"` |

#### Out-of-Scope

Multi-language scanning; struct field docs; package-clause docs; name-prefix
strictness; hunk-level scope; judging comment quality; auto-writing comments; a
docstring-specific baseline (use `audit-baseline`); godoc/Docusaurus site generation;
ratcheting this repo from `warn` to `fail`.

#### Deferred Findings

Backlog checked first — no overlapping entries. Recorded:

- `package-doc-comments`
- `docstring-struct-field-docs`
- `docstring-gate-ratchet-to-fail`

#### Handoff

- **Next role:** senior-engineer.
- Re-measure the violation count with the real gate before writing the TOML comment;
  state the measured number, and if it is 0 then `fail` is the correct severity.
- Do not add a `scope` knob to `[gates.docstring]`.
- `senior-engineer-prompt.md`: ≤8 added lines, mirror byte-identical, same commit.
- `internal/docstring` must import nothing under `github.com/samuelnp/centinela` —
  its leaf classification and the entire layering argument depend on it.
