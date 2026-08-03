# Gatekeeper Report — docstring-gate (round 3)

**Status:** WARNING

Round 2's CRITICAL is genuinely fixed. I reproduced the `actions/checkout` ref
state myself — full-history clone, detached HEAD, local `main` deleted — and the
gate **enforces** there (exit 1, header `since origin/main`), including under the
real CI condition `CI=true` where `diff_mode = "auto"` resolves `ModeFull` and the
gate must self-resolve its own ratchet. I also removed the fallback by hand and
confirmed the replacement acceptance test **fails** without it, at both the
acceptance and unit tiers. The round-2 lesson — a test that could not see its own
claim — is not repeated.

What survives is one wrong-by-2 artifact number, one overstated brief claim with a
real but bounded cross-gate side effect, and four polish items. None of them
falsifies CI enforcement for this repository, so none is a blocker.

## Inputs Read

- `git diff origin/main...HEAD` and the uncommitted tree (`git status --short`)
- `docs/features/docstring-gate.md`
- `specs/docstring-gate.feature`
- `docs/plans/docstring-gate.md`
- `internal/gitdiff/{base.go,resolver.go,base_test.go}`
- `internal/gates/{docstring.go,docstring_report.go,docstring_scope.go,docstring_message.go}`
- `internal/docstring/{scan.go,report.go,filter.go,walk.go,member.go,directive.go,doc.go}`
- `internal/config/{docstring.go,validate_mode.go}`
- `cmd/centinela/{validate_mode.go,pr_gate.go}`
- `tests/acceptance/docstring_gate_{ci,ratchet,helper}_test.go`
- `.github/workflows/validate.yml`, `centinela.toml`
- Output of every command in **Commands Run**

## Refutation Attempts

1. **"CI enforcement is still a false claim."** *Refuted.* In my own reproduced CI
   ref shape the gate exits 1 and names `since origin/main`. Under the real CI
   condition (`CI=true` ⇒ `ModeFull` ⇒ header `Built-in Gates (full scan)`) the
   gate still self-resolves its changed set and reports `✗ docstring-gate`.
2. **"The fallback produces `origin/origin/main` on an already-qualified base."*
   *Refuted.* `diff_base = "origin/main"` resolves on the first attempt; `remoteRef`
   returns `""` for any base containing `/`. Enforced, exit 1.
3. **"A tag or SHA base breaks."** *Refuted.* Both resolve on the first attempt;
   headers read `since v1.0.0` and `since <sha>`, exit 1.
4. **"Total failure silently Passes."** *Refuted.* A base that exists nowhere
   degrades to `diff base "nonexistent" not found (also tried "origin/nonexistent")`
   and the gate reports **Skip**, never Pass. Same for a repo with no `origin`.
5. **"The fallback breaks a legitimately local-only base."** *Refuted.* A repo with
   no remotes at all still resolves `main` locally and enforces, exit 1.
6. **"Local and `origin/<base>` diverge and the unsafe one wins."** *Refuted.* Local
   wins (tried first). With local `main` behind `origin/main` that yields an older
   merge base and therefore a **wider** scope — the fail-closed direction — and the
   header honestly says `since main`.
7. **"The replacement test cannot see its own claim (the round-2 defect)."**
   *Refuted by mutation.* I edited `internal/gitdiff/base.go` to make `remoteRef`
   return `""`, and both `TestDG_CIRefShapeResolvesBaseAndEnforces` and
   `TestDG_CIRefShapeStillRatchets` failed with the exact round-2 symptom
   (`diff base "main" not found`); `internal/gitdiff` unit tests failed too
   (`TestChangedFiles_FallsBackToTheRemoteTrackingRef`,
   `TestChangedFiles_DegradeNamesEveryRefFormTried`). File restored, suite re-green.
8. **"Roots normalization is incomplete."** *Refuted for the stated forms.*
   `src`, `./src`, `src/`, `./src/`, `src/../src`, `.`, a nested root, and a
   duplicated root all enforce identically (one violation, no double count). An
   absolute root is a config error with exit 1. A root matching nothing Skips with
   an honest reason naming the root.
9. **"The ratchet opens legacy under the new fallback."** *Refuted.* CI ref shape
   with a real undocumented `func Legacy()` on the base commit: gate passes, exit 0,
   `across 1 changed Go file(s)`, legacy never named.
