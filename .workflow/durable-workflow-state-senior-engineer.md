# durable-workflow-state — senior-engineer

**Step:** code · **Next role:** qa-senior · **Slices landed:** 1, 2, 3 (4 skipped, see Trade-Offs)

`.workflow/<feature>.json` is now replaced atomically, stamped with a schema
version, and guarded by a compare-and-swap that refuses a stale overwrite
instead of silently discarding another process's update.

## Files Touched

| File | Slice | Change | Lines |
|------|-------|--------|-------|
| `internal/workflow/atomic_write.go` | 1 | **new** — `WriteFileAtomic`, `writeTempSibling`, `syncDir`, `firstNonNil`, the `beforeRename` seam | 92 |
| `internal/workflow/state_io.go` | 1–3 | **new** — `Load`/`Save` moved out of `state.go`; `Save` does one re-read feeding both guards, then `WriteFileAtomic` | 75 |
| `internal/workflow/schema_version.go` | 2 | **new** — `const SchemaVersion = 1`, `legacyVersion`, `versionProbe`, `onDiskVersion`, `errFutureVersion`; migration contract in the doc comment | 70 |
| `internal/workflow/state_cas.go` | 3 | **new** — `digestOf`, `checkNotStale` | 48 |
| `internal/workflow/state.go` | 2–3 | edit — `SchemaVersion` added as the **first** struct field; unexported `loadedDigest`; `Load`/`Save` removed; 4 imports dropped | 97 → 82 |
| `.gitignore` | 1 | edit — `.workflow/.*.json.tmp-*` beside the `.roadmap-digest-*` precedent | +6 |
| `internal/workflow/atomic_write_test.go` | 1 | **new** — replace/mode/no-temp/glob-safety + `stateRepo` helper | 100 |
| `internal/workflow/atomic_write_errors_test.go` | 1 | **new** — rename failure, `Save` write failure, `syncDir` best-effort, `firstNonNil` | 82 |
| `internal/workflow/state_io_crash_test.go` | 1 | **new** — the killed-write proof (re-exec'd child) | 95 |
| `internal/workflow/schema_version_test.go` | 2 | **new** — version rules incl. on-disk-not-in-memory read | 97 |
| `internal/workflow/state_version_compat_test.go` | 2 | **new** — legacy 133-file-shaped fixture round-trip | 84 |
| `internal/workflow/state_cas_test.go` | 3 | **new** — stale refusal, retry recovery, never-loaded exemption | 99 |
| `cmd/centinela/route_edges_test.go` | 1 | **edit (fallout)** — induce save failure via unwritable **directory** | 70 |
| `cmd/centinela/cov2_extra_errors_test.go` | 1 | **edit (fallout)** — same | 74 |

## Architecture Compliance

**Boundaries (G2).** No new import edge out of `internal/workflow`. Every new
file imports stdlib only (`crypto/sha256`, `encoding/hex`, `encoding/json`,
`errors`, `fmt`, `io/fs`, `os`, `path/filepath`) — verified by grepping the four
new source files for `samuelnp/centinela`: zero matches. `internal/evidence`
already imports `internal/workflow`, so `WriteFileAtomic` is exported and
available to it for a later follow-up, but **no call site outside
`internal/workflow` was converted** in this feature.

**G1 (≤100 lines).** Every touched file, source and test, is ≤100. `state.go`
was already at 97, which is why `Load`/`Save` had to move out rather than grow;
`state_cas_test.go` hit 102 on first draft and was trimmed to 99.

**G7 (outer layer).** `cmd/centinela/` gained no logic. The two edits there are
test-only and change how a failure is *induced*, not any production behaviour.

**Naming.** `state_io.go` / `state_cas.go` / `schema_version.go` follow the
package's existing `state_*.go`, `model_routes_*.go` convention.

## Type-Safety Notes

- `WriteFileAtomic(path string, data []byte, perm fs.FileMode) error` — `perm`
  is `fs.FileMode`, not a bare int, and `stateFileMode fs.FileMode = 0o644` is a
  named constant rather than a literal repeated at the call site.
- `SchemaVersion int` with `json:"schemaVersion"` and **no `omitempty`**, so
  every file this binary writes is stamped even at the zero value.
- `versionProbe` is a dedicated single-field struct. Probing with it rather than
  unmarshalling into `Workflow` is what lets a file carrying fields this binary
  cannot model still yield its version — verified in dogfooding with a fixture
  containing `futureFieldThisBinaryCannotModel`.
- `loadedDigest string` is **unexported**, so `encoding/json` ignores it: the
  on-disk shape is unchanged and no fixture needed touching.
- `Save`'s signature is unchanged, so the `var saveWorkflow = workflow.Save`
  seam in `cmd/centinela/complete.go:23` and its stubbing tests still bind.

## Verification Performed

| Check | Result |
|-------|--------|
| `go build ./...` | pass |
| `go vet ./...` | pass |
| `GOOS=windows go build ./...` | pass |
| `gofmt -l internal/ cmd/` | clean |
| `go test ./... -run xxxNONE` (compile inventory) | **48 packages, 0 breakage** |
| `go test ./...` (full suite) | all green |
| `internal/workflow` coverage | **97.2%** (96.5% mid-slice; gate is 95%) |
| `internal/evidence`, `internal/doctor`, `internal/hookpolicy`, `cmd/centinela` | green |

**Three plan predictions were verified rather than trusted:**

1. **The doctor-glob trap (§1.6) — confirmed safe.** Ran `filepath.Glob`
   against a real `os.CreateTemp(".workflow", ".durable-workflow-state.json.tmp-")`
   result: `*.json.tmp` → no match, `*.json` → no match, `.*.json.tmp-*` → match.
   So the temp trips neither `doctor`'s orphan sweep (which would go red with a
   repair that removes nothing, forever) nor `ActiveWorkflows`. Pinned by
   `assertNoGlobMatch` in `atomic_write_test.go` and by
   `TestKilledWriteTempIsNotAnEvidenceOrphan`.
