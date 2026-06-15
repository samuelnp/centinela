# Feature Brief: capability-calibration

> Builds on `governance-telemetry` (the event log) and `model-capability-profiles`
> (class → default enforcement profile). This brief is referenced by
> `docs/plans/capability-calibration.md`.

## Problem

**Who:** The operator who decides which enforcement profile (strict / guided /
outcome) a model should run under on this repo.

**Pain:** Today that assignment is made by intuition. A frontier model defaults to
`outcome`, a limited one to `strict`, and the choice is never revisited against
real behavior. A model that quietly needs *tighter* governance (it keeps tripping
gates and getting verify-rejected under a loose profile) is never recalibrated;
a model that is *over-governed* (it sails through every step under a strict
profile, paying the process tax for no benefit) is never relaxed. There is no
measured answer to "how much scaffolding does *this* model need on *this* repo."

`governance-telemetry` already records the friction events (blocks, gate
failures, verify rejections, complete-rejections, step-advances) — but those
events are **not attributed to a driver model**, so they can never be sliced
per-model. This feature (a) stamps the driver model onto every telemetry event
and (b) adds a calibration analysis + `centinela calibrate` command that groups
events per model, measures friction, compares it against the model's current
class-default profile, and recommends a tighter, looser, or unchanged profile —
each recommendation evidence-backed by the counts that drove it. Output is
**advisory**: a report + recommended profile, never an auto-applied config write.

## User Stories

- As an operator, I run `centinela calibrate` and see, per model, how much
  governance friction it is generating and whether its current profile fits.
- As an operator, I get an explicit recommendation per model — *tighten*,
  *loosen*, or *keep* — with the counts that justify it, so I am not guessing.
- As an operator, I run `centinela calibrate --json` and feed the structured
  report into dashboards or CI without scraping rendered text.
- As an operator on a fresh repo with no telemetry, I get a clean empty-state
  message and a zero exit, not an error or a crash.
- As an operator running an unknown local model, I see it reported as
  *unclassified* (no recommendation) rather than a silent omission or a panic.

## Acceptance Criteria

(Concrete and testable; each maps to a Gherkin scenario in
`specs/capability-calibration.feature`.)

1. **Model stamping carries through.** When telemetry is enabled and a workflow
   is active, every recorded event's `model` field equals that workflow's
   `DriverModel`. With no active workflow (e.g. a `need-init` block), the field
   equals `config.DriverModelFrom("", cfg)` (env/config), or is empty if none.
2. **Back-compat read.** A pre-existing event line without a `model` field
   unmarshals cleanly with `Model == ""` and is bucketed under `"unattributed"`.
3. **Per-model grouping.** `Calibrate` groups events by `Model`; empty Model
   collapses into a single `"unattributed"` bucket; distinct ids stay distinct.
4. **Under-governed → tighten.** A model on a loose profile whose friction rate
   exceeds the high threshold is classified `under-governed` and recommended the
   next-tighter profile (outcome→guided→strict).
5. **Over-governed → loosen.** A model on a tight profile whose friction rate is
   below the low threshold (with enough advances to be meaningful) is classified
   `over-governed` and recommended the next-looser profile (strict→guided→outcome).
6. **Well-calibrated → keep.** Anything between the thresholds (or already at the
   strictness extreme the recommendation would push past) is `well-calibrated`;
   recommended profile equals the current profile.
7. **Unclassified model.** A model whose id has no capability class
   (`CapabilityClassFor` → false) is reported `unclassified` with no
   recommendation; it never crashes and never invents a profile.
8. **Evidence-backed.** Every per-model record carries the raw counts (blocks,
   gate failures, verify rejections, rework, step-advances) that drove its
   classification, so the recommendation is auditable.
9. **Empty / missing log.** No telemetry file, or an empty one, yields an
   empty-state report and exit 0 (text: a "no telemetry yet" line; JSON: a
   well-formed report with zero models).
10. **--json.** `--json` emits the structured `Report` as indented JSON; the
    field names are a stable contract.
11. **Determinism.** Models are emitted in a stable sorted order
    (`"unattributed"` last); no map is ranged in output order; repeated runs over
    the same log produce byte-identical output.

## Edge Cases

- **Empty model** → folded into the `"unattributed"` bucket (still classified by
  friction, but reported `unclassified` since it has no capability class).
- **Model with no capability class** → `unclassified`, no recommendation.
- **Single model** → reported alone; ordering logic still deterministic.
- **Zero step-advanced for a model** → friction *rate* denominator is 0; guard it
  (HasRate=false), classify conservatively as `well-calibrated`/`unclassified`
  (never divide by zero), and surface raw counts so the operator still sees it.
- **Ties in ordering** → secondary sort by model id ascending; fully stable.
- **Back-compat old events** (no `model`) → unattributed, never dropped.
- **Division-by-zero in any rate** → every rate guards its denominator; a 0
  denominator yields `HasRate=false`, not NaN/panic.
- **Already at extreme** → an under-governed model already at `strict` (or an
  over-governed one already at `outcome`) cannot be pushed further; recommend the
  current profile and label `well-calibrated` (with a note that it is maxed).
