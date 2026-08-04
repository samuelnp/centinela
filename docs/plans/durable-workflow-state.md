# Plan: durable-workflow-state

**Phase:** 13 — Lighter Centinela · **Archetype:** canonical · **Next role:** senior-engineer

`.workflow/<feature>.json` is written by an unlocked `os.WriteFile` and carries no
schema version. This plan makes the write atomic, stamps a version, and makes a
stale overwrite fail loudly instead of silently discarding the other writer's
update. It is delivered as four independently revertible slices.

---

## 1. Grounding: what is actually on disk and who touches it

### 1.1 The write today

`internal/workflow/state.go` (97 lines — already near the 100-line cap):

```go
func Save(wf *Workflow) error {
	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil { return err }
	return os.WriteFile(FilePath(wf.Feature), data, 0644)
}
```

`os.WriteFile` opens `O_WRONLY|O_CREATE|O_TRUNC`. Between the truncate and the
write the file is **zero bytes**. A kill in that window loses the feature's
state outright; a short write leaves invalid JSON. Target mode today is **0644**
(verified: the literal in `Save`, subject to umask).

`Load` is permissive about nothing: any read or parse failure is an error. It
has no notion of a version.

### 1.2 Caller inventory

**`Save` — 5 call sites, 5 files.** Two never load first (no conflict possible),
three read-modify-write (the conflict surface):

| # | Call site | Loads first? | Gap between Load and Save | Conflict risk |
|---|-----------|--------------|---------------------------|---------------|
| 1 | `cmd/centinela/start.go:84` | no (`NewWithOrder`) | — | none |
| 2 | `cmd/centinela/hook_autostart.go:42` | no (`NewWithOrder`) | — | none |
| 3 | `cmd/centinela/route_set.go:58` | yes (`:41`) | milliseconds | **yes** |
| 4 | `cmd/centinela/complete.go:59` (via the `saveWorkflow` package var, `:23`) | yes (`:37`) | **the whole validate gate run — minutes** | **yes, the observed one** |
| 5 | `cmd/centinela/revise.go:63` (via `saveWorkflow`) | yes (`:49`) | `RewindTo` + `invalidateDownstream` | **yes** |

`saveWorkflow` is a package var (`var saveWorkflow = workflow.Save`) stubbed by
tests. **`Save`'s signature must not change** or that seam and its tests break.

**`Load` — 18 call sites, 16 files.** 12 in `cmd/` (`revise`, `route_show`,
`active_feature`, `complete`, `route_set`, `mcp`, `verdict`, `status` ×2,
`deliver`, `hook_postwrite`, `verify`), 6 in `internal/`
(`evidence/repair.go:82`, `evidence/roles_retired.go:30`,
`roadmap/roadmap.go:46`, `workflow/active.go:27`, `workflow/contract.go:20` and
`:45`). Every one is read-only. **`workflow/active.go:27` is the hot path** —
see §1.4.

### 1.3 The concurrency case, concretely

`complete` on the validate step loads at T0, runs `runValidateGates` (a full
profiled suite run — minutes, per `centinela.toml`), and saves at T+minutes.
`route set` in another terminal loads at T+1s and saves at T+1s. At T+minutes
`complete` writes its stale copy and the route is gone: `route set` printed
success, emitted telemetry, and changed nothing durable.

**This kills the lock option.** Serializing writers means `complete` holds the
lock across the entire gate run, so `route set` blocks for minutes (the existing
`evidence.LockTimeout` is 2s). An in-process mutex is worse than useless here:
`route set` and `complete` are **separate OS processes**, so a `sync.Mutex` in
`internal/workflow` guards nothing that is actually racing. Say it plainly in
review — the honest answer is that a mutex buys nothing.

The mechanism that fits the shape of the bug is **compare-and-swap**: record what
`Load` read, and refuse to write if the file no longer matches. That is exactly
"detect the conflict and fail loudly".

### 1.4 Blast radius of a refusing `Load` — traced

`cmd/centinela/hook_prewrite.go` → `loadActiveWorkflows()` (`hook_workflows.go`)
→ `workflow.ActiveWorkflows()` (`active.go:27`). `ActiveWorkflows` already
*tolerates* a `Load` error: it warns to stderr and `continue`s. Follow it
through:

