### Merge Steward Report: token-diet
**Date:** 2026-07-30
**Resolution:** git-text-conflict on .workflow/roadmap.json + ROADMAP.md — the recurring roadmap-state divergence class (branch forked before three roadmap mutations landed on main). Resolved by taking main's authoritative roadmap.json, replaying the branch-only records via the CLI (five deferrals: model-capability-spec-scenario-uncovered, legacy-model-specs-have-no-scenario-markers, plan-empty-inputs-message-masked, summary-digest-hashes-nonschedulable-phases, hook-context-panel-diet), and regenerating ROADMAP.md deterministically. No source conflict; no spec conflict (pre-check passed under the fixed detector).
**Confidence:** high — mechanical resolution, no semantic judgment beyond record replay.
