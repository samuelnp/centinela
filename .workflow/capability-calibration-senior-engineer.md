# Senior-Engineer Report — capability-calibration

## Part 1 — telemetry model stamping (telemetry stays a leaf)
Added Event.Model (omitempty, schema unchanged); trailing `model string` on all 5 Record* constructors. Model resolved at the 3 cmd emit sites via cmd-local resolveEmitModel/resolveEmitModelFrom (wf.DriverModel else config.DriverModelFrom), threaded through emitGateFailures/emitVerifyRejection/complete/hook_prewrite/validate. `go list -deps ./internal/telemetry` → only internal/config (leaf purity intact). Fixed 6 existing telemetry test files for the signature change (mechanical).

## Part 2 — internal/calibration + `centinela calibrate`
report.go/friction.go/classify.go/calibrate.go (≤100 each) + cmd/centinela/calibrate.go (thin, --json) + internal/ui/render_calibration.go. Calibrate groups events by Model (empty→unattributed), computes Rework/Advances → Rate (HasRate guard), looks up class/profile, classifies under/over/well/unclassified, recommends one-step tighten/loosen/keep with raw-count evidence. centinela.toml: internal/calibration/** → aggregator layer; PROJECT.md G2 note.

## Verification (orchestrator re-checked)
All files ≤100; gofmt/vet/build clean; go test ./... green (2045). Synthetic multi-model log: loose-model rate=1.00→Undergoverned outcome→guided; tight-model rate=0.25→Overgoverned strict→guided; unattributed→Unclassified last; boundary inclusive; determinism byte-identical; empty-log exit 0; stamped event JSONL carries "model".

## Deviation
Renderer helper named calProfileLine (render_status.go already has profileLine). Behavior per plan.

Handoff → qa-senior.
