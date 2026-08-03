# Edge Cases: evidence-schema-skeleton-legacy-handoff

All rows marked **[probe]** were executed against a scratch binary built from this
worktree (`go build -o /tmp/centinela-ec3 ./cmd/centinela`, `centinela version dev`)
using throwaway fixtures under the session scratchpad. Nothing was written into
this repo's `.workflow/` while probing. Rows marked **[theory]** are reasoned from
source and were not executed.

## Risk Matrix

| # | Case | Impact | Likelihood | Why |
|---|------|--------|------------|-----|
| **E1** | **Feature resolves from the `.worktrees/<feature>` path segment, but `handoffTo` derivation then fails because `workflow.WorkflowDir` is the RELATIVE `.workflow` — so running from any SUBDIRECTORY of the worktree silently falls through `handoffForRole` → `legacyHandoffForRole`** | **Critical** | **High** | **[probe]** From the worktree root: `feature=evidence-schema-skeleton-legacy-handoff, handoffTo=complete` (correct). From `internal/evidence/` in the SAME worktree: `feature=evidence-schema-skeleton-legacy-handoff, handoffTo=documentation-specialist` — the exact legacy value the brief exists to eliminate, now printed next to a *correct* feature slug, which makes it look more authoritative than the pre-fix output (which at least said `<feature-slug>`). Nothing on stdout or stderr distinguishes the two outputs. Likelihood is high: none of the 13 prompt lines that invoke this command (`docs/architecture/*-prompt.md` + the `internal/scaffold/assets` mirror) tell the agent to `cd` to the worktree root first, and agents routinely run commands from a package directory. Same result from inside `.workflow/` itself. |
| **E2** | **`evidence schema` guesses where its sibling `evidence init` refuses** | **High** | **High** | **[probe]** From `<worktree>/sub`: `centinela evidence init demo-solo gatekeeper` exits 1 with `unknown feature "demo-solo" (active: []) — run centinela start demo-solo first`. `centinela evidence schema gatekeeper` from the identical CWD exits 0 and prints the legacy chain. The brief's thesis is "the CLI hands out a value its own gate rejects"; post-fix the command is still the *only* evidence entry point that answers confidently when its own state lookup failed. |
| **E3** | Any `.worktrees/<x>` path segment is trusted unconditionally, with no check that a workflow exists for `<x>` | High | Medium | **[probe]** `~/fakeproj/.worktrees/not-a-feature/deep/er` → `feature='not-a-feature', handoffTo='documentation-specialist'`. SE trade-off #4 calls this "unchanged behaviour, not worse"; it is worse in *signal*, because the resolved `feature` field is now a real-looking slug rather than the obvious `<feature-slug>` slot. Every fabricated segment lands in the legacy branch, never the placeholder branch. |
| **E4** | Nested / repeated `.worktrees` segments resolve the OUTERMOST | Medium | Low | **[probe]** `/repo/.worktrees/outerfeat/.worktrees/innerfeat` → `outerfeat`; `/.worktrees/aaa/sub/.worktrees/bbb/c` → `aaa`. The scan returns on the first match walking left-to-right. If a nested checkout ever exists (a worktree of a worktree, or a vendored repo under a worktree) the *inner*, actually-active feature is shadowed by the outer one — a confidently wrong feature name, then E1/E3 supply a confidently wrong handoffTo on top. Low likelihood, no legitimate workflow creates this today. |
| **E5** | A corrupt/unparseable `.workflow/*.json` turns an AMBIGUOUS situation into a confident single answer | High | Low-Medium | **[probe]** Fixture `b4` = one valid active workflow + one corrupt `demo-corrupt.json` → resolves `demo-active` and derives `complete`. `ActiveWorkflows` warns on stderr and `continue`s, so the corrupt file is invisible to `len(wfs) == 1`. If the corrupt file was the *other* active workflow (parallel sessions — exactly the scenario `ResolveActiveFeature`'s doc comment says it refuses to guess about), the "exactly one" invariant is silently violated. Degrades to a **guess**, not to the placeholder. |
| **E6** | The `workflow warning:` line on stderr breaks the "output is valid JSON" contract for the standard agent capture path | Medium | Medium | **[probe]** In fixture `b4`, `centinela evidence schema gatekeeper 2>&1 \| python3 -m json.tool` → `Expecting value: line 1 column 1`. stdout alone is clean JSON, but Claude Code's Bash tool (and most agent harnesses) merge stderr into the captured text, which is precisely the "embedded in prompts and piped" use the brief calls out as a constraint. Pre-existing behaviour of `internal/workflow`, newly reachable from a command whose entire output contract is "valid JSON". |
| **E7** | `<successor-role>` is silently ACCEPTED by `evidence validate` / the completion gate for any role the pinned contract does not require | Medium | Medium | **[probe]** Fixture `c4` (canonical, no `validateContract` pin → validate step requires `validation-specialist`, not `gatekeeper`). Wrote `handoffTo="<successor-role>"` into `demo-legacy-gatekeeper.json`, filled inputs/outputs → `centinela evidence validate demo-legacy` → `evidence ok for "demo-legacy"`, rc=0. `validateHandoffChain` only iterates `RequiredEvidenceRoles`, so a non-required role's handoffTo is never checked. Same class: a feature with no workflow state at all (`CheckHandoffTo` returns nil when `Load` fails). The spec's scenario "pasting this JSON verbatim … reports a handoffTo issue" is only true for REQUIRED roles — the loud failure the placeholder design depends on is conditional, and the spec does not say so. |
| **E8** | Hostile `.worktrees/<segment>` names flow straight into `feature` and into `filepath.Join`/`Glob` consumers | Low | Low | **[probe]** `has space`, `100%pct`, `ñ-únïcode-☠`, `.hidden`, `-`, `star*glob`, `brack[et]`, and a segment containing a literal newline all resolve as the feature and print as correctly-escaped JSON (`"multi\nline"`). No crash, no injection into the JSON. Empty / `.` / `..` segments are impossible: `filepath.Abs` Cleans them away before the scan (verified in the differential test below), so there is no `.workflow/../x.json` traversal from this path. Residual: a glob-metacharacter slug (`*`, `[`) would misbehave in `evidence.Repair`'s `filepath.Glob`, but that is a different command. |
| **E9** | `EvalSymlinks` can both ADD and REMOVE the `.worktrees/<feature>` segment | Low | Low | **[probe]** A symlink pointing INTO a worktree subdir resolves correctly (`symfeat`) — the documented `/tmp`→`/private/tmp` intent works. Conversely, when the `.worktrees/<feature>` entry is itself a symlink to a directory outside the repo, the resolved real path loses the segment → `<feature-slug>`/`<successor-role>`. That is the SAFE direction (placeholder, not a guess), but it means a symlink-based checkout layout silently loses the worktree signal with no diagnostic. |
| **E10** | `cwd` is exactly the `.worktrees` directory | None | Low | **[probe]** `$B/repo2/.worktrees` → `<feature-slug>`/`<successor-role>`. Correct: the loop bound `i+1 < len(parts)` refuses to read past the end, so a trailing `.worktrees` yields no feature. |
| **E11** | Branch-2 (`exactly one active workflow`) table | — | — | **[probe]** zero active → placeholder ✓; one active → `demo-solo`/`complete` ✓; two active → placeholder, and `grep -c "demo-"` over the whole output is 0 (no candidate leaks) ✓; one active + one `currentStep:"done"` → the active one ✓; only a corrupt file → placeholder ✓ (safe); `feature` field disagreeing with the filename → skipped, placeholder ✓ (safe); mismatch file alongside a genuine active → the genuine one ✓. Only the E5 combination is unsafe. |
| **E12** | Derivation correctness across contracts | — | — | **[probe]** canonical/internal `gatekeeper`→`complete`; canonical/user-facing (`surface: user-facing` in `docs/features/<f>.md`) `senior-engineer`→`ux-ui-specialist`, `ux-ui-specialist`→`qa-senior`, `gatekeeper`→`documentation-specialist`, `documentation-specialist`→`complete`; hotfix `gatekeeper`→`complete`, `senior-engineer`→`qa-senior`, `qa-senior`→`gatekeeper`; spike `planner`→`senior-engineer`, `senior-engineer`→`complete`, `gatekeeper`→`complete`; legacy-pinned (no `validateContract`/`planContract`) `big-thinker`→`feature-specialist`, `planner`→`senior-engineer`, `validation-specialist`→`complete`, `gatekeeper`→`complete`; `merge-steward`→`complete` both with and without a resolved feature; `production-readiness`→`documentation-specialist` (user-facing) / `complete` (internal, hotfix). All 20 combinations match `ExpectedHandoff`, and `evidence init demo-solo planner` writes the identical value `evidence schema planner` prints from the same CWD — "one derivation, three callers" holds **as long as the CWD is the repo/worktree root** (see E1). |
| **E13** | Derivation silently depends on `docs/features/<feature>.md` being readable from the CWD | Medium | Medium | **[theory + probe]** `orchestration.IsUserFacingFeature` reads the relative path `docs/features/%s.md` and returns `false` on any read error. So the same CWD-relativity as E1 applies one layer deeper: a user-facing feature invoked from a directory where `docs/` is not visible derives the INTERNAL chain (`senior-engineer`→`qa-senior` instead of `→ux-ui-specialist`). Masked today by E1 (the `Load` failure fires first at every depth I probed), but it is a second, independent CWD dependency that a fix for E1 must also address. |
| **E14** | Output contract on the error paths | None | — | **[probe]** `evidence schema bogus` → rc=1, **stdout = 0 bytes**, error naming all 11 allowed roles on stderr. Zero args → rc=1, 0 bytes. Two args → rc=1, 0 bytes. Role parsing precedes `os.Getwd()` and all derivation, so no partial JSON is ever emitted. Success path ends `"complete"\n}\n` — 2-space-indented, single trailing newline, safe to embed and to concatenate. |
| **E15** | `worktree.DetectFeatureFromCwd` refactor regression | None | — | **[probe]** `git diff main -- internal/worktree/path.go` shows a verbatim body move; the only token change is the local `Dir` → `worktreepath.Dir`, both `".worktrees"`. Confirmed empirically with a throwaway differential test placed in `internal/worktree` (old inline implementation vs. the delegating one) over 24 inputs — `""`, `"."`, `".."`, `"/"`, `"/.worktrees"`, `"/.worktrees/f"`, `"/a/.worktrees/../b"`, `"/a/.worktrees//f"`, `"relative/.worktrees/f"`, `".worktrees"`, `"/a/.worktrees/f/.worktrees/g"`, `"/a/.worktreesX/f"`, `"/a/x.worktrees/f"`, `"/a/.worktrees/."`, `"/a/.worktrees/.."`, `"~/.worktrees/f"`, `"//a//.worktrees//f//"`, `"/a/.worktrees/ "`, `"/a/.worktrees/ñ"`, `"/a/.worktrees/*"`, … — **0 differences**. Test file removed after the run. The five other call sites (`internal/doctor/context.go:37`, `cmd/centinela/verify.go:52`, `hook_workflows.go:15`, `active_feature.go:13`, `roadmap_defer.go:55`, `hook_postwrite.go:61`) call an unchanged signature and are unaffected. |
| **E16** | New leaf `internal/worktreepath` ships with zero colocated tests | Medium | Certain | **[probe]** `ls internal/worktreepath/` → `path.go` only. Per-package coverage attributes 0% to it (`internal/worktree/path_test.go` exercises it only through the delegation, and coverage here is per-package with no `-coverpkg`). `grep -rl "ResolveActiveFeature" --include='*_test.go'` → nothing: `internal/evidence/schema_active.go` is entirely untested too. |
| **E17** | `tests/acceptance` does not compile on this branch | High | Certain | **[probe]** `go vet ./tests/...` → `tests/acceptance/deterministic_artifact_scaffolds_validate_test.go:87:76: not enough arguments in call to evidence.SchemaSkeleton`. Known and inventoried by the senior-engineer (`tests/` is unwritable during the code step); until it is fixed the whole acceptance tier is a build failure, so any acceptance assertion added for this feature cannot run. |

## Missing or Weak Scenarios

The 7 Gherkin scenarios in `specs/evidence-schema-skeleton-legacy-handoff.feature`
all pass **[probe]**, but every "derive with feature" scenario pins the CWD to the
worktree *root* ("the shell CWD is inside `.worktrees/demo-internal`" was
implemented as the root in the dogfood run). The spec therefore cannot fail on E1.

Specifically missing:

1. **No scenario runs from a worktree SUBDIRECTORY.** This is the single highest-value
   gap: the spec's central claim ("never emits a `handoffTo` the chain gate would
   refuse") is false at any depth > 0. (E1)
2. **No scenario asserts the negative for a resolved-but-stateless feature.** Every
   `.worktrees/<x>` where no `.workflow/<x>.json` is reachable prints the legacy
   chain. There is no test that a resolved feature with no derivable contract
   produces something other than `documentation-specialist`. (E3)
3. **No scenario covers a corrupt workflow JSON alongside a valid one.** The spec
   tests "two or more active" but not "two on disk, one unparseable" — the case
   where the anti-guess invariant actually breaks. (E5)
4. **The "pasted verbatim fails loudly" scenario is only true for required roles.**
   It should be split into a passing case (required role → loud, actionable error)
   and an explicitly acknowledged case (non-required role / no workflow →
   accepted). Today it reads as an unconditional guarantee. (E7)
5. **Nothing asserts stderr is empty on the success path**, or that the combined
   `2>&1` stream stays parseable. (E6)
6. **No nested-`.worktrees` scenario** and no assertion that the scan's
   first-match-wins order is intentional. (E4)
7. **No scenario for `production-readiness`**, the one role in `stepForRole` that is
   not in any `RequiredEvidenceRoles` list, so its derived value is never gate-checked.
8. **No differential/characterization test** locking `worktree.DetectFeatureFromCwd`
   to the pre-refactor behaviour (E15 verified it by hand; nothing in the suite
   would catch a future divergence).

## Proposed/Added Tests

### Unit (colocated, ≤100 lines each — G1 applies to `_test.go` in `internal/` and `cmd/`)

- `internal/worktreepath/path_test.go` — **required, package currently has none (E16).**
  Table over: inside `.worktrees/<f>`; inside `.worktrees/<f>/a/b`; `.worktrees` as the
  final segment (no feature); no `.worktrees` at all; `.worktreesX` / `x.worktrees`
  near-misses; nested `.worktrees` (assert outermost, documenting E4 as intended);
  `..`/`.`/`//` normalization; a symlinked path via `t.TempDir()` + `os.Symlink`
  (both directions from E9); a relative cwd. Assert the returned `root` too, not just
  the feature.
- `internal/worktree/path_characterization_test.go` — golden table asserting
  `DetectFeatureFromCwd` returns exactly the E15 pairs, so the delegation can never
  drift from the pre-refactor semantics.
- `internal/evidence/schema_active_test.go` — `ResolveActiveFeature` with `t.Chdir`
  (Go 1.24+) or the existing `chdirEvidenceTemp(t)` helper: worktree segment beats an
  unrelated single active workflow; exactly one active resolves; zero → `""`; two →
  `""` **and neither candidate slug appears anywhere in the rendered skeleton**;
  one active + one `done` → the active one; **one active + one corrupt JSON → assert
  the intended behaviour explicitly (E5)**; filename/feature mismatch → skipped.
- `internal/evidence/schema_skeleton_test.go` — `SchemaSkeleton("", role, v)` for
  **every** role in `ParseRole`'s allow-list → `<successor-role>`, except
  `merge-steward` → `complete`; `SchemaSkeleton(f, role, v).handoffTo ==
  Skeleton(f, role, v).HandoffTo == workflow.ExpectedHandoff(...)` for a fixture
  feature (the "one derivation, three callers" invariant, asserted rather than
  asserted-by-comment).
- **E1 regression unit test:** `ResolveActiveFeature` resolves `<f>` from
  `<tmp>/.worktrees/<f>/sub`, and `SchemaSkeleton` for that feature must NOT return
  `legacyHandoffForRole(role)` when the workflow is unreachable. Whatever the fix is
  (refuse and placeholder, or resolve the repo root before deriving), this test
  encodes the brief's guarantee at depth > 0. **Highest priority.**

### Integration (`tests/integration/`)

- Drive `runEvidenceSchema` (or the built binary) across a fixture tree with
  `.workflow/` at the root: assert the printed `handoffTo` for the full E12 matrix —
  canonical internal, canonical user-facing, hotfix, spike, legacy-pinned,
  merge-steward, production-readiness.
- **Root-vs-subdirectory equivalence test:** for one fixture, assert the JSON printed
  from `<root>` and from `<root>/pkg/sub` is byte-identical apart from nothing.
  This is the executable form of E1.
- Round-trip test: for each archetype fixture, take the printed `handoffTo`, write it
  with `evidence set`, then assert `workflow.CheckHandoffTo` returns nil — i.e. the
  skeleton and the gate can never disagree.
- stdout/stderr separation: success path → stderr is empty and stdout parses;
  corrupt-sibling path (E6) → stdout still parses on its own, and the test records
  that the merged stream does not.

### Acceptance (`tests/acceptance/`)

- **First, fix the build (E17):** `deterministic_artifact_scaffolds_validate_test.go:87`
  needs `""` as the new first argument to `evidence.SchemaSkeleton`. Nothing else in
  the tier compiles until then.
- Binary-driven acceptance for the 7 Gherkin scenarios, plus a new scenario
  "Derive with feature — invoked from a subdirectory of the worktree" (E1) and
  "Placeholder pasted for a role the contract does not require" (E7). Use a local
  fixture repo under `t.TempDir()`; **no network `git push`** (a prior feature hung
  the suite for hours that way).
- Error-path acceptance: `evidence schema bogus`, no-args, extra-args → rc=1 and
  **zero bytes on stdout** (E14).

### Test-hygiene follow-through (from the senior-engineer's handoff)

- Audit every `os.Chdir` in `cmd/centinela` for a `t.Cleanup` restore. Now that
  `runEvidenceSchema` reads the ambient CWD, a leaked chdir changes another test's
  result. `chdirEvidenceTemp(t)` is the existing pattern — do not add a second one.
- `cmd/centinela/evidence_more_test.go::TestEvidenceSchemaReturnsValidJSON` passes
  today only because it asserts `"step": "plan"` and nothing CWD-dependent. Pin its
  CWD for the same reason.

## Residual Risks

- **E1/E2 are in-scope defects, not deferrals.** The brief's expected outcome is
  "`centinela evidence schema` never emits a `handoffTo` the chain gate would refuse".
  As shipped it does so from any subdirectory of the very worktree it correctly
  identifies. Recommended fix, cheapest first: (a) have `ResolveActiveFeature` — or
  the command — resolve the repo/worktree root and derive from there (`worktreepath.
  DetectFeature` already returns `root`, which is currently discarded at
  `schema_active.go:24`); or (b) treat "feature resolved but no reachable workflow
  state" as *unresolved* and emit the placeholder, which is strictly safer than the
  legacy chain and matches the brief's "the author must decide" principle. Option (a)
  also fixes E13 for free. If neither is done in this feature, the gatekeeper should
  be told the brief's headline guarantee is not met.
- **E3/E4 (unconditional trust in the path segment) are intentional per the plan**
  (`ResolveActiveFeature`'s doc comment: "a worktree names exactly one feature, so no
  guessing is involved"). That reasoning is sound for a real worktree; the residual
  risk is that the code cannot tell a real worktree from any directory whose path
  contains `.worktrees/x`. Mitigated to "harmless" only if E1 is fixed via option (b),
  since a fabricated segment then yields the placeholder instead of the legacy chain.
- **E5 is bounded by how a corrupt `.workflow/*.json` arises** — a crashed write or a
  bad merge. The stderr warning exists; a caller that must not guess should treat
  "any unparseable candidate" as ambiguity. Left as a deferred finding because the
  fix belongs to `workflow.ActiveWorkflows`, whose other consumers (status, cost
  attribution) reasonably prefer the current skip-and-warn.
- **E6 is pre-existing** in `internal/workflow` and affects every command, not just
  this one. Deferred.
- **E7 is a property of the gate, which this feature is explicitly forbidden from
  changing** ("Out of scope: changing the chain-derivation rule itself, or the
  tolerance in `acceptsHandoff`"). Deferred. The actionable part inside this feature
  is documentation honesty: the `Long` help and the Gherkin scenario both promise an
  unconditional loud failure that is in fact conditional on the role being required.
- **E8/E9/E10 need no action.** JSON escaping is correct, `filepath.Abs` removes the
  traversal surface, and the trailing-`.worktrees` bound is right.
- **E15 carries no residual risk** — the refactor is behaviour-preserving, verified
  differentially. The only cost is that `internal/worktree`'s `Dir` is now defined
  in another package; a change to `worktreepath.Dir` silently changes both.
- **Coverage risk:** three touched packages (`internal/worktreepath` at 0%,
  `internal/evidence`'s new file, `cmd/centinela`) all need colocated tests to clear
  the 95% per-package floor; aim ≥97% so a parallel merge cannot tip main red.

## Deferred Findings

Filed to the validate-exempt Backlog with
`--source evidence-schema-skeleton-legacy-handoff/edge-case-tester`:

- `evidence-active-workflow-corrupt-json-ambiguity` — a corrupt `.workflow/*.json`
  lets `ActiveWorkflows` report "exactly one" when two are active, turning
  `ResolveActiveFeature`'s anti-guess invariant into a guess (E5).
- `workflow-warnings-break-json-stdout-capture` — `workflow warning:` on stderr makes
  `centinela evidence schema … 2>&1` unparseable for agent harnesses that merge the
  streams, against the "output stays valid JSON for prompt embedding" constraint (E6).
- `handoff-gate-skips-nonrequired-role-evidence` — `validateHandoffChain` only checks
  roles the pinned contract requires, so a literal `<successor-role>` (or any wrong
  value) in a non-required role's evidence passes `evidence validate` with rc=0 (E7).

**Not deferred — must be resolved inside this feature:** E1 and E2 (subdirectory
invocation reintroduces the legacy `handoffTo` this feature exists to remove), and
E17 (the `tests/acceptance` build break, already inventoried in the senior-engineer
handoff).
