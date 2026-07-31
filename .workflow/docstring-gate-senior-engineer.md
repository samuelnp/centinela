### Senior-Engineer Report: docstring-gate
**Date:** 2026-07-31

#### Orchestrator override applied (plan/brief/spec amended)

The plan proposed `severity = "warn"` because a full scan finds a legacy
backlog. A permanently-warning gate is a permanently-red validator. Shipped
instead at **`severity = "fail"`, scoped to changed files in every mode**:

- `checkDocstring` consumes the run's `*gitdiff.Set` and calls only `Contains`.
- A **nil** filter (CI full-scan mode) does **not** become a whole-repo scan.
  The gate resolves the *same* merge-base set itself, through the same
  `gitdiff.Resolver` and the same `cfg.Validate.DiffBase`. One notion of
  "changed"; the gate simply never opts into the legacy backlog.
- Unresolvable merge base → **Skip** naming the reason. Never Pass, never Fail.
- `centinela docs lint --full` keeps the backlog measurable, report-only, exit 0.

`docs/features/docstring-gate.md`, `specs/docstring-gate.feature` and
`docs/plans/docstring-gate.md` were rewritten so the contract matches: the spec
now asserts fail-on-changed, skip-on-empty/Go-free/unresolvable,
legacy-untouched-never-scanned, ratchet-under-nil-filter, `--full` report-only,
and this repo's shipped `severity = "fail"`.

Deferral re-scoped: `docstring-gate-ratchet-to-fail` removed (moot),
`docstring-full-scan-debt-paydown` deferred in its place.

#### Files Touched
| Path | Reason |
|------|--------|
| internal/docstring/*.go (10 files) | New stdlib-only leaf: Go AST scanner + language seam |
| internal/config/docstring.go | DocstringConfig + NormalizeDocstring + validateDocstring |
| internal/config/{gates_config,defaults,file_size_exceptions}.go | Wire the block into GatesConfig / defaults / validateConfig |
| internal/gates/docstring{,_scope,_report}.go | checkDocstring, ratchet scope, Result mapping |
| internal/gates/gates.go | Register the gate; document the Status consts |
| cmd/centinela/docs_lint{,_full,_json}.go | `centinela docs lint [--changed\|--full] [--json]` |
| centinela.toml | `[gates.docstring] enabled=true severity="fail"` + leaf layer |
| PROJECT.md | `internal/docstring` added to the G2 leaf list |
| docs/architecture/senior-engineer-prompt.md + scaffold mirror | Doc-comment duty, byte-identical |

#### Architecture Compliance
- Boundary checks passed: `internal/docstring` imports **only** the standard
  library (`go/ast`, `go/parser`, `go/token`, `os`, `io/fs`, `path/filepath`,
  `regexp`, `sort`, `strings`, `errors`, `fmt`) — no `github.com/samuelnp/...`
  import at all, so its leaf classification and the whole layering argument
  hold. `gates → docstring` is domain→leaf (allowed); `cmd → gates|docstring`
  is allowed. No cycle is possible.
- G1 file size: every modified and new file ≤ 100 lines, `_test.go` included.
- G7 outer-layer rule: `docs_lint*.go` are thin wrappers; the check itself lives
  in `internal/gates`/`internal/docstring`. `docs lint` calls the same
  `CheckDocstring` entry point `centinela validate` uses — one implementation.
- Docstrings: every exported identifier added or modified is documented; the
  gate was run against this branch's own diff and is green.

#### Type-Safety Notes
- No `any`, no `interface{}`, no reflection. The AST walk switches on concrete
  `*ast.FuncDecl` / `*ast.GenDecl` / `*ast.ValueSpec` / `*ast.TypeSpec` nodes.
- `Scanner` is the only interface, a deliberate one-method seam.
- `include_internal` is a `*bool` so "unset" is distinguishable from "false";
  `IncludesInternal()` resolves the tri-state in one place.
- Severity is validated at config load, so an unknown value cannot reach the
  gate and silently relax it.

#### Trade-Offs
- **Ratchet on nil filter vs. honoring full-scan intent.** Honoring it would
  turn `main` red for 171 legacy gaps at `fail`, and `warn` forever is the
  anti-pattern being removed. Chose the ratchet, documented in code and TOML,
  with `docs lint --full` preserving the measurement. Cost: `audit baseline`
  has nothing legacy to adopt for this gate — acceptable, the ratchet subsumes it.
- **File granularity, not hunk granularity.** A rename or `gofmt` reflow would
  defeat line-range mapping. "Touch a file, document its exports" is honest.
- **No `scope` knob.** A gate-local scope would be a second notion of changed.
- **Exemptions echoed by the CLI.** `ui.RenderGateResult` expands `Details` only
  for `Fail`, so a passing run hid the `//centinela:nodoc` list. Added
  `printDocsLintPassDetails` rather than changing global gate rendering.
- **Struct fields / package docs / name-prefix** stay behind default-off knobs.

#### Deferred Findings
- `docstring-full-scan-debt-paydown` — document the 171 legacy exported
  identifiers in files the ratcheted gate never opens, so
  `centinela docs lint --full` could one day report zero. Recorded via
  `centinela roadmap defer --source docstring-gate/senior-engineer`;
  `docstring-gate-ratchet-to-fail` removed as moot.
- No other new gaps found. `package-doc-comments` and
  `docstring-struct-field-docs` already exist in the Backlog and were left as is.

#### Handoff
- Next role: qa-senior
- Outstanding TODOs: acceptance tests for all spec scenarios (the tests step —
  deliberately not written here), including the two new ratchet scenarios and
  the `--full` report-only scenario.
