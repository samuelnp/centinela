# evidence-schema-skeleton-legacy-handoff — senior-engineer

Step: code. `centinela evidence schema <role>` now derives `handoffTo` through the
same path `evidence init` and the completion gate use when the CWD resolves a
feature unambiguously, and prints the literal `<successor-role>` slot when it
does not. The CLI signature (`schema <role>`, `cobra.ExactArgs(1)`) is unchanged.

## Files Touched

| File | Change | Lines |
|------|--------|-------|
| `internal/worktreepath/path.go` | NEW leaf package: `Dir` + `DetectFeature` — the `.worktrees/<feature>` path vocabulary, stdlib only | 43 |
| `internal/worktree/path.go` | `Dir` re-exports the leaf's const; `DetectFeatureFromCwd` delegates to `worktreepath.DetectFeature`. Public API byte-identical for existing consumers | 41 |
| `internal/evidence/schema_active.go` | NEW: `ResolveActiveFeature(cwd)` — worktree segment, else exactly-one active workflow, else `""` | 31 |
| `internal/evidence/repair.go` | `placeholderFeature` / `unfilledHandoffSlot` consts; `SchemaSkeleton` gains the `feature` first parameter and the merge-steward-exempt placeholder override | 63 |
| `cmd/centinela/evidence_schema.go` | Resolves the feature from `os.Getwd()`; new `Short`/`Long` stating which of the two branches applies | 52 |
| `internal/evidence/repair_more_test.go` | Colocated call site: `SchemaSkeleton("", …)` | 42 |
| `internal/evidence/skeleton_not_poisoned_test.go` | Colocated call site: `SchemaSkeleton("", …)` | 45 |
| `cmd/centinela/evidence_schema_test.go` | `TestEvidenceSchemaEmitsSkeleton` now calls `chdirEvidenceTemp(t)` first — see Trade-Offs 3 | 33 |
| `PROJECT.md` | G2: registers `internal/worktreepath` as a leaf and states why | doc |
| `centinela.toml` | `gates.import_graph.layers` leaf `paths` += `internal/worktreepath/**`, with rationale comment | conf |

Not touched, deliberately: `schema_init.go` (`handoffForRole` /
`legacyHandoffForRole` are already correct for a KNOWN feature and out of scope),
`ExpectedHandoff`, `CheckHandoffTo`, `acceptsHandoff`, `alternateContractRoles` —
no gate weakened. Zero prompt / scaffold-mirror edits, exactly as the plan's §4
inventory predicted: all 13 `centinela evidence schema <role>` invocation lines
stay verbatim correct.

## Architecture Compliance

**The import-edge question (plan §3 flagged it for code-time verification).**
`internal/evidence → internal/worktree` **does not exist today** (`internal/evidence`
imports only `orchestration` + `workflow`), and PROJECT.md G2 sanctions
`internal/worktree` consumers only for `internal/verify`, `internal/doctor` (both
explicitly read-only aggregator allowances) and `cmd/`. `internal/worktree` is
also not a leaf — it imports `internal/config`. So the plan's "import
`internal/worktree` from `internal/evidence`" would have created a **new,
unsanctioned cross-package edge**; the `import_graph` gate would only warn (both
packages are unmapped), which is exactly the kind of silent boundary drift G2
exists to prevent.

