### Planner Report: durable-workflow-state

**Date:** 2026-08-03

#### Problem

`.workflow/<feature>.json` is the single file every Centinela gate consults to
decide what step a feature is on, which roles its contract requires, and — since
`dynamic-model-routing` — which model each role runs at. It carries two
data-integrity defects, both observed rather than theorised.

**1. The write is not atomic.** `internal/workflow/state.go` ends in a bare
`os.WriteFile(FilePath(wf.Feature), data, 0644)`. `os.WriteFile` opens
`O_WRONLY|O_CREATE|O_TRUNC`, so between truncate and write the file is **zero
bytes**. A process killed in that window destroys the feature's state outright:
`complete` cannot load it, and the workflow can be neither advanced nor rewound.
This session killed two long-running `complete` runs by accident (a background
job inheriting a two-minute timeout) and saw `main` briefly carry an unparseable
`.workflow/roadmap.json` from a sibling write path — the same failure mode, one
file over.

**2. Concurrent writers silently lose one side's update.** This is the same
defect's other face, and it needs stating precisely, because the shape of the gap
decides the mechanism. `complete` on the validate step loads at T0, runs
`runValidateGates` — a full profiled suite run, **minutes** per `centinela.toml` —
and saves at T+minutes. A `route set` in another terminal loads at T+1s and saves
at T+1s. At T+minutes `complete` writes its stale copy and the route is gone.
`route set` printed success, emitted telemetry, and changed nothing durable.

**3. There is no schema version.** An older binary unmarshals the file into a
struct lacking the newer fields and re-marshals it on the next save, silently
dropping them. `modelRoutes` is the live example, and the installed binary
routinely lags a worktree build in this project — so this is a normal operating
condition, not an exotic one. Nothing warns; the routing decisions simply vanish.

The blast radius is wide because `Load` is on the hot path of every file write:
`hook_prewrite` → `loadActiveWorkflows()` → `workflow.ActiveWorkflows()` →
`Load`, on every single tool call. Whatever this feature changes about loading is
changed for every write in the repo, which is why §Risks treats a botched `Load`
as the top hazard rather than an afterthought.

#### Scope

**In scope**

- `workflow.Save` becomes an atomic, durable replace: sibling temp file →
  `fsync` → `rename`, plus a best-effort directory `fsync`.
- A `schemaVersion` field, stamped on every save and checked at the mutating
  boundary. Absent means version 1; a higher version than this binary
  understands is refused with an actionable error.
- Concurrent-write detection: a stale save is **refused loudly**, never applied.
- The atomic-write helper is **exported** from `internal/workflow` so future
  adopters need no new import edge.
- Enough test coverage to actually exercise a killed write and a concurrent
  write, per the brief's explicit demand.

**Out of scope**

- Converting the other `.workflow/` writers (`internal/roadmap/rawio.go`,
  `internal/brownmap/write.go`, `internal/evidence/io_write.go`,
  `internal/workflow/stamp.go`). The helper is exposed; the conversions are
  separate work.
- Locking against a hostile writer. `.workflow/` is agent-writable by design;
  this is crash and concurrency safety, not tamper resistance.
- Retrofitting a version into the 132 state files already on disk. No migration
  step, no manual edit — absence is the migration.
- Any change to the meaning of a workflow field, the step order, or any gate.

**Deliberately *not* attempted:** serializing `complete` against `route set`.
See §Behavior Summary — the read-to-write gap is minutes long, so serialization
is the wrong tool and would create a worse operational problem than the one it
solves.

#### Dependencies & Assumptions

**Dependencies**

- None on other roadmap features. The brief states `Depends on: none` and the
  code confirms it: nothing in `internal/workflow`'s save path is under
  concurrent modification by another in-flight feature.
- `internal/evidence` already imports `internal/workflow`
  (`internal/evidence/io_write.go:8`). This one existing arrow is what makes the
  shared-helper decision possible without a cycle — the helper must live in
  `workflow`, and `evidence` may adopt it later for free.

