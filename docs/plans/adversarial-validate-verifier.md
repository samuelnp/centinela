# Plan — adversarial-validate-verifier

> Contract: `docs/features/adversarial-validate-verifier.md`. Phase 13
> "Lighter Centinela". This plan implements that brief and nothing else.

## 1. Problem framing

The `validate` step currently produces two artifacts *about* work and zero
artifacts that *re-derive* it:

- `validation-specialist` (tier `fast`) reads the gatekeeper report, reads
  the production-readiness report, narrates a synthesis. It is primed with
  the implementer's framing and it composes, it does not check.
- `gatekeeper` (tier `fast`, `docs/architecture/gatekeeper-prompt.md`) is a
  spec-conflict checklist: "read all `.feature` files, list findings". Its
  stance is compliance, not refutation.

The only mechanical teeth today are `validateGatekeeper` in
`internal/workflow/validate.go`, which does a bare `os.Stat` on
`.workflow/<feature>-gatekeeper.md`, and `centinela verify`'s claim checks.
A dead subagent that wrote a 200-byte stub passes. A report that says
"Status: SAFE" without having run anything passes. A verdict produced three
fix-rounds ago passes.

This feature converts the validate step's single remaining subagent into an
adversary whose verdict is only admissible when it is *grounded*: it records
the commands it actually executed and the exact tree state it verified, and
the `complete` gate refuses anything else.

## 2. Scope

**In (from the brief):**

1. Refutation-stance rewrite of `gatekeeper-prompt.md` + scaffold mirror.
2. Machine-readable commands-run record + verified-revision binding inside
   `.workflow/<feature>-gatekeeper.md`.
3. `complete`-gate enforcement: verdict, commands record, revision freshness.
4. `RequiredRolesForFeature(f, "validate")` drops `validation-specialist`,
   requires `gatekeeper`; legacy in-flight workflows still complete.
5. Hook directive + statusline wording for the validate step.
6. Agent-config emitter entry for the verifier.

**Out (pre-agreed in the brief):** the plan-role merge; changes to
`centinela verify` mechanics; the production-readiness gate; merge-steward;
mechanical detection of contaminated delegation.

**Out (decided here, deferred — see §9):** `centinela verify` cross-checking
the recorded argv against its own runs; typed gate error codes replacing the
statusline's string matching; `PriorTestRun` reuse in the complete path;
closing the scaffold parity allowlist.

## 3. Locked design decisions

**D1 — the role slug stays `gatekeeper`.** The brief renames the *stance*,
not the artifact. `.workflow/<feature>-gatekeeper.md` is named by AC2, and
the slug is load-bearing in ten places (`internal/memory/capture.go`,
`internal/delivery/deliver_artifacts.go`, `hook_statusline_rules.go`,
`evidence.AllRoles`, `evidence.KindGatekeeper`, `config` override keys, the
scaffold parity allowlist, three acceptance tests). Renaming buys nothing
and costs a wide, risky diff. `docs/architecture/gatekeeper-prompt.md` keeps
its filename; its *content* becomes the adversarial verifier prompt.

**D2 — the verification record lives in a fenced JSON block in the `.md`.**
Per the brief's stated preference. The block is tagged so parsing is
unambiguous and never a prose scan:

````markdown
```json centinela:verification
{
  "revision": "9f2c1ab…",
  "treeDigest": "sha256:4e7d…",
  "commands": [
    {"argv": ["centinela", "validate"], "exitCode": 0, "durationMs": 84210},
    {"argv": ["go", "test", "./..."], "exitCode": 0, "durationMs": 121004}
  ]
}
```
````

Rationale for the `.md` over the `.json` evidence companion: the report is
the artifact the brief names, it is what `centinela revise` invalidates, and
it is what a human reads. The evidence `.json` keeps its existing schema —
**no persisted schema change**, per the brief's Data Model section.

**D3 — freshness = HEAD revision AND a working-tree digest.** A HEAD-only
check is insufficient and would be theater in exactly the scenario the
feature targets. `centinela complete` auto-commits *after* the gate, so HEAD
does not move during the validate step: an operator who fixes code in place
after a CRITICAL verdict leaves HEAD identical, and a HEAD-only comparison
would happily re-admit the stale verdict. The stamp is therefore
`{revision: git rev-parse HEAD, treeDigest: sha256(dirty-state)}`.

