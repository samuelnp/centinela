# Plan: capability-calibration

> Brief: `docs/features/capability-calibration.md`. n-tier; every source file
> ≤100 lines (hard rule); `cmd/` thin only. `internal/calibration` imports
> `internal/telemetry` + `internal/config` (both leaves) + stdlib only — NOT
> `cmd/`, NOT `internal/ui`. Renderer lives in `internal/ui`.

## Design decisions (decided — implement near-verbatim)

### Command verb
`centinela calibrate` (single-word verb, matching `insights` / `doctor` /
`validate`). Flags: `--json` (bool). No `--top` (per-model, not ranked-top-N).

### Friction metric (per model)
- `Rework` = count of `gate-failure` + `verify-rejection` + `complete-rejected`
  events for that model (identical to `insights.reworkType`).
- `Advances` = count of `step-advanced` events for that model (the "green"s).
- **`Rate` = Rework / Advances** = governance friction per successful advance.
  `HasRate = Advances > 0`; when `Advances == 0`, `Rate = 0`, `HasRate = false`
  (no division, no NaN).
- Also surfaced raw for evidence: `Blocks`, `GateFailures`, `VerifyRejections`.

### EXACT thresholds (fixed documented constants in `calibration`)
```
const (
    highFrictionRate = 1.0  // ≥ 1.0 rework events per advance  → high friction
    lowFrictionRate  = 0.25 // ≤ 0.25 rework events per advance  → low friction
    minAdvances      = 3    // need ≥ 3 advances before a rate is trustworthy
)
```
Rationale: ≥1.0 means the model averages at least one gate/verify/complete
rejection for every step it lands — clearly fighting the process. ≤0.25 means
fewer than one rework per four advances — sailing through. `minAdvances` guards
against over-reacting to a 1-or-2-event sample.

