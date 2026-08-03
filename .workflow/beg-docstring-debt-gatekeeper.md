### Adversarial Verifier Report: beg-docstring-debt
**Date:** 2026-08-03
**Status:** SAFE

#### Inputs Read

- `internal/orchestration/evidence.go` (at HEAD, at `main`, and at pre-hotfix `a8de8aa`)
- `centinela.toml` (HEAD vs `main`)
- `internal/workflow/handoff.go`, `internal/workflow/handoff_chain.go`
- `internal/evidence/schema_init.go`
- `.workflow/beg-docstring-debt-senior-engineer.md`
- `.workflow/beg-docstring-debt-qa-senior.md`
- `.workflow/beg-docstring-debt-edge-cases.md`
- `.workflow/beg-docstring-debt.json`

#### Analyzed Specs

None exist for this slug, and none should: `beg-docstring-debt` is a `hotfix`
archetype (`centinela status` confirms `Archetype hotfix`, steps code → tests →
validate), so there is no `docs/plans/` file and no `specs/*.feature`. The
`specs/binding-evidence-gates.feature` on this branch belongs to the inherited
feature and was verified separately; `spec-traceability-gate` passed over its 15
scenarios in my validate run.

#### Refutation Attempts

**Claim attacked:** The gate was satisfied, not weakened — no configuration edit.
**How:** `git diff main -- centinela.toml` against the working tree (not merely the
merge-base three-dot form), plus a sweep of every changed file on the branch for any
config-shaped path.
**Result:** REFUTATION FAILED. `centinela.toml` is **byte-identical to `main`** — the
diff is empty. `[gates.docstring]` is still `enabled = true`, `severity = "fail"`. The
branch changes no `.toml`, no `.yml`/`.yaml` (so no CI `fetch-depth` or scope tampering
either), and no `.centinela*` file. Every changed non-`.workflow` file is Go, Markdown, or
the `.feature` spec. Nothing touched `enabled`, `severity`, `roots`, `check_fields`,
`diff_mode`, `diff_base`, or any exception list.

**Claim attacked:** A per-file exemption was slipped in for the affected file.
**How:** Read the gate's own exemption counter in full-scan mode, on both trees.
**Result:** REFUTATION FAILED. `docs lint --full` reports `(0 exempt)` on HEAD **and** on
`main`. Zero exemptions of any kind exist repo-wide on either side, so none was added —
neither a `[[gates.file_size_exceptions]]`-style block nor a docstring-specific one.

**Claim attacked:** The gate's SCOPE was narrowed rather than satisfied.
**How:** Checked whether the branch touches the scope machinery — `internal/docstring/**`,
`internal/gitdiff/**`, `internal/gates/**`, `internal/config/**` — then independently
recomputed the changed-file set the gate should see and compared it to what the gate
reports.
**Result:** REFUTATION FAILED. None of those packages appears in `git diff main...HEAD
--name-only`. Computing the changed non-test Go files under the configured roots myself
yields exactly **16**, and `internal/orchestration/evidence.go` is among them. The gate
reports the same 16. The scope is genuine and the fixed file is genuinely inside it — the
"fix" is not vacuous.

