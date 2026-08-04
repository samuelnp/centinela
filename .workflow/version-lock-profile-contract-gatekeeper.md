# version-lock-profile-contract — gatekeeper

### Adversarial Verifier Report: version-lock-profile-contract
**Date:** 2026-08-04
**Status:** WARNING

Verified content is `a282f9ee338bb394268d1d9e23b70d31a6992d25` **plus two uncommitted
modifications** (`internal/workflow/state_version_lock_test.go`,
`internal/workflow/state_version_compat_test.go`) — that is the tree `centinela
validate` ran against and the tree the stamp digests. Nothing else was dirty.

#### Inputs Read

- `internal/workflow/state_version_lock_test.go` (the hotfix — golden list, `workflowJSONFields`, justification)
- `internal/workflow/state_version_compat_test.go` (the legacy round-trip canary)
- `internal/workflow/state.go` (the `Workflow` struct the lock describes)
- `internal/workflow/schema_version.go` (`SchemaVersion`, migration contract)
- `internal/workflow/state_io.go` (`Load`/`Save` — the premise the lock rests on)
- `internal/workflow/profile_contract.go` (`UsesGuidedDefault`, blast radius of a dropped pin)
- `tests/acceptance/durable_workflow_state_version_test.go`
- `.workflow/version-lock-profile-contract-{senior-engineer,qa-senior,edge-cases}.md`
- `.workflow/version-lock-profile-contract.json`, `centinela.toml`
- git history: all 138 tags, `v0.55.6` / `v0.56.0` / `v0.56.1` trees

#### Analyzed Specs

None exist for this slug and none are owed: `hotfix` archetype (code → tests →
validate), so no plan file and no `.feature`. The executable contract is
`TestWorkflowStructFieldsAreVersionLocked` plus
`TestLegacyStateRoundTripsWithEveryFieldIntact`; both were attacked directly.

#### Refutation Attempts

Every attempt below was executed, not reasoned about. Struct mutations were applied in
an `rsync` copy of the worktree under the session scratchpad; the worktree itself was
never mutated.

**1. "The lock is now complete — nothing can reach the state file and pass it."**
Two harnesses: a standalone probe replicating `workflowJSONFields` verbatim and
comparing it to real `json.Marshal` key order, and a mutation harness that edits the
REAL `Workflow` struct and runs the two tests.

| Shape planted on `Workflow` | Marshalled? | Lock verdict | Agreement |
|---|---|---|---|
| exported, no json tag (`Sneaky string`) | yes, as `Sneaky` | **FAIL** | correct (previous hole closed) |
| tag with options only (`json:",omitempty"`) | yes, as `Sneaky` | **FAIL** | correct (previous hole closed) |
| `json:"-"` | no | PASS | correct |
| **`json:"-,"`** | **yes, as key `-`** | **PASS** | ***MISMATCH — evades*** |
| **embedded unexported struct type w/ exported tagged field** | **yes, as `sneaky`** | **PASS** | ***MISMATCH — evades*** |
| embedded EXPORTED struct type | yes, as inner keys | **FAIL** | catches (message names the type, not the keys) |
| embedded `*Inner` pointer | yes, as inner keys | **FAIL** (probe) | catches |
| embedded exported struct WITH tag name | yes, as `nested` | PASS | correct |
| unexported field carrying a json tag | no | PASS | correct |
| two fields reordered | order changes | **FAIL** | catches |
| Go field renamed, tag unchanged | unchanged | compile error elsewhere | loud |

Verbatim proof of the two evasions, on the real struct:

```
=== MUTATION: embed-unexported ===
--- PASS: TestWorkflowStructFieldsAreVersionLocked (0.00s)
PROBE marshalled=[... revisions modelRoutes sneaky]
PROBE lockwalk  =[... revisions modelRoutes]

=== MUTATION: dashcomma ===
--- PASS: TestWorkflowStructFieldsAreVersionLocked (0.00s)
PROBE marshalled=[... revisions modelRoutes -]
PROBE lockwalk  =[... revisions modelRoutes]
```

`go vet` flags neither. So the promise the comment makes — "the list could drift from
the real marshalled shape exactly as the lock promises it cannot" — is still not fully
kept. See F1.

