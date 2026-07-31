# Plan: docstring-gate

**Feature brief:** [docs/features/docstring-gate.md](../features/docstring-gate.md)
**Spec:** [specs/docstring-gate.feature](../../specs/docstring-gate.feature)
**Phase:** 13 — Lighter Centinela · **Archetype:** canonical (5 steps)
**Depends on:** `docs-step-markdown-first` (merged)
**Plan roles:** planner (planner-v1 — one context, both lenses)

---

## Lens 1 — Strategy

### 1. Problem framing

`docs-step-markdown-first` shipped and deleted the HTML pipeline: `internal/docgen`,
`centinela docs generate`, the `kb/<feature>.md → .html` contract and the
`docs/project-docs/index.html` completion gate are gone. `centinela docs` is now a
bare cobra parent (`cmd/centinela/docs.go`, 26 lines) with exactly one subcommand,
`docs context <feature>` — and a `RunE` that deliberately errors on an unknown
subcommand rather than printing help with exit 0. The docs step now gates on the
documentation-specialist evidence `outputs` naming a real updated file under `docs/`
or `README.md`. That feature's own out-of-scope list names this one:

> Docstring enforcement (`centinela docs lint`, changed-files ratchet) — separate
> follow-up feature `docstring-gate`.
> Docusaurus/godoc site generation — the redesigned step produces the markdown +
> (later) **docstrings** such a pipeline would consume.

So the hole is precise and pre-agreed: the docs step now produces good *prose*, but
nothing produces or requires the *in-code* documentation a godoc/Docusaurus pipeline
reads. Who hurts: (a) anyone reading this codebase — 215 exported identifiers in
non-test files have no doc comment at all; (b) the future doc pipeline, which would
render a mostly-empty API reference; (c) the code-step agent, which has no instruction
to write doc comments and therefore doesn't.

Why now: the docs step just stopped pretending HTML was documentation. The
replacement contract is "markdown humans read + docstrings machines render". Half of
it shipped; this is the other half, and the dependency edge is already recorded in
`.workflow/roadmap.json`.

### 2. Scope boundaries

**In (v1):**

| # | Deliverable |
|---|-------------|
| 1 | `internal/docstring` — a stdlib-only **leaf** package: Go AST scanner producing `[]Violation` and `[]Exemption` for a set of files |
| 2 | `internal/config` — `DocstringConfig`, `NormalizeDocstring`, `validateDocstring`, `GatesConfig.Docstring` |
| 3 | `internal/gates/docstring.go` — `checkDocstring(cfg, filter)` returning a `Result`, registered in `RunWithFilter` |
| 4 | `cmd/centinela/docs_lint.go` — `centinela docs lint [--changed\|--full] [--json]` |
| 5 | Adoption: `[gates.docstring]` in `centinela.toml` (`enabled = true, severity = "fail"`), `internal/docstring/**` added to the import-graph **leaf** layer |
| 6 | `docs/architecture/senior-engineer-prompt.md` + byte-identical scaffold mirror — the write-doc-comments duty |

**Out (v1):** every item in the brief's Non-goals — multi-language, struct fields,
package-clause docs, name-prefix strictness, hunk-level scope, comment-quality
judgement, a second baseline mechanism, and generating a doc site.

### 3. Dependencies & assumptions

- **`internal/gitdiff`** (leaf) — `Resolver.ChangedFiles(base, includeUntracked)` and
  `ChangedFilesStaged()`. The gate consumes the resulting `*gitdiff.Set` only; it
  never shells to git itself.
- **`internal/gates`** (domain, `allow = ["leaf"]`) — `Result{Name,Status,Message,Details}`,
  `Status` in `{Pass, Fail, Warn, Skip}`, and `RunWithFilter(cfg, filter)`.
- **`internal/config`** (leaf) — `GatesConfig` + the Normalize/validate pattern.
- **`cmd/centinela`** — `resolveDiffFilter(cfg, mode)` and `currentEnv()` in
  `validate_mode.go`; `config.ValidateConfig.ResolveMode(env, flag)`.
