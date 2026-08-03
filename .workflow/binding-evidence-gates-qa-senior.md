# binding-evidence-gates — qa-senior

**Step:** tests · **Status:** done · **Handoff:** gatekeeper

This step was FIX + TEST. The edge-case tester PROBED two verified bypasses in
the very gates this feature exists to close; both are fixed here, each with a
red-then-green test. A third finding (retroactive breakage the plan missed) is
answered with a stated, justified decision and pinned by test. A fourth
(pre-flight/gate disagreement) turned out cheap and is fixed rather than
deferred.

## Regression Guards

### B1 — `"exitCode": null` defeated the commands-schema check (CRITICAL)

**Defect.** `{"argv":["centinela","validate"],"exitCode":null}` satisfied BOTH
halves of `validateCommandEntry`: the key was present, and
`json.Unmarshal("null", &int)` is a documented NO-OP that returns `nil` and
leaves the zero value. The entry decoded to `Command{ExitCode: 0}` — "the gates
passed" — `Assess` returned grounded, and `artifact stamp` wrote it back
verbatim. The check's own comment names an absent key as the defect; `null` is
the same defect four characters wider.

**Fix** (`internal/gatereport/commands_schema.go`). A new `requireInt(field,
raw)` decodes through `*int` and demands non-nil. Because the rule lives in one
place, BOTH sides close at once — `Stamped` (write) and `ParseVerification`
(read) route through `ValidateCommandsSchema`, which is what slice A was for.
`durationMs` gets the same treatment.

**Red → green.** Tests written first against the unfixed tree:

```
TestValidateCommandsSchemaRejectsNullExitCode    FAIL (null + null durationMs accepted)
TestNullExitCodeErrorNamesTheField               FAIL (no error at all)
TestValidateCommandsSchemaNamesTheOffendingIndex FAIL
TestAssessRejectsNullExitCodeRecord              FAIL (verdict was GROUNDED)
TestCommandsSchemaReadWriteParity                FAIL (read=<nil> write=<nil>)
TestStampedBlockStaysBindingAfterHandEdit        FAIL
```

Counterfactual re-run (scratch copy, `requireInt` reverted to the plain int
decode): the same 6 unit tests plus acceptance
`TestBEG_StampRejectsNullExitCode` (`a null exitCode must not stamp: stamped
.workflow/demo-gatekeeper.md`) go red. Restored → all green.

### B2 — a changelog stub behind any preamble passed the docs gate (HIGH)

**Defect.** `validateChangelog` returned `nil` on the FIRST non-blank line that
carried no marker, so `## Changelog\n\n- <FILL: type>: …` was ACCEPTED. Adding a
heading to a report re-opened the gate on an untouched scaffold. Symmetrically,
the raw-substring detector made `- fix: reject an unreplaced <FILL: ...>
marker` — this feature's own most natural changelog line — a false positive.

**The rule, stated.** Two halves, both in the fix:

1. **Content-scoped, not positional.** Every non-blank line is asked the same
   question. An unreplaced slot anywhere fails. This also catches the
   half-filled entry `- fix: <FILL: one-line summary of the change>`, which no
   position-based rule, and no "is there prose beside the marker" rule, can
   refuse.
2. **The generic citation form is not a slot.** `<FILL: ...>` — a description
   that is empty or nothing but ellipsis characters — names no substance to
   fill in, and it is how the CLI's own errors and docs refer to the marker. It
   is read as a quotation. An unterminated `<FILL:` is likewise prose, because
   `FillSlot` always emits the closing bracket.

That pair is what lets "a changelog whose entries are all stubs must fail even
behind a preamble" and "prose that legitimately quotes the marker must still
pass" both hold. The escape hatch is documented in the refusal message itself.

**Verified against the exact line this feature will ship** —
`- fix: reject an unreplaced <FILL: ...> marker behind a preamble` passes, in
both the unit test and the acceptance test.