**Claim attacked:** The numbers the tests-step artifact cites ("22 exported identifiers
across 16 changed files").
**How:** Ran the gate myself at HEAD, both diff-aware and `--full`.
**Result:** REFUTATION FAILED. Diff-aware: **exit 0**, "All 22 exported identifiers across
16 changed Go file(s) are documented." Full: **exit 0** (informational by design), "170
undocumented of 903 exported identifiers across 639 files (0 exempt)". The 22/16 figures
hold exactly.

**Claim attacked:** The documented identifiers are the ones actually reported — nothing
was renamed or unexported to dodge the gate.
**How:** The strongest available test — checked out pre-hotfix `a8de8aa` into a throwaway
worktree and ran the gate there; then diffed the exported surface of `evidence.go`, and of
all 16 changed non-test files, against `main`.
**Result:** REFUTATION FAILED, decisively. At `a8de8aa` the gate **fails, exit 1**, naming
exactly `internal/orchestration/evidence.go:11: type Evidence` and `:25: func
ValidateEvidence`. At HEAD it passes. Those are precisely the two identifiers the hotfix
documented, and `main`'s full scan lists the same two. The `Evidence` struct body is
byte-identical to `main` (no field removed or unexported), `ValidateEvidence`'s signature
is unchanged, and a symbol-level diff across all 16 files shows **nothing removed**: the
four identifiers that left `internal/gatereport/check.go` (`ValidateArgv`,
`assessGrounding`, `assessRecord`, `hasPassingValidate`) all reappear in
`internal/gatereport/grounding.go` — a same-package file split by the inherited feature,
not a lost export that would have "passed" the gate while breaking callers.

**Claim attacked:** The change is genuinely documentation-only.
**How:** Stripped all comment and blank lines from `evidence.go` at `a8de8aa` and at HEAD
and diffed the remainder; separately inspected the raw hotfix diff for deletions and for
comment-shaped *directives*.
**Result:** REFUTATION FAILED. Non-comment content is **identical**. The hotfix diff is 13
added lines, all `//` prose, and **zero** removed lines. No `//go:`, `//nolint`, or
build-constraint line was introduced, and the comments sit after the import block, so no
build tag can be detached. `gofmt -l` is clean. There is no mechanism by which this diff
can alter runtime behavior.

**Claim attacked:** The repo-wide backlog was quietly gamed — new gaps hidden, or the
"fix" offset by fresh debt.
**How:** Measured `docs lint --full` on `main` in a second throwaway worktree and compared
to HEAD.
**Result:** REFUTATION FAILED. `main`: 172 undocumented / 894 exported / 632 files, 0
exempt. HEAD: 170 / 903 / 639, 0 exempt. The branch's net effect on the backlog is exactly
**−2** — the two identifiers this hotfix documented — while adding 9 new exported
identifiers, *all* of them documented. Nothing was hidden. (The `171 / 868 / 618` figure
in `centinela.toml`'s comment is the at-adoption measurement and is now stale against
main's tip; it is prose context, not an assertion any gate checks.)

**Claim attacked:** Shipping no tests leaves a regression path the gate cannot catch.
**How:** Enumerated the regression classes and asked, for each, whether the gate fires.
**Result:** PARTIALLY SUSTAINED — one narrow class is real, but low-impact. Behavior
regression is impossible (comment-only, proven above). Comment *deletion* is caught: any
future edit puts `evidence.go` back in the changed set and the gate re-fires — and this
still holds after the branch merges, because the deleting commit is itself the change that
pulls the file into scope. The one class no gate catches is comment **content accuracy**:
the gate asserts a comment exists, never that it is true, and the new `ValidateEvidence`
doc overstates its own scope (see Findings). A test could not reasonably close this gap
either — pinning comment prose asserts nothing about the program — so the qa-senior "no
new tests" decision is defensible on its merits, and
`.workflow/beg-docstring-debt-edge-cases.md` already discloses the staleness risk under
Residual Risks rather than concealing it.

**Claim attacked:** My own evidence can satisfy this branch's new derived-successor rule.
**How:** Ran `centinela evidence init`, then deliberately set `handoffTo` to the value the
CLI's own `evidence schema gatekeeper` skeleton prints, and re-validated.
**Result:** The rule works and my evidence satisfies it — but the probe surfaced a real
trap. `evidence init beg-docstring-debt gatekeeper` correctly derived `handoffTo:
"complete"` (hotfix archetype: `validate` is the terminal step). Setting it instead to
`documentation-specialist` — exactly what `centinela evidence schema gatekeeper` prints —
is **refused**: `evidence handoffTo for "gatekeeper" is "documentation-specialist", but
this workflow's contract makes "complete" its successor`. Restored to `complete`;
`evidence validate` now reports "evidence ok for \"beg-docstring-debt\"". Root cause:
`handoffForRole` falls back to the hardcoded `legacyHandoffForRole` table when no feature
is supplied to derive from, and `evidence schema <role>` supplies none. Deferred below,
not charged to this hotfix.

#### Commands Run

All run from `/Users/samuelnp/projects/personal/centinela/.worktrees/beg-docstring-debt`
unless noted. Exit codes captured directly via `; echo EXIT=$?`, never inferred from
banner text.

| # | argv | exit | duration |
|---|------|------|----------|
| 1 | `go build -o /tmp/centinela-hv2 ./cmd/centinela` | 0 | ~4s |
| 2 | `/tmp/centinela-hv2 validate` | **0** | **404s** |
| 3 | `/tmp/centinela-hv2 docs lint` | 0 | 1s |
| 4 | `/tmp/centinela-hv2 docs lint --full` | 0 | <1s |
| 5 | `/tmp/centinela-hv2 docs lint` (in `/tmp/hv2-prehotfix` @ `a8de8aa`) | **1** | <1s |
| 6 | `/tmp/centinela-hv2 docs lint --full` (in `/tmp/hv2-main` @ `main`) | 0 | <1s |
| 7 | `/tmp/centinela-hv2 evidence validate beg-docstring-debt` | 0 | <1s |
| 8 | `git diff main -- centinela.toml` | 0 (empty output) | <1s |
| 9 | `gofmt -l internal/orchestration/evidence.go` | 0 (no output) | <1s |
| 10 | `git worktree add --detach /tmp/hv2-prehotfix a8de8aa`; `... /tmp/hv2-main main` | 0 | <1s |
| 11 | `git worktree remove --force` (both) | 0 | <1s |

Command 2 is the single mandated `centinela validate` run. `[validate] commands` in
`centinela.toml` runs `go test ./... -coverprofile=coverage.out` over the whole module
(acceptance, integration, unit, and every colocated package test), so **that one run IS
the full suite run** — no separate suite invocation was needed or made. Its three commands
all passed:

```
✓  go test ./... -coverprofile=coverage.out
✓  COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh
✓  ./scripts/check-fmt.sh
 🛡️👁️  CLI  All gates passed.
```

Built-in gates in that same run: G1 file size ✓ (`evidence.go` is 93 lines), G-Build
cross-compile across 6 release targets ✓, import_graph ⚠ (the pre-existing, non-failing
unmapped-package warning — identical on `main`), spec-traceability ✓ (15 scenarios),
roadmap_drift ✓, **docstring-gate ✓**.

Both throwaway worktrees were removed; `git worktree list` is back to its original four
entries, and `git status` shows only this report as untracked (`coverage.out` is
gitignored). No scratch files remain in the worktree.

#### Findings

**Affected spec:** none (hotfix archetype — no `.feature` spec for this slug)
**Affected scenario:** the doc comments this hotfix added to `Evidence` and
`ValidateEvidence` in `internal/orchestration/evidence.go`
**Risk:** LOW — documentation accuracy; no behavior impact, no gate misled. The new
`ValidateEvidence` comment says the function reports "a role-specific rule (edge cases,
actionable outputs, **handoff successor**)", and the `Evidence` type comment states flatly
that "HandoffTo must name the successor the workflow's own contract derives". Neither is
true of this function as written: `ValidateEvidence` enforces only
`validateStewardHandoff`, a closed literal pair (`complete`/`user`) applying to the
merge-steward role *alone*; the derived-successor check for every other role lives in
`internal/workflow.CheckHandoffTo` and needs workflow state this package cannot see. A
caller trusting this docstring could believe `ValidateEvidence` alone enforces the chain
and skip the `internal/workflow` layer. This is precisely the class of defect the
docstring gate structurally cannot catch — it verifies that a comment exists, never that
it is true — which is the honest answer to "the gate is itself the executable check".
Mitigating: the `validateStewardHandoff` docstring twelve lines below states the split
explicitly ("Every OTHER role's handoffTo is checked against the chain its own workflow
derives ... that check lives in internal/workflow"), so the file is self-correcting for
anyone reading past the first function.
**Suggestion:** Non-blocking — do not hold the step for it. When the file is next touched,
qualify both sentences: on `Evidence`, "…and HandoffTo must name the successor the
workflow's own contract derives (merge-steward excepted: a literal APPLY/ESCALATE pair)";
on `ValidateEvidence`, "…or a role-specific rule (edge cases, actionable outputs, the
merge-steward handoff pair — the derived-successor chain check lives in
internal/workflow)".

No other findings. Specifically no CRITICAL and no WARNING: the gate was satisfied on its
merits, nothing was weakened, exempted, narrowed, renamed, or unexported, and the change
is provably inert at runtime.

#### Deferred Findings

Slug named in report text only. I deliberately did **not** run `centinela roadmap defer`,
because it mutates tracked roadmap state (`.workflow/roadmap.json`) *after* the single
`centinela validate` run that certified `roadmap_drift ✓` — filing it myself would
invalidate the very gate result this report attests to. Recommend the orchestrator file it
after this step closes:

- `evidence-schema-skeleton-legacy-handoff` — `centinela evidence schema <role>` prints a
  `handoffTo` drawn from the hardcoded `legacyHandoffForRole` table (gatekeeper →
  `documentation-specialist`), because no feature is supplied to derive from. On
  archetypes where that step is terminal (hotfix, spike) the new derived-successor gate
  refuses that exact value, so an author following the advertised skeleton hits a refusal
  — the "self-check disagrees with the completion gate" shape the inherited feature exists
  to remove. Belongs to `binding-evidence-gates`, not to this hotfix; impact is blunted
  because the refusal message prints the exact corrective command, and `evidence init`
  (the documented path) derives it correctly.

  `centinela roadmap defer evidence-schema-skeleton-legacy-handoff --summary "evidence schema <role> prints legacy handoffTo that the derived-successor gate refuses on hotfix/spike archetypes" --source beg-docstring-debt/gatekeeper`

#### Recommendation

**APPROVE — advance the validate step.**

I attempted eight distinct refutations of "beg-docstring-debt is complete and correct" and
every one of them failed. The strongest evidence is a red→green pair I produced myself
from two independent trees: the gate **fails with exit 1 at pre-hotfix `a8de8aa`**, naming
exactly `type Evidence` and `func ValidateEvidence`, and **passes with exit 0 at HEAD** —
with `centinela.toml` byte-identical to `main`, `0 exempt` on both sides, the scope
machinery untouched, an independently recomputed changed-set matching the gate's own 16
files, and a repo-wide backlog delta of exactly −2. The change is provably
documentation-only: strip comments and blank lines and `evidence.go` is byte-identical to
its pre-hotfix self, with zero deleted lines and no compiler directives. The claim
survives every attack I could mount.

`centinela validate` passed in full at exit 0 (404s), including the single profiled
`go test ./...` run that constitutes this project's entire suite.

The one finding is a LOW-severity documentation-accuracy overstatement inside the very
comment this hotfix added — non-blocking, self-corrected by the adjacent docstring, and
worth one sentence's edit the next time the file is touched. It does not warrant WARNING
status because it changes no behavior and misleads no gate.

```json centinela:verification
{
  "revision": "7e88ad39e4567d9015c02c15b57a027eb69af0bb",
  "treeDigest": "sha256:96a296d224f285c67bee93c30f8a309157f0daa35dc5b87e410b78630a09cfc7",
  "commands": [
    {"argv": ["go", "build", "-o", "/tmp/centinela-hv2", "./cmd/centinela"], "exitCode": 0, "durationMs": 4000},
    {"argv": ["/tmp/centinela-hv2", "validate"], "exitCode": 0, "durationMs": 404000},
    {"argv": ["/tmp/centinela-hv2", "docs", "lint"], "exitCode": 0, "durationMs": 1000},
    {"argv": ["/tmp/centinela-hv2", "docs", "lint", "--full"], "exitCode": 0, "durationMs": 500},
    {"argv": ["/tmp/centinela-hv2", "docs", "lint"], "exitCode": 1, "durationMs": 500},
    {"argv": ["/tmp/centinela-hv2", "docs", "lint", "--full"], "exitCode": 0, "durationMs": 500},
    {"argv": ["/tmp/centinela-hv2", "evidence", "validate", "beg-docstring-debt"], "exitCode": 0, "durationMs": 500},
    {"argv": ["gofmt", "-l", "internal/orchestration/evidence.go"], "exitCode": 0, "durationMs": 200}
  ]
}
```