### Profile strictness ordering (most → least strict)
`strict (2) > guided (1) > outcome (0)`. Helper `strictnessRank(profile)` and
`tighter(profile)` / `looser(profile)` clamp at the ends (`strict` can't tighten;
`outcome` can't loosen).

### Classification rule (exact)
For each model, look up `class, ok := config.CapabilityClassFor(model, cfg)` and
`current := config.ProfileForCapability(class, cfg)`:
- `!ok` (no class, incl. `"unattributed"`) → **Verdict=Unclassified**,
  **Recommendation=None**, `RecommendedProfile = ""`. (Cannot recommend.)
- `HasRate == false` (Advances < `minAdvances`, includes 0) → **WellCalibrated /
  Keep**, `RecommendedProfile = current`. (Not enough signal.)
- `Rate >= highFrictionRate`:
  - if `current` can be tightened → **Undergoverned / Tighten**,
    `RecommendedProfile = tighter(current)`.
  - else (already `strict`) → **WellCalibrated / Keep** (maxed, noted).
- `Rate <= lowFrictionRate`:
  - if `current` can be loosened → **Overgoverned / Loosen**,
    `RecommendedProfile = looser(current)`.
  - else (already `outcome`) → **WellCalibrated / Keep** (maxed, noted).
- otherwise (between thresholds) → **WellCalibrated / Keep**,
  `RecommendedProfile = current`.

Every `ModelRecord` carries `FrictionStats` (the raw counts + rate) so the
recommendation is evidence-backed and auditable.

### Ordering (deterministic)
`Models` sorted by `(Verdict-priority, model id asc)` — but simplest stable rule:
sort by model id ascending with `"unattributed"` forced last. No map ranged in
output. Repeated runs → byte-identical output.

### v1 scope
- **In:** `Event.Model` field + stamping at the 3 emit sites; `centinela
  calibrate` (+`--json`); per-model friction + classification + single-step
  profile recommendation; unclassified / unattributed / empty-log handling;
  deterministic render.
- **Out (advisory boundary):** auto-applying a profile change to config; writing
  telemetry; multi-step recommendations (only one step tighter/looser at a time);
  per-repo cross-model comparison/ranking; time-windowed/trend analysis.

## File-level breakdown across the 5 steps

Every NEW file named and budgeted ≤100 lines. `go test ./...` ~75s under
`verify_timeout=240`.

### Step 1 — plan
- `docs/features/capability-calibration.md` (this brief) ✓
- `docs/plans/capability-calibration.md` (this file) ✓
- `specs/capability-calibration.feature` — Gherkin scenarios: model stamping
  carries through; back-compat read; per-model grouping; under→tighten;
  over→loosen; well→keep; unclassified; empty log; --json; determinism; maxed-out
  extreme.

### Step 2 — code

**Part 1 — telemetry stamping**
- `internal/telemetry/event.go` — add `Model string \`json:"model,omitempty"\``
  to `Event` (after `Feature`/`Step`, with a doc comment). Tiny edit, stays ≤100.
- `cmd/centinela/telemetry_model.go` (NEW, ~30 lines) — `resolveEmitModel(wf
  *workflow.Workflow, cfg *config.Config) string`: `if wf != nil &&
  wf.DriverModel != "" { return wf.DriverModel }; return
  config.DriverModelFrom("", cfg)`. Plus an overload/wrapper
  `resolveEmitModelFrom(wfs []*workflow.Workflow, cfg)` that picks the first
  active workflow (reuse prewrite's notion) for the prewrite site.
- `cmd/centinela/complete.go` — set the model on each emitted event. Since the
  `telemetry.Record*` constructors don't take a model, the cleanest minimal
  change is to **add a `model` param to the `Record*` constructors** OR stamp
  `e.Model` before `Record`. **Decision:** add a trailing `model string` param to
  the four `Record*` constructors in `internal/telemetry/constructors.go`
  (`RecordBlock`, `RecordGateFailure`, `RecordVerifyRejection`,
  `RecordCompleteRejected`, `RecordStepAdvanced`) — set `Model: model` in the
  Event literal. This keeps stamping at the leaf boundary but the *value* is
  resolved by the caller (telemetry never imports workflow). Update `complete.go`
  call sites to pass `resolveEmitModel(wf, cfg)`.
- `cmd/centinela/hook_prewrite.go` — pass `resolveEmitModelFrom(wfs, cfg)` to the
  two `RecordBlock` calls (need-init uses the env/config fallback since no active
  wf; out-of-step uses the first-active wf's DriverModel).
- `cmd/centinela/telemetry_emit.go` — `emitGateFailures` /
  `emitVerifyRejection` gain a `model string` param (passed from their
  `complete.go` callers) and forward it to the constructors.
- `internal/telemetry/constructors.go` — signature change above (still ≤100).

**Part 2 — calibration package** (`internal/calibration/`)
- `report.go` (NEW, ~70 lines) — `Report`, `ModelRecord`, `FrictionStats`,
  `Verdict`/`Recommendation` enum consts + their `String()`/JSON tags. Package
  doc comment (aggregator-over-leaves note).
- `friction.go` (NEW, ~70 lines) — `reworkType(t)` (mirror insights), group
  events by `Model` (empty→`"unattributed"`), compute per-model `FrictionStats`
  (Blocks/GateFailures/VerifyRejections/Rework/Advances/Rate/HasRate). Guards the
  rate denominator.
- `classify.go` (NEW, ~80 lines) — `strictnessRank`, `tighter`, `looser`, and
  `classify(model, stats, cfg) (Verdict, Recommendation, recProfile, class,
  current)` implementing the exact rule above.
- `calibrate.go` (NEW, ~60 lines) — `Calibrate(events []telemetry.Event, cfg
  *config.Config) Report`: span (reuse a tiny local min/max like insights),
  group, classify each model, sort deterministically (`"unattributed"` last),
  assemble `Report`.

**Part 2 — command + renderer**
- `cmd/centinela/calibrate.go` (NEW, ~55 lines) — thin: `telemetry.ReadDefault()`
  → `calibration.Calibrate(events, cfg)` → `--json` (indented marshal) else
  `ui.RenderCalibration`. Empty/missing log handled inside Calibrate/renderer →
  exit 0. Register on `rootCmd`. Mirrors `insights.go`.
- `internal/ui/render_calibration.go` (NEW, ~80 lines) — house style from
  `render_insights.go`: header (model count + span), one block per model (id,
  class, current→recommended profile, verdict, evidence counts). Empty →
  "no telemetry yet" muted line. Deterministic; no map ranging.

**Part 2 — import graph / G2**
- `centinela.toml` — add `internal/calibration/**` to the `aggregator` layer
  `paths` (alongside doctor/insights). It imports only telemetry + config
  (leaves) + stdlib → fully mapped, no new failing edge.
- `PROJECT.md` G2 prose — one-line note: "`internal/calibration` also joins the
  aggregator layer: a read-only per-model calibration analyzer over telemetry,
  importing the telemetry + config leaves only, imported solely by cmd/ (its
  Report type by internal/ui for rendering)."
- Mirror the PROJECT.md change to `internal/scaffold/assets` if PROJECT.md is
  mirrored there (check parity test scope per project memory).

### Step 3 — tests
Colocated `_test.go` per package (95% per-package gate, no `-coverpkg`; every
test file ≤100 lines per G1-applies-to-tests rule — split if needed):
- `internal/telemetry/record_test.go` (UPDATE) — assert `Model` is recorded when
  set; back-compat unmarshal of a `model`-less line → `Model == ""`.
- `internal/telemetry/constructors_test.go` (NEW if needed) — each constructor
  stamps the passed model.
- `cmd/centinela/telemetry_model_test.go` (NEW) — `resolveEmitModel` /
  `resolveEmitModelFrom`: wf.DriverModel wins; env/config fallback; empty.
- `cmd/centinela/*_test.go` emit-site stamping — assert prewrite/complete emit
  events carrying the resolved model (table-driven; ≤100 lines each).
- `internal/calibration/friction_test.go` — grouping (unattributed bucket),
  rework tally, rate + HasRate guard (zero advances), division-by-zero.
- `internal/calibration/classify_test.go` — every branch: under→tighten,
  over→loosen, well→keep, unclassified, maxed-strict, maxed-outcome,
  below-minAdvances; strictnessRank/tighter/looser clamps.
- `internal/calibration/calibrate_test.go` — end-to-end pure: mixed-model slice →
  deterministic sorted Report; empty slice → empty Report; single model.
- `internal/ui/render_calibration_test.go` — empty-state line; populated render
  contains id/verdict/recommended profile; deterministic (stable across runs).
- `tests/acceptance/capability_calibration_test.go` — one executable test per
  `.feature` scenario (stamp→read→calibrate visible; --json shape; empty-log exit
  0; unclassified). Wire acceptance execution into `validate.commands`.
- `tests/integration/calibration_roundtrip_test.go` — stamp an event via a
  constructor → `telemetry.Read` from a temp dir → `Calibrate` → assert the model
  is attributed and classified end-to-end.
- `.workflow/capability-calibration-edge-cases.md` — enumerate every Edge Case
  from the brief with its covering test.

### Step 4 — validate
- Gatekeeper report `.workflow/capability-calibration-gatekeeper.md` (G1 ≤100,
  import-graph no violation, i18n n/a, leaf purity of telemetry confirmed).
- `centinela validate` green (lint + type + import_graph + full `go test ./...`,
  ~75s < 240 timeout). Production-readiness subagent if gate enabled.

### Step 5 — docs
- Documentation-specialist `.md` + `.json`; generated `docs/project-docs/index.html`.
- `.workflow/capability-calibration-changelog.md` (artifact new early — completion
  requires it).
- Update any user-facing command reference to list `centinela calibrate`.