Resolution — the codebase's own documented precedent, not a new idea:
`internal/acceptance` and `internal/docstring` were each extracted into a
stdlib-only leaf for precisely this reason ("exporting the predicate from
`internal/workflow` would force an `internal/verify → internal/workflow` edge,
while a stdlib-only leaf can be shared by all three"). The `.worktrees/<feature>`
scan is now the `internal/worktreepath` leaf: **one** implementation, shared by
`internal/worktree` (lifecycle) and `internal/evidence` (skeleton resolution),
with no new domain edge and no cycle possible. Verified with `go list`:
`internal/worktreepath` imports `path/filepath` + `strings` and **nothing
internal**, so it satisfies the leaf layer's `allow = []`. The alternative —
re-implementing the 12-line scan inside `internal/evidence` — was rejected: two
copies of the segment/symlink logic drift the day `Dir` or the symlink handling
changes.

- **G2 layers**: the new edges are `internal/worktree → internal/worktreepath` and
  `internal/evidence → internal/worktreepath`, both *→ leaf*, both allowed. Layer
  map and PROJECT.md updated alongside the package — additive registration of a
  new package, not a relaxation of any existing rule.
- **G7 outer layer**: `cmd/centinela/evidence_schema.go` stays a thin orchestrator
  — `os.Getwd()`, one call to `evidence.ResolveActiveFeature`, one call to
  `evidence.SchemaSkeleton`, write stdout. Every decision (which signal wins, when
  to refuse to guess, what the placeholder is) lives in `internal/`.
- **G1 ≤100 lines**: largest touched file is 63 lines (`repair.go`); all new files
  ≤43. No G1 exception needed or claimed.
- **G11 i18n**: N/A — English-only CLI per PROJECT.md; no locale catalogue exists.
- **Docstrings**: every new exported symbol (`worktreepath.Dir`,
  `worktreepath.DetectFeature`, `evidence.ResolveActiveFeature`) carries a doc
  comment; the delegating `worktree.DetectFeatureFromCwd` keeps its original one.

## Type-Safety Notes

- `SchemaSkeleton(feature string, role Role, cliVersion string)` — `feature` is a
  plain `string` deliberately: it is the same untyped slug `Skeleton`,
  `workflow.ExpectedHandoff` and every `.workflow/<feature>` path already speak;
  a bespoke `Feature` type here would only be converted away at all four call
  sites. `role` stays the `evidence.Role` alias, so a transposed
  `SchemaSkeleton(role, feature, …)` is a **compile error**, not a runtime bug —
  the one ordering hazard the new first parameter could have introduced.
- No `interface{}` / `any`, no reflection, no type assertions added.
- `ResolveActiveFeature` returns `string` with `""` as the single, documented
  "not resolved" value; it never returns a partially-resolved or guessed slug, so
  a caller cannot mistake "ambiguous" for "found".
- `worktreepath.DetectFeature` keeps the original named return pair
  `(feature, root string)`; `worktree.DetectFeatureFromCwd`'s signature is
  unchanged, so no consumer of it needed touching.

## Verification (code step — no full `centinela validate`)

```
go build ./...                 → BUILD OK
go vet ./...                   → 1 issue, the expected acceptance call site (below)
go test ./... -run xxxNONE     → 47/48 packages compile; tests/acceptance [build failed]
go test ./internal/evidence/... ./internal/worktree/... ./internal/worktreepath/... \
        ./cmd/centinela/... ./internal/doctor/... ./internal/verify/...
                               → 1021 passed, 0 failed
./scripts/check-fmt.sh         → OK
```

Dogfood, scratch binary `go build -o /tmp/centinela-ess ./cmd/centinela`:

```
A. inside this worktree — centinela evidence schema gatekeeper
   "feature": "evidence-schema-skeleton-legacy-handoff"
   "handoffTo": "complete"          ← pre-fix binary printed "documentation-specialist"

B. temp dir, no worktree, zero active workflows — gatekeeper
   "feature": "<feature-slug>"
   "handoffTo": "<successor-role>"   (valid JSON: json.load() round-trips)

B2. same dir, merge-steward   → "handoffTo": "complete"   (exempted, not placeholder'd)

C. .worktrees/demo-hotfix (archetype hotfix, no docs step) — gatekeeper
   "feature": "demo-hotfix"          "handoffTo": "complete"

D. non-worktree dir, exactly ONE active workflow → "feature": "demo-solo"

E. non-worktree dir, TWO active workflows
   "feature": "<feature-slug>"       "handoffTo": "<successor-role>"
   (grep -c "demo-" over the whole output = 0 — neither candidate leaks)

H. .worktrees/demo-ux, brief with `surface: user-facing` — senior-engineer
   "handoffTo": "ux-ui-specialist"   (same-step hop derived correctly)

F. centinela evidence schema bogus → Error: unknown role "bogus" …, stdout = 0 bytes
   (role parsing still precedes any derivation)

J. placeholder pasted verbatim into a real, resolvable feature's evidence:
   $ centinela evidence validate demo-solo
   [demo-solo/gatekeeper] evidence handoffTo for "gatekeeper" is "<successor-role>",
   but this workflow's contract makes "complete" its successor — fix with:
   centinela evidence set demo-solo gatekeeper handoffTo complete
```

(J) is the whole point: the placeholder is non-empty (clears the generic
completeness check) yet refused with the specific, actionable message — "the
author must decide", never "confidently wrong". All 7 Gherkin scenarios behave as
specified; the tests step must encode them as executable assertions.

## Trade-Offs

1. **New leaf package vs. the plan's direct import.** Costs one more package and a
   two-line delegation hop in `internal/worktree`; buys a sanctioned boundary and
   a single implementation of the scan. The plan explicitly deferred this to code
   time ("verify at code time that `internal/worktree` stays a leaf relative to
   `internal/evidence`"); the verification failed, so the plan's underlying intent
   (reuse the canonical scan, do not duplicate it) is honoured another way.
2. **`SchemaSkeleton` is now CWD-sensitive** through its caller — the inherent
   cost of deriving from the CWD instead of an argument, which the plan accepted.
   `SchemaSkeleton` itself stays pure: only `ResolveActiveFeature` and the command
   touch ambient state, so tests can still pin behaviour by passing the feature.
3. **The plan's §3 prediction about `TestEvidenceSchemaEmitsSkeleton` was wrong,
   and finding out cost a red test.** The plan reasoned that `cmd/centinela`'s
   test CWD "is neither inside a `.worktrees/<feature>` path nor has a
   `.workflow/` dir" — but in worktree mode the checkout *is*
   `…/centinela/.worktrees/<feature>/cmd/centinela`, so the test resolved this
   very feature and asserted `<feature-slug>` against
   `evidence-schema-skeleton-legacy-handoff`. Fixed by having the test pin its own
   CWD via the package's existing `chdirEvidenceTemp(t)` helper (which already
   restores through `t.Cleanup`) rather than inheriting the ambient one. A
   colocated one-line test edit forced by the behaviour change, not new coverage.
4. **The worktree signal is unconditional** (no "does this workflow exist?"
   check), per plan §1. A worktree directory names exactly one feature; if no
   state file exists for it, `handoffForRole` falls back to the legacy static
   value exactly as `evidence init` already does — unchanged behaviour, not worse.

## Deferred Findings

none. (`internal/worktreepath` needing its own colocated test is in scope for the
very next step and is listed under Handoff, not deferred to the roadmap.)

## Handoff

**Next role: qa-senior (tests step).**

### Breakage inventory the tests step MUST clear first

| File | Line | Break | Fix |
|------|------|-------|-----|
| `tests/acceptance/deterministic_artifact_scaffolds_validate_test.go` | 87 | `evidence.SchemaSkeleton(orchestration.RoleBigThinker, "1.0.0")` — `not enough arguments … have (orchestration.Role, string), want (string, evidence.Role, string)`; package `acceptance [build failed]`, and it is the **only** `go vet ./...` issue | add `""` as the first argument. Left untouched on purpose: `tests/` is not writable during the code step |

Nothing else fails to compile. The other three call sites
(`cmd/centinela/evidence_schema.go`,
`internal/evidence/skeleton_not_poisoned_test.go`,
`internal/evidence/repair_more_test.go`) are already updated.

### Tests to write (plan §5, plus what this step learned)

1. `internal/evidence/schema_active_test.go` (colocated, ≤100 lines) —
   `ResolveActiveFeature` table: worktree cwd wins over an unrelated active
   workflow; exactly one active workflow resolves; zero resolves `""`; **two or
   more resolves `""` and neither candidate name appears anywhere** (anti-guess).
2. `SchemaSkeleton` unit tests: `""` + every role → `<successor-role>`, except
   `RoleMergeSteward` → `"complete"`; a real on-disk feature → identical to
   `Skeleton(feature, role, v).HandoffTo` (one derivation, three callers).
3. **`internal/worktreepath/path_test.go` (colocated) — required.** New leaf with
   no colocated tests, so per-package coverage attributes 0% to it even though
   `internal/worktree/path_test.go` exercises it through the delegation. Cover:
   inside `.worktrees/<f>/sub`; a symlinked path (`/tmp` → `/private/tmp`); no
   `.worktrees` segment; `.worktrees` as the final segment (no feature after it).
4. Integration + acceptance for the 7 Gherkin scenarios via the built binary. The
   dogfood commands above translate directly, including the `evidence validate`
   refusal in (J) — that fixture needs `validateContract: adversarial-v1` in the
   workflow JSON for the gatekeeper role to be *required*, otherwise
   `handoffChainHint` skips the role and `evidence validate` reports "evidence
   ok" (a fixture trap I hit while verifying).
5. **Audit every `os.Chdir` in `cmd/centinela`** for a `t.Cleanup` restore (plan
   §5): now that `runEvidenceSchema` reads the CWD, a leaked chdir silently
   changes another test's outcome. `chdirEvidenceTemp` is the correct pattern —
   do not add a second one.
6. `cmd/centinela/evidence_more_test.go::TestEvidenceSchemaReturnsValidJSON`
   passes today but carries the same latent ambient-CWD dependency (it only
   asserts `"step": "plan"`). Pin its CWD too, for the same reason as (5).

### Outstanding TODOs

- Coverage: `internal/evidence`, `internal/worktree`, `internal/worktreepath` and
  `cmd/centinela` are all touched; aim ~97% aggregate (gate floor 95%).
- `.workflow/<feature>-edge-cases.md` is still owed by the tests step.
- No TODO/FIXME left in production code; nothing stubbed or mocked.