**Assumptions, each checked against the tree**

| Assumption | Verification |
|---|---|
| The target file's mode is 0644 today | Confirmed: the literal in `Save`. Matters because `os.CreateTemp` yields **0600**. |
| `Save`'s signature must stay `func(*Workflow) error` | Confirmed: `cmd/centinela/complete.go:23` holds `var saveWorkflow = workflow.Save`, stubbed by tests; `revise.go:63` uses the same seam. |
| No test compares a saved workflow file byte-for-byte | Confirmed: no golden fixtures under `internal/workflow` or `cmd/centinela`. Adding a `schemaVersion` key therefore breaks nothing. |
| Every existing state file is versionless | Confirmed: `grep -L schemaVersion .workflow/*.json` → **132 of 132**. |
| `ActiveWorkflows` tolerates a `Load` error | Confirmed (`active.go:28-34`): it warns to stderr and `continue`s. This tolerance is precisely what turns a `Load`-level refusal into a repo-wide write block — see §Risks. |
| The temp file will not trip `doctor` | Confirmed by reading `internal/doctor/check_evidence.go:35` and `internal/evidence/repair.go:30`. The naming decision exists specifically to sidestep the mismatch between those two globs. |
| `centinela update` is the real upgrade command | Confirmed: `cmd/centinela/update.go:18`. It appears verbatim in the refusal message. |
| An in-process mutex would help | **False, and stated as such.** `route set` and `complete` are separate OS processes. A `sync.Mutex` in `internal/workflow` guards nothing that races. It must not ship and must not be mistaken for a fix. |
| Versioning protects both directions | **False, and stated as such.** It protects *forward only*. Binaries already installed carry no check and will still silently drop `modelRoutes` from files this release writes. The guarantee begins with the first release carrying the check on both sides. |

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| **A botched `Load` change bricks the prewrite hook.** A refusal inside `Load` makes `ActiveWorkflows` skip the feature → `loadActiveWorkflows()` returns empty → `hookpolicy.EvaluatePrewrite` returns `NeedInit` (`prewrite.go:32`) → **every governed write in the repo is blocked**. Worse, `hook_autostart.go:27` fires on the empty list and auto-starts a duplicate `<feature>-2`. | **Critical** — the repo becomes unwritable and grows a phantom workflow | Medium if the refusal is put in the obvious place | **Architectural, not procedural: the refusal lives in `Save`, never in `Load`.** `Load` stays permissive and merely exposes the version. A dedicated scenario asserts a `schemaVersion: 99` file does **not** block `hook prewrite`. |
| **The 132 tracked state files in this repo, plus every downstream project's, stop loading.** | **Critical** — every project on the new binary is bricked at once | Low, but only because the rule is explicit | Absent ⇒ version 1, enforced in `Load` and pinned by a compat test that round-trips a verbatim legacy fixture (versionless, with `modelRoutes` and `revisions`). No migration step exists to get wrong. |
| **A parallel session in this same repo is disrupted.** `.worktrees/guided-by-default` is live right now and shares `.workflow/`. | High — a colleague's feature stalls mid-workflow | Medium | Slices 1–3 are additive and back-compatible; nothing rewrites an existing file until its owner saves. The CAS refusal is a *refusal*, never a corruption — the worst case for the parallel session is one re-run. Land slice by slice with a full validate between, so a regression is attributable. |
| **The temp file makes `centinela doctor` permanently red.** `check_evidence.go:35` globs `*.json.tmp`; `evidence.Repair` globs `<feature>-*.json.tmp`. A temp named `<feature>.json.tmp` matches the first and not the second: doctor reports the repair applied and stays red **forever**. | High — an unfixable red gate that erodes trust in `doctor` | **High** if the obvious name is chosen | Temp is `.workflow/.<feature>.json.tmp-<rand>`, matching neither glob. Asserted by a test that runs `doctor`'s evidence check with a workflow temp present. |
| **Mode regression to 0600.** `os.CreateTemp` creates at 0600; three existing helpers in this tree already have this latent bug. Applied to the file every hook reads, it becomes owner-only. | High — breaks any multi-user or CI-agent setup | High if unguarded | `chmod 0644` on the temp before rename, with a dedicated scenario asserting the mode survives a replace. |
| **A deterministic temp name becomes a corruption vector.** Two writers share one temp path; one can rename the *other's* half-written bytes over the target — strictly worse than the bug being fixed. | Critical | Low (needs true concurrency) | Unique temp per process via `os.CreateTemp`. |
| **CAS false-positive at the end of a long `complete`.** A genuine concurrent `route set` makes a minutes-long validate run end in a refusal. | Medium — real frustration, no data loss | Medium | This is the *designed* behaviour and strictly better than silent loss, but the error text must earn it: it names the file, says another process wrote it, and tells the operator to re-run. A scenario asserts the re-run succeeds and both changes survive. |
| **Slice 4 touches `internal/evidence`.** Relocating the `flock` primitive drags a second package's coverage and tests along. | Medium | Medium | Slice 4 is last and independently revertible. If the verifier objects, drop it: slices 1–3 satisfy the brief on their own. |
| **`internal/workflow/state.go` is already 97 lines** against a hard 100-line cap. | Low — a mechanical gate failure | Certain if ignored | `Load`/`Save` move to `state_io.go` in slice 1, before any field is added. Planned, not discovered. |
| **Coverage gate.** Per-package with no `-coverpkg`; the `tests/` tier does not lift `internal/workflow`. | Medium — validate fails late | Medium | Every new behaviour gets a colocated `internal/workflow/*_test.go` (≤100 lines each). Aim ≥97% on the touched package, per the standing "exceed by ~2%" rule. |
| **Silent forward-only protection is mistaken for full protection.** | Medium — a false sense of safety | Medium | Stated explicitly in the plan and carried into the docs step: already-installed binaries have no check and will still drop fields. |

