# Plan: evidence-schema-skeleton-legacy-handoff

## 1. The decision: how the skeleton learns the feature

**Winner: derive from the CWD (option c), narrowed to two unambiguous signals,
with NO new CLI argument.** `centinela evidence schema <role>` keeps its exact
`Args: cobra.ExactArgs(1)` signature.

Resolution order, in a new `evidence.ResolveActiveFeature(cwd string) string`:

1. `worktree.DetectFeatureFromCwd(cwd)` — cwd is inside `.worktrees/<feature>`.
   Unconditional: a worktree names exactly one feature, no guessing involved.
2. Else `workflow.ActiveWorkflows(workflow.WorkflowDir)`; if **exactly one**
   workflow is active, use it. Zero or ≥2 active workflows both mean "no
   answer" — never picked-most-recent.
3. Else `""` (the genuine no-feature case).

### Why this beats the alternatives

- **(a) optional positional `[feature]`** and **(b) `--feature` flag** both
  require every one of the 8 role prompts that invoke this command to be
  rewritten to pass the slug, and both add a channel through which a
  copy-pasted or stale slug is silently wrong — a *typo'd* feature is exactly
  the "plausible but wrong" failure this feature exists to remove, and it's a
  new failure mode (c) cannot have (CWD cannot be typo'd by the agent). They
  are also strictly more CLI surface, more parsing/validation, and more test
  matrix to cover for zero behavioral gain in the dominant deployment mode —
  see `project_worktree_operational_model` in memory: "with use_worktrees,
  write artifacts into `.worktrees/<feature>/` and run centinela from inside
  it." That is precisely the shape (c) is built to recognize.
- **(d) never derive** (always emit an obvious unfilled slot) satisfies
  requirement 2 in isolation but abandons requirement 1: every ordinary
  worktree-mode invocation — the common case, and the one the bug was
  actually filed against (`beg-docstring-debt`) — would keep getting a
  placeholder it doesn't need, adding manual busywork to literally every
  `evidence schema` call and defeating "the CLI is the single source of
  truth" framing already in 8 prompts. (d)'s answer is *reused*, though — as
  the fallback for when derivation genuinely can't resolve. See §3.
- **Reusing `cmd/centinela/active_feature.go`'s `activeWorkflow(cwd)`
  verbatim was considered and rejected.** That helper falls back to
  "most-recently-touched of several active workflows" when there's no
  worktree signal — a fine heuristic for a cost-attribution *label* (wrong
  label is cosmetic) but wrong for a value a completion gate checks (wrong
  value is exactly the bug). `ResolveActiveFeature` is deliberately narrower:
  it refuses to guess among multiple active workflows. This is the direct
  answer to the "parallel sessions running other features" risk (§4).

## 2. The no-feature `handoffTo` value

**Literal placeholder string `"<successor-role>"`** (sibling to the existing
`"<feature-slug>"` convention SchemaSkeleton already uses for `feature`), NOT
empty string and NOT the literal `"complete"`.

Rejected alternatives:
- **Empty string** bypasses `workflow.CheckHandoffTo`'s chain check entirely
  (`internal/workflow/handoff_chain.go:42`: `if got == "" { return nil }` —
  that function's own docstring says an empty `handoffTo` is deliberately
  "its rule, not this one"). It would only be caught later by
  `orchestration.ValidateEvidence`'s generic `e.HandoffTo == ""` branch, which
  reports the same undifferentiated `"incomplete evidence fields"` message as
  a missing `inputs`/`outputs`/`status` — it doesn't even name the field.
- **Literal `"complete"`** is what `legacyHandoffForRole` already answers for
  several roles today — i.e. it's the exact "plausible but wrong" failure
  mode this feature removes. It would be silently ACCEPTED whenever the real
  successor happens to be terminal (masking the fix having worked) and
  silently WRONG whenever it isn't (repeating the bug).

### Verified consequence (this is the ironic self-check the brief calls out)

Live reproduction, from this very workflow, using the CURRENTLY-INSTALLED
(pre-fix) binary:

```
$ centinela evidence schema gatekeeper
...
  "handoffTo": "documentation-specialist"
```

`documentation-specialist` is `legacyHandoffForRole(RoleGatekeeper)`'s static
answer. But this workflow's *own* contract (canonical archetype, internal
feature — no `surface: user-facing` line in
`docs/features/evidence-schema-skeleton-legacy-handoff.md`) derives something
else: `RequiredRolesForFeature(feature, "docs")` returns `nil` for a
non-user-facing feature (`internal/orchestration/policy.go:52-56`), so
`nextChainStep` skips the docs step entirely and lands on
`TerminalHandoff` = `"complete"`. A gatekeeper report that pasted the
skeleton's `handoffTo` verbatim into `.workflow/<feature>-gatekeeper.json`
would fail `centinela complete` with:

