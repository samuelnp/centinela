# durable-workflow-state — qa-senior

**Step:** tests · **Next role:** gatekeeper · **Mode:** FIX + TEST (edge-case
findings R1/R2/R3/R4/R6/R7/R15 acted on; every claim below was executed)

The edge-case tester's R1 defeated the feature's central design rationale, so
this step is not only tests: `Load` now probes the schema version from the raw
bytes before it unmarshals, and the resulting contract is written down in the
code. R3 (silent resurrection) and R15 (the weaker duplicate atomic write) are
fixed; R4's overclaimed doc comment is corrected to the measured behaviour;
R6/R7 are decisions on record with tests behind them.

---

## Fixes Landed

| # | Fix | Files |
|---|-----|-------|
| **R1** | `Load` probes `schemaVersion` from raw bytes BEFORE the full unmarshal; a future/unreadable version whose body cannot be modelled returns a **degraded** workflow instead of an error | `version_probe.go` (new), `state_future.go` (new), `state_io.go`, `state.go` |
| **R1** | `EvaluatePrewrite` ALLOWS every governed write against an unmodellable workflow — it cannot enforce step rules whose schema it does not understand | `internal/hookpolicy/prewrite.go` |
| **R2** | `Save`'s version guard fails **closed**: a `schemaVersion` present but unreadable is refused (`errUnreadableVersion`), never assumed to be 1 | `schema_version.go`, `state_io.go` |
| **R3** | A missing target is a first write only for a **never-loaded** workflow; for a loaded one it is a conflict (`checkNotDeleted`) | `state_cas.go`, `state_io.go` |
| **R4** | `checkNotStale`'s doc comment now states what the mechanism actually guarantees — staleness, not concurrency; "microseconds" replaced by the measured millisecond window and the 24-writer numbers | `state_cas.go` |
| **R6/R7** | Symlink and umask semantics decided and documented at the constant/helper that implements them | `atomic_write.go`, `state_io.go` |
| **R15** | `stamp.go`'s hand-rolled atomic write deleted; `writeReport` delegates to `WriteFileAtomic` at 0644 (reports were landing 0600) | `stamp.go` |

### R1 — the new `Load` contract (the important one)

`stateVersion(raw) (int, bool)` reads a single raw token and is independent of
`Workflow`:

- absent / `null` / `0` → `(1, true)`: a legacy file.
- an integer → `(n, true)`; `n > SchemaVersion` is a future file.
- present but unreadable as an int (`"2.0"`, `2.0`, `1.5`, `2e2`, `true`,
  `99999999999999999999`) → `(0, false)`. Treated **exactly like a future
  version**: this binary demonstrably does not understand the file.
- not a JSON object at all → `(1, true)`. Corruption is not evidence of the
  future, and diagnosing it is not this function's job.

`Load` then behaves as follows, and this is the contract:

1. Version ≤ ours and the body unmarshals → unchanged.
2. Version future/unreadable and the body unmarshals (a purely ADDITIVE future
   version) → unchanged: full struct, normal step gating, `Save` refuses.