2. **`os.CreateTemp` yields 0600 — confirmed empirically** (`-rw-------`), which
   is why the explicit `Chmod(perm)` is mandatory and not decoration.
3. **Zero test-compile breakage from the version key — confirmed.** 48 packages
   compiled with `-run xxxNONE`, zero failures. All 133 tracked state files are
   versionless and none is byte-compared.

**The D4 Load-cascade was verified end-to-end, not just read.** Static trace
first (`hook_prewrite` → `loadActiveWorkflows` → `ActiveWorkflows` warn+skip →
`hookpolicy.EvaluatePrewrite` `prewrite.go:32-33` `if len(wfs) == 0 { NeedInit }`
→ `hook_autostart.go:27` `len(loadActiveWorkflows()) > 0`), then demonstrated
with the built binary in a scratch repo: corrupting the only state file (so
`Load` genuinely fails) made `hook prewrite` print **BLOCKED WRITE**; restoring
the same file at `schemaVersion: 99` (so `Load` succeeds) made the identical
write **exit 0**. That is the whole argument for putting the refusal in `Save`,
and it now has a demonstration behind it.

**The two "must not be theatre" tests were mutation-tested.** Disabling
`checkNotStale` made `TestStaleSaveIsRefused` and `TestReRunAfterConflictSucceeds`
fail; routing `WriteFileAtomic` straight at the target (the old `O_TRUNC`
behaviour) made `TestKilledWriteLeavesPreviousStateIntact` fail. Both went green
again on revert. They genuinely detect the defects they claim to.

### Dogfood (scratch binary `/tmp/centinela-dws`, trimmed)

Real state file in this worktree — `route set` / `status` round-trip:

```
=== BEFORE ===   {   "feature": "durable-workflow-state",   -rw-r--r--   0 tmp files
 CLI  route set: qa-senior -> balanced (was balanced)
=== AFTER  ===   {   "schemaVersion": 1,   "feature": ...   -rw-r--r--   no .tmp- leftovers
 STATUS  Feature durable-workflow-state / Archetype canonical      (status ok)
 route show: qa-senior  balanced  routed  ...  2026-08-03T18:54:50Z (round-trip ok)
```

Mode stayed 0644, `schemaVersion` became the first key, nothing left behind.
The dogfood route was then reverted — `git status` on the state file is clean.

Scratch repo, `schemaVersion: 99` fixture:

```
1) status delta                       -> loads fine (Feature delta, Archetype canonical)
2) hook prewrite (governed .go write) -> exit=0   (allowed; no "no workflow started")
3) route set delta qa-senior balanced -> exit=1
   cannot save workflow: .workflow/delta.json was written by a newer Centinela
   (schema version 99); this binary understands schema version 1. Refusing to
   write — saving would drop fields it does not know about. Upgrade with
   `centinela update` and re-run
4) diff vs original                   -> IDENTICAL;  ls .workflow/ -> delta.json only
```

## Trade-Offs

**Slice 4 (relocate `flock` into `internal/workflow`, close the TOCTOU) was
NOT done.** The plan marks it last and droppable; here is the grounded reason,
which is stronger than "ran out of room":

- The primitive is **non-blocking** (`syscall.LOCK_EX|LOCK_NB`). Moving
  `tryLockExclusive`/`unlockFile` verbatim gives `Save` a call that returns
  `(false, nil)` when contended — which closes nothing on its own. The
  retry/timeout policy lives one level up in `evidence.Lock` (`LockTimeout`,
  `LockPollInterval`) and is **evidence-specific**: it is keyed by
  `(feature, role)` and its busy error tells the operator to run
  `centinela evidence read`. So slice 4 is not a file move; it is a file move
  plus a second retry loop and a new `Save` failure mode.
- It is the only slice touching a second package, and `internal/evidence`
  currently sits at 97.2%. Moving `lock_unix.go`/`lock_windows.go` out while
  leaving `lock.go` delegating moves **both** packages' per-package coverage at
  once — and per-package coverage is what the gate reads.
- `lock_windows.go` cannot be exercised on this platform, so the relocation
  would be verified only by cross-compilation.
- What it buys is small relative to the risk: the residual TOCTOU is two
  processes both passing CAS and both renaming **microseconds** apart. The bug
  this feature exists to fix is a **minutes**-wide window (`complete` holding a
  read across a full gate run), and slices 1–3 close that completely.