**2. "The golden list matches what the binary actually writes."**
CONFIRMED, and this is the strongest positive result. A fully populated `Workflow`
marshalled through `encoding/json` yields exactly, in order:
`schemaVersion feature startedAt currentStep steps stepOrder orchestrationMode
enforcementProfile archetype worktreePath driverModel validateContract planContract
profileContract revisions modelRoutes` — identical to `wantWorkflowFields` and to the
reflect walk. `profileContract` is recorded in the right position.

**3. "The justification is true."** Attacked against git, not against prose.
- `internal/workflow/schema_version.go` exists in **0 of 138 tags**. Version 1 has
  genuinely not shipped and no released binary carries the version check. TRUE.
- `state_io.go:89` re-marshals from the struct on every `Save`, so the premise ("an
  equal-version Save drops every on-disk key this struct does not model") is TRUE.
- The `v0.55.6` example is the operator's *installed* binary (`centinela --version` →
  0.55.6) and does drop both keys as claimed. It is two releases behind: `v0.56.0` and
  `v0.56.1` DO model `profileContract` and would drop only `schemaVersion`. The
  conclusion is unaffected (no tag has the check), so this is an observation, not a
  finding.
- Consistency: the new paragraph scopes `schema_version.go`'s rule and the `t.Fatalf`
  message ("once version 1 is released…"). It does **not** reconcile
  `fieldsLockedAtVersion`'s own doc, which still says the constant and the list are
  "updated together, never separately" — exactly what this commit does not do. See F4.

**4. "The canary covers `profileContract`."** REFUTED as a canary.
The row is `{"profileContract", after.ProfileContract, before.ProfileContract}` — two
values produced by the same struct through the same tag. Planting `json:"-"` on
`ProfileContract` removes the key from disk entirely and
`TestLegacyStateRoundTripsWithEveryFieldIntact` **still passes**; only the shape lock
fails. Same for a tag rename. The canary therefore adds no coverage the lock did not
already give. Removing the field outright is caught, but by a compile error, not by the
assertion. See F2.

**5. "Nothing else broke."** `workflowJSONFields`'s exportedness rule was checked
against every field currently on `Workflow`: 16 exported tagged fields, no embedded
fields, `loadedDigest`/`unmodellable` unexported and correctly skipped. No
mis-handling. Full suite green (below).

**6. Blast radius if the residual holes ever bite `profileContract` itself.** Dropping
the pin makes `UsesGuidedDefault()` false → resolution stays **strict**. Fails closed.

**7. Gate scope.** File-size gate re-run as a FULL SCAN (validate ran diff-aware over
60 files): **0 of 1622** `.go` files under `internal/` + `cmd/` exceed 100 lines. Both
changed files are 83 and 86 lines. Every `>100` file in the repo is in the `tests/`
tier, which the gate does not govern.

#### Commands Run

All from the worktree root
`/Users/samuelnp/projects/personal/centinela/.worktrees/version-lock-profile-contract`.
Exit codes captured directly with `; echo EXIT=$?`, never inferred from banner text.

| # | argv | exit | duration |
|---|------|------|----------|
| 1 | `go build -o /tmp/centinela-v13 ./cmd/centinela` | 0 | ~19s (wall clock) |
| 2 | `/tmp/centinela-v13 validate` | **0** | **493000 ms** (`date +%s` deltas, 1785836628→1785837121) |
| 3 | `go run .` (standalone shape probe, 11 evasion shapes vs `json.Marshal`) | 0 | ~2s |
| 4 | `bash scratchpad/mutate.sh` (13 mutations of the real struct × 2 tests) | 0 | ~120s |
| 5 | `go test ./internal/workflow/ -run 'TestProbeMarshalledKeys\|TestWorkflowStructFieldsAreVersionLocked' -count=1 -v` (scratch copy; 4 invocations: pristine + 3 mutants) | 0 | ~1s each |
| 6 | `go vet ./internal/workflow/` (scratch copy, both evasion mutants) | 0, no diagnostics | ~3s |
| 7 | `find internal cmd -name '*.go' \| xargs wc -l \| awk '$1>100'` (G1 FULL SCAN) | 0, empty output | <1s |
| 8 | `for t in $(git tag); do git show $t:internal/workflow/schema_version.go; done` | 0 → 0/138 hits | ~4s |
| 9 | `git status --short` (post-cleanup) | 0 | <1s |
| 10 | `/tmp/centinela-v13 artifact new … --force`, `evidence init`, `evidence append` ×18 | 0 | <1s each |

