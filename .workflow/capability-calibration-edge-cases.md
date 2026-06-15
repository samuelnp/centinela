# Edge Cases: capability-calibration

## Covered

- **Division-by-zero / NaN guard** — a model with rework but zero `step-advanced`
  events yields `HasRate=false`, `Rate=0`, verdict `WellCalibrated` (never
  `Undergoverned`). Covered by `friction_test.go::TestFrictionZeroAdvancesGuard`,
  acceptance `capability_calibration_guard_test.go::TestCalZeroAdvancesGuard` and
  `capability_calibration_boundary_test.go::TestCalBoundaryOnlyReworkWellCalibrated`.
- **Threshold boundary inclusivity** — `Rate == 1.0` (highFrictionRate) →
  Undergoverned/Tighten; `Rate == 0.25` (lowFrictionRate) → Overgoverned/Loosen;
  `Rate == 0.0` with advances → Overgoverned if loosenable. Covered by
  `classify_test.go` and `capability_calibration_boundary_test.go`.
- **Maxed-out profiles** — high friction already at `strict`, or low friction
  already at `outcome`, clamps to `WellCalibrated/Keep` (no out-of-range profile
  invented). `tighter`/`looser` clamp at the ends. Covered by
  `classify_test.go::TestTighterLooserClamp`, `TestClassifyTightenMaxed`,
  `TestClassifyLoosenMaxedAndBetween`, and acceptance maxed scenarios.
- **Insufficient evidence** — `Advances < minAdvances (3)` forces `Keep`
  regardless of rate. Covered by `TestClassifyInsufficientAdvances` and
  `TestCalInsufficientAdvances`.
- **Unclassified models** — a model id with no capability class →
  `Unclassified/None`, empty recommended profile (never invents one). Covered by
  `TestClassifyUnclassified` and `TestCalUnclassifiedModel`.
- **Unattributed bucket / back-compat** — events with no `model` key fold into the
  single `"unattributed"` bucket, classified `Unclassified`, and are forced last
  in the deterministic sort. A legacy `model`-less JSONL line parses cleanly with
  `Model == ""`. Covered by `friction_test.go::TestFrictionUnattributedBucket`,
  `model_field_test.go::TestLegacyLineWithoutModelParsesEmpty`,
  `capability_calibration_stamping_test.go::TestCalStampingLegacyBucketsUnattributed`,
  and `TestCalUnattributedLast`.
- **Empty / missing / malformed log** — missing or empty log prints
  `no telemetry yet` and exits 0; malformed JSONL lines are skipped while valid
  events still aggregate; no parse error or stack trace leaks. Covered by
  `capability_calibration_empty_test.go` and `calibrate_test.go`.
- **Determinism** — two runs (human and `--json`) on the same log produce
  byte-identical output; models sort by id ascending with `"unattributed"` last;
  tie-break is by model id. Covered by `calibrate_test.go::TestCalibrateDeterministicSort`,
  `capability_calibration_determinism_test.go`, and `TestCalTieBreakById`.
- **Non-TTY / no ANSI** — piped output (human and `--json`) contains no ANSI
  escape sequences. Covered by `TestCalNonTTYNoANSI` and the `--json` ANSI checks.
- **Model field round-trip** — a stamped `Model` serializes under `"model"`;
  empty `Model` is omitted (`omitempty`) so golden lines stay stable; every
  `Record*` constructor stamps the passed model. Covered by `model_field_test.go`
  and `capability_calibration_stamping_test.go`.
- **Driver-model resolution precedence** — workflow `DriverModel` wins over the
  env/config fallback; empty when nothing is configured;
  `resolveEmitModelFrom` picks the first active workflow. Covered by
  `cmd/centinela/telemetry_model_test.go`.

## Residual Risks

- **Single-step recommendation only** — Calibrate recommends exactly one step
  tighter/looser at a time (no multi-step jumps). This is an intentional v1 scope
  boundary documented in the plan; an extreme-friction model still capped at one
  step is the designed behavior, not a gap.
- **No auto-apply** — the command is advisory: it never writes config or mutates
  telemetry. Applying a recommended profile remains a manual operator action; out
  of scope by design.
- **No time-windowed / trend analysis** — friction is aggregated over the whole
  log span. A model whose behavior improved recently is judged on its full
  history. Out of v1 scope; mitigated by the cited raw counts being auditable.