**Fix.** `orchestration.HasFillMarker` (raw substring) is replaced by
`orchestration.UnreplacedSlot`, which scans slot-by-slot and skips citations;
`workflow.validateChangelog` scans the whole file and reports the offending
line number, keeping the empty-file and missing-file remedies distinct.

**Red → green.** Written first; on the unfixed tree,
`TestValidateChangelogRejectsStubBehindPreamble` failed on 5 shapes (heading,
prose preamble, title+prose, stub after a filled entry, CRLF) and
`TestValidateChangelogAcceptsProseQuotingTheMarker` failed on 3 (this feature's
own line, unicode ellipsis, unterminated marker). Counterfactual re-run with
`UnreplacedSlot` reverted to `strings.Contains`: those 8 subtests plus
`TestUnreplacedSlotSeparatesSlotsFromCitations` (4 rows) and acceptance
`TestBEG_ChangelogStubBehindHeadingFailsDocsGate` go red.

### B3 — the old prefill's retroactive breakage (MEDIUM) — DECISION: do not widen

**The two cases.** On a **user-facing** feature the code step's derived
successor is the SAME-step `ux-ui-specialist`, but the old hardcoded prefill
seeded `senior-engineer → "qa-senior"`; `acceptsHandoff` returns false for a
same-step hop before the tolerance is consulted. On an **internal** feature at
validate, the prefilled `documentation-specialist` meets a derived `complete`
(terminal ⇒ tolerance off). Neither was in the plan's breakage inventory.

**Decision: leave both failing, with the remedy the error already prints.**
Justification, and why this is not the same shape as the alternate-pin
tolerance that IS allowed:

- The alternate-pin tolerance is safe because `handoffTo` there names the
  successor **STEP**, and that step has two legal occupants depending on which
  contract the workflow pinned. `TestContractPinFlipDoesNotWidenAcceptance`
  proves the accepted set is *identical* under both pins — the pin only swaps
  which name is canonical, so it cannot widen anything.
- `senior-engineer → qa-senior` is a different thing entirely: it does not
  misname the successor step's occupant, it names a **later step's** role. It
  asserts the same-step ux-ui-specialist hop does not exist. Accepting it would
  hand back the single property that makes the gate meaningful — that a
  required same-step role cannot be skipped.
- `gatekeeper → documentation-specialist` on an internal feature is the same
  shape at the other end: a handoff to a role the workflow requires no evidence
  from at all. "Internal features ship only a changelog" is precisely the rule
  the derivation exists to encode; tolerating a dangling pointer to a role that
  is never asked for would un-encode it.
- The migration cost is one pasteable command, and
  `TestBEG_OldPrefillOnUserFacingFeatureIsRefusedThenRepairable` proves that
  running exactly what the error prints makes the same gate pass. That test is
  the migration story, made executable.

**Pinned by.** `TestHandoffToleranceDisabledForSameStepAndTerminal` (both
cases, with their positive twins), `TestHandoffErrorNamesExecutableRemedy`,
`TestBEG_ToleranceRefusesAnotherStepsRole`, and — after a counterfactual showed
the first test still passed with the guard removed —
`TestHandoffSameStepGuardBlocksSkippingIntoValidate`.

**That counterfactual is worth recording.** Deleting `if sameStep || want ==
TerminalHandoff { return false }` did NOT turn the first boundary test red:
`alternateContractRoles` returns non-nil only for the `validate` and `plan`
steps, and under the default order the code step's successor is `tests`, whose
alternates are `nil`. The guard is only OBSERVABLE under `BootstrapStepOrder`,
the one shipped order in which the code step's successor is `validate` — the
exact configuration in which a user-facing `senior-engineer` could otherwise
name a validate-step role and skip the ux-ui-specialist outright. The new test
builds that fixture; with the guard removed it fails
(`handoffTo "validation-specialist" must be refused`). The terminal half of the
guard is genuinely unreachable today — a terminal hop yields `nextStep == ""`,
whose alternates are `nil` — so it is belt-and-suspenders, not a load-bearing
branch. The *behaviour* is pinned regardless, by the terminal rows in
`TestHandoffToleranceDisabledForSameStepAndTerminal`.

