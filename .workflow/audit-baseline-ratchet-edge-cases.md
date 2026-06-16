# Edge Cases: audit-baseline-ratchet

## Covered

- **Record on a repo with violations** → baseline captures every Fail-gate
  violation (full scan); exit 0 (unit + acceptance).
- **No-change audit** → all violations baselined; `0 new`; exit 0.
- **New (non-baselined) violation** → ratchet fails (exit 1), names only the new
  one; baselined ones stay tolerated (unit + integration + acceptance).
- **Fixing a baselined violation** → never fails; reported resolved; re-record
  prunes it (ratchet only tightens — pruned-then-reintroduced is new/blocking).
- **Fingerprint stability** → an oversized baselined file growing 130→170 lines
  keeps the SAME Hash, stays baselined (no false "new") — `fingerprint_test.go`,
  the central correctness guard.
- **Missing baseline** → `gate.Check` returns `Skip`; `centinela audit` prints
  "no baseline" and does not block (exit 0).
- **`Save` into a non-existent parent dir** → creates it (`os.MkdirAll`); bug
  found by dogfood, now covered by `baseline_test.go`.
- **Deterministic baseline** → byte-identical on re-record (sorted gates +
  fingerprints, trailing newline).
- **Config gating** → `enabled=false`/`severity=warn` never blocks; Fail vs Warn
  mapped per severity; custom baseline path honored; `validateAuditBaseline`
  rejects bad severity.
- **Participation** → empty `target_gates` = all; explicit allowlist intersects.
- **Stale fingerprint scheme** → `SchemeStale` → Warn "re-run audit baseline".
- **Corrupt/unreadable baseline** → `Load` errors surfaced; `Check` → Fail.
- **Only Fail violations baselined** → import-graph "unmapped" Warn and Skip
  gates excluded.
- **cmd wiring** → `appendAuditGate` no-op when disabled, appends when enabled;
  text + `--json` paths covered.

## Residual Risks

- `Save`/`currentEntries` have a few uncovered error branches (per-func ~82%);
  the aggregate coverage gate (95.1% ≥ 95%) is met and the happy + key error
  paths are exercised.
- v1 parses gate `Details` strings (no structured `Finding` refactor); if a
  gate changes its Detail format across versions, the `scheme` version field
  signals staleness so stale baselines degrade to a non-blocking re-record
  prompt rather than silently mis-matching.