- **Assumption:** the module is Go and `go/parser` can parse every in-scope file. A
  file that fails to parse is a **reported violation of its own kind**, never a
  silent skip (see Risk R3).
- **Assumption (measured, must be re-checked at code time):** 215 violations in the
  v1 scope. The plan's shipping posture depends on this number; the code step
  re-measures with the real gate and the acceptance test asserts the posture, not
  the count.

### 4. Design decisions (the load-bearing ones)

#### 4.1 Ratchet semantics — "changed" means exactly what the other gates mean

There is **one** notion of scope in this codebase and this gate reuses it unchanged:

```
gitdiff.Resolver.ChangedFiles(cfg.Validate.DiffBase, true)
  = git diff --name-only --diff-filter=ACMR $(git merge-base HEAD <base>)
  U git ls-files --others --exclude-standard      (untracked)
```

The gate's signature is `checkDocstring(cfg *config.Config, filter *gitdiff.Set) Result`
— identical in shape to `checkFileSize`, `checkI18nFiltered`, `checkImportGraph`,
`checkSpecTraceability`. It calls `filter.Contains(path)` and nothing else.
Consequences, all free:

- `centinela validate` local → `ModeChanged` → merge-base scope.
- `centinela validate` in CI (`CI=true`) → `ResolveMode` → `ModeFull` → `filter == nil`.
  **This gate does not full-scan.** A nil filter means "the run did not resolve a
  scope", and the docstring gate answers it by resolving the *same* merge-base set
  itself, through the same `gitdiff.Resolver` and the same `cfg.Validate.DiffBase`.
  It is one notion of changed, not two — the gate simply never opts into the legacy
  backlog, because a full scan of a brownfield codebase is a *report*, not a gate
  (§4.3). If git cannot resolve the merge base (shallow clone, detached tree), the
  gate reports **Skip** with the reason — never a confident Pass, never a red `main`.
- `centinela validate --changed` / `--full` override, unchanged.
- `centinela precommit` already builds its filter from `ChangedFilesStaged()`, so the
  gate gets **staged-file** scope with zero extra code.
- `centinela pr-gate` already uses `ChangedFiles(DiffBase, true)`.
- `internal/audit` calls `RunWithFilter(cfg, nil)` — full scan — so the gate is
  baseline-adoptable by construction.

**Rejected: hunk-level scope** ("only identifiers whose lines changed"). It needs
diff line-range → AST position mapping, is defeated by a `gofmt` reflow, and lets a
pure rename dodge the gate. File granularity is the honest ratchet: *touch a file,
document its exports.*

**Rejected: a `scope` config knob.** A second, gate-local notion of scope is exactly
the thing this decision exists to prevent. Scope is `[validate] diff_mode` /
`diff_base` / `--changed` / `--full`, repo-wide, for every gate.

#### 4.2 What counts as documented

| Question | v1 decision | Rationale |
|---|---|---|
| Strictness | **Any non-empty doc comment.** `strings.TrimSpace(node.Doc.Text()) != ""` | Name-prefixing is a deprecated `golint` style rule; it adds 35 violations here for no information. Knob `require_name_prefix` (default `false`) names the seam without shipping it hot. |
| `_test.go` | **Excluded** | Test functions are not API surface. (Note: G1 file-size *does* apply to `_test.go`; docs do not. Stated so the difference is deliberate, not an oversight.) |
| Generated files | **Excluded** | First comment line matching Go's canonical `^// Code generated .* DO NOT EDIT\.$` before the package clause. |
| `internal/` | **Included** (knob `include_internal`, default `true`) | This repo is ~all `internal/`; excluding it makes the gate vacuous here, and the future godoc/Docusaurus pipeline renders the whole module. Consumer projects that only care about their importable API flip the knob. |
| Unexported identifiers | Excluded | `ast.IsExported` only. |
| Methods on **unexported** types | Excluded | Unreachable from outside the package. |
| Interface methods | **Included** (12 undocumented) | They are the contract. |
| Struct fields | **Excluded** (knob `check_fields`, default `false`) | 597 of the 812 non-test violations — 73%. Go convention documents fields via the type doc as often as inline. Including them would make the v1 backlog 4x larger for the weakest convention. |
| Package clause | **Excluded from v1** | 572/599 non-test files lack one, but Go's convention is **one package doc per package** (canonically `doc.go`), not per file — a per-file rule would be wrong, and a per-package rule is a different check. Deferred as `package-doc-comments`. |
| Grouped `const (…)` / `var (…)` / `type (…)` | A doc on the **GenDecl** counts for every spec inside it | Matches how godoc renders block-documented constants. Only a spec with neither its own doc nor a block doc is a violation. |
| `Deprecated:` markers | **Count as documented** | They are doc comments; a deprecation note is documentation. |
| Opt-out | **`//centinela:nodoc`** — Go directive form (no space after `//`), on the declaration or inside its doc comment | Deliberately **not** `//nolint` — that is golangci-lint's namespace and reusing it silently couples our semantics to theirs. |
| Opt-out visibility | Every exemption is listed in the `Result.Details` of a **passing** run | Exactly `checkFileSize`'s `justified` list: an exemption nobody can see is a hole. |

