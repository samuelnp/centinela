### Senior-Engineer Report: adversarial-validate-verifier
**Date:** 2026-07-28

Slices 1–6 of `docs/plans/adversarial-validate-verifier.md` implemented.
Slice 7 (acceptance e2e) belongs to the tests step and was not written.

## Files Touched

| Path | Reason |
|------|--------|
| `internal/gatereport/{doc,verdict,model,block,check,findings,stamp}.go` | Slice 1 — pure parser: Status first-token verdict (D5), tagged `centinela:verification` fence (D2), admissibility `Assess`, in-place stamp rewrite |
| `internal/delivery/verdict.go` | Slice 1 — collapsed to a delegation so one parser owns the Status contract |
| `internal/treestate/{treestate,digest,runner}.go` | Slice 2 — HEAD + working-tree digest with the `.workflow/` exclusion (D3, D3a) behind a stubbable `Runner` seam |
| `internal/workflow/stamp.go` | Slice 2 — `StampVerification`: read report, snapshot, splice, atomic write |
| `cmd/centinela/artifact_stamp.go` | Slice 2 — `centinela artifact stamp <feature>` Cobra wiring only |
| `internal/workflow/{contract,validate_gatekeeper,validate_freshness}.go` | Slice 3 — pinned-contract gate; content checks in `ValidateArtifacts`, git-backed freshness in the `complete` path only (D6) |
| `internal/workflow/{state,order,validate}.go` | Slice 3 — `ValidateContract` pinned at start (D4); `validateGatekeeper` moved out |
| `cmd/centinela/{complete,complete_validate_gates}.go` | Slice 3 — freshness runs first; validate gates extracted to keep `complete.go` ≤100 lines |
| `internal/orchestration/{policy,models,validate}.go` | Slice 4 — `RoleGatekeeper` constant, validate ⇒ `[gatekeeper]`, reasoning tier (D7), `ValidateRoles` extracted |
| `internal/workflow/validate_orchestration.go` | Slice 4 — legacy contract keeps demanding validation-specialist evidence |
| `internal/evidence/{schema_init,invalidation_targets,companion_skeletons,artifact_gatekeeper}.go` | Slice 4 — handoff, invalidation targets, headers, fail-closed stub |
| `docs/architecture/gatekeeper-prompt.md` + scaffold mirror | Slice 5 — refutation-stance rewrite |
| `docs/architecture/{validation-specialist-prompt,evidence-contract}.md` + mirrors | Slice 5 — deprecation banner; gatekeeper contract section |
| `docs/architecture/workflow-enforcement.md`, `CLAUDE.md`, `internal/scaffold/assets/CLAUDE.md` | Slice 5 — validate-row and checklist wording |
| `internal/orchestration/directives.go`, `cmd/centinela/hook_orchestration.go` | Slice 6 — paths-only delegation contract printed for the validate step |
| `cmd/centinela/hook_statusline_{validate,rules}.go` | Slice 6 — validate branch extracted, four new blocks mapped |
| `internal/setup/opencode_agent_config.go` + golden | Slice 6 — gatekeeper agent added via the managed `BuildSyncPlan`/`ApplySync` path, validation-specialist retained |

## Architecture Compliance

- Boundary checks passed (n-tier): `cmd/` imports `internal/*` only;
  `internal/gatereport` imports nothing internal; `internal/treestate`
  imports nothing internal; `internal/workflow` composes both.
- G7: `artifact stamp`'s domain logic lives in `internal/workflow`, not in
  `cmd/centinela/artifact_stamp.go`, which is Cobra wiring only. Minor
  divergence from the plan's file table (see Trade-Offs).
- G1 file size: every new and modified source file ≤ 100 lines, including
  every `_test.go` under `internal/` and `cmd/`. `complete.go` was split
  (98 → 90 plus a new 32-line file) rather than granted an exception.
- Scaffold mirrors for the three parity-enforced docs copied byte-identically
  in this change; `workflow-enforcement.md` deliberately left unmirrored
  (it is on `mirrorParityAllowlist`).

## Type-Safety Notes

- `Verification` / `Command` are concrete structs with explicit JSON tags;
  no `map[string]any` reaches a caller.
- The only `map[string]json.RawMessage` use is in `Stamped`, and it exists
  precisely so the `commands` array survives byte-identically.
- `treestate.Runner` is a named func type, so the git seam is stubbable
  without an interface or reflection.
- Every failure path returns a typed `error`; no boolean fail-open exists.

## Trade-Offs

- **Verdict raw vs normalized.** The plan folds alias normalization into
  `Verdict`. Doing that would have changed the delivery composer's observed
  output (`BLOCKING` → `CRITICAL`) and broken its tests, which the plan also
  requires to pass untouched. Split into `Verdict` (raw token) plus
  `Normalize` (D5 aliases); `Assess` uses both.
- **Stamp logic placement.** The plan puts the report rewrite in
  `cmd/centinela/artifact_stamp.go` (~70 lines). That is business logic in
  the outer layer, so it lives in `internal/workflow/stamp.go` and
  `internal/gatereport/stamp.go` instead.
- **Freshness defers to admissibility.** Dogfooding showed that an
  `artifact new` stub carries an empty-but-present block, so the freshness
  check fired first and reported "stale (verified <empty>)" — masking the
  "no commands-run record" message the spec demands. `VerificationFresh`
  now returns nil whenever `gatereport.Assess` would already block.
- **validation-specialist stays an invalidation target.** No step requires
  the role any more, but a legacy in-flight workflow that rewinds past
  validate must still shed its legacy pair, or the stale pair would satisfy
  its legacy gate on the next pass.
- **`Analyzed Specs` kept as its own stub section.** It is mechanically
  derived from `specs/*.feature` and must stay deterministic, so it was not
  folded into the verifier-authored `Inputs Read`.

## Deferred Findings

- All four §9 deferrals were already recorded on the roadmap during the plan
  step: `verify-crosscheck-verifier-commands`, `typed-gate-error-codes`,
  `reuse-prior-test-run-in-complete-verify`, `mirror-workflow-enforcement-doc`.
- No new out-of-scope gap was found. None recorded.

## Handoff

- Next role: qa-senior
- Outstanding TODOs:
  - Slice 7: `tests/acceptance/adversarial_validate_verifier_test.go`, the
    binary-driven e2e (local bare origin only — never a network remote).
  - Slice 5's `tests/acceptance/adversarial_verifier_prompt_test.go`
    (refutation stance, paths-only contract, no-orchestrator-summaries
    prohibition, `CRITICAL`, the `centinela:verification` fence, the stamp
    instruction).
  - Integration coverage for `complete` blocked/advanced and for the
    legacy-vs-adversarial evidence sets.
