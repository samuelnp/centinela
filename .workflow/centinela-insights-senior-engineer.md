# Senior-Engineer Report — centinela-insights

Implemented `centinela insights` as a pure `internal/insights` aggregator over the telemetry log.

## Files (all ≤100 lines)
- `internal/insights/`: report.go (Report/Count/StepsStat + rankTop stable-sort), blocks.go, gates.go, rework.go, steps.go (mean + span), compute.go (Compute([]Event,topN) Report).
- `cmd/centinela/insights.go` (thin; --top default 5, --json via MarshalIndent).
- `internal/ui/render_insights.go` (human report: header+span+event count, n/a mean, (no events) empty sections).
- `centinela.toml`: aggregator layer paths += internal/insights/**. PROJECT.md G2 prose for internal/insights.

## Metrics (verified)
blocks by "reason · fileType"; gates by Gate; rework = per-feature {gate-failure+verify-rejection+complete-rejected}, empty-Feature excluded; mean steps-to-green = (#complete-rejected+#step-advanced)/#step-advanced, n/a if 0 advances.

## Verification (orchestrator re-checked)
All files ≤100; gofmt/vet/build clean. Synthetic log: gates coverage=2, rework alpha=3/beta=1, mean 1.50=(1+2)/2, malformed line skipped (6/7), determinism byte-identical, empty-log exit 0.

## Deviation
compute.go uses per-metric reducers (file-per-metric mandate) — still linear, no nested scans.

Handoff → qa-senior.