**Real breakage this decision cost, and paid.** The full suite surfaced three
`tests/`-tier fixtures the senior-engineer's inventory missed — all carrying
pre-gate literals, all fixture corrections rather than gate weakening:

| File | Was | Now | Why |
|---|---|---|---|
| `tests/integration/add_ux_ui_specialist_orchestration_integration_test.go` | `senior-engineer → "qa-senior"` | `"ux-ui-specialist"` | user-facing same-step hop — the B3 case itself |
| `tests/acceptance/docs_step_markdown_first_evidence_test.go` (2 tests) | `documentation-specialist → "orchestrator"` | `"complete"` | docs is the last step with required evidence |

### EC-7 — the pre-flight contradicted the gate (addressed, not deferred)

`centinela evidence set … handoffTo banana` succeeded, `evidence validate` then
printed `evidence ok`, and `complete` refused — the agents' own self-check
generating a false success one layer up, which is the thing this framework
exists to remove. It was cheap: `workflow.CheckHandoffTo` is now exported (and
`validateHandoffChain` is written in terms of it, so the two cannot diverge),
and `evidence.hintsForFile` calls it.

Scope is deliberately **identical to the gate's** — only roles in
`RequiredEvidenceRoles(feature, step)` — so the pre-flight can never be
*stricter* than what it previews. That exclusion is what keeps the out-of-band
merge-steward (literal verdict pair, checked in `ValidateEvidence`) and the
optional production-readiness role out of it. Zero test breakage.

Counterfactual (hint removed): `TestValidateFeatureAgreesWithTheCompletionChainGate`
and acceptance `TestBEG_EvidenceValidateAgreesWithComplete` go red.

## Test Inventory

**Unit — `internal/gatereport`** (3 new files)