#### Rollout

Four slices, smallest correct one first. **Atomic write and versioning are
independently revertible** by construction — neither depends on the other.

1. **Slice 1 — atomic write.** New `internal/workflow/atomic_write.go`
   (`WriteFileAtomic`), `Load`/`Save` moved out of the 97-line `state.go` into
   `state_io.go`, `Save` rewired to the helper. **Zero file-format change**, so
   it ships and reverts with no compatibility consequence whatsoever. Revert is
   one line: point `Save` back at `os.WriteFile`. Fixes the truncation half of
   defect 1.
2. **Slice 2 — schema version.** `const SchemaVersion = 1`, the field as the
   *first* struct member (so `head -3` of any state file shows it), `Save`
   stamps it and refuses a higher on-disk version. `Load` untouched in its
   permissiveness. Independent of slice 1. Fixes defect 2. *Note on revert: the
   reverted binary drops the field silently — lossy in exactly the way this
   feature exists to prevent. Record that on the revert path.*
3. **Slice 3 — compare-and-swap.** An unexported `loadedDigest` on `Workflow`
   (invisible to `encoding/json`, so the on-disk shape and all 48 hand-written
   fixtures are untouched), set by `Load`, checked by `Save` against a re-read of
   the target — the *same* re-read slice 2 already performs. Empty digest (a
   `New()`-built workflow that was never loaded) skips the check, which is what
   keeps `start` and `hook autostart` working. Fixes the silent-loss half of
   defect 1.
4. **Slice 4 — TOCTOU close + one `flock` implementation.** Optional hardening.
   Move the build-tagged `flock` primitive from `internal/evidence` into
   `internal/workflow` (one implementation, no new edge, no signature change to
   `evidence.Lock`) and take a **microsecond** lock across `Save`'s re-read →
   rename. Never held across a gate run. Drop the whole slice if contentious.

Land 1 → validate → 2 → validate → 3 → validate. **Do not batch.** The point of
the ordering is that a regression in the version rules cannot be confused with a
regression in the write path.