**D3a — the digest MUST exclude `.workflow/`.** The verifier writes its own
report and evidence into the working tree, so a naive digest is
self-invalidating: whatever the verifier stamps is stale the moment it
finishes writing. `.workflow/` is the *output* of verification, never its
input. Excluding it also makes `centinela artifact stamp` idempotent and
order-independent.

**D4 — legacy back-compat is state-dated, not either-set.** The brief's edge
case asks for "legacy-complete only if written before migration". `Workflow`
already carries pinned-at-start fields with the same shape
(`EnforcementProfile`, `Archetype`, `DriverModel`: empty ⇒ pre-existing).
Add `ValidateContract string` set to `"adversarial-v1"` by
`workflow.NewWithOrder`; empty ⇒ legacy path (existence-only gate, and the
`validation-specialist` evidence pair satisfies the validate step). This is
strictly better than the file-presence "either-set" rule, which would let a
*fresh* feature dodge the new format by hand-authoring legacy-named files.
Clocks are never consulted.

**D5 — verdict vocabulary.** New canonical set: `SAFE` / `WARNING` /
`CRITICAL`. `BLOCKING` and `UNSAFE` are retained as accepted aliases that
normalize to `CRITICAL` (legacy reports, and `production-readiness`, which
still emits `BLOCKING` and is out of scope). Parsing keeps the PR #59
contract exactly: first token of the `**Status:**` line, no prose scan.
Missing or unparseable status ⇒ blocked, never fail-open.

**D6 — the freshness check runs in the `complete` path only.** It shells out
to `git`; `workflow.ValidateArtifacts` is called by `hook context` for every
active workflow on *every prompt*, and adding two `git` invocations per
workflow per prompt is an unacceptable latency tax on the host session.
Content checks (verdict + commands record) are pure file reads and stay in
`ValidateArtifacts`; the git-backed freshness check is a separate exported
function called from `runComplete`.

**D7 — the gatekeeper's tier moves `fast` → `reasoning`.** AC6 requires the
directive to name "its reasoning-tier model". Refutation is the hardest
judgment in the workflow; it also now replaces the `validation-specialist`
context entirely, so the net subagent spend still drops.

## 4. Slices

Ordering is smallest-correct-first. Each slice compiles, tests green, and is
independently revertible.

### Slice 1 — report parser (`internal/gatereport`), zero behavior change

New package, pure, no filesystem or git.

| File | Lines | Contents |
|------|-------|----------|
| `internal/gatereport/doc.go` | ~15 | package doc: the Status contract + block contract |
| `internal/gatereport/verdict.go` | ~50 | `Verdict(report string) string`; `**Status:**` line, first token, alias normalization (D5) |
| `internal/gatereport/model.go` | ~35 | `Verification{Revision, TreeDigest string; Commands []Command}`, `Command{Argv []string; ExitCode, DurationMS int}` |
| `internal/gatereport/block.go` | ~60 | `ParseVerification(report string) (Verification, error)` — locate the ```` ```json centinela:verification ```` fence, unmarshal, typed errors for absent/malformed |
| `internal/gatereport/check.go` | ~65 | `Assess(report string) error` — verdict admissible, commands non-empty, every entry well-formed (argv non-empty, exitCode present), ≥1 entry whose argv joins to `centinela validate` with `exitCode == 0`, revision + treeDigest non-empty |

Also in this slice: `internal/delivery/verdict.go` collapses to a delegation
(`func gatekeeperVerdict(r string) string { return gatereport.Verdict(r) }`,
~10 lines) so there is one parser, not two. The delivery composer's observed
behavior is unchanged for `SAFE`/`WARNING`/`BLOCKING` — asserted by the
existing delivery tests, which must keep passing untouched.

**Tests** (colocated, each ≤100 lines): `verdict_test.go` (SAFE/WARNING/
CRITICAL/BLOCKING alias/UNSAFE alias/lowercase/`Status:` without bold/prose
containing "warning" elsewhere/missing line), `block_test.go` (well-formed,
absent fence, untagged ```json fence ignored, malformed JSON, multiple
fences → first wins), `check_test.go` (empty commands, argv-less entry,
`centinela validate` non-zero exit, missing revision, happy path).