#### 4.3 Severity, config shape, and how it ships here

```toml
[gates.docstring]
enabled  = true
severity = "fail"
# roots               = ["internal", "cmd", "pkg", "lib", "src", "app"]
# include_internal    = true
# check_fields        = false
# require_name_prefix = false
```

- **Default when the block is absent: `enabled = false`.** Every optional built-in
  (`spec_traceability`, `roadmap_drift`, `security`, `import_graph`, `audit_baseline`)
  defaults off; a new gate must never turn itself on in someone's existing project.
- **Default severity: `fail`** (`NormalizeDocstring`), matching
  `validateSpecTraceability`'s default. `validateDocstring` rejects any severity
  other than `fail`/`warn` **and only when enabled** — the exact shape of
  `validateSpecTraceability`, so a typo cannot silently relax the gate. `warn`
  remains available as a config value for adopting projects; it is not what this
  repo ships.
- **In this repo it ships `enabled = true, severity = "fail"`.** A permanently
  warning gate is a permanently-red validator, which is the anti-pattern Phase 13
  exists to remove, and lowering a gate to make it green is forbidden by this
  project's standing rule. The ratchet is what makes `fail` safe: the 215 legacy
  violations live in files this gate never opens. The measured 215 is written into
  the TOML comment as the *backlog* number, with a pointer to
  `docstring-full-scan-debt-paydown`.
- **No new baseline is built.** The gate emits `Details`, and `[gates.audit_baseline]`
  with empty `target_gates` participates all detail-emitting gates — although with a
  ratcheted scope there is nothing legacy left for a baseline to adopt.
- The `warn` path is fully implemented and unit-tested; it is a *config value for
  other projects*, not this repo's posture.

#### 4.4 Reporting — Skip is a first-class outcome

Following `truthful-validators`, a gate that inspected nothing may not report a
confident pass. Two distinct Skips:

1. `filter != nil && filter.Len() == 0` → `Skip`, "no changed files — nothing inspected".
2. Filter non-empty but **zero in-scope Go files** after root/test/generated filtering
   (a docs-only or TOML-only branch) → `Skip`, "no Go files in scope". This one is
   the subtle case; without it a docs-only branch reports a green check for a gate
   that opened no file.

Otherwise: no violations → `Pass` (message carries the identifier count inspected and
the `Details` exemption list); violations → `Warn` or `Fail` per severity, one
`Details` line per violation as `path:line: <kind> <Name> has no doc comment`.

#### 4.5 Language seam (named, not built)

`internal/docstring` exposes:

```go
type Scanner interface { Scan(files []string, opts Options) (Report, error) }
func Register(lang string, s Scanner)   // v1 registers "go" only
func For(lang string) (Scanner, bool)
```