```
evidence handoffTo for "gatekeeper" is "documentation-specialist", but this
workflow's contract makes "complete" its successor — fix with: centinela
evidence set evidence-schema-skeleton-legacy-handoff gatekeeper handoffTo
complete
```
— exactly the failure this feature exists to fix. (By contrast, this same
run's `centinela evidence schema planner` happens to print
`"handoffTo": "senior-engineer"`, which for THIS workflow is correct — legacy
and derived agree by coincidence for the planner role, so the bug does not
reproduce on every role every time; that's what makes it insidious.)

Once the fix lands, `centinela evidence schema gatekeeper` run with no known
feature will instead print `"handoffTo": "<successor-role>"`. Pasted verbatim
into a REAL feature's evidence, that value is non-empty (so it clears the
generic completeness check) but is not any real role name, so
`workflow.CheckHandoffTo` refuses it with the same specific, actionable
message shown above (naming the TRUE successor and the exact fix command) —
the failure mode becomes "the author must decide," never "confidently
wrong." Residual limitation, not a regression: for a feature with **no**
workflow state file at all, `ExpectedHandoff` returns `ok=false` by design
(no contract to derive from), so `CheckHandoffTo` is a no-op and the
placeholder would pass structurally — but `legacyHandoffForRole`'s old
static value was equally unchecked in that same no-workflow-state corner, so
this is strictly no worse, and the placeholder is far more likely to be
*noticed* on manual review than a plausible-looking role name.

## 3. File-by-file plan

### `internal/evidence/repair.go` (35 → ~55 lines)
- Add two package consts: `placeholderFeature = "<feature-slug>"` (extracted
  from the inline literal already used) and
  `unfilledHandoffSlot = "<successor-role>"`.
- Change `SchemaSkeleton(role Role, cliVersion string)` to
  `SchemaSkeleton(feature string, role Role, cliVersion string)`.
  - `f := feature; if f == "" { f = placeholderFeature }`
  - `skel := Skeleton(f, role, cliVersion)`
  - `if feature == "" && role != orchestration.RoleMergeSteward { skel.HandoffTo = unfilledHandoffSlot }`
    (merge-steward's `handoffForRole` already answers the fixed literal
    `"complete"` independent of feature — see `schema_init.go:44-46` — so it
    is deliberately exempted from the override; overriding it would replace an
    already-correct answer with a placeholder that then fails validation for
    no reason, breaking the out-of-band merge-steward scenario).
  - `return skel.MarshalJSON()`
  - `Repair()` is untouched.

### `internal/evidence/schema_active.go` (NEW, ~30 lines)
- `func ResolveActiveFeature(cwd string) string` implementing the resolution
  order in §1. Imports `internal/workflow` (already an `internal/evidence`
  dependency via `schema_init.go`) and `internal/worktree` (new dependency;
  confirmed no import cycle — `internal/worktree` does not import
  `internal/evidence` or `internal/workflow`... verify at code time that
  `internal/worktree` stays a leaf relative to `internal/evidence`).
- Pure given `cwd` and ambient `.workflow/` disk state — no other hidden
  inputs — so it is unit-testable with the existing `chdirToTemp` /
  `chdirEvidenceTemp` + `writeFakeWorkflow` fixtures already used by
  `cmd/centinela/evidence_init_test.go` and `internal/evidence/io_test.go`.

### `cmd/centinela/evidence_schema.go` (33 → ~45 lines)
- `runEvidenceSchema`: `cwd, _ := os.Getwd()`, then
  `feature := evidence.ResolveActiveFeature(cwd)`, then
  `evidence.SchemaSkeleton(feature, role, Version)`.
