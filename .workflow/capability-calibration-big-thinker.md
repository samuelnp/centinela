# Big-Thinker Report — capability-calibration

## Decision (user-approved): stamp model on telemetry, then calibrate
Part 1 — add `Event.Model` (omitempty, schema unchanged); resolve at the 3 cmd emit sites (wf.DriverModel else config.DriverModelFrom) and pass into Record* — telemetry stays a config-only leaf (never imports workflow). Part 2 — `internal/calibration` Calibrate(events,cfg) groups by model, computes friction, classifies, recommends a profile.

## Rules (precise)
Rate = Rework/Advances; Rework = gate-failure+verify-rejection+complete-rejected; Advances = step-advanced. HasRate = Advances>0 (guarded). Thresholds high=1.0, low=0.25, minAdvances=3. Strictness strict(2)>guided(1)>outcome(0), clamp. Class: no class/unattributed→Unclassified; <3 advances→Keep; Rate>=1.0 & tightenable→Tighten; Rate<=0.25 & loosenable→Loosen; else Keep. Evidence = raw counts.

## Command / scope
`centinela calibrate` (+--json), thin. In: stamping, per-model grouping, single-step tighten/loosen/keep, unclassified/empty handling, deterministic render (unattributed last). Out: auto-apply, multi-step jumps, time-windows/trends.

## Layer
internal/calibration joins aggregator layer (telemetry+config leaves, fully mapped) + G2 note.

Handoff → feature-specialist.