1. The only active feature's file fails to load → it is skipped.
2. `loadActiveWorkflows()` returns an **empty slice**.
3. `hookpolicy.EvaluatePrewrite` hits `if len(wfs) == 0 { return
   PrewriteDecision{NeedInit: true} }` (`prewrite.go:32`) → **every write to a
   governed file type is blocked** with "no workflow started".
4. Worse: `hook_autostart.go:27` fires on `len(loadActiveWorkflows()) > 0` being
   false, so the next prompt **auto-starts a duplicate workflow**
   (`uniqueFeatureName` → `<feature>-2`).

**Decision (D4): the version refusal never goes in `Load`.** `Load` stays
permissive — it parses, exposes the version, and returns. Refusal lives in the
single mutating choke point, `Save`. A future-version file then costs the
operator nothing but the ability to *advance* that feature, which is the correct
scope of the protection.

### 1.5 Existing atomic-write helpers (reuse before inventing)

Four already exist. None is directly reusable, for a reason worth recording:

| Helper | fsync? | Mode after rename | Reusable from `internal/workflow`? |
|--------|--------|-------------------|-------------------------------------|
| `internal/evidence/io_write.go` `writeBytesAtomic` | file only | **0644** (explicit `O_CREATE` perm) | **No — import cycle.** `evidence` already imports `workflow` (`io_write.go:8`). |
| `internal/roadmap/rawio.go` `writeAtomic` | none | **0600** (`os.CreateTemp` default — latent mode regression) | No — would create a `workflow → roadmap` edge PROJECT.md G2 does not sanction. |
| `internal/workflow/stamp.go` `writeReport` | none | **0600** (same latent bug) | Same package, but markdown-specific and does not fsync. |
| `internal/brownmap/write.go` | none | 0600 | No — layer. |

Two findings fall out:

- The `evidence` helper is the best of the four and is the one to model on, but
  the dependency arrow points `evidence → workflow`, so the shared helper must
  live **in `internal/workflow`**, not the other way round.
- `os.CreateTemp` creates at **0600**. Three of the four helpers silently
  downgrade the mode of the file they replace. Our helper must set 0644
  explicitly or it reintroduces that bug in the one file every hook reads.

### 1.6 Temp-file naming vs. `centinela doctor` — a verified trap

`internal/doctor/check_evidence.go:35` globs `.workflow/*.json.tmp` and reports
any match as `Error: orphaned evidence *.json.tmp files from a crashed write`,
offering a repair marked `Safe, Idempotent`. That repair calls
`evidence.Repair(featurePrefix(base))`, and `evidence.Repair`
(`internal/evidence/repair.go:30`) globs `<feature>**-**\*.json.tmp` — **with a
hyphen**, because it expects `<feature>-<role>.json.tmp`.

So if the workflow temp were named `.workflow/<feature>.json.tmp`:

- `orphanedTmps()` matches it → doctor goes **red**;
- `featurePrefix` finds no role suffix → returns the bare feature;
- `evidence.Repair(feature)` globs `<feature>-*.json.tmp` → **does not match**;
- repair returns `removed: []`, `err: nil` → doctor reports the fix applied and
  the check stays red **forever**.

A permanently-red doctor with a repair that silently does nothing. The temp name
must therefore stay out of that glob.

### 1.7 Hand-written fixture inventory (the compatibility canaries)

**48 `_test.go` files hand-author a `.workflow/*.json` fixture containing
`currentStep`** — 7 in `cmd/centinela/`, 3 in `internal/`, 38 across
`tests/{unit,integration,acceptance}/`. Plus **132 real, git-tracked
`.workflow/<feature>.json` files in this repo**, and every state file in every
downstream project.

`grep -L schemaVersion .workflow/*.json` → **132 of 132 are versionless.**

These are the breakage inventory *and* the proof set:

- They break **only if `Load` starts requiring the version field.** The design
  forbids that (absent ⇒ 1), so the expected breakage is **zero**.
- They cannot break on the CAS check either: none of them exercises a
  load-then-stale-save.
- No test compares a saved workflow file byte-for-byte (checked: no golden
  fixtures under `internal/workflow` or `cmd/centinela`), so adding a
  `schemaVersion` key to the marshalled output breaks nothing.
