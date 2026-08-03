# binding-evidence-gates — senior-engineer

**Step:** code · **Status:** done · **Handoff:** qa-senior

Three fail-open gates now bind. Implemented as the plan's three independent,
independently revertible slices, in order: A (stamp commands schema), B
(changelog stub rejection), C (handoffTo chain validation).

## Files Touched

| File | Slice | Change |
|---|---|---|
| `internal/gatereport/commands_schema.go` (new, 68) | A | `ValidateCommandsSchema` + `validateCommandEntry` — the ONE place that knows the `commands` shape |
| `internal/gatereport/block.go` (71) | A | `ParseVerification` schema-checks raw `commands` before decoding into the typed slice |
| `internal/gatereport/stamp.go` (62) | A | `Stamped` validates before splicing, on BOTH the existing-block and created-block branches |
| `internal/gatereport/check.go` (33) | A | reduced to `Assess`/`criticalError`; grounding moved out to stay under the size cap |
| `internal/gatereport/grounding.go` (new, 75) | A | `assessGrounding`/`assessRecord`/`hasPassingValidate` + new `groundingParseError` |
| `internal/orchestration/fill_marker.go` (new, 29) | B | `FillMarker` const, `FillSlot`, `HasFillMarker` — relocated so both sides can reach it |
| `internal/evidence/fill.go` (14) | B | `FillMarker`/`FillSlot` become thin re-exports; every existing caller unchanged |
| `internal/workflow/validate_docs.go` (49) | B | `validateChangelog` rejects an entry line still carrying `<FILL:` |
| `internal/orchestration/evidence.go` (80) | C | `validateStewardHandoff` — merge-steward's literal `complete`/`user` pair |
| `internal/orchestration/handoff_read.go` (new, 27) | C | `ReadHandoffTo` — one-field evidence read for the workflow layer |
| `internal/workflow/handoff.go` (new, 78) | C | `ExpectedHandoff`, `handoffTarget`, `nextChainStep` — the derivation |
| `internal/workflow/handoff_chain.go` (new, 61) | C | `validateHandoffChain`, `acceptsHandoff` |
| `internal/workflow/validate_orchestration.go` (88) | C | wires the chain check in after `ValidateRoles`; adds `alternateContractRoles` |
| `internal/evidence/schema_init.go` (98) | C | `handoffForRole(feature, role)` now derives; static table kept as `legacyHandoffForRole` fallback |
| `docs/architecture/evidence-contract.md` (+ scaffold mirror) | C | global rule 4 states the derivation; 3 stale per-role lines corrected; mirror `cmp`-verified byte-identical |
| `internal/evidence/roles_test.go` | C | call site follows the rename (`legacyHandoffForRole`); table unchanged |
| `cmd/centinela/complete_cmd_test.go`, `cmd/centinela/complete_branches_test.go`, `internal/workflow/validate_orchestration_docs_test.go` | C | the 3 fixtures the plan predicted: `handoffTo: "orchestrator"` → `"complete"` |

## The successor rule

`ExpectedHandoff(feature, step, role)` derives the successor from the
workflow's own contract, never a literal chain:

1. the next required role WITHIN the same step, when one follows it there
   (`senior-engineer → ux-ui-specialist`, legacy `big-thinker →
   feature-specialist`);
2. else the first required role of the next step whose
   `RequiredEvidenceRoles` is non-empty — steps requiring none are SKIPPED,
   which collapses "archetype omits the step" and "internal feature, docs
   needs no role" into one case needing no special-casing;
3. else the literal `"complete"`.

Both inputs already carry the knowledge: `RequiredEvidenceRoles` is
contract-pin aware, `OrderedSteps()` is archetype aware. `merge-steward` never
appears in `OrderedSteps()` and so has no derivable successor; its rule is a
self-contained literal pair in `orchestration.ValidateEvidence`.

**One deliberate looseness (deviation from the plan — see Trade-Offs):** a
NEXT-step hop also accepts the successor step's occupant under the *other*
contract pin (`alternateContractRoles`). A same-step hop and a terminal
`complete` remain exact-match.

## Architecture Compliance

- **Boundaries (n-tier).** No new package edges. The marker moved *down* to
  `internal/orchestration` (imports nothing internal) precisely to avoid the
  `workflow → evidence` edge that would cycle. Verified before writing code,
  not assumed: `go list -deps` confirms `evidence → workflow → orchestration`,
  and `orchestration` has no internal imports. `internal/workflow` reads
  evidence through `orchestration.ReadHandoffTo` rather than duplicating JSON
  parsing or importing `internal/evidence`.
- **G1 (≤100 lines).** Every touched and new file is ≤98 lines. `check.go`
  crossed the cap at 105 while wiring slice A and was split into
  `check.go` (verdict assembly, 33) + `grounding.go` (admissibility, 75) along
  the seam that already existed in the code. Largest file touched:
  `internal/evidence/schema_init.go` at 98.
- **G7 (no hardcoded user-facing strings).** This is a Go CLI with no i18n
  layer; all new strings are operator-facing gate errors written in the same
  voice as their neighbours, each naming the exact remediation command.