10. **"Invalid config fails open."** *Refuted.* `severity = "bogus"` and `"FAIL"`
    both exit 1 with a clear error; empty defaults to `fail` and enforces.
11. **"An unparseable-only scope is swallowed as a Skip."** *Refuted.* The Skip
    branch is guarded by `rep.Files == 0 && rep.OK()`, so a parse-error-only scope
    falls through to Fail/Warn and is reported.
12. **"`--full` is not report-only, or its counts are invented."** *Refuted.*
    `docs lint --full` exits 0 and reports `171 undocumented of 868 exported
    identifiers across 618 files`. My independent walk yields exactly **618**.
    (The count in `centinela.toml` does not — see Finding 3.)
13. **"Violations and parse errors are still conflated."** *Refuted.* They are
    counted and named separately, singular/plural agree, and a Warn ends by pointing
    at `centinela docs lint` rather than a colon introducing an unrendered list.
14. **"Scope silently widened somewhere, turning main red or hiding violations."**
    *Partly sustained, in the narrowing direction.* No consumer keys logic off
    `Summary.Base` (only the display header), and the required `validate` job is
    `ModeFull` in CI so every other gate still full-scans there. But `pr-gate` calls
    `ChangedFiles` unconditionally, and its CI scope narrows — see Finding 2.
15. **"The fallback covers every base-branch shape."** *Sustained.* It does not
    cover hierarchical branch names — see Finding 1.
16. **"Repo-escaping roots are validated like absolute ones."** *Sustained.* See
    Finding 5.
17. **"The new CI spec scenario's clauses match what its test asserts."**
    *Partly sustained.* Four of five clauses are genuinely asserted; the ratchet
    clause is not — see Finding 4.

## Commands Run

`validate` was executed exactly once, in the worktree, in the background
(wall clock derived from the log file's create/modify timestamps, 15:18:21 →
15:25:19). `go test ./...` was timed by `time`. Probe durations are wall-clock
approximations rounded to the second; each probe ran in its own throwaway repo
under `/tmp/dg3-*`, which is why the same argv appears more than once with
different exit codes — those are different working directories, not reruns.

Repeated argv, disambiguated:
- `docs lint --changed` exit 1 in `/tmp/dg3-probes/cloneA` — the reproduced CI ref
  shape, enforcing. Exit 0 in `cloneB` (no `origin`), `cloneC` (remote named
  `upstream`), `cloneG` (`release/2.0`) and `cloneH` (nonexistent base) — the four
  honest Skips.
- `go test ./internal/gitdiff/` appears twice: exit 1 with `base.go` mutated,
  exit 0 after restore.
- `pr-gate` appears twice on the identical repo `/tmp/dg3-blast/clone`: exit 1 with
  the pre-fallback binary (`/tmp/centinela-preFallback`, full repo scan, G1 fail)
  and exit 0 with the shipped binary (diff-aware, G1 pass). That pair is Finding 2.

`centinela evidence set/append`, `centinela evidence validate` and
`centinela artifact stamp` necessarily run *after* this file is written and so
cannot appear below; their exit codes are reported to the caller. `roadmap defer`
(×6) and `roadmap generate` ran before stamping and are recorded.

## Findings

### Finding 1 — MEDIUM — the `origin/<base>` retry does not apply to a hierarchical base branch

`remoteRef` treats **any** base containing `/` as already remote-qualified, so a
base branch named `release/2.0`, `develop/next` or `feature/x` gets no retry and
reproduces the round-2 failure exactly. Verified: in a CI-shape clone where
`refs/remotes/origin/release/2.0` exists, `diff_base = "release/2.0"` degrades with
`diff base "release/2.0" not found` and the gate Skips at exit 0 — and the notice
omits the `also tried` clause, so nothing hints a retry was skipped.

Not triggered by this repo's shipped `diff_base = "main"`, which is why CI
enforcement here is real. But `internal/gitdiff` is shared by every diff-aware
gate, `validate --changed`, `precommit`, `pr-gate` and `audit`. The correct
qualification test is whether the first path segment names an actual remote, not
whether the string contains a slash.

### Finding 2 — MEDIUM — `pr-gate`'s CI scope silently narrows, and the brief overstates the repair