- The one thing to watch: tests that assert `Save` output *contents* by
  substring. Grep before the code step; none found today.

---

## 2. The four decisions

### D1 — Atomic write mechanics

Temp file in `.workflow/`, `fsync`, `rename`, best-effort directory `fsync`.

| Question | Decision | Why |
|----------|----------|-----|
| Temp name | `os.CreateTemp(WorkflowDir, "."+feature+".json.tmp-")` → `.workflow/.alpha.json.tmp-3417829` | Unique per process (a deterministic name lets two writers clobber each other's temp — one could rename the *other's* half-written bytes over the target, a corruption vector worse than the bug we are fixing). Dot-prefixed and suffixed `.tmp-<rand>`, so it matches neither `.workflow/*.json` (`ActiveWorkflows`) nor `.workflow/*.json.tmp` (§1.6). |
| Temp mode | `chmod 0644` on the temp before rename | `os.CreateTemp` yields 0600; without this the state file every hook reads silently becomes owner-only after the first save. Matches today's `Save` literal. |
| fsync the file | **Yes**, before close | Guarantees the bytes reach the device before the rename publishes them. |
| fsync the directory | **Yes, best-effort** — open `.workflow`, `Sync()`, ignore any error | Without it, a power loss can leave the directory entry unpersisted. Best-effort because opening a directory for sync is not portable (Windows); a failure there must never fail a save that already succeeded. This is a deliberate improvement over the `evidence` helper. |
| Temp on failed write | Removed on every in-process error path (`defer os.Remove(tmp)`, a no-op once the rename has moved it). A **SIGKILL** leaves one behind — unavoidable, and harmless: it is a hidden dotfile, it trips no glob (§1.6), and it will be gitignored. |
| gitignore | Add `.workflow/.*.json.tmp-*`, mirroring the existing `.workflow/.roadmap-digest-*` precedent | Keeps `git status` clean after a kill. |

Note the ordering guarantee this buys: a reader either sees the complete old file
or the complete new one, because `rename(2)` is atomic within a filesystem and
the temp is a **sibling** (no `EXDEV`).

### D2 — Concurrency: compare-and-swap, not a lock

**Chosen: optimistic compare-and-swap, detect and fail loudly.**

- `Load` records `sha256(raw file bytes)` into an **unexported** `loadedDigest`
  field on `Workflow`. Unexported ⇒ `encoding/json` ignores it ⇒ the on-disk
  shape is untouched and no fixture changes.
- `Save` re-reads the target. If `loadedDigest` is non-empty and the current
  bytes hash differently, the save is **refused**.
- If `loadedDigest` is empty (the workflow was built by `New`/`NewWithOrder` and
  never loaded — call sites 1 and 2), **no check runs**. This is what keeps
  `start`, `hook autostart`, and all 48 fixtures untouched.

Rejected alternatives, with reasons to state in review:

- **Lockfile serialization** — `complete` would hold the lock across a
  minutes-long gate run, blocking `route set` far past any sane timeout. The
  lock cannot be *narrowed* to the write, because the stale data was read
  minutes earlier; a narrow lock leaves the bug entirely intact.
- **In-process mutex** — `route set` and `complete` are separate processes. It
  guards nothing. Do not ship it and do not let it be mistaken for a fix.

**Residual TOCTOU:** two processes could both pass the CAS and both rename,
microseconds apart. Slice 4 closes it with a short `flock` held only across
re-read → rename (microseconds, never across a gate run). It is the *last* slice
precisely because the feature is correct and shippable without it.

### D3 — Schema version

| Question | Decision |
|----------|----------|
| Field | `SchemaVersion int \`json:"schemaVersion"\`` — **no `omitempty`**, so every file this binary writes carries it |
| Placement | **First field in the struct**, so it is the first key in the marshalled JSON and `head -3` of any state file shows it |
| Constant | `const SchemaVersion = 1` in a new `internal/workflow/schema_version.go`, with the migration contract in its doc comment |
| Absent / `0` on load | **Version 1.** Load succeeds unchanged. |
| Equal on load | Load succeeds. |
| **Greater** than this binary understands | **Load still succeeds** (D4 — refusing here bricks the prewrite hook). `Save` refuses. |
| **Less** than this binary understands (a future v2 reading a v1) | Loads; `Save` stamps the current version — a **silent one-way upgrade**. Every field added so far is back-compat-by-absence (`Archetype`, `ValidateContract`, `PlanContract`, `ModelRoutes`), so "absent means the documented default" is the whole migration. A future v2 that needs more than defaulting must add its own migration next to the constant; record that obligation in the doc comment now. |
| Where the check reads the version | **From the on-disk bytes at save time**, not only from the in-memory struct — the file may have been upgraded by a newer binary between our `Load` and our `Save`. This shares the single re-read that D2's CAS already performs. |