- **One rule, one place.** `ValidateCommandsSchema` is called from the reader
  and the writer. `assessRecord`'s surviving empty-argv loop is the
  pre-existing *admissibility* check, not a second copy of the shape rule —
  the distinction is stated in its doc comment.

## Type-Safety Notes

- `ValidateCommandsSchema` takes `json.RawMessage` and decodes through
  `[]map[string]json.RawMessage` rather than the typed `Command`, because the
  typed decode is lossy in exactly the direction that matters: a missing
  `exitCode` becomes `0` (= passed) and a non-object entry vanishes. The check
  demands the KEY be present, not merely decodable — closing a smaller
  instance of the same accepts-what-it-should-reject class.
- Unknown keys are ignored on purpose: the block is a forward-compatible
  record, so a future field must not brick an older binary.
- `ExpectedHandoff` returns `(string, bool)` rather than `(string, error)`:
  "no workflow state to derive from" is a legitimate absence the prefill must
  branch on, not a failure to report.
- `handoffTarget` returns `sameStep bool` so `acceptsHandoff` cannot
  accidentally apply next-step tolerance to a same-step hop.
- No `interface{}`/`any` introduced anywhere.

## Trade-Offs

**1. Next-step handoffs accept either contract pin's occupant.** The plan
specified exact-match against the single derived successor. That would have
retro-broken the parallel in-flight `docstring-gate` session: it is at the
`tests` step *right now* with real, on-disk `qa-senior` evidence carrying
`handoffTo: "validation-specialist"` (seeded by the very hardcoded prefill
this feature fixes) on an `adversarial-v1` workflow whose derived successor is
`gatekeeper`. The plan's risk table judged this "only bites at a terminal
step"; it bites mid-chain too, because the stale prefill — not the stale doc —
is the source. Retro-failing it violates the brief's "fail-closed but not
retroactively brittle" constraint, so `acceptsHandoff` treats `handoffTo` as
naming the successor STEP and accepts either legal occupant of it
(`alternateContractRoles`, written as the literal other arm of
`RequiredEvidenceRoles`' two pin branches and kept adjacent to it). Nothing is
weakened elsewhere: `banana`, `orchestrator`, a role from the wrong step, a
same-step hop, and a terminal `complete` are all still exact. Verified against
the real file, not a replica (dogfood S/S2). Sunset deferred — see below.

**2. Malformed-block errors now surface their cause.** Slice A made
`ParseVerification` reject an empty-`argv` entry at parse time, which
previously surfaced at `assessRecord`. Rather than let `Assess` degrade to the
generic "no commands-run record", `groundingParseError` keeps that phrase (the
statusline classifier in `cmd/centinela/hook_statusline_validate.go` matches on
it) and appends what is actually malformed. "No record" alone would send a
verifier hunting for a fence it can plainly see.

**3. Stub detection stays changelog-only.** Per plan §3d; `HasFillMarker` is
applied to the changelog ENTRY line only, so a changelog whose later prose
discusses the marker still passes (dogfood P).

**4. `validateHandoffChain` reloads the workflow.** `validateOrchestration`
does not hold the `*Workflow` that `strictOrchestrationEnabled` already
loaded. Threading it through would have touched a signature for no behavioural
gain; the load is a single file read on an already-I/O-heavy gate.

## Deferred Findings

- `retire-handoff-alternate-pin-tolerance` — deferred to Backlog via
  `centinela roadmap defer retire-handoff-alternate-pin-tolerance --summary "..." --source binding-evidence-gates/senior-engineer`.
  Once no in-flight or downstream workflow carries a pre-gate `handoffTo`,
  tighten `acceptsHandoff` to exact-match and delete `alternateContractRoles`.
- `broaden-fill-marker-stub-detection` — already on the Backlog from the
  planner (plan §3d); not re-deferred.

## Dogfood results

Scratch binary `go build -o /tmp/centinela-beg ./cmd/centinela`, run against
isolated sandboxes (never against the live worktrees).

Defect 1 — handoffTo chain:

```
handoffTo=banana        -> Error: evidence handoffTo for "qa-senior" is "banana", but this
                           workflow's contract makes "gatekeeper" its successor — fix with:
                           centinela evidence set dg qa-senior handoffTo gatekeeper
handoffTo=documentation-specialist (right role, WRONG step) -> same refusal
handoffTo=gatekeeper    -> Step "tests" completed for "dg".
spike  [plan,code]      -> senior-engineer expected "complete"   (banana refused, complete accepted)
hotfix [code,tests,validate] -> senior-engineer expected "qa-senior"
legacy (no adversarial-v1 pin) -> qa-senior expected "validation-specialist"
user-facing code step   -> senior-engineer expected "ux-ui-specialist", NOT "qa-senior"
internal feature        -> gatekeeper at validate expected "complete" (docs needs no role)
merge-steward: orchestrator -> refused ("must be \"complete\" (APPLY) or \"user\" (ESCALATE)")
merge-steward: complete/user -> evidence ok
```

Real parallel-session evidence (`.workflow/docstring-gate-qa-senior.json`
copied verbatim into a sandbox):

```
handoffTo=validation-specialist (its ACTUAL on-disk value) -> Step "tests" completed
handoffTo=banana (same file, corrupted)                    -> refused, names "gatekeeper"
```

Prefill, which previously seeded values the new gate would reject:

```
qa-senior  (adversarial-v1)   -> "gatekeeper"  [was hardcoded "validation-specialist"]
gatekeeper (internal feature) -> "complete"    [was hardcoded "documentation-specialist"]
```

Defect 2 — stamp commands schema:

```
[{"argv":["centinela","validate"]}]          -> commands[0]: exitCode is required (an absent
                                                key would silently read as a pass)
[{"argv":[],"exitCode":0}]                   -> commands[0]: empty argv (a recorded command
                                                with no argv is not evidence)
[{"argv":"centinela validate","exitCode":0}] -> commands[0]: argv must be an array of strings
["centinela validate"]                       -> commands must be an array of objects
[{"argv":[...],"exitCode":"0"}]              -> commands[0]: exitCode must be an integer
[{"argv":["centinela","validate"],"exitCode":0,"durationMs":1200}] -> stamped;
   revision + treeDigest recorded, commands array preserved byte-verbatim
```

Defect 3 — changelog stub:

```
centinela artifact new dg changelog -> "- <FILL: type>: <FILL: one-line summary of the change>"
docs gate on that stub -> Error: changelog entry is still a template placeholder for "dg":
   .workflow/dg-changelog.md (replace the <FILL: ...> slots with a real one-line summary)
docs gate on "- feat: bind the evidence gates so a stub can no longer pass" -> completes
docs gate on a filled entry whose LATER prose mentions <FILL: type>        -> completes
```

## Handoff

**Next role:** qa-senior (tests step).

**Verification state at handoff:** `go build ./...` clean, `go vet ./...`
clean, `gofmt -l` clean, `go test ./... -run xxxNONE` clean (no test-compile
breaks), full suite **2926 passed, 0 failed across 43 packages**. The tree is
GREEN, not transiently red — the plan's 3 predicted fixtures were the only
breakage and are fixed. A full `centinela validate` was NOT run (code step).

**Compile/runtime breakage inventory (all resolved; listed so the fixes can be
confirmed as fixture corrections, not gate weakening):**

1. `cmd/centinela/complete_cmd_test.go` (`f2`) — `handoffTo:"orchestrator"` →
   `"complete"`. Predicted by plan §6.1.
2. `cmd/centinela/complete_branches_test.go` (`f`) — same. Plan §6.2.
3. `internal/workflow/validate_orchestration_docs_test.go` — same. Plan §6.3.
4. **NOT predicted:** `internal/evidence/roles_test.go` failed to COMPILE —
   `handoffForRole` gained a `feature` parameter. Call site retargeted to
   `legacyHandoffForRole`; the assertion table is byte-identical, and the test
   was renamed `TestLegacyHandoffForRoleCoverage` to say what it now covers.
5. **NOT predicted:** `internal/gatereport/check_argv_test.go::TestAssessEmptyArgvEntryBlocks`
   failed at runtime — an empty `argv` is now caught at parse time, not at
   `assessRecord`. Fixed in PRODUCTION code, not the test: `groundingParseError`
   surfaces the malformed detail, and the schema message reads "empty argv" so
   the existing assertion still holds unmodified.

**For the tests step:**

- `internal/evidence/artifact_changelog_test.go::TestRenderTemplateChangelogEmitsNonBlankOneLiner`
  still PASSES, but its comment ("so the docs gate passes") is now false — the
  stub deliberately fails that gate. Correct the comment and add the negative
  test (plan §6, "non-breaking update").
- 12 spec scenarios need executable assertions with `// Acceptance:` +
  `// Scenario:` comments. Scenarios 1–6 map to
  `validateHandoffChain`/`acceptsHandoff`/`validateStewardHandoff`, 7–9 to
  `ValidateCommandsSchema` via `Stamped`, 10–11 to `validateChangelog`.
- Coverage: `internal/workflow/handoff.go` + `handoff_chain.go` and
  `internal/gatereport/commands_schema.go` are new and currently exercised only
  indirectly. Per-package ≥97% on touched packages is the constraint; colocated
  `_test.go` files are what move that number, and they are subject to the same
  ≤100-line cap.
- Worth an explicit test each: the alternate-pin tolerance
  (`qa-senior → validation-specialist` on an `adversarial-v1` workflow MUST
  pass, `qa-senior → documentation-specialist` MUST fail); the same-step hop
  under UX; the spike/hotfix terminal `"complete"`; and `handoffForRole`'s
  no-workflow-state fallback.
- `internal/scaffold` parity: the `evidence-contract.md` mirror is
  `cmp`-verified byte-identical; keep it that way if the doc is touched again.

**Outstanding TODOs:** none. No commented-out code, no TODO markers, no
skipped assertions introduced.
