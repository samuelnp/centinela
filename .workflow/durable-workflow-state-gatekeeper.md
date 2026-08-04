### Adversarial Verifier Report: durable-workflow-state
**Date:** 2026-08-04
**Status:** WARNING

#### Inputs Read

- `git diff main...HEAD` (50 files, +4215/-56) plus the uncommitted working tree
  (6 modified, 4 new source/test files — the fixes from the previous round).
- `docs/features/durable-workflow-state.md` — the brief, i.e. the contract I
  judged "complete" against.
- `specs/durable-workflow-state.feature` (14 scenarios, current text).
- `docs/plans/durable-workflow-state.md`.
- Shipped source read line-by-line: `internal/workflow/{state.go,state_io.go,
  state_future.go,schema_version.go,version_probe.go,state_cas.go,
  atomic_write.go,active.go,stamp.go}`, `internal/hookpolicy/{prewrite.go,
  prewrite_actors.go,applypatch.go}`, `cmd/centinela/{hook_prewrite.go,
  hook_prewrite_block.go,hook_autostart.go,hook_workflows.go,complete.go}`,
  `internal/ui/render_stale_binary.go`, `.gitignore`,
  `internal/gates/spec_traceability*.go`.
- Tests read for substance (never accepted as proof):
  `internal/workflow/{state_io_crash_test.go,state_delete_conflict_test.go}`,
  `internal/hookpolicy/{prewrite_future_test.go,prewrite_degraded_scope_test.go}`,
  `tests/unit/durable_workflow_state_future_test.go`,
  `tests/acceptance/durable_workflow_state_{version,write,legacy}_test.go`.
- `.workflow/roadmap.json`, `.workflow/durable-workflow-state-edge-cases.md`.
- The output of every command I ran myself (recorded below).

**Flagged, per the input contract:** the task prompt DID contain a narrative
summary of the implementation (a five-point "What the feature claims" list). I
treated it strictly as a list of claims to attack, not as evidence. Every verdict
below rests on source I read or a command I ran. I did not accept any role's
`.workflow/*.md` narrative as proof; the edge-cases artifact was read only to
harvest deferral slugs and to check its assertions against observed behavior.

#### Analyzed Specs

`specs/durable-workflow-state.feature` — all 14 scenarios, across four groups:
atomic/durable writes (5), schema version (5, including the two that replaced the
single ambiguous "does not block file writes" scenario), concurrent writers (3),
and the never-loaded exemption (1). Each was checked against observed binary
behavior; the mapping is in the Refutation Attempts section. Acceptance-tier
coverage of these scenarios is Finding 3.

#### Refutation Attempts

**A. Permissiveness vs `main` — the central question.**
Built `/tmp/centinela-main` from a `git archive main` extract and
`/tmp/centinela-v11` from the current tree, then drove `hook prewrite` on both
across 27 shapes (a code target, a plan-doc target, and a `.workflow/*.json`
target). Result: **there is no shape in which this branch is more permissive than
main.** Every divergence is in the strict direction — main returns `NeedInit`
where the branch returns `StaleBinary`; both exit 2:

| shape | branch | main |
|---|---|---|
| real wf @plan / @code | OUTOFSTEP / ALLOW | same |
| only degraded (`v99`, `steps` array) | **STALEBIN** | NEEDINIT |
| only degraded (`schemaVersion:"2.0"`) | **STALEBIN** | NEEDINIT |
| degraded + real @plan | OUTOFSTEP (names `real`) | OUTOFSTEP |
| degraded + real @code | ALLOW | ALLOW |
| degraded beside a `done` wf | **STALEBIN** | NEEDINIT |
| two degraded files | **STALEBIN** | NEEDINIT |
| `outcome`-profile wf, ± degraded | ALLOW | ALLOW |
| degraded newest / oldest (mtime order both ways) | OUTOFSTEP | OUTOFSTEP |
| parsable `v99` @plan / @code | OUTOFSTEP / ALLOW | same |
| `v99` with `feature` / `currentStep` not a string | **STALEBIN**, names `delta` (the file's own name) | NEEDINIT |
| `v99` with `feature` != basename (phantom) | NEEDINIT | NEEDINIT |
| garbage non-JSON / truncated JSON | NEEDINIT | NEEDINIT |
| real wf + future-schema `roadmap.json` | OUTOFSTEP | OUTOFSTEP |
| no state file | NEEDINIT | NEEDINIT |

The specific bypass the code comments name — "write a future version over your
own state file and the gate opens" — is **closed**: I wrote a degraded file over
the only workflow and the code write was refused, exit 2.

**B. `hook autostart` duplicate fork.** Drove the real `hook autostart` with a
matching prompt. Degraded-only repo: the branch creates **nothing**; main forks
`payment-gateway.json` alongside `delta.json`. The no-duplicate claim holds for
the degraded path and is a real improvement over main. (Non-object corruption
still forks on both — see Deferred Findings.)

**C. Did the refusal break legitimate use?** The rendered panel is actionable:
headline `Next action: run `centinela update`, then retry the write`, and
`centinela update --help` confirms the command exists and is real (release
resolution, SHA256SUMS verification, atomic binary replace). The panel names the
feature and the file, never says "no workflow", and never suggests
`centinela start`. `centinela start <new>` still works from a stale-binary repo
(the escape hatch is intact) and grants no more permission than on main — I
checked: after `start`, a code write is still blocked at the new workflow's plan
step, on both binaries.

**D. Full `schemaVersion` token matrix, driven through `centinela route set`.**
absent / `null` / `0` / `1` / `-1` → save succeeds, file stamped
`"schemaVersion": 1` as the first key. `2` / `99` → exit 1, FUTURE refusal, file
**byte-unchanged**, unknown field preserved. `"1"` / `"2.0"` / `1.5` / `1.0` /
`true` / `[]` / `{}` / `99999999999999999999` → exit 1, UNREADABLE refusal, file
byte-unchanged. The guard fails **closed** on every unreadable token. (`-1` is
accepted as "lower" and normalised to 1 — nonsense input, no data at risk; noted
only for completeness.)

**E. Backward compatibility, end to end.** Versionless legacy file → `route set`
→ `route set`: both succeed, `schemaVersion: 1` lands first, `startedAt`, both
pinned contracts and the prior route all preserved. A fresh `centinela start` on
an empty repo writes a stamped v1 file. No migration is asked of the operator.

**F. Atomicity / durability, driven as the real binary.**
- Unwritable `.workflow/` (0500): exit 1, the error names `.workflow/f.json` (the
  target, not the temp), source file intact, **zero** leftover temps.
- State path is a directory: exit 1, clean error, zero temps.
- 0400 read-only state file: save succeeds, result `-rw-r--r--`.
- `umask 077`: result still `-rw-r--r--` — the explicit chmod does its job.
- Symlinked state FILE: replaced, not followed — link gone, original target
  untouched. Matches the documented, deliberate behavior.
- Symlinked `.workflow/` DIRECTORY: save succeeds, link preserved, bytes land in
  the resolved directory — the documented workaround genuinely works.
- Kill-before-rename: the `beforeRename` seam plus a re-exec'd `os.Exit` child in
  `state_io_crash_test.go` is a real proof — it asserts the target is
  byte-identical and still parses, and that the seam actually fired (exit-3
  sentinel), not merely that the test did not crash.
- Temp naming `.<base>.tmp-<rand>` matches neither `.workflow/*.json` nor
  `.workflow/*.json.tmp`, and `.gitignore` now covers it.

**G. CAS / concurrency, driven as the real binary.**
- 24 fully-overlapping `route set` processes on one file: 12 exit 0, 12 refused
  with "changed on disk since this command read it", final file parses, zero
  temps. CAS is genuinely armed.
- 5 concurrent `route set` for distinct roles: 2 were legitimate writes (the
  other 3 are domain rejections — confirmed by re-running them serially, where
  they fail identically), and **both exited 0 while only one route survived** —
  one silently lost update, reproduced. See Finding 2.
- Delete-between-load-and-save: `state_delete_conflict_test.go` and the
  `tests/unit` counterpart both assert refusal and non-resurrection; the
  never-loaded exemption is separately pinned.
- **Does CAS break `centinela complete`?** `complete` Loads, runs the whole
  validate gate for minutes, then Saves — the widest possible window. I recorded
  `.workflow/durable-workflow-state.json`'s mtime before and after a full
  `centinela validate` (485 s, full suite + gates): **unchanged**. Nothing in a
  gate run writes the state file, so CAS cannot spuriously refuse a `complete`.
  There are only three `workflow.Save` call sites (`start`, `hook_autostart`,
  `route_set`) plus `complete`'s `saveWorkflow`, and none double-loads.

**H. Do artifacts match behavior?** I re-derived every comment in the named files
against observed behavior rather than assuming the previous round's fixes landed
correctly:
- `state_future.go` — the degraded contract now says SKIP/REFUSE and matches
  observation. Its "note the boundary precisely" paragraph correctly scopes the
  no-brick guarantee to **unparseable** bodies. The salvage rules are verified by
  probe (shapes 16/17: `feature` falls back to the file's own name; the panel
  showed `delta`). "Save REFUSES" is verified structurally: a degraded workflow
  can only exist when the version is future or unreadable, both of which Save
  refuses.
- `schema_version.go` — the migration contract matches the matrix in D exactly,
  including the deliberately-stated forward-only limitation.
- `state_cas.go` — the measured-numbers paragraph is directionally confirmed by
  my own runs and, unusually, is honest about what it does NOT guarantee.
- `atomic_write.go` — every claim (EXDEV, random suffix, dot prefix, doctor glob,
  symlink replacement, errors naming the target) verified by probe or by reading
  the globs.
- `prewrite.go`, `prewrite_actors.go`, `hook_prewrite_block.go`,
  `render_stale_binary.go` — all match observed behavior. The test that
  previously asserted the opposite is renamed and inverted correctly
  (`TestUnmodellableWorkflowRefusesAndNamesTheRemedy`), and so is the acceptance
  test (`TestAccUnmodellableFutureVersionRefusesAndNamesTheUpgrade`).
- `specs/durable-workflow-state.feature` — the old "does not block file writes"
  scenario is correctly split into the modellable (allow) and unmodellable
  (refuse) cases; both match observed behavior.
- I grepped every changed file for surviving "never block / must be allowed /
  does not block / ALLOWS" language: **none left** in shipped source, spec, or
  tests.
- Deferral slugs cited in shipped source: exactly one —
  `close-state-save-toctou` at `internal/workflow/state_cas.go:49`. It is **NOT**
  in `.workflow/roadmap.json`. See Finding 1.
- `.workflow/durable-workflow-state-edge-cases.md` still maps a probe to the spec
  scenario *"A future-version state file does not block file writes"*, which no
  longer exists; `docs/plans/…md` names `tests/acceptance/durable_workflow_state_test.go`,
  which does not exist. See Finding 4.

**I. File-size gate, FULL SCAN.** Ran a full-repo scan independently of the
diff-aware gate: 30 `.go` files exceed 100 lines and **every one is under
`tests/`** (the exempt tier). Zero files in `internal/` or `cmd/` exceed the cap,
including `_test.go` files.

#### Commands Run

All run from the worktree root
`/Users/samuelnp/projects/personal/centinela/.worktrees/durable-workflow-state`.
Exit codes were captured directly with `; echo EXIT=$?` or `$?` per process,
never inferred from banner text. The long-running commands were instrumented with
`date +%s`; the short probes were run in batches and are recorded with their exit
codes and an approximate wall time.

| # | argv | exit | duration |
|---|------|------|----------|
| 1 | `git rev-parse HEAD` | 0 | <1 s |
| 2 | `git status --porcelain` | 0 | <1 s |
| 3 | `git diff --stat main...HEAD` | 0 | <1 s |
| 4 | `git diff` (uncommitted) | 0 | <1 s |
| 5 | `centinela validate` | **0** | **485 s** |
| 6 | `go build -o /tmp/centinela-v11 ./cmd/centinela` | 0 | ~9 s |
| 7 | `git archive main \| tar -x -C <scratch>/maintree` | 0 | ~2 s |
| 8 | `go build -o /tmp/centinela-main ./cmd/centinela` (in maintree) | 0 | ~11 s |
| 9 | `bash <scratch>/diffattack.sh` — 27 shapes × 2 binaries of `<bin> hook prewrite` | 0 (driver); per-shape codes tabulated above | ~30 s |
| 10 | `<bin> hook autostart` × 8 (4 shapes × 2 binaries) | 0 each | ~5 s |
| 11 | `<bin> hook prewrite` self-disarm / escape matrix × 8 | 2 and 0 as tabulated | ~5 s |
| 12 | `/tmp/centinela-v11 route set …` × 15 (schemaVersion token matrix) | 0 ×5, 1 ×10 | ~15 s |
| 13 | `/tmp/centinela-v11 route set …` durability probes (unwritable dir, directory target, 0400, umask 077, symlinked file, symlinked dir) | 0 / 1 as tabulated | ~8 s |
| 14 | `/tmp/centinela-v11 route set …` × 24 concurrent | 12 ×0, 12 ×1 | ~12 s |
| 15 | `/tmp/centinela-v11 route set …` × 5 concurrent, then × 5 serial | as tabulated | ~6 s |
| 16 | `/tmp/centinela-v11 start greenfield-thing` | 0 | ~1 s |
| 17 | `/tmp/centinela-v11 update --help` | 0 | <1 s |
| 18 | `centinela --version` (0.55.6) and `/tmp/centinela-v11 --version` (dev) | 0 | <1 s |
| 19 | full-scan `find . -name '*.go'` + per-file `wc -l` (G1 full scan) | 0 | ~2 s |
| 20 | `python3` re-implementation of the spec-traceability matcher over the changed specs | 0 | <1 s |
| 21 | `python3` roadmap slug lookups over `.workflow/roadmap.json` | 0 | <1 s |
| 22 | `stat -f '%Sm %N' .workflow/durable-workflow-state.json` (before and after validate) | 0 | <1 s |
| 23 | `centinela artifact new durable-workflow-state gatekeeper --force` | 0 | <1 s |
| 24 | `centinela evidence init durable-workflow-state gatekeeper` | 0 | <1 s |

`centinela validate` was run **exactly once**, per the mandate; its
`Validate Commands` section ran the full suite
(`go test ./... -coverprofile=coverage.out`), the coverage check, and the fmt
check, so the suite was not run a second time. The known `tests/acceptance`
10-minute-timeout flake did **not** occur: the single attempt completed at 485 s
with exit 0, so there is no second attempt to report. Verbatim output:

```
Built-in Gates (diff-aware: 56 files changed since main)
✓ G1: File Size  All in-scope files are within the 100-line cap (per-file exceptions are configurable under [[gates.file_size_exceptions]]).
✓ G-Build: Cross-Compile  All 6 release targets compile.
⚠ import_graph  Packages match no configured layer:
⚠ spec-traceability-gate  Scenarios without acceptance coverage:
✓ roadmap_drift  ROADMAP.md is in sync.
✓ docstring-gate  All 14 exported identifiers across 13 changed Go file(s) are documented.

Validate Commands
✓  go test ./... -coverprofile=coverage.out  go test (non-verbose) carries no skip data — add -json or -v to make skips detectable
✓  COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh
✓  ./scripts/check-fmt.sh

 🛡️👁️  CLI  All gates passed.
EXIT=0
DURATION_S=485
```

The `import_graph` warning is pre-existing and unchanged: `centinela.toml` is not
touched by this branch, no new package is introduced, and the config's own
comments document the unmapped-package warning as expected. The
`spec-traceability-gate` warning is new and is Finding 3.

Scratch artefacts (both binaries, the `main` archive extract, the attack repos)
live only under `/tmp` and the session scratchpad. `git status --porcelain` after
all probing shows exactly the six modified and four new source files of this
feature plus this report — nothing I invented.

#### Findings

**1. WARNING — a deferral slug cited in permanent shipped source has no roadmap
record.** `internal/workflow/state_cas.go:49` tells the next reader that the
remaining write-write race is "(deferred: `close-state-save-toctou`)". That name
does not exist in `.workflow/roadmap.json` — features there are keyed by `name`,
and I enumerated all of them. The pointer in shipped source is dangling, so the
only trackable record of a *reproduced* data-loss window is a code comment and a
step artifact that will never be scheduled. Six further deferrals named in
`.workflow/durable-workflow-state-edge-cases.md` are equally absent (listed under
Deferred Findings). The fix is one roadmap record per slug; I did not create them,
per instruction.

**2. WARNING — feature-brief outcome #2 is only partially delivered, and the
shortfall is reproducible.** The brief states: *"Concurrent writes do not
silently lose updates. Either serialize writers or detect the conflict and fail
loudly. Silently losing a routing decision or a step advance is the outcome to
eliminate."* Two concurrent `centinela route set` processes writing different
roles to the same file **both exited 0 and only one route survived** — a silently
lost routing decision, exactly the named outcome. What ships closes the
*stale-read* window (minutes wide; the one `complete` opens) and closes it well
(12 of 24 refused under contention). It does not close the *overlapping-write*
window (marshal → CreateTemp → write → fsync → rename). `state_cas.go`'s doc
comment says so plainly and the plan parks it as optional slice 4, so this is
disclosed rather than concealed — but "complete and correct" against the brief as
written is an overstatement, and Finding 1 means the remainder is untracked. This
is not a regression: `main` loses **every** such update and detects nothing.

**3. WARNING — three of this feature's own scenarios have no acceptance-tier
coverage, and the gate that says so renders no detail.** `spec-traceability-gate`
warns with the bare headline "Scenarios without acceptance coverage:" and drops
its per-scenario detail lines (`reportTraceability` fills `r.Details`; the
renderer never prints them). I re-implemented the gate's matcher
(`// Acceptance: specs/<slug>.feature` plus a standalone `// Scenario: <name>`
over `tests/acceptance`, whitespace/case/trailing-dot normalised) and computed the
gap for the only changed spec — this feature's:

- *A killed write leaves the previous state intact*
- *A stale save is refused rather than silently overwriting a newer one*
- *Re-running a refused command after the conflict succeeds*

These are precisely the three the brief singles out ("Tests must actually
exercise the failure modes — a killed write and a concurrent write — rather than
asserting the happy path twice") and that the plan's own §4 titles "Two tests
that must be real, not theatre". They ARE tested, and tested well, at the
colocated tier (`state_io_crash_test.go`, `state_cas_test.go`) and the
integration tier (`durable_workflow_state_cas_integration_test.go`) — so this is
a tier-placement and traceability gap, not an untested claim. Severity is `warn`
by config so it does not block; I record it because the rendered output makes it
invisible and CLAUDE.md's tests-step contract asks for executable acceptance
artifacts.

**4. WARNING — two step artifacts point at names that no longer exist.**
`.workflow/durable-workflow-state-edge-cases.md` (the required edge-cases
artifact, ~line 143) maps a probe to the spec scenario *"A future-version state
file does not block file writes"*, renamed to *"A future-version state file this
binary can still model keeps governing"* — the traceability row is dangling.
`docs/plans/durable-workflow-state.md` (~line 345) names the acceptance file
`tests/acceptance/durable_workflow_state_test.go`, which does not exist; the
suite shipped as four files (`_legacy_`, `_start_`, `_version_`, `_write_`).
Neither affects behavior. I separately verified the edge-cases artifact's
*behavioral* claims against the shipped binary and they hold — in particular its
"hook prewrite exits 0" line describes a **well-typed** `v99` file whose body
parses, which I confirmed still exits 0; it is not a stale claim about the
degraded path. Its R15 entry (stamp.go hand-rolls a weaker write) describes a
defect that was subsequently fixed in this same branch.

**No CRITICAL finding.** I tried to break the three load-bearing claims and could
not: (a) I found no shape where this branch is more permissive than `main`;
(b) the refusal is actionable, names a command that exists, never claims "no
workflow", and does not fork a duplicate; (c) atomicity, the version matrix and
backward compatibility hold under every probe, including unwritable directories,
symlinks, directory targets and a killed write.

#### Deferred Findings

Slugs named in this feature's artifacts or in shipped source that are **absent
from `.workflow/roadmap.json`** and should be recorded (I did not run
`centinela roadmap defer`, per instruction):

1. `close-state-save-toctou` — **cited in shipped source**
   (`internal/workflow/state_cas.go:49`). A short flock held across re-read →
   rename, closing the overlapping-write window. Measured here: 1 of 2 concurrent
   legitimate writers silently lost. *Highest priority — a permanent code comment
   points at it.*
2. `autostart-duplicate-on-unloadable-state` — a state file that is not a JSON
   object at all (truncated, non-JSON) still empties the active set and lets
   `hook autostart` fork a duplicate workflow. Confirmed live on **both**
   binaries, so pre-existing and not a regression; the new degraded path covers
   only JSON objects carrying a state marker.
3. `doctor-check-future-schema-version` — `status`, `status-all` and `verdict`
   render a `v99` file with no warning; the operator only learns at the point of
   refusal, and `verdict`'s JSON carries no schema version for CI consumers.
4. `doctor-detect-corrupt-state-file` — `doctor` reports the workflow-state check
   green for a truncated state file; the only signal is a stderr warning.
5. `preflight-schema-version-in-complete` — `complete` refuses a future-version
   file only *after* `runValidateGates`, costing a full multi-minute gate run
   before an instant-to-compute refusal.
6. `atomic-writes-for-roadmap-and-brownmap` — both still hand-roll the write;
   `WriteFileAtomic` exists now, but the layering question is unsettled.
7. `doctor-repair-non-role-json-tmp` — `orphanedTmps()` globs `*.json.tmp` while
   `evidence.Repair` globs `<feature>-*.json.tmp`, so a non-role temp makes the
   check red forever with a repair that removes nothing. This feature's naming
   sidesteps the trap; the trap remains for the next writer.

Already tracked, no action needed: `workflow-state-file-disarms-its-own-gates`
(an agent can still clear its own governance by writing `"currentStep":"done"` —
I reproduced it; identical on `main`) and
`acceptance-tier-exceeds-default-test-timeout` (did not fire on this run).

Implemented in this branch, so no roadmap record needed:
`load-probes-version-before-unmarshal`,
`save-refuses-on-deleted-target-when-loaded`, `stamp-report-atomic-write`.

#### Recommendation

**WARNING — may proceed.** The engineering is sound and, on the question that
mattered most, decisively so: driven as a real binary against a real `main` build
across 27 shapes, this branch is never more permissive than `main`, closes the
self-service bypass it claims to close, stops `hook autostart` forking a
duplicate off an unreadable state file, fails closed on every unreadable version
token, and survives kills, unwritable directories, symlinks and directory targets
without leaving a partial file. The previous rounds' inverted comment, test name,
acceptance test and spec scenario are all correctly fixed — I checked each against
observed behavior rather than trusting that the fix was right.

The claim "complete and correct" is nonetheless overstated on one axis: brief
outcome #2 asks that concurrent writes not *silently* lose updates, and I
reproduced a silent loss between two overlapping writers. That would be
acceptable as a disclosed, deferred limitation, except the slug the source points
at does not exist in the roadmap, so the remainder is untracked. Before merge:

1. Create the roadmap record for `close-state-save-toctou`, and ideally the six
   other slugs listed above, so the shipped comment stops dangling.
2. Either add acceptance-tier coverage for the three scenarios in Finding 3 or
   consciously accept the warn — they are the feature's headline failure modes.
3. Optional: correct the two stale names in Finding 4.

None of these is a correctness defect in shipped behavior, which is why this is
WARNING and not CRITICAL.

```json centinela:verification
{
  "revision": "9d924433fdc4cae462665c50e1227027a6d1f1b4",
  "treeDigest": "sha256:f546f3daaebf0ee5484952f18a9c5288d9f4047fb1f41a6d9b41c4f5275613eb",
  "commands": [
    {"argv": ["centinela", "validate"], "exitCode": 0, "durationMs": 485000},
    {"argv": ["go", "build", "-o", "/tmp/centinela-v11", "./cmd/centinela"], "exitCode": 0, "durationMs": 9000},
    {"argv": ["go", "build", "-o", "/tmp/centinela-main", "./cmd/centinela"], "exitCode": 0, "durationMs": 11000},
    {"argv": ["/tmp/centinela-v11", "hook", "prewrite"], "exitCode": 2, "durationMs": 120},
    {"argv": ["/tmp/centinela-main", "hook", "prewrite"], "exitCode": 2, "durationMs": 120},
    {"argv": ["/tmp/centinela-v11", "hook", "autostart"], "exitCode": 0, "durationMs": 130},
    {"argv": ["/tmp/centinela-main", "hook", "autostart"], "exitCode": 0, "durationMs": 130},
    {"argv": ["/tmp/centinela-v11", "route", "set", "f", "senior-engineer", "balanced", "--reason", "config-only change"], "exitCode": 0, "durationMs": 150},
    {"argv": ["/tmp/centinela-v11", "route", "set", "f", "senior-engineer", "balanced", "--reason", "config-only change"], "exitCode": 1, "durationMs": 150},
    {"argv": ["/tmp/centinela-v11", "start", "greenfield-thing"], "exitCode": 0, "durationMs": 400},
    {"argv": ["/tmp/centinela-v11", "update", "--help"], "exitCode": 0, "durationMs": 60},
    {"argv": ["git", "status", "--porcelain"], "exitCode": 0, "durationMs": 60},
    {"argv": ["centinela", "artifact", "new", "durable-workflow-state", "gatekeeper", "--force"], "exitCode": 0, "durationMs": 80},
    {"argv": ["centinela", "evidence", "init", "durable-workflow-state", "gatekeeper"], "exitCode": 0, "durationMs": 80}
  ]
}
```