Durations for #1 and #3–#10 are tool wall-clock approximations; #2 is measured from
`date +%s` on both sides of the run.

`centinela validate` was run **exactly once**; its `[validate] commands` block runs the
full suite, so the suite was not re-run against the worktree. The `tests/acceptance`
10-minute-timeout flake did **not** occur (the whole validate finished in 493s); no
competing `go test`/`centinela` process appeared in `ps` during the run and no re-run
was needed. Verbatim tail:

```
Built-in Gates (diff-aware: 60 files changed since main)
✓ G1: File Size   ✓ G-Build: Cross-Compile   ⚠ import_graph   ⚠ spec-traceability-gate
✓ roadmap_drift   ✓ docstring-gate
Validate Commands
✓  go test ./... -coverprofile=coverage.out
✓  COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh
✓  ./scripts/check-fmt.sh
 🛡️👁️  CLI  All gates passed.
EXIT=0
```

The two `⚠` are advisory and pre-existing on this branch (`import_graph`: packages match
no configured layer; `spec-traceability-gate`: scenarios without acceptance coverage —
this hotfix archetype ships no `.feature` by design). Neither is caused by the hotfix.

Scratch artifacts (`/tmp/centinela-v13`, the rsync copy, the probe module, the mutation
script, `/tmp/validate-run1.log`) live outside the worktree; the scratch
`zz_probe_test.go` was deleted and the scratch `state.go` restored from its pristine
copy. `git status` after cleanup shows only the two hotfix test files (M) and this
report (??).

#### Findings

**F1 — WARNING — the shape lock's stated guarantee is still not airtight (two proven evasions).**
An embedded unexported struct type with an exported tagged field, and an exported field
tagged `json:"-,"`, both marshal into `.workflow/<feature>.json` while
`TestWorkflowStructFieldsAreVersionLocked` passes. This is the same failure class the
hotfix set out to close (a field reaches the state file yet passes the lock); the fix
closed the two likely shapes and left two exotic ones. Probability is low — nothing
embeds into `Workflow` today and `json:"-,"` is a deliberate escape hatch — but the
comment states the guarantee absolutely ("the list could drift … exactly as the lock
promises it cannot") and `-senior-engineer.md` restates it ("it cannot drift from the
real marshalled shape without failing"). Either weaken the wording or close it properly:
compare against the keys of `json.Marshal` on a fully populated `Workflow`, which I
verified equals the golden list exactly today and is immune to every shape in the table.
NOT blocking.

**F2 — WARNING — the round-trip canary's `profileContract` row is vacuous.**
It compares `after.ProfileContract` to `before.ProfileContract`, both read back through
the same struct, so it cannot observe the field ceasing to be persisted: with
`ProfileContract` retagged `json:"-"` the key vanishes from disk and the canary still
PASSES. The file's own doc claims it proves "nothing … may drop a field the operator's
file already carries" — it does not; the shape lock is what caught every drop I planted.
Assert against the literal in `legacyState` (`"guided-v1"`, better still
`ProfileContractGuidedDefault`). The weakness applies to all nine rows and is
pre-existing — but adding `profileContract` to this table was one of the three things
this round delivered, and it adds no coverage. NOT blocking.

**F3 — WARNING — role evidence describes the superseded implementation.**
`.workflow/version-lock-profile-contract-senior-engineer.md` was committed before the
current (uncommitted) revision and now misdescribes it: *Files Touched* lists only
`state_version_lock_test.go` while the tree also changes `state_version_compat_test.go`;
it states the file is "66 lines" (it is 83); and its Type-Safety note — "the golden list
stays `[]string` compared against reflected **struct tags**" — is precisely the
tag-based rule this round replaced with exportedness, while its "cannot drift" claim is
the one F1 refutes. `-edge-cases.md` repeats "derived from struct tags at run time". The
evidence record no longer matches the delivered diff. NOT blocking, but refresh it
before the docs step so the artifact trail is not wrong in merged history.

**F4 — LOW — one unreconciled statement of the strict rule.**
`fieldsLockedAtVersion`'s doc still reads "The two are updated together, never
separately", which this commit contradicts by construction (list updated, constant
held). The new paragraph reconciles `schema_version.go` and the `t.Fatalf` message by
scoping them to "once version 1 is released", but not this one; and a reader who lands
in `schema_version.go` first sees "Adding a field to Workflow without bumping this
constant is a silent data-loss release" with no caveat and no pointer to the exception.
One line in each place fixes it.

