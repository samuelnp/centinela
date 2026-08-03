# Edge Cases: binding-evidence-gates

Adversarial edge-case pass over the three newly-bound gates. Every row marked
**PROBED** was executed against a scratch binary (`go build -o /tmp/centinela-ec2
./cmd/centinela`) or a probe module linked directly against the worktree's
`internal/...` packages, on throwaway temp fixtures — the real repo's
`.workflow/` was never written to during probing. Rows marked *reasoned* are
static analysis only.

Two of the three gates have a **verified bypass that lands in exactly the class
the feature exists to close**, and both are one keystroke wide.

---

## Risk Matrix

| # | Case | Impact | Likelihood | Why |
|---|---|---|---|---|
| **EC-1** | **PROBED — `"exitCode": null` fully defeats the new commands-schema check.** `{"argv":["centinela","validate"],"exitCode":null}` passes `ValidateCommandsSchema` (present-key check satisfied; `json.Unmarshal("null", &int)` is a documented no-op returning `nil`), passes `Stamped`, passes `ParseVerification`, decodes to `Command{ExitCode: 0}`, and `Assess` returns `nil` — i.e. **grounded**. Confirmed at the CLI too: `centinela artifact stamp` accepted it and spliced it back verbatim alongside a real revision/treeDigest. | **Critical** — the gate's own comment says "an absent key would silently read as a pass"; an explicit `null` is that exact defect with four more characters. A verifier whose `centinela validate` exited non-zero can record the run as grounded. | Medium — not something an honest verifier types, but it is the first thing a hand-shaped or LLM-generated block produces when the exit code is unknown, and it is indistinguishable from an honest record on review. | `validateCommandEntry` tests key *presence*, then type-decodes. Go's JSON `null` satisfies both. Fix: reject `null` explicitly (`if string(rawExit) == "null"`), or decode into `*int` and require non-nil. Same hole exists on `durationMs` (harmless). |
| **EC-2** | **PROBED — the changelog stub check only inspects the FIRST non-blank line.** `## Changelog\n\n- <FILL: type>: <FILL: one-line summary of the change>` → **ACCEPT**. So does `Changelog:\n- <FILL: …>`. The literal `centinela artifact new <f> changelog` output is correctly rejected only because it happens to be a single line. | **High** — defect 3 is "an unfilled changelog template passes the docs gate", and it still does behind any one line of preamble. | Medium-High — adding a markdown heading to a report file is one of the most natural things an agent does, and doing so re-opens the gate without touching the stub. | `validateChangelog` returns `nil` on the first non-blank line that has no marker. The comment ("later lines are free prose") justifies not scanning *the rest*; it does not justify letting the first line *be* prose. Fix: scan for the first line that looks like an entry (e.g. leading `-`/`*`), or reject a marker anywhere in the first N lines / anywhere in the file with a documented escape. |
| **EC-3** | **PROBED — in-flight user-facing workflows at the `code` step newly break, and the tolerance deliberately does not cover them.** Old prefill seeded `senior-engineer → "qa-senior"` unconditionally. Derived value on a user-facing feature is `"ux-ui-specialist"` (same-step hop ⇒ `acceptsHandoff` returns `false` before the tolerance is consulted). Verified REJECT. Same shape at validate on an **internal** feature: old prefill `"documentation-specialist"`, derived `"complete"` (terminal ⇒ tolerance also disabled) — REJECT. | **High** for the affected workflow (`centinela complete` blocked), zero for correctness. Self-repairable: the error string carries the exact `centinela evidence set …` fix. | Medium — bites any parallel worktree (e.g. `docstring-gate`) sitting at the `code` step of a user-facing feature, or at `validate` on an internal one. | The plan's §6 breakage inventory covers only three *test fixtures* and the risk table anticipates only the validate/plan **contract-pin** cases. The user-facing `code`-step case is neither in the inventory nor in the risk table, and the tolerance was scoped to next-step hops so it structurally cannot cover it. Not a defect in the tolerance — a gap in the migration story. |
| **EC-4** | **PROBED — the tolerance itself holds up under attack: it cannot be used to skip a step.** Next-step hop into `validate` on an `adversarial-v1` workflow accepts exactly `{gatekeeper, validation-specialist}`. Rejected: `qa-senior`, `documentation-specialist`, `planner`, `senior-engineer`, `ux-ui-specialist`, `big-thinker`, `feature-specialist`, `merge-steward`, `production-readiness`, `complete`, `user`, `banana`, `GATEKEEPER`, `Gatekeeper`, `" gatekeeper"`, `"gatekeeper "`, `"gatekeeper\n"`, `"  "`, and a 2000-char value. | Low (no bypass found) | — | `alternateContractRoles` is keyed on the **same** `nextStep` that produced `want`, and both branches return only that step's occupants. `default: return nil` means every other step (`code`, `tests`, `docs`) admits no alternate at all — verified: a `code`-step hop into `tests` accepts `qa-senior` and nothing else. |
| **EC-5** | **PROBED — flipping the contract pin cannot WIDEN the accepted handoff set.** Under `adversarial-v1`: `want=gatekeeper`, `alt=validation-specialist`. Under the legacy pin: `want=validation-specialist`, `alt=gatekeeper`. The accepted **union is identical**; the pin only swaps which name is canonical. Same symmetry for `plan`. | Low | — | Answers the "can a feature flip which pin it uses to widen acceptance" question directly: **no**, not for the handoff gate. (It *does* widen elsewhere — see EC-6.) |
| **EC-6** | **PROBED — the derivation's trust root is agent-writable, and one blanked field disarms every check this feature adds.** `orchestrationMode: ""` (or deleting `.workflow/<feature>.json` outright) makes `strictOrchestrationEnabled` false ⇒ `validateOrchestration` returns `nil` ⇒ the entire handoff chain check is skipped. Separately, `validateContract: ""` makes `validateGatekeeper` early-return ⇒ a narrated "I looked at it, trust me" gatekeeper report with **no verification block at all** passes the validate gate (verified `nil`). | **High** as a bound on the word "binding"; **not this feature's regression** — both predate it. | Low as an honest accident; the point is that "the workflow's own contract" is a file the agent under test can edit. | Mitigating nuance (verified with `git check-ignore`): `.workflow/<feature>.json` **is tracked**, so a pin edit shows up in review. But `.workflow/*-<role>.json` **is gitignored**, so a hand-edited `handoffTo` never appears in a diff. Deferred as `workflow-state-file-disarms-its-own-gates`. |
| **EC-7** | **PROBED — write-side / pre-flight asymmetry for defect 1.** `centinela evidence set demo senior-engineer handoffTo banana` succeeds silently; `centinela evidence validate demo` then prints `evidence ok for "demo"`; `centinela complete demo` rejects. Defect 2 got both a write-side check (`Stamped`) and an equally strict read-side one — verified: hand-editing the block *after* stamping to drop `exitCode` is caught on read. Defect 1 got read-side only. | Medium — the pre-flight command the role prompts tell agents to run gives a green light the gate then contradicts, which is the "false success" signal the framework exists to eliminate, one layer up. | Medium-High — any role agent that self-checks with `evidence validate` before handing off. | `runEvidenceValidate` → `evidence.ValidateFeature` → `orchestration.ValidateEvidence` per role; the chain check lives in `internal/workflow` and is only reached via `ValidateArtifacts`. Deferred as `evidence-validate-skips-handoff-chain`. |
| **EC-8** | **PROBED — a legitimate changelog that quotes the marker is a false positive.** `- fix: reject an unreplaced <FILL: ...> marker` → REJECT. | Low | **Concrete and immediate** — that is the single most natural changelog line for *this* feature, so binding-evidence-gates is odds-on to hit its own gate at its own docs step. | The plan (§4 risk table) rates this "very low / implausible" on the grounds that "changelog lines are one-line prose summaries, not documentation about templating". That reasoning does not hold for the feature that *is* about templating. Practical mitigation: write the entry without the literal marker. |
| **EC-9** | **PROBED — substance floor is still just "non-empty".** After the marker is gone, `-` alone and `x` alone both ACCEPT. | Low | Low | Removing the marker is sufficient; no conventional-commit shape or minimum length is required. Defensible for v1, but it should be a *named* residual rather than an unexamined one. Deferred as `changelog-substance-floor-is-non-empty`. |
| **EC-10** | **PROBED — marker detection is exact-substring and case-sensitive.** `<fill:`, `< FILL:` and `<FILL type>` all ACCEPT. | Low | Very low | The scaffold only ever emits `<FILL:`, so this bites only a hand-retyped marker. Deferred as `fill-marker-detection-is-exact-substring`. |
| **EC-11** | **PROBED — commands-schema residuals (all correct or by design).** ACCEPT: `[]`, `null`, absent key (all deferred to `Assess`'s admissibility rule, per the doc comment), extra unknown keys, duplicate identical entries, `argv: [""]`. REJECT: `commands` as object/string/number, entries as `null`/array/string, `argv` absent/`[]`/`null`/string/non-string-elem/nested, `exitCode` as string/bool/`1.5`/`0.0`/`1e2`/bignum, `durationMs` as string. | Low | — | `argv: [""]` cannot satisfy `hasPassingValidate` (`filepath.Base("")` is `"."`). Duplicate JSON keys within one entry resolve **last-wins in both** the raw `map[string]json.RawMessage` and the typed `Command`, so there is no schema-vs-decode divergence to exploit — checked explicitly. |
| **EC-12** | **PROBED — `ExpectedHandoff` is silently total for inputs that should have no answer.** An unknown step returns `("complete", true)`; an empty/missing `stepOrder` falls back to `DefaultStepOrder`, so a hotfix workflow that lost its order derives `plan`/`docs` successors its archetype never had. Also: `handoffForRole` now seeds the non-required `production-readiness` role from the *next step* (`documentation-specialist` on a user-facing feature) where it used to seed `complete`. | Low | Low | `ok` is `false` only when `Load` fails. Cosmetic today (`production-readiness` is never chain-checked), but it is an undocumented behavior change from `legacyHandoffForRole` and a silent-default the feature's own thesis argues against. |
| **EC-13** | **PROBED — merge-steward literal check is correct and strict.** `complete`/`user` accept; `orchestrator`, `COMPLETE`, `" user"`, arbitrary reject with the two-verdict message; `""` is caught upstream by `incomplete evidence fields`. | Low | — | Self-contained in `validateStewardHandoff`, no workflow context, no tolerance. Clean. |
| **EC-14** | *Reasoned* — past steps are never re-validated. The chain runs only over `RequiredEvidenceRoles(feature, currentStep)`; once a step completes, its `handoffTo` can be rewritten and nothing re-reads it. | Low | Low | Consistent with `validateOrchestration`'s long-standing current-step-only scope and with the brief's "no retrofitting". Worth stating plainly: the chain is proven **pairwise at completion time**, never end-to-end. |
| **EC-15** | *Reasoned* — whitespace-only `handoffTo` (`"  "`) slips past `ValidateEvidence`'s `== ""` test but is **PROBED**-rejected by the chain check, which quotes the raw value in the error. Values are never trimmed, so `" gatekeeper"` also fails. Fail-closed in both directions. | Low | Low | Correct behavior; noted only because the two checks own adjacent halves of the same field and neither normalizes. |

---

## Missing or Weak Scenarios

Measured against `specs/binding-evidence-gates.feature` and the current tree
(`internal/workflow/handoff.go`, `handoff_chain.go` and
`internal/gatereport/commands_schema.go` have **no test files at all** yet — the
tests step owns all of them).

1. **No scenario for a `null` exit code (EC-1).** The spec's defect-2 block
   covers "missing `exitCode`" and "empty `argv`" — the two cases the code
   already handles. The one shape that *defeats* the check is absent from the
   spec, so a green suite would certify the hole.
2. **No scenario for a changelog stub that is not on line 1 (EC-2).** The
   defect-3 block tests exactly the single-line scaffold and exactly one filled
   line. Every real-world shape in between is unspecified.
3. **No scenario asserting the tolerance's boundary.** The spec has the
   *positive* legacy case ("A valid mid-chain handoff on a legacy workflow") but
   nothing pinning what the tolerance must **refuse** — no same-step case, no
   terminal case, no other-step role. Without those, a future widening of
   `alternateContractRoles` would not fail a single test.
4. **No pin-flip symmetry assertion.** Nothing locks in the property that makes
   the tolerance safe: that the accepted set for a next-step hop is identical
   under both contract pins, so flipping the pin cannot widen it.
5. **No migration scenario for the old prefill on a user-facing feature
   (EC-3).** The spec's "same-step handoff when UX is required" tests the *new*
   correct value; nothing states what must happen to evidence carrying the old
   `qa-senior` value. That decision (accept, or reject with a remedy) is
   currently made implicitly by code with no test.
6. **No read/write parity assertion for the commands schema.** The design claim
   is "one shape, one place, shared by `Stamped` and `ParseVerification`".
   Nothing tests that the two agree; they could drift the moment either grows a
   caller-specific branch.
7. **No test that the CLI's own scaffold fails the CLI's own gate.**
   `changelogBody()` and `validateChangelog` must stay in lockstep; the existing
   `artifact_changelog_test.go` still only asserts the stub is non-blank.
8. **No archetype-subset coverage of the derivation.** hotfix (`code,tests,
   validate`), spike (`plan,code`) and `BootstrapStepOrder` (`plan,code,validate,
   docs`) each exercise a different branch of `nextChainStep`; none is
   specified. Bootstrap order in particular is the only shipped order in which
   `code`'s successor step is `validate`, i.e. the only one where the tolerance
   is reachable from the code step.
9. **No coverage of the `evidence validate` / `complete` disagreement (EC-7).**
10. **No negative test for `HasFillMarker`'s aliasing.** `evidence.FillMarker`
    is now a re-export; nothing asserts the two constants stay equal, which is
    the only thing keeping the renderer and the detector on the same string.

---

## Proposed / Added Tests

Priority order: **P0 tests must go red on the current tree.**

### Unit — `internal/gatereport`

- **P0 `TestValidateCommandsSchemaRejectsNullExitCode`** — table over
  `exitCode: null`, absent, `"0"`, `true`, `1.5`, `0.0`, `1e2`, bignum.
  `null` must reject. *(red today)*
- **P0 `TestAssessRejectsNullExitCodeRecord`** — an end-to-end report whose only
  command is `{"argv":["centinela","validate"],"exitCode":null}` must be
  inadmissible. Guards the actual laundering path, not just the shape rule.
  *(red today)*
- **P1 `TestCommandsSchemaReadWriteParity`** — for each malformed raw array,
  assert `Stamped(...)` and `ParseVerification(...)` return the *same* wrapped
  error text. Locks the "one shape, one place" claim.
- **P1 `TestValidateCommandsSchemaTable`** — the full accept/reject matrix from
  EC-11, including duplicate keys within an entry (last-wins in both decoders)
  and `argv: [""]`.
- **P2 `TestValidateCommandsSchemaDefersAdmissibility`** — `[]`, `null` and an
  absent key must pass the *schema* and be caught by `Assess`, pinning the
  deliberate division of labour documented in the file header.

### Unit — `internal/workflow` (handoff)

- **P0 `TestExpectedHandoffDerivationMatrix`** — table: canonical
  internal/user-facing × every step; hotfix; spike; `BootstrapStepOrder`; legacy
  validate pin; legacy plan pin; unknown step; empty `stepOrder`; missing state
  file (`ok == false`). One table replaces the whole §3a prose argument.
- **P0 `TestHandoffToleranceIsStepScoped`** — a next-step hop into `validate`
  accepts exactly `{derived, alternate}`; assert rejection of `qa-senior`,
  `documentation-specialist`, `planner`, `complete`, `user`, `merge-steward`,
  case variants, whitespace-padded and a 2000-char value.
- **P0 `TestHandoffToleranceDisabledForSameStepAndTerminal`** — user-facing
  `senior-engineer → qa-senior` must reject (successor is `ux-ui-specialist`);
  internal `gatekeeper → documentation-specialist` must reject (successor is
  `complete`). These two are the tolerance's stated boundary and the EC-3
  migration cases.
- **P1 `TestContractPinFlipDoesNotWidenAcceptance`** — compute the accepted set
  for a next-step hop into `validate` under both pins and assert set equality.
  The single test that makes the tolerance auditable.
- **P1 `TestHandoffChainSilentOnUnreadableEvidence`** — missing / unparseable /
  empty-`handoffTo` evidence must produce no chain error (the no-double-report
  contract), while `ValidateRoles` still reports it.
- **P2 `TestHandoffErrorNamesExecutableRemedy`** — assert the error contains a
  literally runnable `centinela evidence set <feature> <role> handoffTo <want>`.

### Unit — `internal/workflow` (changelog) and `internal/orchestration`

- **P0 `TestValidateChangelogRejectsStubBehindPreamble`** — `## Changelog\n\n` +
  stub, and `Changelog:\n` + stub, must reject. *(red today)*
- **P0 `TestScaffoldedChangelogFailsItsOwnGate`** — feed `changelogBody(feature)`
  straight into `validateChangelog` and require rejection; then the same content
  with the slots replaced and require acceptance. Keeps renderer and gate in
  lockstep forever.
- **P1 `TestValidateChangelogTable`** — CRLF, CR-only, BOM + stub, no trailing
  newline, empty file, whitespace-only file, blank lines before the stub,
  filled-first-line-with-later-stub (documents the accepted behavior),
  marker-in-prose (documents the accepted false positive), bare `-`.
- **P1 `TestStewardHandoffLiterals`** — `complete`/`user` accept;
  `orchestrator`, `COMPLETE`, `" user"` reject with the two-verdict message.
- **P2 `TestFillMarkerAliasParity`** — `evidence.FillMarker ==
  orchestration.FillMarker` and `evidence.FillSlot(x) ==
  orchestration.FillSlot(x)`.

### Integration — `cmd/centinela`

- **P0 `TestArtifactStampRejectsNullExitCode`** — real temp git repo, block with
  `exitCode: null`, `centinela artifact stamp` must fail at write time. *(red
  today; the absent-key sibling already passes)*
- **P1 `TestStampedBlockStaysBindingAfterHandEdit`** — stamp a well-formed
  report, hand-edit the block to drop `exitCode`, assert the validate-step gate
  rejects it. Pins read-side strictness as equal to write-side.
- **P1 `TestCompleteRejectsOutOfChainHandoffAndRemedyFixesIt`** — `complete`
  fails on `handoffTo: banana`, then run the exact `evidence set` command from
  the error message and assert `complete` proceeds. Makes the error message
  itself a tested contract.
- **P1 `TestEvidenceValidateAgreesWithComplete`** — `centinela evidence validate`
  must not report `evidence ok` for a feature `complete` will reject on the
  chain. *(red today — pairs with the `evidence-validate-skips-handoff-chain`
  deferral; if that stays deferred, invert this into a documented-divergence
  test so the gap is at least pinned.)*
- **P2 `TestArchetypeSubsetsCompleteWithTerminalHandoff`** — spike completes at
  `code` with `complete`; hotfix completes at `validate` with `complete`.

### Acceptance — `tests/acceptance`

Drive the installed/worktree binary; keep the `// Acceptance:` + `// Scenario:`
comment convention the spec header mandates.

- **A1** the three existing spec blocks, unchanged.
- **A2 (new spec scenario)** *A null exit code is rejected at stamp time* —
  add to the defect-2 block in `specs/binding-evidence-gates.feature`.
- **A3 (new spec scenario)** *A stub behind a heading fails the docs gate* —
  add to the defect-3 block.
- **A4 (new spec scenario)** *The tolerance refuses a role from another step* —
  the negative twin of the existing "valid mid-chain handoff on a legacy
  workflow" scenario.
- **A5 (new spec scenario)** *Evidence seeded by the old prefill on a
  user-facing feature* — pins whatever EC-3 is resolved to, so the migration
  behavior is stated rather than emergent.

Every new acceptance file must run under the existing acceptance runner already
listed in `validate.commands`, and must use a **local bare repo** if it needs an
`origin` — a network push in an acceptance test hangs the suite and turns
claim verification into a TIMEOUT.

---

## Residual Risks

- **The gate is only as trustworthy as `.workflow/<feature>.json`** (EC-6). The
  derivation's premise — "the workflow's own contract" — is a file the agent
  being gated can rewrite, and one blanked field turns the whole strict
  orchestration path into a no-op. Deferred, not fixed here, but the feature's
  claim should be stated as *"handoffTo is bound to the declared contract"*
  rather than *"handoffTo cannot be forged"*.
- **The role evidence files are gitignored**, so a hand-edited `handoffTo` is
  invisible in review; only the state file's contract pins are tracked. Reviewer
  attention cannot be relied on as a second layer here.
- **The chain is pairwise, never end-to-end** (EC-14). A completed step's
  evidence can be rewritten afterwards with no re-validation. Out of scope per
  the brief's "no retrofitting", but it means "the chain is enforced" is true
  only at each completion instant.
- **`hasPassingValidate` still proves nothing about the world.** A well-formed,
  schema-valid `{"argv":["centinela","validate"],"exitCode":0}` written by hand
  is indistinguishable from an honest one. The stamp binds *shape*, and the
  revision/treeDigest bind *tree state* — neither binds *"this command actually
  ran and exited 0"*. Correctly out of scope (the plan says so explicitly), but
  it is the ceiling on defect 2's value and should not be overclaimed in the
  changelog.
- **Changelog substance floor** (EC-9) and **marker spelling** (EC-10) are both
  deliberate v1 narrowings; both deferred with slugs so they are recoverable.
- **The false positive on a changelog quoting the marker** (EC-8) is accepted,
  and this feature is the most likely thing ever to trip it. Write the docs-step
  changelog entry without the literal `<FILL:` string.
- **Broadening stub detection to all companion reports** stays deferred per plan
  §3d; this report is itself an example of a legitimate document that discusses
  the marker at length and would be a false positive under a blanket rule.

---

## Deferred Findings

Registered via `centinela roadmap defer <slug> --summary "…" --source
binding-evidence-gates/edge-case-tester`:

- `workflow-state-file-disarms-its-own-gates` — `orchestrationMode: ""` or a
  deleted `.workflow/<feature>.json` silently disables the whole
  strict-orchestration gate, and `validateContract: ""` disables the
  grounded-verdict gate; the agent-writable state file is the trust root for
  both.
- `evidence-validate-skips-handoff-chain` — `centinela evidence validate` and
  `evidence set` both accept an out-of-chain `handoffTo` that `centinela
  complete` rejects, so the agents' own pre-flight command disagrees with the
  gate.
- `changelog-substance-floor-is-non-empty` — once the marker is gone,
  `validateChangelog` accepts a bare `-` or `x` as an entry; no
  conventional-commit shape or minimum substance is required.
- `fill-marker-detection-is-exact-substring` — `HasFillMarker` matches only the
  literal `<FILL:`, so `<fill:`, `< FILL:` and `<FILL type>` pass, while a
  legitimate changelog quoting the marker in prose is a false positive.

**Not deferred — these three belong to this feature and should be fixed before
it ships:** EC-1 (`null` exit code), EC-2 (stub behind a preamble line), and a
stated decision on EC-3 (old prefill on user-facing `code`-step evidence).