#### Behavior Summary

**Writing.** `Save` marshals as today, then writes to a unique sibling temp
(`.workflow/.<feature>.json.tmp-<rand>`), chmods it to 0644, `fsync`s it, closes
it, and `rename`s it over the target — atomic within the filesystem, no `EXDEV`
because the temp is a sibling. It then `fsync`s `.workflow/` best-effort (opening
a directory for sync is not portable to Windows, and a failure there must never
fail a save that already succeeded). Any in-process error removes the temp. A
`SIGKILL` leaves one behind: a hidden dotfile that trips no glob and will be
gitignored via `.workflow/.*.json.tmp-*`, mirroring the existing
`.workflow/.roadmap-digest-*` precedent.

**Before `rename`, the reader sees the complete old file; after, the complete new
one. There is no window in which it sees neither.**

**Versioning.** `SchemaVersion int` carries `json:"schemaVersion"` with **no
`omitempty`**, so every file this binary writes is stamped.

| On-disk version | `Load` | `Save` |
|---|---|---|
| absent / `0` | succeeds; treated as **1** | stamps 1 |
| `== 1` | succeeds | stamps 1 |
| `> 1` (a newer binary wrote it) | **succeeds** — refusing here bricks the prewrite hook | **refused**, actionable error, file left byte-identical |
| `< current` (a future v2 reading v1) | succeeds | stamps the current version — a **silent one-way upgrade** |

The silent upgrade is correct for this codebase: every field added so far is
back-compat-by-absence (`Archetype`, `ValidateContract`, `PlanContract`,
`ModelRoutes`), so "absent means the documented default" *is* the whole
migration. A future v2 needing more than defaulting must add its own migration
next to the constant; that obligation is recorded in the constant's doc comment
now, not left to be rediscovered.

The version is read from the **on-disk bytes at save time**, not only from the
in-memory struct — the file may have been upgraded by a newer binary between our
`Load` and our `Save`.

**Concurrency — compare-and-swap, chosen over locking.** `Load` records
`sha256` of the raw bytes it read into an unexported field. `Save` re-reads the
target; if the digest differs, the write is refused. A workflow that was never
loaded has an empty digest and skips the check entirely.

The rejected alternatives, and why, because this is the decision most likely to
be second-guessed:

- **Serializing lock** — `complete` holds its loaded copy across a minutes-long
  gate run, so a lock that actually serializes would block `route set` for
  minutes (the existing `evidence.LockTimeout` is 2 seconds). Narrowing the lock
  to the write itself leaves the bug *completely* intact, because the stale data
  was read minutes earlier. A lock cannot solve a stale-read problem.
- **In-process mutex** — `route set` and `complete` are separate OS processes. It
  guards nothing.

A residual TOCTOU remains: two processes could both pass the CAS and both
`rename`, microseconds apart. Slice 4 closes it with a lock held only across
re-read → rename. The feature is correct and shippable without it, which is
exactly why that slice is last.

#### Gherkin Scenarios

Written in full to `specs/durable-workflow-state.feature` — **13 scenarios**,
complete now because `specs/` writes are blocked during the tests step.

*Atomic, durable writes (5)*

1. **A killed write leaves the previous state intact** — a process is killed
   after the replacement bytes are written but before the rename; the state file
   still parses and still records the old step, with no zero-byte or truncated
   file left behind.
2. **A completed write replaces the state file in one step** — the new step is
   recorded and no temporary file remains beside it.
3. **An abandoned temporary file is not mistaken for orphaned evidence** —
   `doctor`'s evidence check stays green with a workflow temp present. This is
   the §Risks trap, asserted.
4. **The replaced state file keeps its readable file mode** — still 0644 after a
   save.
5. **A write that cannot be completed reports the state file path** — the error
   names `.workflow/alpha.json` and no partial file is created.

*Schema version (5)*