Error text (must name file, its version, our version, and the fix):

```
.workflow/delta.json was written by a newer Centinela (schema version 99);
this binary understands schema version 1. Refusing to write — saving would
drop fields it does not know about. Upgrade with `centinela update` and re-run.
```

(`centinela update` is the real command — `cmd/centinela/update.go:18`.)

**Honest limitation to state in the report and the docs step:** versioning only
protects *forward*. Binaries already installed have no check, so they will still
silently drop `modelRoutes` from a file this release writes. The guarantee
begins with the first release that carries the check on both sides. Nothing can
retrofit that, and the plan should not pretend otherwise.

CAS conflict error text:

```
.workflow/epsilon.json changed on disk since this command read it — another
centinela process wrote it (a concurrent `route set`, `complete`, or `revise`).
Refusing to write so that update is not lost. Re-run this command to apply
your change on top of the current state.
```

### D4 — Refusal lives in `Save`, never in `Load`

Traced in §1.4. `Load` stays permissive; `Save` is the single choke point every
mutator already goes through and whose error every one of the 5 call sites
already wraps and surfaces. Consequence: a future-version file leaves the
project fully writable and fully readable — `status`, `dashboard`, `verify`,
`hook prewrite`, `hook postwrite` all keep working — and only *advancing* that
one feature is refused. That is the correct blast radius.

Deferred (not this feature): a `doctor` check that *reports* a future-version
state file, so the operator learns about it before hitting the refusal.

### D5 — Export the helper

**Yes: `workflow.WriteFileAtomic(path string, data []byte, perm fs.FileMode) error`.**

Justification: `internal/evidence` **already imports `internal/workflow`**
(`io_write.go:8`), so it can adopt the shared helper with **no new import edge
and no cycle** — which is exactly the constraint the brief sets. `internal/roadmap`
and `internal/brownmap` do **not** import `workflow`; adopting there would create
edges PROJECT.md G2 does not sanction, so they are left alone (converting them
is out of scope per the brief anyway).

Exporting costs nothing now and is what makes the follow-up cheap. **No call
site outside `internal/workflow` is converted in this feature.**

---

## 3. File-by-file plan

Every file below is ≤100 lines including tests (`cmd/` and `internal/`).
`internal/workflow/state.go` is already at **97 lines**, so `Load`/`Save` move
out of it — that is a hard constraint, not a preference.

### Slice 1 — Atomic write (no schema change, no behaviour change)

| File | Change | Notes |
|------|--------|-------|
| `internal/workflow/atomic_write.go` | **new** | `WriteFileAtomic(path, data, perm)`; unexported `writeTempFile` (open `O_WRONLY\|O_CREATE\|O_EXCL`, write, chmod, `Sync`, close) and `syncDir` (best-effort). Wrap every error with the **target** path. ~55 lines. |
| `internal/workflow/state_io.go` | **new** | `Load` and `Save`, moved verbatim out of `state.go`. `Save` now ends in `WriteFileAtomic(FilePath(wf.Feature), data, 0o644)`. ~35 lines. |
| `internal/workflow/state.go` | trim | Delete `Load`/`Save`; drop the now-unused `encoding/json`, `errors`, `io/fs`, `os` imports. Drops to ~65 lines, leaving room for slices 2–3. |
| `internal/workflow/atomic_write_test.go` | **new** | Round-trip; mode is 0644 after replacing a 0644 file; error names the target when the dir is unwritable; no temp remains after success; temp name matches neither `*.json` nor `*.json.tmp`. |
| `internal/workflow/state_io_crash_test.go` | **new** | The killed-write proof — see §4.1. |
| `.gitignore` | edit | Add `.workflow/.*.json.tmp-*` under the existing `.roadmap-digest-*` block, with a one-line comment. |

