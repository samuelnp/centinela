# Implementation Plan — `centinela doctor`

> Feature brief: `docs/features/centinela-doctor.md`.
> Spec: `specs/centinela-doctor.feature`.

A new `internal/doctor/` domain package holds a `Check` abstraction; each check
(and its repair) lives in its own ≤100-line file. `cmd/centinela/doctor.go` is a
THIN orchestrator that runs the registry and renders via `internal/ui`. We reuse
existing logic — no reimplementation of drift, evidence repair, worktree
enumeration, hook wiring, or config load.

## Exit-code + `--fix` safety model (DECIDED)

- **Read-only by default.** `centinela doctor` only diagnoses + renders. It
  mutates **nothing**.
- **Exit codes.** Exit `0` when no check is `Error` (OK and WARN both pass).
  Exit `1` when any check is `Error`. WARN never fails the command — it is
  advisory. This keeps CI gating simple: "doctor green = no errors."
- **`--fix` applies ONLY safe + idempotent repairs** (`Repair.Safe == true`):
  re-wire missing hooks, regenerate `ROADMAP.md`, strip a phase-name glyph,
  remove orphaned `*.json.tmp`. After `--fix`, doctor re-runs the diagnoses and
  renders the post-fix report; exit code reflects the post-fix state.
- **Destructive actions are NEVER auto-applied.** Deleting `.workflow` state or
  removing worktrees: `Repair.Safe == false`, `Apply == nil`, `Command` set to
  the exact user-runnable command. `--fix` prints these, never executes them.
- **Partial-failure rule.** Under `--fix`, every safe repair is attempted; a
  failing `Apply` marks that check `Error` (with the error in `Details`) and the
  rest still run. Exit `1` if any post-fix `Error` remains.

## Layer / import-graph decision (CALL-OUT)

`internal/doctor` is a **new domain orchestrator** that imports several existing
domains (`config`, `gates`, `roadmap`, `evidence`, `workflow`, `worktree`,
`setup`) read-only and is itself imported only by `cmd/`. PROJECT.md G2 currently
maps three layers (`leaf`, `domain`, `cmd`); `internal/doctor` does not fit the
existing `domain` layer (which may import only `leaf`). Decision:

1. Add a **new `[gates.import_graph]` layer** `aggregator` (or `doctor`) in
   `centinela.toml`: `paths = ["internal/doctor/**"]`,
   `allow = ["domain", "leaf", "internal/roadmap", "internal/evidence",
   "internal/worktree", "internal/setup", "internal/workflow"]` — but since the
   current matrix only encodes mapped layers and warns on unmapped packages, the
   pragmatic v1 is: add the `aggregator` layer allowing `domain` + `leaf`, and
   add the multi-domain leaf packages it needs (`roadmap`, `evidence`,
   `worktree`, `setup`, `workflow`) as their own mapped layers OR keep them
   unmapped (current behavior = non-failing WARN). **v1 choice:** map
   `internal/doctor` as `aggregator` allowing `domain` + `leaf`; leave
   `roadmap/evidence/worktree/setup` unmapped (they already are) so doctor's
   edges to them surface as the existing non-failing warning rather than a hard
   fail. This avoids a large matrix rewrite.
2. **Update PROJECT.md G2 prose** to document the `aggregator` layer and that
   `internal/doctor` may import the named domains read-only — mirroring the
   existing `internal/verify` allowance (`verify` already imports `config`,
   `evidence`, `orchestration`, `worktree`).
3. Coordinate with the deferred Backlog item
   **`roadmap-import-graph-layer-mapping`** — it proposes mapping
   `internal/roadmap` as a real layer; when that lands, doctor's `roadmap` edge
   becomes mechanically enforced. Note this dependency in the gatekeeper report.

## Step 1 — plan (this step)

Artifacts: this plan, `docs/features/centinela-doctor.md`,
`specs/centinela-doctor.feature` (Gherkin: one scenario per acceptance criterion
above — clean project all-OK; hook drift + `--fix` re-wire; roadmap glyph detect
+ strip; abandoned worktree report-only; orphaned tmp sweep; config drift WARN;
version skew WARN; `--fix` idempotency; destructive refusal; non-git degrade).

## Step 2 — code (new files, each ≤100 lines)

**Shared scaffolding (`internal/doctor/`):**

| File | Budget | Responsibility |
|------|--------|----------------|
| `doctor.go` | ~70 | Types: `Status` (OK/Warn/Error), `Diagnosis`, `Repair`, `Check` iface, `Context` (repo root + `*config.Config`). |
| `registry.go` | ~80 | Ordered `[]Check` registry; `Run(ctx) []Diagnosis` (pure); `Fix(ctx) []Diagnosis` (apply safe repairs then re-diagnose); `ExitError(diags) bool`. |
| `context.go` | ~40 | Resolve repo root (reuse `worktree.DetectFeatureFromCwd` to climb out of a worktree), load config. |