3. Version future/unreadable and the body does **not** unmarshal (a future
   version that changed a TYPE or a SHAPE) → **`Load` succeeds** with a degraded
   workflow: `Unmodellable() == true`, `Feature`/`CurrentStep` salvaged when
   still plain strings (else the file's own name and `UnknownStep`),
   `SchemaVersion` = the probed version or 0. `Save` still refuses.
4. Version NOT future and the body does not unmarshal → genuine corruption,
   still an error (unchanged; detection is the `doctor-detect-corrupt-state-file`
   deferral).

`ActiveWorkflows` keeps case 3 (it is a real workflow with a real feature name),
so the active set is never emptied, `EvaluatePrewrite` never falls to `NeedInit`
and `hook_autostart` never forks a duplicate. `EvaluatePrewrite` allows the
write rather than blocking on a step it cannot read. **A v99 file — well-typed
or not — cannot block `hook prewrite`.**

Boundary, stated rather than discovered: the degraded path requires the document
to carry one of `feature`, `currentStep`, `steps`, `stepOrder`. That is what
stops a future `roadmap.json` schema from becoming a phantom workflow that
`doctor` would then offer to `rm` (pinned by
`TestFutureVersionInNonStateJSONIsNotAWorkflow`). A future version that renames
all four keys is not recognisable as workflow state by this binary at all.

### R6/R7 — decisions

**Mode ignores umask; the explicit 0644 stays.** `f.Chmod` is not filtered by
umask where `os.WriteFile`'s perm argument was, so under `umask 077` the old
writer produced 0600 and any hook or subprocess running under another uid (CI
images, shared checkouts) could not read the state file. The file's readability
is a property of Centinela, not of the operator's shell; the 0644 literal always
stated that intent and umask silently overrode it. The file holds feature names,
steps and model routes — coordination metadata, never secrets. A 0400 state file
is likewise normalised to 0644, which the atomic replace requires anyway
(`rename(2)` needs write permission on the DIRECTORY). Pinned by
`TestSaveNormalisesModeRegardlessOfUmask` and `TestSaveNormalisesAReadOnlyStateFile`.

**A symlinked state FILE is replaced, not followed — accepted.** Writing through
the link would re-open the `O_TRUNC` torn-write window this feature exists to
close, and the link's target may sit on another filesystem where `rename` fails
with EXDEV and atomicity is unavailable at all. The supported way to share state
between checkouts is symlinking the `.workflow/` **directory**, which still works
because the temp lands in the resolved directory. Both halves pinned
(`TestSymlinkedStateFileIsReplacedNotFollowed`,
`TestSymlinkedStateDirectoryIsWrittenThrough`) and documented on `WriteFileAtomic`.

---

## Test Inventory by Tier

### Colocated — `internal/workflow` (this is what moves the coverage gate)

| File | Tests | Pins |
|------|-------|------|
| `state_future_test.go` | `TestFutureVersionNeverBricksLoad` (9-case table over type AND shape), `TestActiveWorkflowsKeepsFutureVersionFile` (same table) | R1 — Load and the active-set invariant directly, not through the hook |
| `state_future_guard_test.go` | `TestFutureVersionInNonStateJSONIsNotAWorkflow`, `TestFutureVersionEvidenceJSONKeepsItsFeature`, `TestUnreadableVersionIsRefusedNotAssumedLegacy`, `TestCorruptFileIsStillAParseError`, `TestStateVersionReadsEveryTokenShape` (12 tokens) | R1 boundary, R2 |
| `state_delete_conflict_test.go` | `TestDeleteBetweenLoadAndSaveIsAConflict`, `TestNeverLoadedWorkflowStillCreatesAMissingFile`, `TestDeleteAfterSaveIsAlsoAConflict` | R3 + the `start`/autostart exemption |
| `state_mode_unix_test.go` | `TestSaveNormalisesModeRegardlessOfUmask`, `TestSaveNormalisesAReadOnlyStateFile` | R7 (both halves) |
| `state_symlink_test.go` | `TestSymlinkedStateFileIsReplacedNotFollowed`, `TestSymlinkedStateDirectoryIsWrittenThrough` | R6 |
| `state_version_lock_test.go` | `TestWorkflowStructFieldsAreVersionLocked` | the process guard: an equal-version Save drops unmodelled keys, so a forgotten bump is a silent data-loss release |
| `stamp_write_test.go` (+1) | `TestWriteReportKeepsTheReportReadable` | R15 |
| `schema_version_test.go` (edited) | `TestStateVersionToleratesGarbage` (was `TestOnDiskVersionToleratesGarbage`) | corruption ≠ future version |

### Colocated — `internal/hookpolicy`

`prewrite_future_test.go`: `TestUnmodellableWorkflowNeverBlocksAWrite` (fixture
sits on `plan`, where a code file is normally BLOCKED, so the allow can only
come from the new rule) and `TestModellableFutureWorkflowKeepsStepGating` (an
additive future file is still governed — the fix does not silently disable
enforcement).

### Unit — `tests/unit/`

`durable_workflow_state_unit_test.go`: `TestSchemaVersionRules` (5-row table:
absent/null/zero/equal/higher, load + save + byte-identity), `TestSaveStampsVersionFirst`.
`durable_workflow_state_future_test.go`: `TestFutureVersionStaysInTheActiveSet`,
`TestFutureVersionSaveIsRefusedWithAnActionableMessage`,
`TestDeletedWorkflowIsNotResurrected`.

### Integration — `tests/integration/`

- `TestDoctorStaysGreenWithAWorkflowTemp` — asserted on **`doctor.Run`'s own
  diagnoses**, not on `filepath.Glob`: the evidence check must be OK with the
  exact message, no check may name the temp in its details, and `evidence.Repair`
  must not remove it (reaping a live write's temp mid-flight would corrupt the
  very write this feature made atomic).
- `TestSerialisedConcurrentWritersLoseNothing` — 12 concurrent `Load`s,
  serialised saves: exactly 1 ok, 11 refused, and the on-disk route count equals
  the successes. The honest, executable form of the CAS guarantee.
- `TestDoctorIsStillBlindToACorruptStateFile` — an **executable statement of the
  `doctor-detect-corrupt-state-file` deferral**, not an endorsement. It fails
  (with a message naming the deferral and telling the reader to delete it) the
  moment someone teaches `doctor` to report a corrupt state file.

### Acceptance — `tests/acceptance/` (binary built from `./cmd/centinela`, local fixtures only)

| Test | Scenario |
|------|----------|
| `TestAccFutureVersionSaveIsRefused` | A future-version state file is refused on save with an actionable message |
| `TestAccFutureVersionDoesNotBlockWrites` | A future-version state file does not block file writes |
| `TestAccUnmodellableFutureVersionDoesNotBlockWrites` | (same scenario, the R1 case: a v99 this binary cannot unmarshal) |
| `TestAccLegacyVersionlessFileLoadsAndIsStamped` | A versionless legacy workflow file loads unchanged |
| `TestAccCompletedWriteLeavesNoTemp` | A completed write replaces the state file in one step |
| `TestAccReplacedStateFileKeepsItsMode` | The replaced state file keeps its readable file mode |
| `TestAccUnwritableStateDirReportsThePath` | A write that cannot be completed reports the state file path |
| `TestAccSameVersionRoundTripKeepsEveryField` | A same-version file round-trips without losing fields |
| `TestAccStartStampsTheSchemaVersion` | A newly started workflow is stamped with the current schema version |
| `TestAccStartNeverLoadsAndIsFencedByExistence` | A workflow that was never loaded saves without a conflict check |
| `TestAccDoctorIgnoresAnAbandonedWorkflowTemp` | An abandoned temporary file is not mistaken for orphaned evidence |

---

## Acceptance Wiring

Every acceptance file carries `// Acceptance: specs/durable-workflow-state.feature`
and a standalone `// Scenario: <name>` matching the spec verbatim (checked by
diffing the spec's scenario list against the mapped names: 10 of 13 covered,
0 mismatched). The binary is built once per suite via the existing `dmrBuildBin`
`sync.Once` helper into a temp dir — never the installed binary, which predates
the schema version; fixtures are `t.TempDir()` projects; nothing touches the
network.

**Three scenarios have no acceptance mapping, deliberately** (the gate is
`severity = "warn"`, and all three are proven colocated):

- *A killed write leaves the previous state intact* — needs the `beforeRename`
  package seam plus a re-exec'd child that calls `os.Exit`. The shipped binary
  exposes no crash seam; `state_io_crash_test.go` is the proof.
- *A stale save is refused rather than silently overwriting a newer one* and
  *Re-running a refused command after the conflict succeeds* — need a second
  writer to land **inside** one command's load→save window. Probed: `complete`
  does not run `[validate] commands` at the code step, so there is no
  CLI-reachable pause between a load and its save; every other command loads and
  saves within milliseconds. `state_cas_test.go` (two `Load`s, one process) is
  the shape that actually reproduces it. Filed below as
  `acceptance-seam-for-crash-and-stale-save`.

---

## Regression Guards — red→green evidence

Every fix was reverted in place, the tests re-run, and the fix restored. No test
survived the revert of the defect it claims to detect.

| Fix reverted | Went red |
|--------------|----------|
| `Load`'s version-probe branch (back to unmarshal-then-fail) | `TestFutureVersionNeverBricksLoad` (**9/9 subtests**), `TestActiveWorkflowsKeepsFutureVersionFile` (**9/9**), `tests/unit TestFutureVersionStaysInTheActiveSet` (4/4), `internal/hookpolicy TestUnmodellableWorkflowNeverBlocksAWrite`, `tests/acceptance TestAccUnmodellableFutureVersionDoesNotBlockWrites` |
| `EvaluatePrewrite`'s `Unmodellable` clause only | `TestUnmodellableWorkflowNeverBlocksAWrite`; `TestAccUnmodellableFutureVersionDoesNotBlockWrites` — the binary printed **`HOOK BLOCKED WRITE`**, which is the bricking itself |
| `Save`'s `errUnreadableVersion` (back to fail-open) | `TestUnreadableVersionIsRefusedNotAssumedLegacy`; plus the save half of `TestFutureVersionNeverBricksLoad` for all 6 unreadable-version rows |
| `checkNotDeleted` call | `TestDeleteBetweenLoadAndSaveIsAConflict`, `TestDeleteAfterSaveIsAlsoAConflict`, `tests/unit TestDeletedWorkflowIsNotResurrected` |
| `writeReport` back to the hand-rolled temp | `TestWriteReportKeepsTheReportReadable` — `mode = -rw-------` |
| `f.Chmod(perm)` in `writeTempSibling` | `TestSaveNormalisesModeRegardlessOfUmask` (`-rw-------` under umask 077), `TestSaveNormalisesAReadOnlyStateFile` |
| Added `NewFieldNobodyBumpedFor` to `Workflow` | `TestWorkflowStructFieldsAreVersionLocked`, with the golden diff and the "bump SchemaVersion" instruction |

Not one of the new tests passed with its fix reverted. R4 is a doc-comment
correction with no code change; its executable counterpart is
`TestSerialisedConcurrentWritersLoseNothing`, which states the guarantee that
does hold (1 ok / N−1 refused / 0 lost) rather than the one that does not.

---

## Coverage Gaps

| Package | Coverage | Note |
|---------|----------|------|
| `internal/workflow` | **97.8%** | up from 97.2%; gate is 95% |
| `internal/hookpolicy` | **99.0%** | |
| `internal/doctor` | 96.0% | untouched by this step |

Two uncovered blocks remain in the touched code, both deliberate:

1. `writeTempSibling`'s coalesced write/chmod/sync/close failure path — no
   portable trigger; contorting production code to reach it is worse than the
   uncovered line (the senior-engineer's judgement, re-confirmed).
2. `Save`'s `json.MarshalIndent` error — `Workflow` contains no unmarshalable
   type, so it is unreachable without a fake. Pre-existing.

Everything else added this step is at 100%.

---

## Deferred Findings — NOT filed

`centinela roadmap defer` was **not** run: from inside this worktree the record
lands on the feature branch. **Orchestrator: please file these from the primary
checkout.** Items 1–6 carry forward from the edge-case report; 7 and 8 from that
report are now **FIXED IN THIS STEP** and are listed only to close them out; 9
and 10 are new.

1. `harden-createtemp-mode-regression` — `internal/roadmap/rawio.go` and
   `internal/brownmap/write.go` still replace a 0644 file with a 0600 one via
   `os.CreateTemp`. (`internal/workflow/stamp.go`, the third site, was fixed here
   — same package, no layering question.)
2. `doctor-repair-non-role-json-tmp` — `orphanedTmps()` globs `*.json.tmp` but
   `evidence.Repair` globs `<feature>-*.json.tmp`, so a non-role temp makes the
   check red forever with a repair that removes nothing. This feature's naming
   sidesteps it; the trap remains.
3. `doctor-check-future-schema-version` — the operator only learns at the point
   of refusal; `status`, `status-all` and `verdict` all render a v99 file with no
   warning, and `verdict`'s JSON carries no state-file schema version.
4. `autostart-duplicate-on-unloadable-state` — carried forward from the
   edge-case tester's probe. **Verified this step that a future-version file no
   longer reaches it** (`hook autostart` over a v99 `"2.0"` file started nothing
   — R1's fix). My own scratch-repo probe with a genuinely TRUNCATED file also
   started nothing, so the trigger conditions are narrower than first reported;
   the finding stands but its reproducer needs re-deriving.
5. `atomic-writes-for-roadmap-and-brownmap` — both still hand-roll it;
   `WriteFileAtomic` makes the follow-up cheap once the layer rule is settled.
6. `close-state-save-toctou` — plan slice 4 (flock across re-read → rename).
   Measured, not estimated: 23/24 updates lost under fully-overlapping saves,
   0/24 when saves are serialised. Now stated in `checkNotStale`'s doc comment
   instead of the old "microseconds" claim.
7. ~~`load-probes-version-before-unmarshal`~~ — **fixed this step** (R1).
8. ~~`save-refuses-on-deleted-target-when-loaded`~~ — **fixed this step** (R3).
9. `doctor-detect-corrupt-state-file` *(carried, now with an executable marker)*
   — `doctor` still prints `✓ workflow-state no orphaned .workflow state` and
   `0 error` for a truncated state file. `TestDoctorIsStillBlindToACorruptStateFile`
   fails the day this is fixed and names the slug.
10. `acceptance-seam-for-crash-and-stale-save` *(new)* — three spec scenarios
    (killed write, stale save refused, re-run after conflict) cannot be driven
    through the built binary: there is no crash seam and no CLI-reachable pause
    between a command's load and its save. A hidden test-only env seam (or a
    `--simulate-crash-before-rename` debug flag) would let the acceptance tier
    prove them end-to-end instead of only in-process.
11. `preflight-schema-version-in-complete` *(carried)* — `runComplete` refuses
    only after `runValidateGates`, so a future-version file costs a full gate run
    before an instant-to-compute refusal.
12. `acceptance-tier-exceeds-default-test-timeout` *(known)* — see Suite Result.

---

## Suite Result

```
go test ./... -coverprofile=coverage.out            → EXIT 0
  48 packages ok · 0 FAIL · 0 panic
  tests/acceptance 425.5s · tests/integration 11.5s · tests/unit (cached)
  internal/workflow 97.8% · internal/hookpolicy 99.0%

COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh → EXIT 0
  coverage gate passed: 97.3% >= 95.0%
```

**No `test timed out after 10m0s` flake** in any run this step — the known
`acceptance-tier-exceeds-default-test-timeout` deferral did not fire; the
acceptance tier landed at 425–457s.

**Honest note on the run count.** Three full profiled runs happened, not one,
and only because the test suite kept growing under me: run 1 (green, EXIT 0,
acceptance 453.9s) preceded the last three test files; run 2 (green, EXIT 0,
acceptance 457.3s) preceded one comment-level edit that removed a no-op mutex
from `TestSerialisedConcurrentWritersLoseNothing`; run 3 above is the
authoritative one and covers the tree exactly as handed off. Runs 2 and 3 reused
Go's per-package cache for everything unchanged, so the extra cost was the
acceptance tier only.

Also run: `go vet ./...` (clean), `gofmt -l internal cmd tests` (clean),
`centinela docs lint` → `✓ docstring-gate All 13 exported identifiers across 9
changed Go file(s) are documented`. Every file touched or added is ≤100 lines,
`_test.go` included, verified by a repo-wide scan of `internal/` and `cmd/`.
`centinela validate`'s remaining gates were NOT run here — that is the
gatekeeper's, and re-running them would have meant a fourth full suite.

---

## Handoff

**Next role: gatekeeper.**

- **No spec change was needed or made.** `specs/durable-workflow-state.feature`
  is untouched; every fix here strengthens an existing scenario rather than
  changing one. The one contract the spec does not state — what `Load` returns
  for a future file it cannot parse — is documented in `state_future.go` and
  `schema_version.go`, and enforced by `TestFutureVersionNeverBricksLoad`.
- **New production behaviour to review**, one line each: `Load` may now return a
  workflow for a file it cannot unmarshal (only when that file claims a future
  version AND looks like state); `EvaluatePrewrite` allows writes against such a
  workflow; `Save` refuses an unreadable version and refuses to recreate a
  deleted target for a loaded workflow; `writeReport` now fsyncs and chmods.
- **The three deferrals worth your attention** are 6 (`close-state-save-toctou`
  — the residual TOCTOU is now documented honestly rather than understated), 9
  (`doctor-detect-corrupt-state-file` — this feature prevents new corruption and
  ships no detection for old corruption) and 10 (the missing acceptance seam for
  the three unmapped scenarios).
- **If you write a test that induces a save failure, chmod the DIRECTORY, not
  the file** — a read-only state file is now writable and normalised to 0644, by
  decision.
