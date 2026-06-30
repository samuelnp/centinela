# fix-init-managed-sync-drift — validation-specialist

## Gates Run

`centinela validate` (diff-aware, 15 files) — all green: G1, Cross-Compile (6),
`go test ./...`, acceptance, coverage (95.0% ≥ 95.0%), fmt. roadmap_drift in
sync. import_graph + spec-traceability warnings empty-body, non-blocking.

## Synthesis

One-function bug fix routing init's `setupOpenCode()` through the managed-sync
path so a freshly-init'd project no longer reports spurious pending migrations.
Gatekeeper: SAFE. Proven end-to-end: `centinela init` then `centinela migrate`
reports 0 pending and AGENTS.md/plugin carry the managed-version header. The
init→migrate idempotency that the bug violated is now pinned by an acceptance
test (previously untested).

## Decision

PASS → documentation-specialist.