- **Non-TTY** → lipgloss auto-strips ANSI; piped output is plain and parseable.
- **Threshold boundary at exactly Rate=1.0** → `Rate >= highFrictionRate` is
  inclusive; a model with exactly 3 advances and 3 rework events (Rate=1.0) is
  classified `Undergoverned`, not `WellCalibrated`.
- **Threshold boundary at exactly Rate=0.25** → `Rate <= lowFrictionRate` is
  inclusive; a model with 4 advances and 1 rework event (Rate=0.25) is
  classified `Overgoverned` if loosenable, not `WellCalibrated`.
- **Advances with zero rework (Rate=0.0)** → a model with ≥3 advances and
  zero rework events has `Rate=0.0`, `HasRate=true`; `0.0 <= 0.25` triggers
  `Overgoverned` if loosenable (e.g., `haiku` on `strict` → recommend `guided`).
- **Only rework events, zero advances** → `HasRate=false`; the classification
  short-circuits to `WellCalibrated/Keep` before evaluating Rate, so a model
  with 10 gate-failures but 0 advances is NOT classified `Undergoverned`.
- **`--json` on empty log** → well-formed JSON `Report` with `ModelCount=0`
  and `Models=[]`, exit 0; never a parse error or null.
- **Single event of one type (not step-advanced)** → Advances=0, `HasRate=false`,
  `WellCalibrated/Keep`; raw counts still surfaced so the operator sees it.

## Data Model

- **`telemetry.Event.Model string \`json:"model,omitempty"\`** — additive,
  back-compat. Schema stays `centinela.telemetry/v1`. Old lines → `""`.
- **`calibration.Report`** — `{ ModelCount int; SpanStart, SpanEnd string;
  Models []ModelRecord }`. `Models` sorted deterministically.
- **`calibration.ModelRecord`** — `{ Model string; Class string;
  CurrentProfile string; Friction FrictionStats; Recommendation Recommendation;
  RecommendedProfile string; Verdict Verdict }`.
- **`calibration.FrictionStats`** — `{ Blocks, GateFailures, VerifyRejections,
  Rework, Advances int; Rate float64; HasRate bool }`. `Rework` = gate-failure +
  verify-rejection + complete-rejected (mirrors `insights.reworkType`). `Rate` =
  `Rework / Advances` (friction per successful advance), `HasRate=false` when
  `Advances==0`.
- **`calibration.Verdict`** (enum) — `Undergoverned | Overgoverned |
  WellCalibrated | Unclassified`.
- **`calibration.Recommendation`** (enum) — `Tighten | Loosen | Keep | None`.

## Integration Points

- **Telemetry log + emit sites.** `internal/telemetry/Event` gains `Model`.
  The three cmd/ emit sites (`complete.go`, `hook_prewrite.go`,
  `telemetry_emit.go`) resolve the model and stamp it. Resolution rule: prefer
  the active workflow's `DriverModel`; fall back to `config.DriverModelFrom("",
  cfg)`. A small cmd/-local helper (`resolveEmitModel`) centralizes this.
- **Capability config (leaf).** `calibration` uses `config.CapabilityClassFor`
  (class lookup) and `config.ProfileForCapability` / `config.DefaultProfileForModel`
  (current/default profile) for each model. These are pure leaf reads.
- **--json consumers.** The `Report` JSON shape is a stable contract for
  dashboards/CI; renderer is presentation-only and never the data source.

## Risks

| Risk | Note |
|------|------|
| Thresholds arbitrary / misleading | Thresholds are **fixed, documented constants** in `calibration` (see plan) and every recommendation cites the raw counts, so an operator can audit the call and override. Recommendations are advisory only — never auto-applied. |
| Telemetry leaf purity | `internal/telemetry` MUST NOT import `internal/workflow`. The model is resolved at cmd/ emit sites and passed in; `Event.Model` is set before `Record`. `calibration` imports only `telemetry` + `config` (both leaves) + stdlib. |
| ≤100-line splits | Each metric/classification/recommendation step is its own ≤100-line file; renderer split if needed. |
| Schema back-compat | `Model` is `omitempty` and additive; old lines unmarshal with `""`; schema constant unchanged; existing telemetry/insights readers untouched. |

## Decomposition

**Part 1 — Telemetry model stamping (extend governance-telemetry).**
- Add `Event.Model`.
- Stamp at the three cmd/ emit sites via a cmd/-local `resolveEmitModel(wfs/wf, cfg)`
  helper (telemetry stays a leaf; model is set on the Event, not resolved inside).
- Update/extend telemetry tests for the stamped field; add emit-site stamping tests.

**Part 2 — Calibration analysis + command.**
- `internal/calibration/`: types (`Report`/`ModelRecord`/`FrictionStats`/enums),
  per-model friction compute, classification + recommendation (with thresholds),
  `Calibrate(events, cfg) Report`.
- `cmd/centinela/calibrate.go`: thin command, `--json`, empty-state exit 0.
- `internal/ui/render_calibration.go`: house-style renderer, deterministic order.
- `centinela.toml` + PROJECT.md G2: add `internal/calibration/**` to the
  `aggregator` layer (telemetry + config leaves) with a one-line note.