**Revert = one line**: point `Save` back at `os.WriteFile`. Nothing else in the
tree depends on the helper yet.

### Slice 2 — Schema version

| File | Change | Notes |
|------|--------|-------|
| `internal/workflow/schema_version.go` | **new** | `const SchemaVersion = 1`; `readOnDiskVersion(path) (int, []byte, error)` (unmarshals into a `struct{ SchemaVersion int }` probe — tolerant of every other field); `errFutureVersion(path string, onDisk int) error` with the D3 text. Document the v1→v2 migration contract here. ~50 lines. |
| `internal/workflow/state.go` | edit | Add `SchemaVersion int \`json:"schemaVersion"\`` as the **first** struct field, with a doc comment stating: absent ⇒ 1; a higher value is refused on save, never on load. |
| `internal/workflow/state_io.go` | edit | `Load` leaves an absent version as `0` and normalises it to `SchemaVersion` on the returned struct (in-memory only — no rewrite of the file). `Save` stamps `wf.SchemaVersion = SchemaVersion`, and refuses when the **on-disk** version exceeds it. |
| `internal/workflow/schema_version_test.go` | **new** | absent ⇒ 1; equal round-trips; a `schemaVersion: 99` file loads fine but refuses on save; the error names path + 99 + 1 + `centinela update`; the refused file is byte-identical afterwards. |
| `internal/workflow/state_version_compat_test.go` | **new** | A verbatim copy of a real 132-file-style legacy fixture (no `schemaVersion`, with `modelRoutes` and `revisions`) loads and round-trips with every field intact. |

**Revert = drop the field and the check.** Files already stamped with
`schemaVersion: 1` still load in the reverted binary — it is just an unknown key
that `encoding/json` ignores. *Reverting is lossy in exactly the way this
feature exists to prevent, so note it on the revert path.*

### Slice 3 — Compare-and-swap

| File | Change | Notes |
|------|--------|-------|
| `internal/workflow/state.go` | edit | Add unexported `loadedDigest string` (not serialized) with a doc comment explaining why it is unexported. |
| `internal/workflow/state_cas.go` | **new** | `digestOf([]byte) string` (sha256, hex); `checkNotStale(path string, wf *Workflow, current []byte) error` returning `errStaleWrite` with the D3 CAS text. ~40 lines. |
| `internal/workflow/state_io.go` | edit | `Load` sets `wf.loadedDigest`. `Save` performs **one** read of the target and feeds both the version check (slice 2) and the staleness check. A missing target is not a conflict (first write). |
| `internal/workflow/state_cas_test.go` | **new** | Two `Load`s, first `Save` wins, second is refused; the error names the file and says to re-run; the first writer's change survives; a `New()`-built workflow saves with no check; reloading after a refusal then saving succeeds. |

**Revert = delete the digest field and the call.** Slices 1 and 2 stand alone.

### Slice 4 (last, optional) — Close the TOCTOU; one flock implementation

Only if slices 1–3 land green with room to spare. Independently revertible and
the only slice that touches a second package.

| File | Change |
|------|--------|
| `internal/workflow/filelock_unix.go` / `filelock_windows.go` | **new**, build-tagged. Moved verbatim from `internal/evidence/lock_unix.go` / `lock_windows.go`, exported as `TryLockExclusive(*os.File) (bool, error)` / `UnlockFile(*os.File) error`. |
| `internal/evidence/lock_unix.go`, `lock_windows.go` | **delete**; `internal/evidence/lock.go` delegates to `workflow.TryLockExclusive` / `workflow.UnlockFile`. No new edge (`evidence → workflow` already exists), no signature change to `evidence.Lock`, no behaviour change. |
| `internal/workflow/state_io.go` | `Save` takes a short flock on `.workflow/<feature>.lock` around re-read → rename only. Already covered by the `.workflow/*.lock` gitignore rule. Hold time: microseconds. **Never** across a gate run. |
| tests | Move `internal/evidence`'s lock tests with the primitive; add a test asserting the lock is released even when the version/CAS check refuses the write. |

**If the verifier objects to touching `internal/evidence`, drop slice 4
entirely** — slices 1–3 satisfy the brief.

### Test tiers (the tests step writes these; specs are already complete)