This mirrors `internal/importgraph`'s provider seam (`go` | `node` | `python` |
`script`), which is the precedent in this repo for language-pluggable gates. v1
registers the Go scanner and the gate resolves `"go"` unconditionally; manifest-based
auto-selection is a later feature. **No second language is implemented.**

#### 4.6 Layering

`internal/docstring` imports only `go/ast`, `go/parser`, `go/token`, `path/filepath`,
`strings`, `regexp`, `sort` — stdlib only. It is therefore a **leaf**, added to
`[[gates.import_graph.layers]] name = "leaf"` `paths` in `centinela.toml` with the
standard justification comment (the `internal/golist` / `internal/acceptance`
precedent). `internal/gates` is `domain` with `allow = ["leaf"]`, so
`gates → docstring` is legal; `docstring` imports nothing internal, so no cycle.
`cmd/**` allows `domain|leaf|aggregator`, so `docs lint` may call the gate directly.

### 5. Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| R1 — Enabling at `fail` turns `main` red on merge (215 legacy violations) | **High** | High if the gate full-scanned | The gate never full-scans: nil filter resolves the same merge-base set, so a legacy file is never opened. Acceptance tests assert (a) this repo's configured severity is `fail`, (b) an unchanged undocumented file is not reported, (c) an unresolvable scope Skips |
| R2 — A second notion of "changed" drifts from the other gates | High | Medium | Gate takes `*gitdiff.Set` and calls only `Contains`; no git access, no `scope` knob; acceptance test drives it through `RunWithFilter` |
| R3 — An unparseable `.go` file silently disappears from the scan (false green) | High | Low | Parse error is reported as its own `Details` line and contributes to Warn/Fail — never a silent skip. This is the failure mode `truthful-validators` exists to prevent |
| R4 — Empty/Go-free diff scope reports a confident Pass | Medium | Medium | Two explicit `Skip` branches (§4.4), each with its own scenario |
| R5 — `senior-engineer-prompt.md` edit breaks the 130-line budget or mirror parity | Medium | Medium | 109 lines today → **21 lines headroom**; duty is ≤8 lines; mirror copied byte-for-byte (the file is *not* on `mirrorParityAllowlist`) — `TestScaffoldArchitectureMirrorParity` + `promotedPromptLineBudget` both catch it |
| R6 — Scanning the whole repo in CI is slow | Low | Low | `go/parser` with `ParseComments` over 1879 files ran in well under a second during plan-time measurement; no type-checking, no `go/packages` |
| R7 — G1: an AST walker naturally grows past 100 lines | Medium | High | Pre-split by construction: `scan.go` (walk), `decl.go` (GenDecl/FuncDecl extraction), `filter.go` (roots/test/generated), `directive.go` (nodoc), `report.go` (types) |
| R8 — Coverage dips below the 95% gate (aim ≥97%) on a new package | Medium | Medium | Colocated `_test.go` per file (coverage is per-package — the `tests/` tier does not move the gate) |
| R9 — `centinela docs lint` duplicates the gate logic | Medium | Low | `docs_lint.go` is a wrapper that builds the filter via `resolveDiffFilter` and calls the same `gates` entry point; acceptance test asserts CLI and gate agree on the same fixture |

### 6. Rollout sequence

**Slice 1 — `internal/docstring` (leaf, no wiring).** `Options`, `Violation`,
`Exemption`, `Report`, the Go scanner, generated-file + `_test.go` + root filtering,
`//centinela:nodoc`, grouped-GenDecl doc inheritance, parse-error reporting. Colocated
tests. Nothing user-visible; nothing can regress.

**Slice 2 — config.** `DocstringConfig` + `NormalizeDocstring` + `validateDocstring`
+ `GatesConfig.Docstring`. Gate still not registered. Default `enabled = false` keeps
every existing project byte-identical in behavior.

**Slice 3 — the gate.** `internal/gates/docstring.go`: `checkDocstring(cfg, filter)`,
both `Skip` branches, `Pass`/`Warn`/`Fail`, registration in `RunWithFilter` behind
`cfg.Gates.Docstring.Enabled`.

