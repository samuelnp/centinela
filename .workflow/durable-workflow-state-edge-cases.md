# Edge Cases: durable-workflow-state

**Step:** tests · **Role:** edge-case-tester · **Method:** probed, not reasoned.

Every row marked **[probe]** was executed: package-internal probes against a
throwaway copy of this worktree (`internal/workflow/zzprobe_*_test.go`,
discarded), and CLI probes driving `go build -o /tmp/centinela-ec4
./cmd/centinela` inside `mktemp -d` scratch repos. The real `.workflow/` was
never written to during probing. Rows marked **[reasoned]** are code-trace only
and say so.

Regression baseline re-run at the start of this step:
`go test ./internal/workflow/... ./internal/doctor/... ./internal/evidence/...
./cmd/centinela/... -count=1` → **all four packages green**.

---

## Risk Matrix

| # | Case | Impact | Likelihood | Why |
|---|------|--------|-----------|-----|
| R1 | **A future binary that changes the TYPE of any field — including `schemaVersion` itself — bricks this binary completely.** `Load` unmarshals the whole `Workflow` struct *before* anything consults the version. **[probe]** `"schemaVersion":"2.0"`, `2.0`, `1.5`, `2e2`, `true`, `99999999999999999999` all make `Load` fail → `ActiveWorkflows` returns **0** → `EvaluatePrewrite` → `NeedInit` → every governed write blocked + `hook autostart` starts a duplicate. A well-typed `v99` that merely *reshapes* `steps.status` to an array does the same. | **Critical** | Low now, rises with every future schema change | The whole stated rationale for refusing in `Save` and never in `Load` is "a Load failure cascades". That cascade is still fully reachable; the real guarantee is *"a future version may only ADD fields, and may never change one"* — a constraint written nowhere and enforced by nothing. |
| R2 | **`onDiskVersion` fails OPEN.** Unparseable bytes and an unparseable version both return `legacyVersion = 1`, so `Save`'s future-version guard passes. **[probe]** A completely corrupt file reports version 1. Today `Load` fails first on those inputs, so the only reachable exploit is the never-loaded path (`start` / `hook autostart`), which is fenced by an `os.Stat` upstream — but the fence is in `cmd/`, not in `Save`. | High | Low | Fail-open is the wrong default for a guard whose entire job is refusing to write what it cannot understand. The safe reading of "I could not parse the version of a file that exists" is *refuse*, not *assume 1*. |
| R3 | **Delete-then-save silently resurrects.** **[probe]** `Load("alpha")` → `rm .workflow/alpha.json` → `Save` returns **nil** and recreates the file. `Save` treats `ErrNotExist` as "first write" even for a struct that carries a `loadedDigest`. | High | Medium | An operator abandoning a workflow, a `git clean -xfd`, a stale-worktree cleanup, or a `git checkout` during a minutes-long `complete` is silently undone. One-line fix: a non-empty `loadedDigest` plus a missing file **is** a conflict. |
| R4 | **The residual TOCTOU is not "microseconds" and loses updates at scale.** **[probe]** Using the `beforeRename` seam to land B's entire `Save` inside A's pre-rename window: both saves return `nil`, B's update is gone. Under 24 fully-overlapping writers with distinct mutations: **reportedSuccess=24, refused=0, silently lost=23** (two of three runs; the third: 10 ok / 14 refused / 9 lost). With saves serialised — the realistic multi-process shape — detection is **perfect**: 1 ok / 23 refused / **0 lost**. | High (bounded) | Low in practice | CAS detects *staleness*, not *concurrency*. It fully closes the minutes-wide read-modify-write window this feature was built for and provides ~zero protection against a genuine write-write race. The `checkNotStale` doc comment calls the window "microseconds"; the measured window (marshal + CreateTemp + write + **fsync** + rename) is milliseconds and empirically wide enough to lose 23/24. Correct the claim, not necessarily the code. |
| R5 | **A refusal costs a whole gate run.** `runComplete` = `Load` → `runValidateGates` (minutes) → `Complete` → `saveWorkflow`. **[reasoned — complete.go:37-61]** Both the future-version refusal and the CAS refusal fire *after* the expensive work, and the remedy ("re-run this command") means paying for it twice. | Medium | Medium | A pre-flight `onDiskVersion` check at the top of `runComplete` is nearly free and turns a 10-minute failure into an instant one. CAS cannot be pre-flighted (that is the point); the *version* can. |
| R6 | **A symlinked state file is clobbered, not followed.** **[probe]** `.workflow/alpha.json` as a symlink to `../elsewhere.json`: `Load` reads through it, `Save` **replaces the symlink with a regular file** and the destination keeps the stale content. The old `os.WriteFile` wrote *through* the link. | Medium | Low | Genuine, undocumented semantic change from `rename(2)`. Anyone sharing state between checkouts via a symlink gets a silent fork of the data. (A symlinked `.workflow/` **directory** is fine — probed, works.) |
| R7 | **Mode normalises to 0644 unconditionally, ignoring umask.** **[probe]** Under `umask 077` the state file is created `-rw-r--r--`; old `os.WriteFile(path, data, 0644)` would have produced `0600`. **[probe]** A deliberately `0400` state file is silently made writable and rewritten to `0644`. | Low–Medium | Medium | The flagged change is real and is *two* changes, not one: read-only files are now writable (correct — `rename` needs directory write), **and** `f.Chmod` bypasses umask, so on a shared box the file is now group/world-readable where the operator's umask said otherwise. `0644` is required for hooks, so this is defensible — but it must be stated in the docs step, not discovered. |
| R8 | **A corrupt state file is invisible to `doctor`, which reports a false all-clear.** **[probe]** With a truncated `.workflow/delta.json` (exactly what the pre-atomic writer left behind) `centinela doctor` prints `✓ workflow-state  no orphaned .workflow state` and `5 ok, 2 warn, 0 error`. The only trace is a bare `workflow warning:` line on stderr belonging to no check and affecting no exit code. | Medium | Medium (files already corrupted by the old writer exist) | This feature prevents *new* corruption and ships no detection or repair for existing corruption — while `doctor` actively asserts the opposite. |
| R9 | **`hook autostart` still forks a duplicate workflow off a corrupt file.** **[probe]** Truncated `delta.json` as the only state file → `workflow warning: … unexpected end of JSON input` → `CENTINELA DIRECTIVE: auto-started workflow "for-user-login"`; `.workflow/` then holds both. | Medium | Low | Pre-existing (senior-engineer deferred #4), reproduced live here. Worth restating because R1 makes it reachable from a *future-version* file too, not only a corrupt one. |
| R10 | **The digest is bound to the bytes, not to the path.** **[probe]** `wf, _ := Load("alpha"); wf.Feature = "beta"; Save(wf)` → refused, but with the **wrong diagnosis**: *"beta.json changed on disk since this command read it — another centinela process wrote it"*. Nothing changed it; we simply never read it. If the target does **not** exist, the same cross-file save **succeeds silently**. | Low | Very low (no live caller mutates `.Feature`) | Latent. Storing `loadedPath` beside `loadedDigest` and treating a path change as never-loaded (or as an error) makes the guard honest. |
| R11 | **Leaked temps are invisible and immortal.** **[probe]** With `.workflow/.delta.json.tmp-1234567890` present, `doctor` reports `✓ evidence  no orphaned evidence temp files`, no check mentions it, and it survives `doctor` untouched. It is dot-prefixed, so `ls` hides it too. **[probe]** `git check-ignore` confirms `.gitignore:62 .workflow/.*.json.tmp-*` covers it, and `internal/treestate` excludes `.workflow/` from verifier digests. | Low | Low | The isolation is correct and verified. The cost is that nothing ever notices or reaps them; they can only accumulate. |
| R12 | **The `.gitignore` rule is hard-coded to `.json` targets.** `WriteFileAtomic` derives its temp from `filepath.Base(path)`, so a `.md` target yields `.<name>.md.tmp-N`, which line 62 does **not** match. | Low | Medium (the moment anyone reuses the helper) | The helper is exported and advertised for reuse; the ignore rule silently does not follow it. |
| R13 | **CAS refuses on a byte-difference, not a meaning-difference.** **[probe]** Appending a single `\n` externally is enough to refuse the save. | Low | Low | Correct and deliberate (fields this binary cannot model must count), but it means a reformat, an editor's trailing newline, or a `git checkout .workflow/x.json` mid-command produces a "another centinela process wrote it" message naming a cause that did not happen. |
| R14 | **`start`/`hook autostart` are exempt from CAS and would clobber if their upstream `os.Stat` ever moved.** **[probe]** At package level a never-loaded `Save` over an existing `validate`-step file succeeds and resets it to `plan`, with no error. **[probe]** At CLI level both callers are fenced: `start omega` twice → `Error: workflow for "omega" already exists`; `hook autostart` picks `<feature>-2`. | Medium | Very low today | The invariant that makes the exemption safe lives in `cmd/`, is duplicated in two places, and is itself a stat-then-write TOCTOU. Nothing in `internal/workflow` enforces it, and no test pins it. |
| R15 | **`internal/workflow/stamp.go` still hand-rolls a *weaker* atomic write, one file away from the new helper.** `writeReport` uses `os.CreateTemp(".stamp-*")` with **no chmod** (so the gatekeeper report lands at `0600`), **no fsync**, and **no directory fsync**. | Medium | Certain (every `artifact stamp`) | Deferred findings #1/#5 correctly park the *cross-package* conversions on a layering question. This one has no layering question at all — same package, same file cap, three lines. Leaving it makes the package self-inconsistent about the exact defect the feature exists to fix. |

### Confirmed as designed (probed, no defect)

- **Anti-bricking holds for a well-typed future version.** **[probe]**
  `schemaVersion: 99` plus an unmodellable field: `status` renders, `status-all`
  renders, `hook prewrite` on a governed write **exits 0** with no "no workflow
  started" — and `route set` refuses with the full message, leaving the file
  **byte-identical** and no temp behind.
- **The refusal message is actionable and complete.** It names the file, the
  on-disk version, this binary's version, and `centinela update` — and
  **[probe]** `centinela update` really exists (`--check` is read-only).
- **Version matrix.** **[probe]** absent / `0` / `null` → 1, load ok, save ok.
  `1` → load, save. `-1` → loads, saves, silently normalised to 1. `2`, `99` →
  load, save refused.
- **`Save`'s guard read fails CLOSED on a read error.** **[probe]** A
  `0000`-mode target and a target that is a **directory** both abort at the
  guard read (`permission denied` / `is a directory`) with the target named and
  **no temp left behind**.
- **Failure paths leave the target untouched.** **[probe]** unwritable
  `.workflow/` (`0500`) → error, target byte-identical; missing `.workflow/` →
  error; refused save (CAS or version) → original byte-identical, `.workflow/`
  containing only the state file.
- **Crash seam.** **[probe]** the shipped `TestKilledWriteLeavesPreviousStateIntact`
  (re-exec'd child, `os.Exit` inside `beforeRename`) passes; 32-way concurrent
  saves leave a file that still parses and **zero** temp leftovers.
- **CAS core.** **[probe]** two `Load`s → two `Save`s: second refused, first
  writer's step survives. External modification → refused. `Load`→`Save`→`Save`
  → succeeds. A refused struct **keeps** refusing on retry (correct); re-`Load`
  plus save merges both updates.
- **End-to-end regression.** **[probe]** `start omega` → `route set omega
  qa-senior balanced --reason …` → `status` → `start omega` (refused):
  `schemaVersion` is the first key, mode `-rw-r--r--`, `modelRoutes` survives
  the round-trip, `.workflow/` clean of temps.

---

## Missing or Weak Scenarios

The 13 spec scenarios cover the happy paths and the three headline failures
well. What no scenario asks:

1. **No scenario constrains what a future schema version is ALLOWED to change.**
   *"A future-version state file does not block file writes"* pins the guarantee
   for one hand-picked shape. R1 shows the guarantee evaporates the moment a
   future version changes a type rather than adding a field. The migration
   contract needs that constraint written down and a test that fails when it is
   broken.
2. **No scenario covers a file whose version cannot be parsed** (R2) — the
   fail-open branch of `onDiskVersion` is reachable and untested as a *policy*
   decision (it is covered as a line, not as a behaviour).
3. **No scenario covers deletion during the read-modify-write window** (R3).
   The concurrency scenarios all assume the competing writer *wrote*; none
   assumes it *removed*.
4. **No scenario bounds the CAS guarantee.** *"A stale save is refused"* proves
   detection in the serialised case. Nothing states that overlapping saves are
   **not** protected (R4), so a reader of the spec will over-trust it.
5. **No scenario covers the mode/umask half of the flagged semantic change**
   (R7). *"The replaced state file keeps its readable file mode"* asserts
   0644→0644; it does not assert 0400→0644 or umask-independence, which are the
   parts that actually changed.
6. **No scenario covers a symlinked state file** (R6).
7. **No scenario asserts that a corrupt state file is DETECTED** (R8/R9) — only
   that a temp file is not mistaken for evidence.
8. **The `Load`-cascade proof is asserted only through `hook prewrite`.** The
   more valuable assertion is the direct one: `ActiveWorkflows` over a directory
   containing only a v99 file returns **1**, not 0. That is the invariant; the
   hook is one consumer of it.

---

## Proposed / Added Tests

Priority order. Each is written to fail against today's code where a defect is
claimed, and to pin behaviour where it is correct.

### Unit — `tests/unit/durable_workflow_state_unit_test.go` (+ colocated)

| P | Test | Pins |
|---|------|------|
| **1** | `TestFutureVersionNeverBricksLoad` — table over `"2.0"`, `2.0`, `1.5`, `2e2`, `true`, `99999999999999999999`, and a v99 whose `steps.status` is an array. Assert `Load` succeeds — **or**, if the team accepts R1 as-is, assert the current failure explicitly and name the constraint in the migration contract, so it stops being invisible. | R1 |
| **1** | `TestActiveWorkflowsKeepsFutureVersionFile` — write only a v99 file, assert `len(ActiveWorkflows(dir)) == 1`. The direct form of the anti-bricking invariant. | R1 |
| **1** | `TestDeleteBetweenLoadAndSaveIsAConflict` — `Load` → `os.Remove` → `Save` must refuse. **Fails today** (returns nil, recreates the file). | R3 |
| **2** | `TestUnparseableVersionIsRefusedNotAssumedLegacy` — a file whose `schemaVersion` cannot be decoded must not be treated as v1 by `Save`'s guard. **Fails today.** | R2 |
| **2** | `TestSaveNormalisesModeRegardlessOfUmask` — `syscall.Umask(0o077)`, save, assert `0644`; and `chmod 0400` → save succeeds → `0644`. Pins the flagged semantic change in both halves. (Unix-only build tag.) | R7 |
| **2** | `TestSymlinkedStateFileIsReplacedNotFollowed` — pin the `rename` semantics deliberately, whichever way the team decides. | R6 |
| **3** | `TestResidualTOCTOULosesAnUpdate` — drive `beforeRename` to land a full competing `Save` inside the window; assert both succeed and one update is lost. An **executable statement of the documented residual risk** — it starts failing the day someone adds the flock, which is exactly when it should be revisited. Restore `beforeRename` in a `defer` (package var). | R4 |
| **3** | `TestFeatureRenameAfterLoadIsNotAStaleWrite` — assert the cross-file save is either refused with an honest message or bound by `loadedPath`. **Fails today** (misdiagnoses, or writes silently). | R10 |
| **3** | `TestWorkflowStructFieldsAreVersionLocked` — golden list of `Workflow`'s JSON keys; adding a field without bumping `SchemaVersion` fails the test. This is the only mechanism that makes the whole version scheme *work*: **[probe]** a same-version save provably drops unknown fields (`unknownFieldKept=false`), so a forgotten bump is a silent data-loss release. | R1 (process) |

### Integration — `tests/integration/durable_workflow_state_integration_test.go`

| P | Test | Pins |
|---|------|------|
| **1** | `TestDoctorStaysGreenWithAWorkflowTemp` — run `internal/doctor`'s evidence check (not `filepath.Glob`) with `.<feature>.json.tmp-<rand>` present; assert green **and** that `evidence.Repair` does not remove it. Asserting on `doctor` is the point — the glob assertion already exists colocated. | Spec scenario 3 / R11 |
| **1** | `TestDoctorReportsACorruptStateFile` — truncated `.workflow/x.json`; assert `doctor` reports an **error**, not `✓ no orphaned .workflow state`. **Fails today.** | R8 |
| **2** | `TestSerialisedConcurrentWritersLoseNothing` — N goroutines `Load` concurrently, saves serialised; assert `ok == 1 && refused == N-1 && lost == 0`. The measured, honest form of the CAS guarantee. | R4 |
| **2** | `TestSaveFailurePathsLeaveTargetByteIdentical` — table: unwritable dir, missing dir, target-is-dir, target mode `0000`, stale digest, future version. For each: the error names `.workflow/<feature>.json`, the original is byte-identical, and `.workflow/` is free of `.*tmp-*`. | R-various |
| **3** | `TestStampVerificationUsesTheAtomicWriter` — assert the gatekeeper report lands at `0644`. **Fails today** (`0600`); converting `writeReport` to `WriteFileAtomic` is the fix. | R15 |

### Acceptance — `tests/acceptance/durable_workflow_state_*_test.go`

Each needs `// Acceptance: specs/durable-workflow-state.feature` plus a
standalone `// Scenario: <exact name>` matching the spec verbatim.

| P | Test | Scenario |
|---|------|----------|
| **1** | v99 fixture → `centinela status` renders; `hook prewrite` on a governed write exits **0** with no "no workflow started". | *A future-version state file does not block file writes* |
| **1** | v99 fixture → `centinela route set …` (with `[orchestration] routing_mode = "dynamic"`) exits **1**; assert all four message parts (file, 99, 1, `centinela update`) and `diff` the file against the original. | *A future-version state file is refused on save…* |
| **1** | Versionless legacy fixture → `centinela status` succeeds; then a save-bearing command stamps `"schemaVersion": 1` as the **first** key. | *A versionless legacy workflow file loads unchanged* / *…stamped with the current schema version* |
| **2** | `start` → `route set` → `status` → `start` (refused) round-trip; assert mode `-rw-r--r--`, `modelRoutes` preserved, `.workflow/` free of `.*tmp-*`. Guards the never-loaded exemption's real fence. | *A workflow that was never loaded saves without a conflict check* / R14 |
| **3** | `centinela doctor` with a leaked `.<feature>.json.tmp-<rand>` → evidence check green, exit unchanged. | *An abandoned temporary file is not mistaken for orphaned evidence* |

**Do not** chase `writeTempSibling`'s coalesced write/chmod/sync/close branch —
there is no portable trigger and contorting the production code for it is worse
than the uncovered line. Note also that `tests/` tier files do **not** move
`internal/workflow`'s per-package coverage (no `-coverpkg`); the unit rows above
must be duplicated colocated, ≤100 lines per file including `_test.go`.

---

## Residual Risks

1. **Write-write concurrency is not protected** (R4) — measured, not assumed.
   The feature closes the read-modify-write window it was built for and leaves
   the overlapping-save window open. Accepting that is defensible; leaving the
   doc comment's "microseconds" in place is not.
2. **Forward compatibility is only additive** (R1). Until `Load` probes the
   version before it unmarshals, "a newer file never bricks an older binary" is
   true only for files that add fields and change none.
3. **Nothing repairs what the old writer already broke** (R8/R9). Every
   `.workflow/<feature>.json` corrupted before this release stays corrupted,
   invisible to `doctor`, and now also unreachable via `start` (which sees the
   file exists and refuses) — so recovery is entirely manual.
4. **Versioning protects forward only** — already stated honestly by the author.
   Installed binaries predating this release still silently drop `modelRoutes`.
   Unfixable; belongs in the docs step, prominently.
5. **Leaked temps accumulate with no reaper** (R11) — bounded and harmless per
   occurrence, unbounded in aggregate.
6. **The never-loaded exemption is enforced in `cmd/`, not in the package**
   (R14). Correct today; one new `New(...)` + `Save(...)` caller re-opens it,
   and no test would notice.

---

## Deferred Findings

Not filed — `centinela roadmap defer` from inside this worktree lands the record
on the feature branch. **Orchestrator: please file these from the primary
checkout.** Items 1–6 are the senior-engineer's list, carried forward and
re-verified where noted; 7–10 are new from this step.

1. `harden-createtemp-mode-regression` — `internal/roadmap/rawio.go`,
   `internal/workflow/stamp.go`, `internal/brownmap/write.go` each replace a
   0644 file with a 0600 one via `os.CreateTemp`. **Re-confirmed this step**
   (`os.CreateTemp` → `-rw-------`). Pre-existing.
2. `doctor-repair-non-role-json-tmp` — `orphanedTmps()` globs `*.json.tmp` but
   `evidence.Repair` globs `<feature>-*.json.tmp`, so a non-role temp makes the
   check red forever with a repair that removes nothing. This feature's naming
   sidesteps it; the trap remains for the next writer who picks the obvious
   name.
3. `doctor-check-future-schema-version` — the operator only learns at the point
   of refusal. **[probe]** `status`, `status-all` and `verdict` all render a v99
   file with no warning whatsoever, and `verdict`'s JSON carries no state-file
   schema version — a CI/Magallanes consumer cannot tell its verdict was
   computed from a partially-understood file.
4. `autostart-duplicate-on-unloadable-state` — **reproduced live this step**: a
   truncated `delta.json` as the only state file yields `auto-started workflow
   "for-user-login"` alongside it.
5. `atomic-writes-for-roadmap-and-brownmap` — both still hand-roll it;
   `WriteFileAtomic` makes it cheap but the layer rule must be settled first
   (neither imports `internal/workflow`).
6. `close-state-save-toctou` — plan slice 4 (flock across re-read → rename), not
   implemented. **Now measured** rather than estimated: 23/24 updates lost under
   fully-overlapping saves; 0/24 lost when saves are serialised.
7. `load-probes-version-before-unmarshal` *(new, highest value)* — make `Load`
   read `schemaVersion` first and treat a future version as "load permissively",
   so the anti-bricking guarantee survives a future type or shape change and not
   only a field addition. R1.
8. `save-refuses-on-deleted-target-when-loaded` *(new)* — a non-empty
   `loadedDigest` plus a missing file is a conflict, not a first write. R3. One
   line, and the cheapest real fix on this list.
9. `doctor-detect-corrupt-state-file` *(new)* — `doctor` currently prints
   `✓ workflow-state  no orphaned .workflow state` and `0 error` for a truncated
   state file, with the only signal a check-less stderr warning. R8.
10. `preflight-schema-version-in-complete` *(new)* — `runComplete` refuses only
    after `runValidateGates`, so a future-version file costs a full gate run
    before an instant-to-compute refusal. R5.

Two items are small enough that they may belong in **this** feature rather than
a deferral, at the orchestrator's discretion:

- Convert `internal/workflow/stamp.go`'s `writeReport` to `WriteFileAtomic`
  (R15) — same package, no layering question, removes a weaker duplicate of the
  exact primitive this feature introduces.
- Correct the "microseconds" claim in `checkNotStale`'s doc comment to the
  measured behaviour (R4), and widen `.gitignore` beyond `.json` targets (R12).