**Two `cmd/centinela` tests changed behaviour, deliberately.**
`TestCov2RunCompleteSaveError` and `TestRunRouteShow_MissingWorkflowAndUnwritableState`
made the **target file** read-only to induce a save error. Atomic replace makes
that no longer fail: `rename(2)` needs write permission on the **directory**, not
the file. Both were switched to chmod the state directory `0500`, which is also
what the spec actually describes ("Given the workflow state directory cannot be
written to"). Note the resulting semantic change for review: a read-only state
file is now writable by Centinela, and its mode is normalised to 0644 on save.
Nothing in the codebase makes the state file read-only, and 0644 is the mode
every hook needs to read it, so this is judged correct rather than a regression.

**Optimistic CAS, not a lock — restated for review.** A lock cannot fix this:
`complete` reads minutes before it writes, so serializing writers either blocks
`route set` for a whole gate run (past the 2s `LockTimeout`) or, narrowed to the
write, leaves the lost update entirely intact. An in-process mutex guards
nothing, because `route set` and `complete` are separate OS processes. This is
written into the `checkNotStale` doc comment so it cannot be "helpfully" undone.

**Self-conflict avoided.** After a successful write `Save` refreshes
`wf.loadedDigest` to the bytes it just published. Without this, a second `Save`
on the same in-memory struct would refuse against the file it had itself
written. Not in the plan; found while reasoning about the call sites.

**Honest limitation, for the docs step.** Versioning protects **forward only**.
Already-installed binaries carry no refusal, so they can still silently drop
`modelRoutes` from a file this release writes. The guarantee begins with the
first release carrying the check on both sides. Nothing retrofits that.

## Deferred Findings

**none filed** — no `centinela roadmap defer` was run. `.workflow/roadmap.json`
is committed state that belongs on `main`, and deferring from inside this
worktree lands the record on the feature branch instead, where it is easily
lost. Listing them here for the **orchestrator to file from the repo root**
(items 1–5 are the plan's §6; item 6 is new from this step):

1. `os.CreateTemp` **mode regression** in three existing helpers —
   `internal/roadmap/rawio.go`, `internal/workflow/stamp.go`,
   `internal/brownmap/write.go` each replace a 0644 file with a 0600 one.
   Confirmed real this step (`os.CreateTemp` → `-rw-------`). Pre-existing.
2. `doctor` **cannot repair a non-role `*.json.tmp`** — `orphanedTmps()` globs
   `*.json.tmp` but `evidence.Repair` globs `<feature>-*.json.tmp`, so a
   non-role temp makes the check red forever with a repair that removes nothing.
   Our naming sidesteps it; the trap remains for the next writer who picks the
   obvious name.
3. **No `doctor` check for a future-version state file** — the operator only
   learns at the point of refusal.
4. `hook_autostart` **creates a duplicate workflow** when the only state file
   fails to load. Demonstrated live this step (a corrupt file empties the active
   set). Independent of this feature; reproducible today.
5. `internal/roadmap` / `internal/brownmap` still hand-roll atomic writes.
   `WriteFileAtomic` makes the follow-up cheap, but the layer rule must be
   settled first (neither imports `internal/workflow`).
6. **The residual TOCTOU** — plan slice 4, not implemented. See Trade-Offs.

## Handoff

**Next role: qa-senior.**

**Breakage inventory: two files, both already fixed and green.** `go test ./...
-run xxxNONE` reports **zero** compile breakage across 48 packages; the full
suite is green. The only behavioural fallout was
`cmd/centinela/route_edges_test.go` and `cmd/centinela/cov2_extra_errors_test.go`
(read-only target no longer fails a save) — described under Trade-Offs. **If you
write any new test that induces a save failure, chmod the DIRECTORY, not the
file.**

**Seams available to you (all already exercised, extend rather than re-invent):**

- `beforeRename` (`atomic_write.go`) — package var, no-op in production, called
  after the temp is written+fsynced and before `os.Rename`. The deterministic
  killed-write seam. Drive it exactly as `state_io_crash_test.go` does: a helper
  test gated on `CENTINELA_CRASH_TEST=1` that sets it to `os.Exit(1)`, re-exec'd
  via `exec.Command(os.Args[0], "-test.run", ...)`. `os.Exit` skips deferred
  cleanup, which is faithful to SIGKILL for the only question that matters.
  **It is a package var — restore it if you set it in an in-process test.**
- `writeTempSibling(dir, base, data, perm)` — unexported; call it directly to
  materialise a temp file without a rename, for glob-safety assertions.
- `stateRepo(t)` (`atomic_write_test.go`) — `t.Chdir` into a temp repo with
  `.workflow/` created. `writeRawState(t, feature, body)`
  (`schema_version_test.go`) — hand-authored on-disk bytes.
- **Concurrency needs no processes**: two `Load`s of the same feature is the
  whole shape (`state_cas_test.go`). Both key tests were mutation-tested.
- `legacyState` (`state_version_compat_test.go`) — the versionless fixture
  shaped like the 133 tracked state files.

**Still owed by the tests step** (per plan §3, not written here — `tests/` is
blocked during the code step):

- `tests/unit/durable_workflow_state_unit_test.go` — the version rules table.
- `tests/integration/durable_workflow_state_integration_test.go` — `Load` →
  mutate → `Save` over the real `.workflow/` layout, plus the assertion that
  `internal/doctor`'s evidence check stays **green** with a workflow temp
  present (this is the §1.6 trap; assert on `doctor`, not just on the glob).
- `tests/acceptance/durable_workflow_state_test.go` — drives the built binary
  for the three dogfooded behaviours: a legacy versionless file loads via
  `centinela status`; `schemaVersion: 99` is refused by `centinela route set`
  with the actionable message; the same file does **not** make
  `centinela hook prewrite` block. Each needs `// Acceptance:` plus a standalone
  `// Scenario: <exact name>` matching `specs/durable-workflow-state.feature`
  verbatim.
- `.workflow/durable-workflow-state-edge-cases.md`.

**Note on coverage:** per-package coverage is what moves the gate (no
`-coverpkg`), so the `tests/` tier will not lift `internal/workflow`. It is
already at **97.2%** from the colocated tests; keep it there. The one uncovered
branch left is `writeTempSibling`'s coalesced write/chmod/sync/close failure
path, which has no portable trigger — do not contort the production code to
reach it.

**Spec coverage:** all 13 scenarios have a colocated counterpart already, except
the three requiring the built binary or the `doctor` check (acceptance /
integration tier above).