**Slice 4 — the CLI.** `centinela docs lint` with `--changed`/`--full`/`--json`,
exit 1 only on `Fail`. (Note: `docsCmd.RunE` errors on unknown subcommands, so the
new verb must be registered via `docsCmd.AddCommand` in `init()` like `docs context`.)

**Slice 5 — adoption + prompt.** `centinela.toml`: `[gates.docstring] enabled = true,
severity = "fail"` with the measured-backlog comment, plus `internal/docstring/**` on
the leaf layer and `internal/docstring` added to PROJECT.md's G2 leaf list.
`senior-engineer-prompt.md` gains the duty; the scaffold mirror is updated
byte-identically.

**Deliberately later, not now:** paying down the 215-item legacy backlog so a full
scan could be clean (`docstring-full-scan-debt-paydown`); struct fields; package
docs; a second language; auto-selecting the language by manifest.

---

## Lens 2 — Spec

### 7. Behavior summary

When `[gates.docstring]` is enabled, `centinela validate` (and `precommit`, `pr-gate`,
`verify`, and the new `centinela docs lint`) walk the Go AST of every changed file
and report each exported `func`, `type`, method, interface method, `const`, and `var`
that carries no doc comment. Scope is always the changed-file set: the `*gitdiff.Set`
the run already resolved (merge-base locally, staged files under `precommit`,
merge-base under `pr-gate`), or — when the run resolved none because it asked for a
full scan — the same merge-base set the gate resolves itself. A file nobody touched
is never opened, which is what makes `severity = "fail"` safe on a codebase with a
legacy backlog. Files that are `_test.go`, generated, or outside the configured roots
are never inspected; identifiers marked `//centinela:nodoc` are exempt and are listed
in the passing report so the exemption is visible. With no violations the gate
reports `Pass` naming how many identifiers it inspected; with violations it reports
`Fail` (or `Warn` where a project configured that) with one line per violation
carrying path, line, kind and name. When the scope is empty, contains no Go files, or
cannot be resolved at all, the gate reports `Skip` — never a green pass for a scan
that opened nothing. The whole-repo backlog stays visible through
`centinela docs lint --full`, a report-only surface that always exits 0.

### 8. Gherkin scenarios

Written as `specs/docstring-gate.feature`; each maps to an executable acceptance test
under `tests/acceptance/` carrying `// Acceptance: specs/docstring-gate.feature` and
`// Scenario: <exact name>` (the `spec_traceability` gate matches on those comments).

1. **Undocumented exported function in a changed file fails the gate** — happy path
   for detection. Fixture package with `func Exported() {}` and no doc → `Fail`, one
   `Details` line naming the function.
2. **Documented exported identifiers pass** — same fixture with doc comments → `Pass`,
   message names the inspected count.
3. **Warn severity reports the same violations without failing** — identical fixture,
   `severity = "warn"` → `Warn`, same `Details`, `gates.AllPassed` still true.
4. **An empty diff scope reports Skip, not Pass** — `filter.Len() == 0` → `Skip`.
5. **A diff scope with no Go files reports Skip** — filter contains only `README.md`
   → `Skip`; the negative path that a docs-only branch must not fake a green gate.
6. **A legacy unchanged file with undocumented exports is never scanned** — the
   ratchet itself: legacy file absent from the filter → `Pass`.
6b. **The gate ratchets to changed files even when the run asks for a full scan** —
   nil filter → the gate resolves the same merge-base set; legacy files outside it
   are never reported.
6c. **An unresolvable scope reports Skip, not Pass and not Fail** — nil filter plus a
   git failure → `Skip` naming the reason.
7. **A grouped const block's doc covers every constant in the block** — `const ( A; B )`
   with a block doc → no violations.
8. **Test files and generated files are never reported** — `foo_test.go` and a file
   headed `// Code generated by x. DO NOT EDIT.` both full of undocumented exports →
   `Pass`.
9. **A `//centinela:nodoc` identifier is exempt and the exemption is reported** →
   `Pass` with the exemption in `Details`.
10. **Struct fields are not reported when check_fields is false** — undocumented
    exported fields on a documented type → `Pass`.