`cmd/centinela/pr_gate.go` calls `ChangedFiles` unconditionally, independent of
`ResolveMode`. Before this change it degraded in CI and full-scanned the repo;
now it resolves and runs diff-aware. Isolated by differential build on one repo
(legacy 150-line file on the base commit, PR adds one documented file):

- pre-fallback binary: `centinela pr-gate: diff base "main" not found — full repo scan`, `| G1: File Size | ❌ fail |`
- shipped binary: `| G1: File Size | ✅ pass |`

So `pr-gate` PR comments stop surfacing pre-existing repo-wide violations. Main's
protection is intact — the required `validate` job resolves `ModeFull` in CI
(`diff_mode` unset ⇒ `auto`, `CI=true`) and still full-scans — which is also why
the brief's claim that the resolver change "repairs **every** diff-aware gate in
CI" is overstated: in this repo's CI no other gate was routing through the
resolver at all. The one gate it does change, it changes by *reducing* scope, and
that side effect is undocumented.

### Finding 3 — LOW — a round-3 "corrected" artifact claim is still wrong

`centinela.toml:70` reads `171 undocumented of 868 exported identifiers across
**616** non-test files`. The shipped scanner reports **618**, and my independent
walk (`find internal cmd pkg lib src app -name '*.go' -not -name '*_test.go'`
minus the exclusion set) yields exactly **618**. Round 3 correctly bumped
`867 → 868` for the newly exported `ExcludedDir` but did not bump `616 → 618` for
the two non-test files it added (`internal/gitdiff/base.go`,
`internal/gates/docstring_message.go`). One of the four claims certified as
corrected is therefore still false. Identifier and exemption counts verified correct.

### Finding 4 — LOW — the new CI scenario's ratchet clause is not asserted by its test

Spec clause: `And a legacy file on the base commit is still never scanned`.
`TestDG_CIRefShapeStillRatchets` asserts only `"All 1 exported identifiers"`, and
`setupCIShapeRepo` seeds `src/legacy.go` as `package a\n` — **zero** exported
identifiers. Scanning legacy would leave `Inspected` at 1 and the assertion would
still pass; only the unasserted file count would move. The behavior itself is
correct (verified independently with a real undocumented `func Legacy()` on the
base commit). Fix: seed an undocumented legacy body as `setupDocstringRepoWithLegacy`
already does, and assert `across 1 changed Go file(s)` as the sibling ratchet test does.

### Finding 5 — LOW — the new absolute-root config error is incomplete by its own rationale

Round 3 rejects absolute roots because "scanned paths arrive repo-relative, so an
absolute root can never match". A `..`-escaping root has identical semantics and is
accepted: with `roots = ["../outside"]` the gate Skips honestly, but
`docs lint --full` walks **outside the repository** and reports
`../outside/out.go:3: func OutsideUndoc has no doc comment`. This is a second
gate/report scope divergence, so the deferred `docstring-full-scan-symlink-divergence`
note calling symlinks "the one residual" divergence is inaccurate.

### Finding 6 — LOW — the Pass message ignores the pluralization round 3 added

`docstringProblemMessage` gained `plural()`; `docstringPassMessage` did not, so a
single-identifier pass reads `All 1 exported identifiers across 1 changed Go
file(s) are documented`. Observed live, and asserted verbatim by
`TestDG_CIRefShapeStillRatchets` and `TestDG_FullScanStillRatchets`, which pins the
disagreement in place.

## Deferred Findings

Confirmed present and **not** re-litigated: `docstring-generated-banner-visibility`,
`docstring-full-scan-empty-roots-honesty`, `docstring-ratchet-content-change-only`,
`docstring-gate-scenario-clause-coverage`, `gate-pass-details-invisible`,
`docstring-full-scan-debt-paydown`, `package-doc-comments`,
`docstring-struct-field-docs`, `docstring-nodoc-spaced-spelling`,
`docstring-full-scan-symlink-divergence`.

Added this round (all six findings above):
`gitdiff-remote-fallback-skips-hierarchical-base`,
`pr-gate-ci-scope-narrowed-by-remote-fallback`,
`docstring-adoption-file-count-stale`,
`docstring-ci-scenario-ratchet-clause-unasserted`,
`docstring-roots-parent-escape-unvalidated`,
`docstring-pass-message-plural-disagreement`.

