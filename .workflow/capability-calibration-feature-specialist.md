# Feature-Specialist Report — capability-calibration

## Spec
`specs/capability-calibration.feature` — 27 scenarios, 1:1 acceptance-traceable: stamping (pinned model in JSONL / empty / legacy), classification (under/over/well/keep at ends), threshold boundaries (Rate=1.0 incl, =0.25 incl, =0.0 overgoverned, zero-advance guarded), unclassified + unattributed-last, empty/missing/malformed log, --json stable, determinism, non-TTY, multi-model single pass.

## Edge cases added to brief
Rate=1.0/0.25 inclusivity; advances+zero-rework Rate=0.0 → Overgoverned-if-loosenable; only-rework zero-advance HasRate=false → Keep; --json empty (ModelCount=0, Models=[]); single non-advance event → Keep.

Handoff → senior-engineer.