- Add a `Long` description (satisfies requirement 3 — "the command's help
  states which of the two applies") explaining both branches: derived when
  CWD resolves a single active feature; otherwise `handoffTo` is the
  `<successor-role>` placeholder and is expected to fail validation until
  replaced. This keeps the explanation in `--help` rather than in
  `docs/architecture/*.md`, so **no scaffold-mirror edit is needed for this**
  (see §4 prompt inventory — zero prompt files change).
- `Args: cobra.ExactArgs(1)` unchanged.

### Call-site / test breakage inventory (must all be updated together)

`SchemaSkeleton`'s signature change breaks every direct caller. None of these
assert a specific `handoffTo` value today, so none need behavioral
re-verification beyond adding the new first argument — but all four must
compile:

| File | Call site | Fix |
|------|-----------|-----|
| `cmd/centinela/evidence_schema.go:27` | `evidence.SchemaSkeleton(role, Version)` | becomes `evidence.SchemaSkeleton(feature, role, Version)` (production; see above) |
| `internal/evidence/skeleton_not_poisoned_test.go:22` | `SchemaSkeleton(orchestration.RoleBigThinker, "1.0.0")` | add `""` first arg |
| `internal/evidence/repair_more_test.go:11` | `SchemaSkeleton(orchestration.RoleFeatureSpecial, "v1")` | add `""` first arg |
| `tests/acceptance/deterministic_artifact_scaffolds_validate_test.go:87` | `evidence.SchemaSkeleton(orchestration.RoleBigThinker, "1.0.0")` | add `""` first arg |

Tests that call the COMMAND (`runEvidenceSchema`, unaffected signature) but
whose assertions depend on the NO-feature branch actually resolving to `""`
in the test's CWD — must be re-verified, not just recompiled, because
`SchemaSkeleton` becomes CWD-sensitive for the first time:

| File | Test | Why it should still pass, and the new dependency it now carries |
|------|------|----|
| `cmd/centinela/evidence_schema_test.go` | `TestEvidenceSchemaEmitsSkeleton` | Asserts `"feature": "<feature-slug>"` for role `big-thinker` with no chdir. Passes because `go test`'s CWD is the `cmd/centinela` package directory, which is neither inside a `.worktrees/<feature>` path nor has a `.workflow/` dir with active workflows relative to it — `ResolveActiveFeature` returns `""`. **New fragility**: this now depends on ambient CWD/filesystem state rather than being a pure function of its arguments; a stray `os.Chdir` left uncleaned by an earlier test in the same package (or a `.workflow/` directory ever appearing under `cmd/centinela/`) would silently change this test's outcome. The qa-senior step must audit every `os.Chdir` in this package for a `t.Cleanup` restore (the existing `chdirEvidenceTemp` helper already does this correctly) rather than add a new one. |
| `cmd/centinela/evidence_more_test.go` | `TestEvidenceSchemaReturnsValidJSON` | Same reasoning; only asserts `"step": "plan"`, unaffected either way. |
| `internal/evidence/skeleton_not_poisoned_test.go` | `TestSchemaSkeletonRepairInputsEmpty` | Only asserts absence of `"docs/plans/"`; unaffected by feature/handoffTo value. |
| `internal/evidence/repair_more_test.go` | `TestSchemaSkeletonReturnsPlaceholder` | Asserts presence of `"<feature-slug>"`; still emitted for the empty-feature branch. |
| `tests/acceptance/deterministic_artifact_scaffolds_validate_test.go` | `TestDAS_SkeletonStaysEmpty` | Same as above. |
| `tests/acceptance/agent_evidence_contract_acceptance_test.go` | `TestPromptsLinkToEvidenceContract` | Checks `strings.Contains(s, "centinela evidence schema "+role)` — a SUBSTRING match. Unaffected: the prompt text is not changing, and even if it grew a `[feature]` mention it would still contain this substring. Confirmed no prompt wording changes are planned (§4), so this test needs no changes at all. |

None of the existing tests hardcode a specific pre-fix legacy `handoffTo`
value from `SchemaSkeleton`/`runEvidenceSchema`, so there is no test asserting
the OLD (wrong) behavior that must be deleted — only the four signature call
sites need a one-argument edit.

## 4. Prompt / scaffold-mirror inventory

Searched every `docs/architecture/*.md` (and `*.template`) file for
`evidence schema` invocations. **8 files, 13 invocation lines, mirrored
byte-for-byte under `internal/scaffold/assets/docs/architecture/`** (16 files
total counting mirrors):

| Prompt | Invocation lines | Wording change needed? |
|--------|----|----|
| `planner-prompt.md` | 1 (L32) | No |
| `senior-engineer-prompt.md` | 2 (L29, L96) | No |
| `qa-senior-prompt.md` | 2 (L31, L102) | No |
| `ux-ui-specialist-prompt.md` | 2 (L33, L90) | No |
| `validation-specialist-prompt.md` | 2 (L38, L99) | No |
| `documentation-generator-prompt.md` | 2 (L33, L80) | No |
| `gatekeeper-prompt.md` | 1 (L41) | No |
| `production-readiness-prompt.md` (+ `.template` mirror) | 1 (L21) | No |

**Zero wording changes required**, and therefore **zero scaffold-mirror
edits required**, because the winning design (§1) adds no new required or
optional argument to the invocation — every one of these 13 lines already
reads `centinela evidence schema <role>` verbatim, and that remains the exact,
complete, correct invocation after the fix. None of the surrounding prose
("print the JSON skeleton — it is no longer embedded in this prompt" / "the
CLI is the single source of truth") makes a claim that becomes false;
`docs/architecture/evidence-contract.md` (also checked, incl. its mirror)
already documents the chain-derivation rule correctly and independent of this
CLI's no-feature fallback, so it is untouched too.

This zero-file outcome is itself evidence for the decision in §1: options
(a)/(b) would have required editing all 13 invocation lines (26 counting
mirrors) to add the new argument/flag, plus re-verifying
`TestPromptsLinkToEvidenceContract`'s substring match still holds — strictly
more surface, more mirror-parity risk, for a capability (c) provides for
free in the dominant operational mode.

## 5. Test plan (for the tests step)

- **Unit** (`internal/evidence/schema_active_test.go`, colocated,
  ≤100 lines): `ResolveActiveFeature` table — worktree cwd wins over an
  unrelated active workflow; non-worktree cwd with exactly one active
  workflow resolves it; non-worktree cwd with zero active workflows resolves
  `""`; non-worktree cwd with two-or-more active workflows resolves `""`
  (never picks either).
- **Unit** (`internal/evidence/repair_test.go` extension or new file):
  `SchemaSkeleton("", role, v)` emits `<successor-role>` for every role
  except merge-steward; `SchemaSkeleton("", RoleMergeSteward, v)` still emits
  `"complete"`; `SchemaSkeleton(feature, role, v)` for a real on-disk
  workflow matches `Skeleton(feature, role, v).HandoffTo` (the same value
  `evidence init` would prefill — no divergence between the three callers).
- **Integration**: build the real binary, `cd` into a temp dir laid out as
  `.worktrees/<feature>/`, run `centinela evidence schema <role>` from inside
  it, assert the printed `feature` and `handoffTo` match what
  `centinela evidence init <feature> <role>` would write.
- **Acceptance**: `tests/acceptance/evidence_schema_derives_feature_test.go`
  (or extend `deterministic_artifact_scaffolds_validate_test.go`) driving the
  six Gherkin scenarios in `specs/evidence-schema-skeleton-legacy-handoff.feature`
  via the built binary, matching `// Acceptance:` / `// Scenario:` comments.
- Update the four broken call sites (§3 table) as part of this step's first
  commit so the package compiles before writing new assertions.

## 6. Constraints carried into code/tests

- 100-line cap (incl. `_test.go`): `repair.go` grows to ~55, new
  `schema_active.go` ~30, `evidence_schema.go` ~45 — all comfortably under.
  `schema_init.go` (98 lines today) is **not touched** — no room left in it,
  and no edit is needed there (its derivation logic for a KNOWN feature is
  already correct and out of scope).
- No gate weakened: `ExpectedHandoff`, `CheckHandoffTo`, `acceptsHandoff`,
  `alternateContractRoles` are untouched.
- `legacyHandoffForRole` keeps working unchanged — still the fallback inside
  `handoffForRole` for the genuine no-workflow-state case that `evidence
  init` relies on; this feature does not touch it.
- Output stays valid JSON: both branches go through
  `RoleEvidence.MarshalJSON()` unchanged; only the string VALUE of one field
  differs.
