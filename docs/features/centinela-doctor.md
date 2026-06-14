# Feature Brief — `centinela doctor`

> One command that **diagnoses** (and, where safe, **repairs**) project-health
> problems that are otherwise invisible until they bite mid-workflow: broken
> hook wiring, roadmap drift, abandoned worktrees, stale `.workflow` state,
> orphaned evidence, config drift, and binary version skew. It generalizes the
> single-purpose `evidence repair` into a holistic health check.

## Problem

**Who.** The developer/operator running a Centinela-governed project — the
person who runs `centinela start`, drives the 5-step workflow, and lives in the
`.worktrees/` + `.workflow/` machinery every day.

**Pain.** Centinela's enforcement leans on a web of side-state — `.claude/
settings.json` hook wiring, `roadmap.json`/`ROADMAP.md`, `.worktrees/*`,
`.workflow/*.json`, `centinela.toml`, and the installed binary. When any of
these drifts, the failure is **silent until it blocks you mid-workflow**, with a
confusing error far from the root cause. There is no single command to surface
these problems. Real incidents from this very feature's setup session:

- **Binary version skew** — the installed `centinela` (`~/.local/bin`) lagged at
  `0.15.0` while the repo Makefile was newer; it *silently* blocked greenfield
  starts with no hint that the binary was the culprit.
- **Roadmap glyph baked into a phase name** — a live-status glyph was authored
  into a phase title (`"✅ Phase 0: Bootstrap"`), which broke
  `isBootstrapPhaseName` (it lowercases + checks `HasPrefix("phase 0")`, so the
  leading glyph defeats the prefix match) and blocked **all** greenfield starts.
- **Abandoned worktrees** — several `.worktrees/*` for merged/complete features
  had to be hand-cleaned with `git worktree remove`.
- **Config drift** — `verify_timeout` was too low (60s) for the real ~75s test
  suite, producing spurious claim-verifier timeouts; it had to be raised to 240.
- **Orphaned evidence** — crashed atomic writes leave `.json.tmp` files;
  `evidence repair` already cleans these per-feature but nothing sweeps holistically.

A `doctor` command turns "mysterious mid-workflow failure" into "one read-only
report up front, with a safe `--fix` for the obvious stuff."

## User Stories

- **As a developer**, I run `centinela doctor` and get a single per-check report
  (✓/⚠/✗ + message) plus a summary line, so I can see all project-health
  problems at a glance without running five different inspection commands.
- **As a developer**, I run `centinela doctor --fix` and have the *safe,
  idempotent* problems repaired automatically (re-wire missing hooks, regenerate
  `ROADMAP.md`, strip a phase-name glyph, sweep orphaned `.json.tmp`), while
  destructive actions are only *reported* with the exact command to run.
- **As an operator on CI / a non-TTY shell**, I rely on `centinela doctor`'s
  exit code (non-zero iff any ERROR) so I can gate automation on project health.
- **As a developer who just merged a feature**, I run `centinela doctor` and am
  told which worktrees and `.workflow` states are now abandoned, with the exact
  `git worktree remove` / removal command — but they are never auto-deleted.
- **As an onboarding developer**, I run `centinela doctor` after cloning and
  immediately learn my installed binary is stale (`make install`) or my hooks
  aren't wired (`--fix`).

## Acceptance Criteria

(Concrete/testable; map to Gherkin in `specs/centinela-doctor.feature`.)

1. `centinela doctor` runs every enabled v1 check and prints one line per check
   with a status glyph (✓ ok / ⚠ warn / ✗ error), the check name, and a
   message, in a **deterministic, fixed order**, followed by a summary line
   (`N ok, M warn, K error`).
2. `centinela doctor` is **read-only**: it never mutates any file. It exits `0`
   when no check is ERROR (OK/WARN allowed) and exits `1` when any check is ERROR.
3. `centinela doctor --fix` applies only the repairs flagged safe+idempotent on
   their diagnoses; running it twice in a row produces no further changes on the
   second run (idempotency) and the second run's report is all-OK for repaired checks.
4. `--fix` **never** performs a destructive action (deleting `.workflow` state,
   removing worktrees). Those checks always render as a report with the exact
   command the user must run themselves.
5. The hook-wiring check detects missing/stale centinela hook entries in
   `.claude/settings.json` (and the OpenCode equivalent when present) and, under
   `--fix`, re-wires them via the existing setup sync (idempotent).
6. The roadmap check detects (a) `roadmap.json`↔`ROADMAP.md` drift (reusing the
   existing drift logic) and (b) a live-status glyph baked into a phase name;
   under `--fix` it regenerates `ROADMAP.md` and strips the offending glyph.
7. The evidence check sweeps orphaned `*.json.tmp` across **all** features under
   `.workflow/`; under `--fix` it removes them (reusing `evidence.Repair`).
8. The config check flags `verify_timeout` lower than a configurable floor,
   gates referencing missing directories, and unknown TOML keys (report only).
9. The binary-skew check compares the installed `centinela --version` against the
   repo's Makefile `VERSION` and reports skew with `make install` (report only —
   doctor cannot safely self-reinstall).
10. With no `.claude/` dir, not inside a git repo, or with no worktrees, the
    relevant checks degrade to a clear WARN/OK rather than crashing.

## Edge Cases

- **No `.claude/` directory** — hook-wiring check reports WARN ("hooks not
  configured; run `centinela setup`"), not an error/crash.
- **Not in a git repo** — worktree + version-skew checks that shell out to git
  degrade to WARN/SKIP with a clear message; no panic.
- **No worktrees** — abandoned-worktree check is OK (nothing to report).
- **Clean project (all OK)** — every check ✓, summary `N ok, 0 warn, 0 error`,
  exit `0`.
- **`--fix` idempotency** — second consecutive `--fix` run is a no-op.
- **Partial repair failure** — if one repair errors under `--fix`, others still
  run; the failed check renders ✗ with the error; exit `1`. No half-applied
  silent state.
- **Destructive-action refusal** — `--fix` on an abandoned-worktree finding does
  NOT remove it; it prints the command. Verified by an acceptance test.
- **Non-TTY output** — report is plain, parseable, deterministic (no spinner,
  stable ordering); exit code is the contract for automation.
- **Unknown config keys** — surfaced as WARN with the offending key names.
- **Run from a worktree vs repo root** — doctor resolves the repo root (walking
  out of `.worktrees/<feature>` via the existing `worktree.DetectFeatureFromCwd`
  logic) so checks operate on the canonical repo, not the worktree subtree.
- **Multiple simultaneous problems** — all problems are reported in a single
  invocation; doctor never stops at the first finding. If hooks, roadmap drift,
  and orphaned tmp files all exist, all three appear in one output. (Added by
  feature-specialist — omitted from AC enumeration but load-bearing for UX.)
- **`centinela` binary not on PATH** — the version-skew check degrades to WARN
  with a clear "binary not found" message; it does NOT error/crash. Distinct
  from the version-mismatch case where the binary exists but is older. (Added by
  feature-specialist.)
- **Doctor does not require an active workflow** — `centinela doctor` must run
  successfully with no active feature workflow. The prewrite hook must not block
  it; it is a read-only diagnostic command. (Added by feature-specialist —
  critical for onboarding and CI use-cases where no workflow is in progress.)
- **Check dependency missing at runtime** (e.g., `centinela.toml` has a syntax
  error, or a referenced path is inaccessible) — the affected check degrades to
  ERROR with a clear message naming the failure; all other checks that do not
  share that dependency still produce their normal diagnosis. No panic. (Added by
  feature-specialist — ensures doctor is maximally informative even in degraded
  environments.)
- **Multiple fixable problems in one `--fix` run** — all safe repairs are
  applied within the same invocation, not sequentially across separate runs. The
  post-fix report reflects the state after ALL repairs have been attempted. (Added
  by feature-specialist — clarifies the atomicity expectation for `--fix`.)

(Full enumeration mirrored into `.workflow/centinela-doctor-edge-cases.md`.)

## Data Model

A small `Check` abstraction in `internal/doctor/`:

```go
type Status int // OK | Warn | Error

type Repair struct {
    Safe      bool              // true => eligible for --fix
    Idempotent bool             // documents the re-run guarantee
    Apply     func() error      // nil for report-only/destructive checks
    Command   string            // user-runnable command for report-only fixes
}

type Diagnosis struct {
    Name    string   // stable identifier, drives ordering + rendering
    Status  Status
    Message string
    Details []string
    Repair  *Repair  // nil when nothing to fix
}

type Check interface {
    Name() string
    Run(ctx Context) Diagnosis   // pure diagnose; never mutates
}
```

- `Context` carries the resolved repo root + loaded `*config.Config` so each
  check is pure and unit-testable with a temp dir.
- `Apply` is only invoked under `--fix` and only when `Repair.Safe`. Destructive
  remediations leave `Apply == nil` and set `Command`.
- Ordering is the registry order (deterministic). Status precedence for the
  exit code: any `Error` ⇒ exit 1; else 0.

## Integration Points

- **`.claude/settings.json`** + OpenCode config — via `internal/setup`
  (`BuildSyncPlan`/`ApplySync`, `buildHookSettings`, `mergeHooks`).
- **git worktrees** — `internal/worktree` (`Dir`, `Path`, `Exists`,
  `DetectFeatureFromCwd`, `Remove` for the report command); enumeration of
  `.worktrees/*` + `git worktree list --porcelain` to detect merged/complete.
- **`roadmap.json` / `ROADMAP.md`** — `internal/gates` drift logic +
  `internal/roadmap` (`Load`, `RenderMarkdown`, `isBootstrapPhaseName`
  rationale, phase-name glyph strip).
- **`centinela.toml`** — `internal/config` (`Load`, `Verify.TimeoutSeconds`,
  gates dirs, unknown-key detection).
- **`.workflow/*.json`** + `*.json.tmp` — `internal/workflow` (`WorkflowDir`,
  `Load`) + `internal/evidence` (`Repair`).
- **installed binary** — shell out to `centinela --version`; compare to Makefile
  `VERSION` parsed from the repo.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| **False positives** (e.g. flagging a legitimately custom hook setup) | Medium | Conservative checks: WARN not ERROR when ambiguous; reuse battle-tested setup/drift logic rather than reimplementing. |
| **Destructive repairs** | High | Hard rule: `--fix` never deletes worktrees/`.workflow`. Report-only with explicit command. Acceptance test asserts refusal. |
| **Layer-import explosion** | Medium | `internal/doctor` imports many domains; assign it a NEW top layer in PROJECT.md G2 + `[gates.import_graph]` (see plan). Keep imports read-only. |
| **≤100-line file splits** | Medium | One check + its repair per file; thin `cmd/` orchestrator; registry in its own file. Budgeted in the plan. |
| **Non-determinism** | Medium | Fixed registry order; sort enumerations (worktrees, tmp files); pure `Run` so tests are reproducible. |
| **Version-skew false alarm in dev** | Low | Report-only WARN; never blocks; message names both versions and `make install`. |

## Decomposition

**Cohesive, not a grab-bag:** every check answers the same question — "is this
project's Centinela side-state healthy?" — and shares one `Check`/`Diagnosis`
abstraction, one renderer, one exit-code policy. The variety of *sources*
(hooks, roadmap, worktrees, config) is exactly why a single command is valuable;
splitting them into separate commands would recreate the "no single command"
pain this feature exists to kill.

**Per-check work units (each its own ≤100-line file + colocated `_test.go`):**

1. `check_hooks.go` — hook-wiring diagnosis + safe re-wire repair.
2. `check_roadmap.go` — drift + phase-name-glyph diagnosis + safe regenerate/strip.
3. `check_worktrees.go` — abandoned-worktree diagnosis (report-only command).
4. `check_workflow_state.go` — stale/orphaned `.workflow` diagnosis (report-only).
5. `check_evidence.go` — orphaned `.json.tmp` sweep diagnosis + safe repair.
6. `check_config.go` — `verify_timeout` floor + missing dirs + unknown keys (report).
7. `check_version.go` — installed-vs-Makefile skew (report-only, `make install`).

Plus shared scaffolding: `doctor.go` (types: `Check`, `Diagnosis`, `Repair`,
`Status`, `Context`), `registry.go` (ordered registry + `Run`/`Fix` drivers),
the `cmd/centinela/doctor.go` thin orchestrator, and `internal/ui/render_doctor.go`.

**v1 IN:** checks 1–7 (all motivated by real incidents). **DEFERRED FINDINGS:**
none of the seven are deferred; but the `roadmap-import-graph-layer-mapping`
Backlog item is *related* — mapping `internal/roadmap` (and now `internal/
doctor`) as explicit import_graph layers — and this feature's layer assignment
should be coordinated with it.

See the implementation plan at `docs/plans/centinela-doctor.md`.