Finding 4 is adjacent to the deferred `docstring-gate-scenario-clause-coverage` but
concerns a scenario introduced in round 3, so it is tracked separately rather than
folded in.

## Recommendation

**Proceed to the docs step.** `centinela validate` passes (exit 0, all gates green,
all three validate commands green) and `go test ./...` passes in full. The round-2
CRITICAL is fixed for real and is now pinned by a test that provably fails without
the fix — I verified that by mutation rather than by reading the transcript.

The two MEDIUM findings are tracked, not blocking: Finding 1 is latent for this
repo's configuration, and Finding 2 leaves main's enforcement intact because the
required `validate` job still full-scans in CI. Finding 3 should be corrected
before the docs step publishes the number, since it is a one-token edit to a comment
that this feature has now gotten wrong twice.

```json centinela:verification
{
  "revision": "d327665d2724598252c2493a48887a264993e525",
  "treeDigest": "sha256:aa46e57ff1532b63828ed0ce2214b679bde771e4016fad4a3111323bff449b05",
  "commands": [
    {"argv": ["go", "build", "-o", "/tmp/centinela-verify-dg3", "./cmd/centinela"], "exitCode": 0, "durationMs": 12000},
    {"argv": ["/tmp/centinela-verify-dg3", "validate"], "exitCode": 0, "durationMs": 418000},
    {"argv": ["go", "test", "./..."], "exitCode": 0, "durationMs": 417040},
    {"argv": ["go", "test", "./tests/acceptance/", "-run", "TestDG_CIRefShape", "-v"], "exitCode": 1, "durationMs": 41000},
    {"argv": ["go", "test", "./internal/gitdiff/"], "exitCode": 1, "durationMs": 3000},
    {"argv": ["go", "test", "./internal/gitdiff/"], "exitCode": 0, "durationMs": 3000},
    {"argv": ["/tmp/centinela-verify-dg3", "docs", "lint", "--changed"], "exitCode": 1, "durationMs": 900},
    {"argv": ["/tmp/centinela-verify-dg3", "docs", "lint", "--changed"], "exitCode": 0, "durationMs": 900},
    {"argv": ["/tmp/centinela-verify-dg3", "validate"], "exitCode": 1, "durationMs": 6000},
    {"argv": ["/tmp/centinela-verify-dg3", "validate"], "exitCode": 0, "durationMs": 6000},
    {"argv": ["/tmp/centinela-verify-dg3", "pr-gate"], "exitCode": 0, "durationMs": 5000},
    {"argv": ["/tmp/centinela-preFallback", "pr-gate"], "exitCode": 1, "durationMs": 5000},
    {"argv": ["/tmp/centinela-verify-dg3", "docs", "lint", "--full"], "exitCode": 0, "durationMs": 2400},
    {"argv": ["git", "for-each-ref", "--format=%(refname)"], "exitCode": 0, "durationMs": 40},
    {"argv": ["find", "internal", "cmd", "pkg", "lib", "src", "app", "-name", "*.go", "-not", "-name", "*_test.go"], "exitCode": 0, "durationMs": 120},
    {"argv": ["/tmp/centinela-verify-dg3", "roadmap", "defer", "gitdiff-remote-fallback-skips-hierarchical-base"], "exitCode": 0, "durationMs": 600},
    {"argv": ["/tmp/centinela-verify-dg3", "roadmap", "defer", "pr-gate-ci-scope-narrowed-by-remote-fallback"], "exitCode": 0, "durationMs": 600},
    {"argv": ["/tmp/centinela-verify-dg3", "roadmap", "defer", "docstring-adoption-file-count-stale"], "exitCode": 0, "durationMs": 600},
    {"argv": ["/tmp/centinela-verify-dg3", "roadmap", "defer", "docstring-ci-scenario-ratchet-clause-unasserted"], "exitCode": 0, "durationMs": 600},
    {"argv": ["/tmp/centinela-verify-dg3", "roadmap", "defer", "docstring-roots-parent-escape-unvalidated"], "exitCode": 0, "durationMs": 600},
    {"argv": ["/tmp/centinela-verify-dg3", "roadmap", "defer", "docstring-pass-message-plural-disagreement"], "exitCode": 0, "durationMs": 600},
    {"argv": ["/tmp/centinela-verify-dg3", "roadmap", "generate"], "exitCode": 0, "durationMs": 700}
  ]
}
```