11. **An unparseable Go file is reported, not silently skipped** → not `Pass`;
    `Details` names the file.
12. **An unknown severity is a configuration error** — `severity = "nope"` with
    `enabled = true` → config validation error naming `gates.docstring.severity`.
13. **`centinela docs lint` exits 1 on Fail and 0 on Warn** — CLI surface agrees with
    the gate on the same fixture.
13b. **`centinela docs lint --full` reports the legacy backlog without failing** —
    report-only surface, exit 0, prints the whole-repo undocumented count.
13c. **This repository ships the gate enforcing at fail severity** — `centinela.toml`
    has `[gates.docstring] enabled = true, severity = "fail"` and no scope knob.
14. **The senior-engineer prompt carries the doc-comment duty in both copies** —
    `docs/architecture/` and `internal/scaffold/assets/…` byte-identical and within
    the 130-line budget.

### 9. UX states

| State | Trigger | Surface |
|-------|---------|---------|
| loading | n/a — synchronous AST walk, sub-second | — |
| empty (Skip) | no changed files, no Go files in scope, or an unresolvable merge base | `centinela validate` gate line: `docstring-gate — no Go files in scope; nothing inspected` (Skip) |
| success (Pass) | zero violations | `docstring-gate — 885 exported identifiers documented` + exemption lines in `Details` |
| warning (Warn) | violations, `severity = "warn"` (not this repo's posture) | `docstring-gate — undocumented exported identifiers:` + one line per violation; validate still exits 0 |
| error (Fail) | violations, `severity = "fail"` | same body at Fail; `gates.AllPassed` false; `centinela validate` and `centinela docs lint` exit 1 |
| config error | unknown severity while enabled | config load fails: `gates.docstring.severity must be fail or warn, got "nope"` |

### 10. Out-of-scope

- Multi-language docstring scanning (seam named, Go-only implementation).
- Struct field doc comments (knob present, default off).
- Package-clause doc comments (572 files) — excluded entirely from v1.
- Name-prefixed-comment strictness (knob present, default off).
- Hunk-level / line-range diff scope.
- Judging whether a doc comment is *good* — presence only.
- Auto-writing missing doc comments.
- A docstring-specific baseline (use `audit-baseline`).
- Generating a godoc/Docusaurus site.
- Paying down the 215-item legacy backlog so a whole-repo `--full` scan could be
  clean (`docstring-full-scan-debt-paydown`).

### 11. Deferred findings

New discoveries measured during planning (not pre-agreed exclusions), recorded with
`--source docstring-gate/planner`; Backlog checked first for overlaps — none found:

- `package-doc-comments` — 572 of 599 non-test Go files carry no package-clause doc
  comment; a per-*package* (not per-file) rule is a separate check from v1.
- `docstring-struct-field-docs` — 597 undocumented exported struct fields, 73% of the
  non-test backlog; excluded from v1 behind `check_fields`.
- `docstring-full-scan-debt-paydown` — document the 215 legacy exported identifiers
  in files this ratcheted gate never opens, so `centinela docs lint --full` could one
  day report zero. (Replaces the plan-time `docstring-gate-ratchet-to-fail`, which is
  moot: the gate ships enforcing at `fail`.)

### 12. Handoff

**Next role:** `senior-engineer`.

Outstanding questions for the code step:

1. **Re-measure before adopting.** Run `centinela docs lint --full` on this repo and
   record the actual backlog count in the TOML comment. The number is context for
   `docstring-full-scan-debt-paydown`; it does not change the shipped posture, which
   is `severity = "fail"` scoped to changed files.
2. **Do not add a `scope` knob** to `[gates.docstring]` under any pressure; scope is
   `[validate] diff_mode`/`diff_base` repo-wide.
3. **`senior-engineer-prompt.md` has 21 lines of headroom.** Budget ≤8 for the duty
   and copy the mirror byte-for-byte in the same commit.
4. `internal/docstring` must import nothing under `github.com/samuelnp/centinela` —
   its leaf classification and the whole layering argument depend on it.
