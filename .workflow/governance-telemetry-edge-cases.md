# Edge Cases: governance-telemetry

## Covered

- **Disabled is a no-op** — `[telemetry] enabled=false` writes no file, no dir (`TestGT_DisabledNoOp`).
- **Default-on (opt-out)** — absent `[telemetry]` (Enabled nil) records events (`TestGT_DefaultEnabled`, `IsEnabled` nil→true).
- **Non-blocking on I/O error** — a file occupying `.workflow` makes `MkdirAll` fail; `Record` warns and returns without panicking, host command unaffected (`TestGT_IOErrorDoesNotFail`, `TestRecord_IOErrorIsSwallowed`).
- **Lenient `Read`** — a corrupt/garbage line is skipped; surrounding valid events still returned (`TestGT_ReadSkipsCorruptLine`).
- **Missing log** — `ReadDefault` on a fresh tree returns `(nil, nil)` (`TestGT_ReadMissingLog`).
- **Schema + timestamp on every event** — every line carries `schema="centinela.telemetry/v1"` and an RFC3339 UTC timestamp (`TestGT_SchemaAndTimestamp`).
- **block need-init vs out-of-step** — distinct `reason`; need-init carries no feature/step (`TestGT_BlockNeedInit`, `TestGT_BlockOutOfStep`).
- **gate-failure carries no feature** — validate isn't feature-scoped; one event per failing gate (`TestGT_GateFailure`, `TestGT_GateFailurePerGate`). Downstream attributes it to the co-occurring `complete-rejected{reason:gates}` by proximity.
- **complete-rejected reasons** — `gates` and `verify` variants carry feature+step (`TestGT_CompleteRejectedGates/Verify`).
- **verify-rejection checks** — the failing claim/role/status/detail set is preserved (`TestGT_VerifyRejection`).
- **Append-only ordering** — multiple records accumulate in call order; two sequential records both land intact under `O_APPEND` (`TestGT_AccumulateInOrder`, `TestGT_TwoSequentialIntact`).
- **Rework is derivable** — N `complete-rejected` for a (feature,step) before its `step-advanced` (`TestGT_ReworkDerivable`).

## Residual Risks

- **Absolute-path classification nuance** — `ClassifyFile` matches segments like `/tests/` and configured code dirs; a relative path can fall through to `TypeOther` (not a block). Real hooks receive absolute paths, so this only affects synthetic callers. Mitigation: emission sites pass the hook-provided path verbatim.
- **gate-failure ↔ feature join is proximity-based** — bare `centinela validate` runs produce feature-unattributed gate-failures (intended: global "which gates bite" stats). Per-feature attribution relies on the same-process `complete-rejected`. Mitigation: documented join contract for downstream readers.
- **Concurrency across non-worktree processes** — `O_APPEND` is atomic per write on local FS; pathological NFS/network mounts are out of scope (local, git-tracked design).
- **No retention/rotation in v1** — the log grows unbounded; rotation deferred to a downstream feature (this repo `.gitignore`s its own log to avoid commit growth).