**Why first:** it is the dependency of slices 2 and 3, it is pure, and
landing it alone provably changes nothing.

### Slice 2 — tree stamp (`internal/treestate`) + `centinela artifact stamp`

| File | Lines | Contents |
|------|-------|----------|
| `internal/treestate/treestate.go` | ~55 | `Stamp(root string, run Runner) (Snapshot, error)`; `Snapshot{Revision, Digest string}` |
| `internal/treestate/digest.go` | ~55 | digest = sha256 over `git status --porcelain=v1` lines **and** `git diff HEAD` body, both filtered to drop any path under `.workflow/` (D3a); stable ordering |
| `internal/treestate/runner.go` | ~25 | `Runner func(dir string, args ...string) (string, error)` seam + `NewExecRunner()` |
| `cmd/centinela/artifact_stamp.go` | ~70 | `centinela artifact stamp <feature>`: read the gatekeeper report, compute the snapshot, rewrite `revision`/`treeDigest` inside the verification block (creating the block when absent), leave `commands` byte-identical, write atomically |

`artifact stamp` is a sibling of the existing `artifact new` (both operate on
`.workflow/<feature>-<kind>` stubs), so no new top-level command surface.

**Tests:** `treestate/digest_test.go` + `treestate/treestate_test.go` with a
stubbed `Runner` (clean tree, dirty tree, `.workflow/`-only dirt must produce
the *same* digest as clean — the D3a regression test, and the single most
important assertion in this slice); `cmd/centinela/artifact_stamp_test.go`
(block created, block updated in place, `commands` preserved verbatim,
missing report → clear error).

### Slice 3 — the gate (`internal/workflow`)

`internal/workflow/validate.go` is 91 lines; the new logic must not land in
it. Extract:

| File | Lines | Contents |
|------|-------|----------|
| `internal/workflow/validate_gatekeeper.go` | ~70 | `validateGatekeeper(feature)`: legacy contract ⇒ existence only (today's behavior verbatim); `adversarial-v1` ⇒ existence + `gatereport.Assess`, error echoes the verdict reason |
| `internal/workflow/validate_freshness.go` | ~55 | `VerificationFresh(feature, root string, run treestate.Runner) error` — parse block, compare `revision`/`treeDigest` to a fresh `treestate.Stamp`, distinct error for each mismatch |
| `internal/workflow/state.go` | +6 | `ValidateContract string \`json:"validateContract,omitempty"\`` with the same back-compat comment style as `Archetype` |
| `internal/workflow/contract.go` | ~30 | `func (wf *Workflow) UsesAdversarialVerifier() bool`; `const ValidateContractAdversarial = "adversarial-v1"` |
| `internal/workflow/order.go` or wherever `NewWithOrder` lives | +2 | pin the contract at start |
| `internal/workflow/validate.go` | −7 | `validateGatekeeper` body moves out |
| `cmd/centinela/complete.go` | +4 | in the `current == "validate"` branch, call `workflow.VerificationFresh` before `executeValidation()` (cheapest check first, and it is the one that tells the operator to re-verify before they pay for a full suite run) |

Failure messages must be actionable and name the remedy, in the house style
already used by `checkPRStatus`:

- `gatekeeper verdict: CRITICAL — <first finding line>. Re-verify with a FRESH verifier context after fixing.`
- `gatekeeper report has no commands-run record — a narrated verdict is not evidence. Re-run the verifier (see docs/architecture/gatekeeper-prompt.md).`
- `gatekeeper verification is stale (verified <sha7>, tree is now <sha7>). Spawn a FRESH verifier; the previous verdict cannot certify a changed tree.`

**Tests:** `internal/workflow/validate_gatekeeper_test.go` (legacy contract +
old-format report ⇒ pass; adversarial contract + old-format report ⇒ block;
CRITICAL ⇒ block with reason; WARNING ⇒ pass; missing Status ⇒ block),
`validate_freshness_test.go` (matching stamp, revision skew, digest skew,
absent block), each ≤100 lines, split across more files if needed.

### Slice 4 — orchestration policy + evidence wiring

| File | Change | Lines |
|------|--------|-------|
| `internal/orchestration/policy.go` | add `RoleGatekeeper Role = "gatekeeper"`; `RequiredRoles("validate")` → `[]Role{RoleGatekeeper}`; keep `RoleValidationSpec` (legacy evidence + config keys still reference it) | net +4 |
| `internal/orchestration/models.go` | `RoleGatekeeper: TierReasoning` (D7); it is no longer "out-of-band" — update the comment; `RoleValidationSpec` keeps `TierFast` for legacy | ±6 |
| `internal/orchestration/validate.go` | extract `ValidateRoles(feature, step string, roles []Role, uiPaths []string) error`; `ValidateStep` delegates with `RequiredRolesForFeature` | +10 |
| `internal/workflow/validate_orchestration.go` | choose the role set from the pinned contract: adversarial ⇒ `RequiredRolesForFeature`; legacy ⇒ `[]Role{RoleValidationSpec}`; call `ValidateRoles` | +12 |
| `internal/evidence/schema_init.go` | `handoffForRole(gatekeeper)` → `documentation-specialist` (today it falls through to `"complete"`) | +2 |
| `internal/evidence/invalidation_targets.go` | drop the now-duplicated `gatekeeper` append for `validate` (it arrives via `RequiredRolesForFeature`); keep `production-readiness` | −1 |
| `internal/evidence/companion_skeletons.go` | gatekeeper headers → `{"Inputs Read", "Refutation Attempts", "Commands Run", "Findings", "Recommendation"}` | ±1 |
| `internal/evidence/artifact_gatekeeper.go` | stub body: `SAFE \| WARNING \| CRITICAL` vocabulary, the four new sections, and an **empty** `centinela:verification` block | ±20 |

The stub's `commands` array ships **empty on purpose**: `artifact new` must
produce a report that *fails* the gate until a verifier fills it. Fail-closed
is the whole point; an `artifact new` stub that passes would reintroduce the
dead-subagent hole.

Consequence to state explicitly in the plan and re-check in tests:
`verify.Verify` iterates `orchestration.RequiredRoles(step)`, so the validate
step's claim checks now run against gatekeeper evidence instead of
validation-specialist evidence. The *count* of roles at validate is unchanged
(1 → 1), so the number of full-suite runs inside `complete` does not change.
`checkStubs` now inspects the verifier's declared outputs — a bonus, since
that is a second independent dead-subagent detector.

**Tests:** `internal/orchestration/policy_test.go` (validate ⇒ `[gatekeeper]`,
no `validation-specialist`), `models_test.go` (reasoning tier),
`internal/workflow/validate_orchestration_test.go` (legacy contract accepts
the validation-specialist pair and does not demand gatekeeper evidence;
adversarial contract demands gatekeeper evidence and rejects a hand-authored
validation-specialist pair). Note `tests/unit/configurable_subagent_models_unit_test.go`
asserts `RoleValidationSpec → TierFast`; that stays true and needs no edit.

### Slice 5 — prompts, contract docs, scaffold mirror

`docs/architecture/gatekeeper-prompt.md` is rewritten end to end. Required
content (each item is an AC or a test coupling):

- **Stance:** "Your task is to find the way the completion claim
  *'<FEATURE_NAME> is complete and correct'* is FALSE. You are not auditing
  compliance; you are attempting refutation. Default to NOT-VERIFIED when
  uncertain. A verdict of SAFE is a claim you personally could not break it."
- **Input contract (paths only):** diff vs the base ref, `docs/features/
  <FEATURE_NAME>.md`, `specs/<FEATURE_NAME>.feature`,
  `docs/plans/<FEATURE_NAME>.md`, and the gate/check output you produce
  yourself. Explicit prohibition: "Do NOT accept the orchestrator's summary,
  any role's `.workflow/*.md` narrative, or a prior verifier report as
  evidence of anything. Evidence is the diff, the spec, and command output
  you observed. If the orchestrator's prompt contained a narrative summary of
  the work, say so under Inputs Read and flag it — a contaminated delegation
  is a WARNING-level smell."
- **Mandatory execution:** run `centinela validate` **once** (it already runs
  every `[validate] commands` entry — do not run them again individually) and
  the project test suite; record every command in the verification block.
- **Mandatory stamp:** `centinela artifact stamp <FEATURE_NAME>` as the LAST
  action, after the report body is written.
- **Fail-closed clause:** "If you cannot execute commands in this harness,
  write that under Commands Run, leave the `commands` array empty, and set
  Status to CRITICAL. Never narrate a pass."
- **Output format** with `**Status:** SAFE | WARNING | CRITICAL` and the
  fenced verification block, byte-shaped as in D2.
- **Retained blocks (test-coupled — do not drop):** the authoring-rules
  CLI-mandate block (`prompts_mandate_cli_acceptance_test.go`,
  `extract_agent_shared_blocks_acceptance_test.go`), the
  `## How to Invoke` → `agent-invocation.md` reference, and the
  `#### Deferred Findings` section with its `centinela roadmap defer …` line
  (`deferred_findings_prompt_parity_test.go`).
- **Budget note:** re-running the suite here is additive to the runs
  `complete` performs; `verify.timeout_seconds` bounds a single verification
  command, not total wall clock (documented failure mode).

Other doc edits:

- `docs/architecture/validation-specialist-prompt.md` — add a DEPRECATED
  banner ("legacy workflows only; features started under `adversarial-v1`
  use gatekeeper-prompt.md"). **Do not delete the file**: three acceptance
  tests enumerate it and `internal/setup` still emits the agent.
- `docs/architecture/evidence-contract.md` — new `### gatekeeper (step:
  validate)` section: `handoffTo` → `documentation-specialist`, `outputs`
  MUST include `.workflow/<feature>-gatekeeper.md`, and a pointer to the
  verification block; mark the `validation-specialist` section legacy.
- `docs/architecture/workflow-enforcement.md` line ~145 — validate row gains
  "…with a non-empty commands-run record and a current verified revision".
- `CLAUDE.md` — the Gatekeeper Subagent section and the quick-reference row
  rename to "Validate agent — adversarial verifier"; the Gate Keepers
  Checklist row becomes "Verifier verdict: SAFE or WARNING (CRITICAL blocks)".
- **Mirror every changed `docs/architecture/*.md` into
  `internal/scaffold/assets/docs/architecture/` byte-identically in the same
  commit.** `gatekeeper-prompt.md`, `validation-specialist-prompt.md` and
  `evidence-contract.md` are all parity-enforced.
  `workflow-enforcement.md` is on `mirrorParityAllowlist` (not mirrored at
  all) — its mirror does not exist and must not be created here; see §9.

**Tests:** a new `tests/acceptance/adversarial_verifier_prompt_test.go`
asserting the prompt carries the refutation stance, the paths-only input
contract, the no-orchestrator-summaries prohibition, the `CRITICAL` token,
the `centinela:verification` fence, and the stamp instruction. Existing
parity tests cover the mirror.

### Slice 6 — directives, statusline, agent-config emitters

| File | Change | Lines |
|------|--------|-------|
| `internal/orchestration/policy.go` (or a new `directives.go`) | `DelegationContract(step) string` → for `validate`: "pass ONLY the feature slug and file paths; do NOT summarize the implementation into the verifier's prompt"; `""` elsewhere | ~20 |
| `cmd/centinela/hook_orchestration.go` | print the contract line when non-empty (AC6: role name + reasoning tier already come from `annotateRoles`) | +3 |
| `cmd/centinela/hook_statusline_validate.go` (new) | extract the validate branch out of `hook_statusline_rules.go` (65 lines today, would exceed 100); map the new blocks: `MISSING_GATEKEEPER`→`run-verifier`, `VERDICT_CRITICAL`→`fix-and-reverify`, `MISSING_COMMANDS_RECORD`→`rerun-verifier`, `STALE_VERIFICATION`→`reverify-fresh-context` | ~55 |
| `cmd/centinela/hook_statusline_rules.go` | delegate the validate branch | −8 |
| `internal/setup/opencode_agent_config.go` | **add** a `gatekeeper` agent (adversarial description + prompt); **keep** `validation-specialist` (`adapt_opencode_support_test.go` asserts it, and legacy workflows still use it) | +5 |

`internal/setup` changes go through `BuildSyncPlan`/`ApplySync` — never a
legacy `Ensure*` writer — or existing projects report perpetual pending
drift (known regression class). `internal/config/orchestration_models.go`
already allows the `gatekeeper` override key; no change needed.

**Tests:** `internal/orchestration/directives_test.go`,
`cmd/centinela/hook_statusline_validate_test.go`,
`internal/setup/opencode_agent_config_test.go` (gatekeeper entry present,
validation-specialist retained, task permission allowed).

### Slice 7 — spec + acceptance wiring (tests step, listed for completeness)

`specs/adversarial-validate-verifier.feature` (authored by the
feature-specialist in this same plan step) drives:
`tests/acceptance/adversarial_validate_verifier_test.go` — a binary-driven
end-to-end: init a temp repo with a **local bare origin** (never a network
remote — a real push hangs `go test` for hours and times out claim
verification), start a feature, fast-forward to validate, write a stub
report, assert `complete` blocks; fill the commands record and stamp, assert
`complete` advances; mutate a tracked file, assert `complete` blocks as
stale; mutate only `.workflow/`, assert it still advances (D3a).

## 5. File-by-file summary

New source files (all ≤100 lines): `internal/gatereport/{doc,verdict,model,
block,check}.go`, `internal/treestate/{treestate,digest,runner}.go`,
`cmd/centinela/artifact_stamp.go`, `internal/workflow/{validate_gatekeeper,
validate_freshness,contract}.go`, `cmd/centinela/hook_statusline_validate.go`,
`internal/orchestration/directives.go`.

Modified source files: `internal/delivery/verdict.go`,
`internal/workflow/{validate,state,validate_orchestration}.go` and the
`NewWithOrder` site, `internal/orchestration/{policy,models,validate}.go`,
`internal/evidence/{schema_init,invalidation_targets,companion_skeletons,
artifact_gatekeeper}.go`, `cmd/centinela/{complete,hook_orchestration,
hook_statusline_rules}.go`, `internal/setup/opencode_agent_config.go`.

Modified docs: `docs/architecture/{gatekeeper-prompt,
validation-specialist-prompt,evidence-contract,workflow-enforcement}.md`,
`CLAUDE.md`, and the byte-identical `internal/scaffold/assets/…` mirrors of
the three parity-enforced prompt/contract docs.

Every new `_test.go` under `internal/` and `cmd/` is ≤100 lines (G1 applies
to test files; diff-aware `validate` hides violations that the CI full scan
catches). Coverage is per-package with no `-coverpkg`, so every new package
carries **colocated** tests; `tests/` tier files do not move the 95% gate.
Target ≥97% on `internal/gatereport`, `internal/treestate`,
`internal/workflow`, `internal/orchestration`, `cmd/centinela`.

## 6. Test strategy

| Slice | Unit (colocated) | Integration | Acceptance |
|-------|------------------|-------------|------------|
| 1 | verdict/block/check tables | — | — |
| 2 | stubbed-Runner digest + stamp rewrite | `artifact stamp` on a real temp git repo | — |
| 3 | gate matrix × contract | `complete` blocked/advanced via `ValidateArtifacts` | e2e in slice 7 |
| 4 | policy/tier/handoff | legacy vs adversarial evidence sets | e2e in slice 7 |
| 5 | — | — | prompt-content + scaffold parity |
| 6 | directive string, statusline mapping, emitter | managed-sync idempotency | opencode/codex config parity |

The single highest-value assertion in the whole feature: **a report whose
only change since verification is under `.workflow/` is still fresh, and a
report whose tracked source changed is not.** Both directions must be tested;
only asserting the blocking direction leaves D3a's self-invalidation bug
undetected until a real run deadlocks.

Second-highest: **`centinela artifact new <f> gatekeeper` followed
immediately by `centinela complete <f>` must FAIL.** That is the
dead-subagent regression in one line.

## 7. Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Verifier theater — refutation prompt still rubber-stamps | High | Medium | Commands record + revision/digest are enforced mechanically in the gate, not by convention; missing/unparseable ⇒ blocked, never fail-open |
| Stale-verdict hole via uncommitted fixes (HEAD unchanged) | High | High | D3 tree digest, not HEAD alone; `complete` auto-commits only *after* the gate, so HEAD-only would be near-useless |
| Digest self-invalidation (report write dirties the tree) | High | Certain if naive | D3a `.workflow/` exclusion, with an explicit both-directions test |
| Legacy in-flight workflow bricked mid-validate | Medium | Medium | D4 contract pinned at start; empty ⇒ today's behavior verbatim; existing `complete_cmd_test.go` / `complete_branches_test.go` fixtures exercise the legacy path unchanged |
| Prompt-doc test coupling (3 acceptance tests enumerate `validation-specialist-prompt.md`) | Medium | High | Keep the file + its CLI-mandate and Deferred-Findings blocks; deprecate in prose only |
| Scaffold mirror drift | Medium | High | Edit source + mirror in the same commit; parity test is a required gate; `workflow-enforcement.md` is unmirrored by allowlist — do not "fix" it here |
| Wall-clock blowout: verifier suite run + `complete`'s runs exceed the job cap | Medium | Medium | Prompt mandates `centinela validate` ONCE (it already runs `[validate] commands`); do not add a redundant `go test ./tests/acceptance/...` to `validate.commands`; run long `complete`s backgrounded and poll |
| `git` latency in the prompt hook | Low | Medium | D6 — freshness check lives in the `complete` path only; `ValidateArtifacts` stays pure file I/O |
| Blocking loop — over-zealous CRITICAL on a healthy feature | Medium | Low | WARNING semantics preserved (advances, finding lands in the memory ledger); `centinela revise` is the escape valve; no skip flag exists, by design |
| Coverage dip below the 95% gate on touched packages | Low | Medium | Colocated `_test.go` ≤100 lines each; aim ≥97% so a parallel merge cannot tip main red |
| Parallel-merge semantic conflict with `unified-plan-specialist` | Low | Medium | Only overlap is the CLAUDE.md quick-reference table and `opencode_agent_config.go` map — textual, trivial; run the real gate on the merged tree before merging second |

## 8. Rollout

1. **Slice 1** — parser package + delivery delegation. Provably inert.
2. **Slice 2** — `treestate` + `artifact stamp`. New command, nothing gates
   on it yet. Dogfood by building a scratch binary from `./cmd/centinela`
   (the installed binary lags the worktree).
3. **Slice 3** — gate, behind the pinned `ValidateContract`. In-flight
   features are untouched; only features started after this lands are gated.
4. **Slice 4** — policy + evidence wiring. First slice that changes what the
   validate step *requires*.
5. **Slice 5** — prompt rewrite + mirrors + contract docs. Must be in the
   same branch as slices 3–4 so the first feature started under
   `adversarial-v1` has a prompt that produces a conforming report.
6. **Slice 6** — directives, statusline, emitters. Cosmetic-adjacent; last.
7. **Slice 7** — spec-driven acceptance e2e.

This feature dogfoods itself: it is started under the *old* contract, so its
own validate step runs the legacy path. That is correct and intended — but it
means the new gate is never exercised by this feature's own `complete`. The
acceptance test in slice 7 is therefore not optional; it is the only place
the new gate is proven end to end before the next feature depends on it.

## 9. Deferred (recorded on the roadmap)

- `verify-crosscheck-verifier-commands` — have `centinela verify` compare the
  verifier's recorded argv/exit codes against its own claim-verification runs
  (the brief lists this as an *optional* integration point; v1 enforces the
  record's existence and shape, not its truthfulness).
- `typed-gate-error-codes` — `hook_statusline_rules.go` classifies validate
  failures by substring-matching the error message, so any non-BLOCKING error
  is misreported as `MISSING_PROD_READINESS`. This slice adds cases rather
  than fixing the pattern; typed error codes are the real fix.
- `reuse-prior-test-run-in-complete-verify` — `cmd/centinela/complete_verify.go`
  never sets `verify.Deps.PriorTestRun`, so `complete` re-runs the suite even
  though `executeValidation()` just ran it. The seam exists and is unused;
  wiring it is the cheapest available cut to validate-step wall clock, which
  this feature makes more valuable.
- `mirror-workflow-enforcement-doc` — `workflow-enforcement.md` is on
  `mirrorParityAllowlist` as "not mirrored"; scaffolded projects get no copy,
  and edits to it (including this feature's) silently never reach them.