- `tests/unit/durable_workflow_state_unit_test.go` — version rules table.
- `tests/integration/durable_workflow_state_integration_test.go` — `Load` →
  mutate → `Save` across the real `.workflow/` layout; the doctor-glob assertion
  (`internal/doctor` evidence check stays green with a workflow temp present).
- `tests/acceptance/durable_workflow_state_test.go` — drives the **built
  binary**: a legacy versionless file loads via `centinela status`; a
  `schemaVersion: 99` file is refused by `centinela route set` with the
  actionable message; the same file does **not** make `centinela hook prewrite`
  block a write. Each carries `// Acceptance:` plus a standalone
  `// Scenario: <exact name>` line matching the `.feature` verbatim.

Remember: **per-package coverage is what moves the gate** (no `-coverpkg`), so
the `tests/` tier does not lift `internal/workflow`. The colocated
`internal/workflow/*_test.go` files above are the ones that must carry it to
≥97%.

---

## 4. Two tests that must be real, not theatre

The brief is explicit: *"Tests must actually exercise the failure modes — a
killed write and a concurrent write — rather than asserting the happy path
twice."*

### 4.1 The killed write

`rename` is instantaneous, so racing a real signal is nondeterministic. Make it
deterministic with a **helper subprocess** and `os.Exit`:

- A package var `beforeRename = func() {}` in `atomic_write.go`, no-op in
  production.
- `TestHelperCrashesBeforeRename` runs only when `CENTINELA_CRASH_TEST=1`: it
  sets `beforeRename` to `func() { os.Exit(1) }`, then calls `Save`.
- The parent test writes a state file at step `code`, re-execs itself
  (`os.Args[0] -test.run=TestHelperCrashesBeforeRename`) with the env var, waits
  for the non-zero exit, and asserts the target **still parses** and **still
  says `code`**.

`os.Exit` skips deferred cleanup — faithful to SIGKILL for the only question
that matters here: what is on disk afterwards. The same test asserts the
leftover temp does not match `.workflow/*.json.tmp`.

### 4.2 The concurrent write

No processes needed; the *shape* is what matters, and it is two `Load`s:

```
wfA, _ := Load("epsilon")   // stands in for `complete`
wfB, _ := Load("epsilon")   // stands in for `route set`
wfB.SetModelRoute(...)
Save(wfB)                    // succeeds
wfA.CurrentStep = "docs"
Save(wfA)                    // MUST fail with the stale-write error
```

Then assert `wfB`'s route survived. That is the observed bug, reproduced
deterministically, and it fails against today's `Save`.

---

## 5. Rollout and revert

1. **Slice 1 (atomic write)** — smallest correct slice. Fixes defect 1. Zero
   file-format change, so it can ship and be reverted with no compatibility
   consequence at all.
2. **Slice 2 (schema version)** — fixes defect 2. Independent of slice 1.
3. **Slice 3 (CAS)** — fixes the silent-loss half of defect 1. Reuses slice 2's
   read.
4. **Slice 4 (flock + primitive relocation)** — hardening only. Drop if
   contentious.

Land 1 → validate → 2 → validate → 3 → validate. Do not batch: the whole point of
the ordering is that a regression in the version rules cannot be confused with a
regression in the write path.

---

## 6. Deferred findings (file as roadmap entries, do not fix here)

1. **`os.CreateTemp` mode regression in three existing helpers** —
   `internal/roadmap/rawio.go`, `internal/workflow/stamp.go`,
   `internal/brownmap/write.go` all replace a 0644 file with a 0600 one. Real,
   pre-existing, out of scope.
2. **`doctor` cannot repair a non-role `*.json.tmp`** — §1.6. Our naming
   sidesteps it, but the latent trap remains for any future writer that picks
   the obvious name.
3. **No `doctor` check for a future-version state file** — the operator only
   learns at the point of refusal.
4. **`hook_autostart` creates a duplicate workflow when the only state file
   fails to load** — §1.4 step 4. Independent of this feature; a corrupt file
   triggers it today.
5. **`internal/roadmap` / `internal/brownmap` still hand-roll atomic writes** —
   the brief scopes them out; `WriteFileAtomic` makes the follow-up cheap, but
   the layer rule has to be settled first.