**Observation (not a finding)** — the comment's `v0.55.6` example is true and is the
operator's installed binary, but it is two releases old; `v0.56.1` models
`profileContract` and would drop only `schemaVersion`. The argument it supports survives
unchanged, since 0 of 138 tags carry the version check.

**Verified correct** — the substance of the claim holds: the golden list is identical to
real `json.Marshal` output including `profileContract` in the right position; holding
`SchemaVersion` at 1 is safe and a bump would buy nothing (no released binary has the
check); the previous round's exported-untagged and options-only holes are genuinely
closed (both now fail red); reorder, embedded-exported, tag-rename, tag-drop and field
removal all fail loudly; `json:"-"` and unexported-tagged correctly pass; the
exportedness rule mis-handles no field currently on the struct; and dropping the
`profileContract` pin fails closed (strict), never open.

#### Deferred Findings

Not filed with `centinela roadmap defer` (it would land on this branch and stale the
stamp). Listed here for the operator to file from the primary checkout after merge:

1. **`workflow-shape-lock-marshal-derived`** — replace the reflect walk in
   `workflowJSONFields` with (or cross-check it against) the key order of
   `json.Marshal` on a fully populated `Workflow`. Closes F1 for every shape, including
   ones nobody has thought of yet. Verified today that the two lists agree, so the swap
   is a no-op on the current struct.
2. **`legacy-roundtrip-canary-asserts-literals`** — make
   `TestLegacyStateRoundTripsWithEveryFieldIntact` compare against the literal values
   embedded in `legacyState` rather than against the pre-`Save` struct, so a field that
   stops being persisted is caught by the canary itself (F2).
3. **`state-shape-lock-ignores-field-types`** — the lock pins key NAMES only. Changing a
   key's Go type (e.g. `string` → struct) keeps it green while breaking the unmarshal in
   every binary that models the old type. Consider recording `name:kind` pairs.

#### Recommendation

**WARNING — ship.** The claim under test ("record `profileContract` in the golden field
list and hold `SchemaVersion` at 1") is correct, and I verified it against real
`json.Marshal` output and against all 138 release tags rather than against the prose.
`centinela validate` passes end to end (exit 0, 493s, one run). The four findings are an
incompletely-closed guarantee (F1), a test row that asserts less than its comment claims
(F2), stale role evidence (F3), and one unreconciled sentence (F4) — none of which makes
the delivered tree wrong or unsafe. Refreshing `-senior-engineer.md` (F3) is the one I
would do before the docs step; F1/F2 are the deferrals above.

```json centinela:verification
{
  "revision": "a282f9ee338bb394268d1d9e23b70d31a6992d25",
  "treeDigest": "sha256:55d1b699e0fe3c84000429d5da39e96a1aacd30495ad50f631bf285fb075f5d4",
  "commands": [
    {"argv": ["go", "build", "-o", "/tmp/centinela-v13", "./cmd/centinela"], "exitCode": 0, "durationMs": 19000},
    {"argv": ["/tmp/centinela-v13", "validate"], "exitCode": 0, "durationMs": 493000},
    {"argv": ["bash", "scratchpad/mutate.sh"], "exitCode": 0, "durationMs": 120000},
    {"argv": ["go", "test", "./internal/workflow/", "-run", "TestProbeMarshalledKeys|TestWorkflowStructFieldsAreVersionLocked", "-count=1", "-v"], "exitCode": 0, "durationMs": 1000},
    {"argv": ["go", "vet", "./internal/workflow/"], "exitCode": 0, "durationMs": 3000},
    {"argv": ["git", "status", "--short"], "exitCode": 0, "durationMs": 100}
  ]
}
```