**Per-check files (each pairs the diagnosis + its repair):**

| File | Budget | Reuses | Repair kind |
|------|--------|--------|-------------|
| `check_hooks.go` | ~90 | `setup.BuildSyncPlan`/`ApplySync`/`buildHookSettings` | SAFE (re-wire) |
| `check_roadmap.go` | ~95 | `roadmap.Load`/`RenderMarkdown`, drift compare, `isBootstrapPhaseName` rationale | SAFE (regen + strip glyph) |
| `check_worktrees.go` | ~95 | `worktree.Dir`/`Path`/`Exists` + `git worktree list --porcelain` | REPORT (`git worktree remove`) |
| `check_workflow_state.go` | ~95 | `workflow.WorkflowDir`/`Load`, glob `.workflow/*.json` | REPORT (deletion) |
| `check_evidence.go` | ~60 | `evidence.Repair` (sweep all features) | SAFE (remove `.json.tmp`) |
| `check_config.go` | ~95 | `config.Load`, `Verify.TimeoutSeconds`, gates dirs, unknown-key scan | REPORT |
| `check_version.go` | ~80 | exec `centinela --version`, parse Makefile `VERSION` | REPORT (`make install`) |

If `check_roadmap.go` exceeds budget, split the glyph-strip repair into
`check_roadmap_glyph.go`. The unknown-key scan in `check_config.go` may need a
helper file `config_keys.go` (re-decode TOML into a map, diff against known keys).

**Renderer (`internal/ui/`):**

| File | Budget | Responsibility |
|------|--------|----------------|
| `render_doctor.go` | ~70 | `RenderDiagnosis(d)` (✓/⚠/✗ + name + message + indented details + repair command) and `RenderDoctorSummary(diags)` (`N ok, M warn, K error`). Follows `render_gates.go` house style + `StyleGreen/Yellow/Red`. |

**Thin orchestrator (`cmd/centinela/`):**

| File | Budget | Responsibility |
|------|--------|----------------|
| `doctor.go` | ~70 | Cobra `doctorCmd` with `--fix` bool flag; build `doctor.Context`; call `doctor.Run`/`doctor.Fix`; render each diagnosis + summary via `ui`; return non-nil error (→ exit 1) iff `doctor.ExitError`. NO business logic. |

## Step 3 — tests

- **Unit (colocated `_test.go`, each ≤100 lines)** — one per check + registry +
  context + renderer. CRITICAL: the **95% per-package coverage gate has no
  `-coverpkg`**, so coverage must come from `internal/doctor/*_test.go` and
  `internal/ui/render_doctor_test.go` colocated with the code under test — NOT
  from `tests/` tier files. Each check tested against a temp dir fixture:
  healthy ⇒ OK, drifted ⇒ Warn/Error, `Repair.Apply` ⇒ fixes + idempotent on
  second call, destructive checks ⇒ `Apply == nil` + `Command` set.
- **Integration (`tests/integration/`)** — a full `doctor` run on a seeded
  fixture repo (all checks), plus a `--fix` round-trip: dirty fixture →
  `doctor --fix` → second `doctor` is all-OK for repaired checks (idempotency).
- **Acceptance (`tests/acceptance/`)** — executable artifacts, one per Gherkin
  scenario; `validate.commands` MUST include their execution. Assert exit codes
  (0 clean / 1 with error) and that destructive findings are reported-not-applied.
- **`.workflow/centinela-doctor-edge-cases.md`** — enumerate the brief's edge
  cases with the test that covers each.
- **Timing.** `go test ./...` runs ~75s; `verify_timeout = 240` accommodates it.

## Step 4 — validate

Gatekeeper report at `.workflow/centinela-doctor-gatekeeper.md`; `centinela
validate` (lint + type + full suite + import_graph gate) passes. Confirm the new
`aggregator` import_graph layer + PROJECT.md G2 prose update are consistent and
that no doctor file exceeds 100 lines. Production-readiness subagent if gated.

## Step 5 — docs

Documentation-specialist `.md` + `.json`; regenerate `docs/project-docs/
index.html`; `.workflow/centinela-doctor-changelog.md`. Document `centinela
doctor` + `--fix` in the command reference and the exit-code/safety model.

## Final v1 check list

**IN (v1):** hook-wiring, roadmap (drift + phase-glyph), abandoned-worktrees,
stale `.workflow` state, orphaned-evidence, config-drift, binary-version-skew.
**DEFERRED:** none of the seven; related Backlog item
`roadmap-import-graph-layer-mapping` to be coordinated with the layer decision.