- `commands_schema_test.go` — `TestValidateCommandsSchemaRejectsNullExitCode`
  (12-row table: null, absent, string, bool, 1.5, 0.0, 1e2, bignum, null
  durationMs, plus the three that must pass), `TestNullExitCodeErrorNamesTheField`,
  `TestValidateCommandsSchemaDefersAdmissibility` (absent/empty/null ARRAY stay
  Assess's job — the documented division of labour).
- `commands_schema_table_test.go` — `TestValidateCommandsSchemaTable` (19-row
  EC-11 matrix incl. duplicate keys last-wins and `argv: [""]`),
  `TestValidateCommandsSchemaNamesTheOffendingIndex`.
- `commands_schema_parity_test.go` — `TestAssessRejectsNullExitCodeRecord`,
  `TestCommandsSchemaReadWriteParity` (identical error TEXT from `Stamped` and
  `ParseVerification` over 6 malformed arrays),
  `TestStampedBlockStaysBindingAfterHandEdit`.

**Unit — `internal/workflow`** (7 new files)

- `handoff_fixture_test.go` — the shared `hoCase`/`hoRepo`/`hoEvidence` fixture.
- `handoff_matrix_test.go` — `TestExpectedHandoffDerivationMatrix` (12 rows:
  canonical internal across every step, user-facing same-step and docs, both
  legacy pins), `TestExpectedHandoffAcrossArchetypeOrders` (hotfix, spike,
  bootstrap), `TestExpectedHandoffTerminatesOutsideTheOrderedSteps` (merge and
  unknown step — EC-12's silent totality, now a stated behaviour),
  `TestExpectedHandoffReportsAbsentWorkflowState`.
- `handoff_tolerance_test.go` — `TestHandoffToleranceIsStepScoped` (2 accepted,
  19 refused incl. case variants, whitespace-padded, newline-suffixed,
  2000-char), `TestHandoffToleranceDisabledForSameStepAndTerminal`.
- `handoff_migration_test.go` — `TestHandoffSameStepGuardBlocksSkippingIntoValidate`,
  `TestHandoffErrorNamesExecutableRemedy`.
- `handoff_chain_test.go` — `TestContractPinFlipDoesNotWidenAcceptance`,
  `TestHandoffChainSilentOnUnreadableEvidence` (missing / unparseable / empty
  handoffTo / absent workflow — the no-double-report contract).
- `alternate_contract_roles_test.go` — `TestAlternateContractRolesMirrorsOnlyThePinnedBranches`
  (7 rows: both pins across validate/plan, and nil for code/tests/docs),
  `TestCheckHandoffToIgnoresAnUnsetField`.
- `validate_changelog_stub_test.go` — `TestValidateChangelogRejectsStubBehindPreamble`
  (9 shapes), `TestValidateChangelogAcceptsProseQuotingTheMarker` (6 shapes),
  `TestValidateChangelogKeepsEmptyAndMissingDistinct`.

**Unit — `internal/orchestration`** (2 new files)

- `fill_marker_test.go` — `TestFillSlotRendersTheCanonicalMarker`,
  `TestUnreplacedSlotSeparatesSlotsFromCitations` (14-row table),
  `TestScaffoldedSlotsAreDetectable`.
- `handoff_read_test.go` — `TestReadHandoffToReadsTheField`,
  `TestReadHandoffToReportsUnreadableFiles`, `TestStewardHandoffLiterals`.

**Unit — `internal/evidence`** (2 new files + 1 corrected)

- `schema_init_derived_test.go` — `TestSkeletonPrefillDerivesFromTheWorkflowContract`
  (5 roles on a user-facing feature), `TestSkeletonPrefillTerminatesOnInternalFeature`
  (4 roles), `TestFillMarkerAliasParity` (the re-export is the only thing
  keeping renderer and detector on one string).
- `validate_handoff_test.go` — `TestValidateFeatureAgreesWithTheCompletionChainGate`,
  `TestValidateFeatureStaysSilentOnInChainHandoffs`,
  `TestValidateFeatureSkipsRolelessFilenames`,
  `TestValidateFeatureSkipsRolesTheStepDoesNotRequire`.
- `artifact_changelog_test.go` — the plan's predicted non-breaking update: the
  comment claiming the stub exists "so the docs gate passes" is corrected (it
  deliberately does NOT), and an assertion is added that the scaffold still
  reads as a template under the gate's own predicate.

**Integration — `tests/integration`** — fixture correction only (see B3 table).

**Acceptance — `tests/acceptance`** (7 new files, all `TestBEG_*`)

- `binding_evidence_gates_helper_test.go` — real temp repos, real workflow
  state, real role evidence with real output files, so the STRUCTURAL check
  passes and each scenario turns on the rule under test.
- `..._handoff_test.go`, `..._steward_test.go`, `..._stamp_test.go`,
  `..._changelog_test.go` — the 11 spec scenarios.
- `..._migration_test.go`, `..._preflight_test.go` — the behaviours whose spec
  scenarios are blocked (below).
- `TestBEG_ScaffoldedChangelogFailsItsOwnGate` feeds the REAL
  `evidence.RenderTemplate(KindChangelog, …)` output straight into the REAL
  `workflow.ValidateArtifacts` docs gate, then the same body with its slots
  replaced — renderer and gate locked together end to end.
- The four stamp scenarios drive a binary built from `./cmd/centinela` in a
  real local git repo (`avvBuildBin` / `avvFixture`). No fixture ever contacts
  a network remote.

## Acceptance Wiring

Every file carries `// Acceptance: specs/binding-evidence-gates.feature`. All
**11** scenarios in the spec have an exact `// Scenario:` marker:

| Scenario | Test |
|---|---|
| A rejected banana handoff | `TestBEG_RejectedBananaHandoff` |
| A valid terminal handoff on an internal feature | `TestBEG_TerminalHandoffOnInternalFeature` |
| A valid mid-chain handoff on a legacy workflow | `TestBEG_MidChainHandoffOnLegacyWorkflow` |
| A valid same-step handoff when UX is required | `TestBEG_SameStepHandoffWhenUXRequired` |
| A rejected merge-steward handoff outside {complete, user} | `TestBEG_StewardHandoffOutsideVerdictPairRejected` |
| A valid merge-steward escalation handoff | `TestBEG_StewardEscalationHandoffAccepted` |
| A malformed commands array is rejected at stamp time | `TestBEG_StampRejectsEntryMissingExitCode` |
| A commands entry with an empty argv is rejected at stamp time | `TestBEG_StampRejectsEmptyArgv` |
| A well-formed commands array is accepted at stamp time | `TestBEG_StampAcceptsWellFormedCommands` |
| An unfilled changelog template fails the docs gate | `TestBEG_UnfilledChangelogFailsDocsGate` |
| A filled-in changelog passes the docs gate | `TestBEG_FilledChangelogPassesDocsGate` |

### BLOCKED: the 4 new spec scenarios

The prewrite hook refuses `specs/` writes during the tests step:

```
✖ Write blocked by workflow policy
Can't write "plan" files during "tests" step.
File  specs/binding-evidence-gates.feature
```

Per instruction I stopped rather than fought it. The behaviours are all
implemented and covered by tests (marked in-file as proposed scenarios); only
the Gherkin text is outstanding. Ready to paste — the first two into the
defect-2 and defect-3 blocks, the last two into the defect-1 block:

```gherkin
  Scenario: A null exit code is rejected at stamp time
    Given a gatekeeper report whose centinela:verification block has a
      commands entry {"argv": ["centinela", "validate"], "exitCode": null}
    When "centinela artifact stamp demo" runs
    Then the stamp is rejected
    And the error says exitCode must be an integer, not null, because a null
      decodes to 0 and reads as a passing run

  Scenario: A changelog stub behind a heading fails the docs gate
    Given ".workflow/demo-changelog.md" contains a "## Changelog" heading
      followed by the literal scaffolded stub
    When the docs step artifact gate runs for "demo"
    Then validation is rejected, because the rule is scoped to the entry's
      content rather than to the first non-blank line
    And a changelog that quotes the marker in its generic "<FILL: ...>"
      citation form still passes

  Scenario: The tolerance refuses a role from another step
    Given a feature "demo" at the "tests" step with qa-senior evidence
    And the evidence's handoffTo is "documentation-specialist"
    When "centinela complete demo" runs
    Then completion is rejected
    And the tolerance that lets the successor STEP be named under either
      contract pin does not admit it, because it is scoped to that one step

  Scenario: Evidence seeded by the old prefill on a user-facing feature
    Given a user-facing feature "demo-ux" at the "code" step whose
      senior-engineer evidence carries the pre-gate literal "qa-senior"
    When "centinela complete demo-ux" runs
    Then completion is rejected, because the derived successor is the
      SAME-step "ux-ui-specialist" and same-step hops are exact by design
    And the error names a runnable "centinela evidence set demo-ux
      senior-engineer handoffTo ux-ui-specialist"
    And running that command makes the same completion succeed
```

## Coverage Gaps

One profiled run: `go test ./... -coverprofile=coverage.out` — **46 packages
ok, 0 FAIL** — then `COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh`
→ `coverage gate passed: 97.3% >= 95.0%`.

| Touched package | Coverage | Note |
|---|---|---|
| `internal/gatereport` | 98.6% | `ValidateCommandsSchema`, `validateCommandEntry`, `requireInt` all 100% |
| `internal/orchestration` | 98.6% | was 92.7% entering this step — `UnreplacedSlot`, `FillSlot`, `ReadHandoffTo` had NO colocated tests |
| `internal/workflow` | 97.3% | `ExpectedHandoff`, `handoffTarget`, `nextChainStep`, `validateHandoffChain`, `acceptsHandoff`, `validateChangelog` all 100% |
| `internal/evidence` | 97.1% | `handoffForRole` 60% → 100% |
| `cmd/centinela` | 96.6% | untouched by this step |

Remaining gaps in `internal/evidence` are all pre-existing I/O error paths
(`writeTempFile` 80%, `Lock` 85%, `tryLockExclusive` 83%, `Repair` 87.5%) —
none belongs to this feature. Every line of code this feature added or changed
is covered.

Honest note: `tests/`-tier files do not move per-package coverage, so all of
the above comes from colocated `_test.go` files, each ≤100 lines.

## Deferred Findings

Already on the Backlog from the edge-case tester and still deferred:

- `workflow-state-file-disarms-its-own-gates` — `orchestrationMode: ""` or a
  deleted state file silently disables the whole strict-orchestration path;
  `validateContract: ""` disables the grounded-verdict gate. The trust root is
  agent-writable. Predates this feature; out of scope per the brief.
- `changelog-substance-floor-is-non-empty` — once no slot remains, a bare `-`
  or `x` is accepted. Unchanged here, deliberately: a substance floor is a
  different rule from a template check, and conflating the two is what makes
  the half-filled-entry case unsolvable.
- `fill-marker-detection-is-exact-substring` — `<fill:`, `< FILL:` and
  `<FILL type>` still pass. **Partially narrowed** here: the citation form
  `<FILL: ...>` is now a deliberate, documented non-match rather than an
  accident, and the false positive the slug also named (EC-8) is FIXED.
- `broaden-fill-marker-stub-detection` — the changelog is still the only
  consumer. This report and the edge-case report are both live examples of
  documents that discuss the marker at length; under the new citation rule a
  blanket application would be safer than before, but it is still unbuilt.
- `retire-handoff-alternate-pin-tolerance` (senior-engineer) — unchanged; the
  new `TestContractPinFlipDoesNotWidenAcceptance` and
  `TestAlternateContractRolesMirrorsOnlyThePinnedBranches` make the eventual
  tightening a one-line, fully-pinned change.

Resolved rather than deferred: `evidence-validate-skips-handoff-chain` — fixed
above; the Backlog entry can be closed.

New, not deferred but worth stating: the terminal half of the
`sameStep || want == TerminalHandoff` guard is currently unreachable (see B3).
It is correct and cheap; it is simply not the branch doing the work.

## Handoff

**Next role:** gatekeeper (validate step) — derived, not assumed:
`binding-evidence-gates` is internal (no `surface:` line in its brief), its
tests step requires only `qa-senior`, and its next step with required evidence
is validate, which this workflow pins to `adversarial-v1`.

**Dogfood of this feature's own gates against this feature's own evidence,**
using a binary built from the worktree (`go build -o /tmp/cent-qa
./cmd/centinela` — the installed 0.54.1 predates the fixes and seeded
`handoffTo: "validation-specialist"` on init):

```
/tmp/cent-qa evidence validate binding-evidence-gates -> evidence ok (exit 0)
```

That now includes the new chain check, so this evidence is verified against the
gate `complete` will run, not merely against the schema.

**For the verifier:**

- The three `tests/`-tier fixture corrections in the B3 table are the ONLY
  test-file behaviour changes; everything else in `tests/` and
  `internal/*_test.go` is additive. No assertion was relaxed and no gate
  threshold was touched.
- `internal/gatereport/commands_schema.go`, `internal/orchestration/fill_marker.go`,
  `internal/workflow/validate_docs.go`, `internal/workflow/handoff_chain.go` and
  `internal/evidence/validate.go` are the only production files changed in this
  step; `gofmt -l` and `go vet ./...` are clean, and every file under
  `internal/` and `cmd/` is ≤100 lines.
- No `docs/architecture/` file was touched in this step, so the
  `internal/scaffold/assets` mirror is untouched and stays byte-identical.
- **The changelog line this feature ships must avoid a bare unreplaced slot but
  MAY quote the marker generically.** `- fix: reject an unreplaced <FILL: ...>
  marker behind a preamble` is verified to pass its own gate
  (`TestValidateChangelogAcceptsProseQuotingTheMarker`,
  `TestBEG_ChangelogStubBehindHeadingFailsDocsGate`).
- The 4 spec scenarios above still need to be pasted into
  `specs/binding-evidence-gates.feature` at a step where `specs/` writes are
  permitted.

**Outstanding TODOs:** none. No skipped tests, no `t.Skip`, no commented-out
assertions introduced.