6. **A versionless legacy workflow file loads unchanged** — no `schemaVersion`
   field, load succeeds, treated as version 1, no migration step required of the
   operator. This scenario stands for all 132 tracked files and every downstream
   project.
7. **A newly started workflow is stamped with the current schema version.**
8. **A same-version file round-trips without losing fields** — a recorded model
   route (the live example of a field an older binary drops) survives a
   load-and-save, and every other field is unchanged.
9. **A future-version state file is refused on save with an actionable message**
   — the error names the file, the version it carries, the version this binary
   understands, and that upgrading is the fix; the file is left byte-for-byte
   unchanged.
10. **A future-version state file does not block file writes** — with that file
    as the only workflow, the prewrite hook still allows a governed write and
    does not claim no workflow has been started. **This is the anti-bricking
    guard**; it is the scenario that pins the refusal-in-`Save` decision in place.

*Concurrent writers (3)*

11. **A stale save is refused rather than silently overwriting a newer one** —
    the `complete`-versus-`route set` case verbatim: the long-running command's
    save is refused, the error explains the file changed since it was read and
    says to re-run, and the model route is still present.
12. **Re-running a refused command after the conflict succeeds** — both the model
    route and the step advance end up in the file. The refusal must be
    recoverable, not a dead end.
13. **A workflow that was never loaded saves without a conflict check** —
    `start` and `hook autostart` keep working.

**Two tests must be real, not theatre** — the brief demands it. The killed write
is made deterministic with a helper subprocess: a package var `beforeRename`
(no-op in production) is set to `os.Exit(1)` in a re-exec'd child, which skips
deferred cleanup exactly as a `SIGKILL` would; the parent then asserts the target
still parses and still holds the old step. The concurrent write needs no
processes at all, because the *shape* is what matters: two `Load`s, the second
saves and wins, the first must be refused, and the second's route must survive.
That test fails against today's `Save`.

#### UX States

**`centinela route set` / `complete` / `revise` — future-version refusal (exit ≠ 0):**

```
.workflow/delta.json was written by a newer Centinela (schema version 99);
this binary understands schema version 1. Refusing to write — saving would
drop fields it does not know about. Upgrade with `centinela update` and re-run.
```

Names the file, the version it carries, the version we understand, and the fix —
the four things the brief requires, in that order.

**`centinela route set` / `complete` / `revise` — stale-write refusal (exit ≠ 0):**

```
.workflow/epsilon.json changed on disk since this command read it — another
centinela process wrote it (a concurrent `route set`, `complete`, or `revise`).
Refusing to write so that update is not lost. Re-run this command to apply
your change on top of the current state.
```

Every one of the five `Save` call sites already wraps the error (`cannot save
workflow: %w`) and surfaces it, so no command needs new plumbing.

**`centinela hook prewrite` — unchanged in every case.** A future-version or
conflicted state file produces **no hook output and no block**. This is the
designed non-event, and scenario 10 pins it.

**`centinela status` / `dashboard` / `verify` / `deliver` — unchanged.** All are
read-only `Load` callers; a future-version file renders normally.

**`centinela doctor` — unchanged.** No new check, and the evidence check stays
green with a leaked workflow temp present (scenario 3).

**Success paths print nothing new.** `start`, `complete`, and `route set` keep
their current output verbatim. Durability is not a feature the operator should
have to notice; the only new text in the product is a refusal.

#### Edge Cases

Recorded in full in `.workflow/durable-workflow-state-planner.json`
(`edgeCases`, 15 entries). The load-bearing ones:

- A `Load`-level refusal empties `ActiveWorkflows`, which makes `hook_prewrite`
  return `NeedInit` and block every governed write — and makes `hook_autostart`
  create a duplicate `<feature>-2`.
- A temp named `<feature>.json.tmp` matches `doctor`'s glob but not
  `evidence.Repair`'s, producing a permanently red check with a no-op repair.
- `os.CreateTemp` creates at 0600 against a 0644 target.
- A deterministic temp name lets one writer rename another's half-written bytes
  over the target.
