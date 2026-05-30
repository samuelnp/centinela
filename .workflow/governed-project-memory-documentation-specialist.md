# governed-project-memory — documentation-specialist

## Summary

Knowledge base entry and project documentation generated for the governed-project-memory feature.

### What was written

- **KB entry**: `docs/project-docs/kb/governed-project-memory.md` — A 3-section user-facing guide covering: (1) What it does — the feature automatically harvests lessons, verdicts, and decisions into a git-tracked ledger and recalls them at plan time; (2) When you'd use it — every feature start, to avoid re-learning and to browse project history; (3) How it behaves — with 6 bullets covering capture automation, idempotence, graceful error handling, plan-time ranking (dependencies > tags > recency), byte/count caps, and config gating.
- **Generated HTML**: `docs/project-docs/index.html` (102.6K) and `docs/project-docs/kb/governed-project-memory.html` (7.0K) rendered from the markdown via `centinela docs generate`.

### Scenario coverage

All 13 Gherkin scenarios (SC-01 through SC-13) reflected:
- **Capture** (SC-01, SC-02, SC-03, SC-04): Three sources, optional decisions, automatic at step boundaries.
- **Idempotence** (SC-05): Re-completing a step does not duplicate entries.
- **Non-blocking failures** (SC-06, SC-07): Missing or malformed artifacts emit warnings, never block.
- **Recall** (SC-08, SC-09, SC-10, SC-11): Plan-time injection, deterministic ranking, empty ledger safety, count/byte caps.
- **Config** (SC-12): Memory disabled makes capture/recall no-ops.
- **Concurrency** (SC-13): Per-file writes prevent clobbering across worktrees.

### Roadmap & artifact lineage

Roadmap phase: Phase 1 — Harness Capabilities. Feature brief documents 7 locked decisions (D1–D7) spanning capture sources, trigger points, recall determinism, and out-of-scope items. All decisions preserved in the KB as user-visible behavior and context.