- `complete` holds its read for the whole gate run, which is why a lock cannot
  work and an in-process mutex is meaningless across two processes.
- `Save`'s signature is pinned by the stubbed `saveWorkflow` package var.
- 132 tracked state files and 48 hand-written `currentStep` fixtures are all
  versionless: absent must mean 1.
- Versioning is forward-only; already-installed binaries still drop fields.
- `internal/evidence` importing `internal/workflow` fixes which package may own
  the shared helper.
- A `New()`-built workflow has an empty digest and must skip the CAS check.
- Directory `fsync` must be best-effort for Windows portability.

#### Out-of-Scope

- Converting `internal/roadmap/rawio.go`, `internal/brownmap/write.go`,
  `internal/evidence/io_write.go`, or `internal/workflow/stamp.go` to the shared
  helper. The helper is **exported** so the follow-up is cheap; `roadmap` and
  `brownmap` do not import `workflow`, so adopting there needs a PROJECT.md G2
  layer decision first.
- Tamper resistance. `.workflow/` is agent-writable by design.
- Retrofitting a version into files already on disk.
- Any change to `Load`'s permissiveness, any gate, any step order, or the meaning
  of any workflow field.
- A `doctor` check that reports a future-version state file (deferred below).
- Fixing the pre-existing 0600 mode regression in the three sibling helpers
  (deferred below).

#### Deferred Findings

1. **`os.CreateTemp` mode regression in three existing helpers.**
   `internal/roadmap/rawio.go:59`, `internal/workflow/stamp.go:39`, and
   `internal/brownmap/write.go:45` each replace a 0644 file with a 0600 one.
   Real, pre-existing, and out of scope here — but it is the same bug this
   feature guards against in the one file it owns.
2. **`doctor` cannot repair a non-role `*.json.tmp`.** `check_evidence.go:35`
   globs `*.json.tmp` while `evidence.Repair` globs `<feature>-*.json.tmp`; any
   temp not matching `<feature>-<role>` makes doctor permanently red with a
   repair that silently removes nothing. Our naming sidesteps it; the trap
   remains armed for the next writer who picks the obvious name.
3. **No `doctor` check for a future-version state file.** The operator only
   learns at the point of refusal. A read-only check would surface it at session
   start, where `doctor` already runs.
4. **`hook_autostart` creates a duplicate workflow when the only state file fails
   to load.** `hook_autostart.go:27` fires on an empty active list, and
   `ActiveWorkflows` empties on any load failure — so a *corrupt* file already
   triggers this today, independent of this feature.
5. **`internal/roadmap` and `internal/brownmap` still hand-roll atomic writes,
   neither with an `fsync`.** `roadmap.json` was seen unparseable on `main` this
   session — the same failure mode as defect 1, one file over. The layer rule
   has to be settled before `WriteFileAtomic` can be reused there.

#### Handoff

**Next role: senior-engineer.**

Start with **slice 1 only** and validate before touching slice 2. The three
things most likely to be got wrong, in order:

1. **Do not put the version refusal in `Load`.** It bricks the prewrite hook.
   The trace is in the plan, §1.4.
2. **Do not name the temp `<feature>.json.tmp`.** It makes `doctor` permanently
   red with a repair that does nothing. The trace is in the plan, §1.6.
3. **Chmod the temp to 0644 before the rename.** `os.CreateTemp` gives 0600.

`internal/workflow/state.go` is at 97 of 100 lines — move `Load`/`Save` out
*first*, before adding any field. `Save`'s signature must not change
(`cmd/centinela/complete.go:23` holds it in a stubbed package var). Colocated
`internal/workflow/*_test.go` files are the only thing that moves the coverage
gate; the `tests/` tier does not.

Artifacts: `docs/plans/durable-workflow-state.md` (file-by-file plan, caller and
fixture inventories, slice ordering) and `specs/durable-workflow-state.feature`
(13 scenarios, complete — `specs/` is write-blocked during the tests step).
