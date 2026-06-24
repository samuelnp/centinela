### Feature-Specialist Report: adoption-baseline
**Date:** 2026-06-24

#### Behavior Summary

`centinela adopt [--force] [--json]` is the one-time brownfield adoption act: it records the
current full-repo gate violations as the accepted audit baseline (`.workflow/audit-baseline.json`)
via a new `audit.Adopt(cfg, force) (Outcome, error)` orchestrator that reuses the shipped
`audit.Load` (existence check), `audit.Record`, and `audit.Save`. It differs from the existing
`centinela audit baseline` in exactly three deliberate ways, all asserted by the spec: (1) it
**refuses to overwrite an existing baseline** by default — skip-if-exists — and only re-records
with `--force` (the opposite of `audit baseline`, which always overwrites for the ongoing
ratchet); (2) it prints a **human adoption report** — per-gate accepted-violation counts, the
total, and a "you are starting with N accepted findings; ratchet to zero over time" framing — so
the team consciously accepts the debt; and (3) it is the **named, documented step** between
`centinela roadmap brownfield` and the first `centinela start`. The written file is byte-identical
to what `audit.Record` + `audit.Save` already produce — adopt adds semantics (the skip rule and
the report), never different data. After adoption, an unchanged-repo `centinela audit` reports
zero new violations, so day-one `validate` is usable instead of drowning in pre-existing debt.

#### Gherkin Scenarios   (reference specs/adoption-baseline.feature)

Eight scenarios, each with concretely assertable Then-steps (exit code, baseline file written at
the configured path, report lines, byte-unchanged file on skip, byte-identical determinism):

1. **First adoption on a repo with pre-existing violations** — `centinela adopt` writes
   `.workflow/audit-baseline.json` capturing every current violation, exits 0, and prints per-gate
   counts + total + the ratchet-to-zero framing.
2. **Post-adoption ratchet is clean** — a fresh `centinela audit` on the unchanged repo exits 0,
   reports "0 new", every pre-existing violation is baselined/tolerated, no error/stack trace.
3. **Skip-if-exists** — re-running `centinela adopt` when a baseline exists exits non-zero, prints
   "baseline already exists" + a use-`--force` instruction, and leaves the file byte-identical.
4. **`--force`** — re-runs and overwrites the existing baseline, exits 0, reports the new total.
5. **Clean repo** — `centinela adopt` writes a zero-finding baseline and the report says
   "0 accepted findings" (nothing to ratchet).
6. **`--json` (fresh)** — emits valid JSON `{adopted:true, skipped:false, path, total, per_gate}`
   instead of the human prose.
7. **`--json` (skip)** — when a baseline exists, emits `{adopted:false, skipped:true}`, exits
   non-zero, file byte-unchanged.
8. **Determinism** — the baseline adopt writes is byte-identical to `audit.Record` + `audit.Save`
   on the same unchanged repo, entries in stable deterministic order.

#### UX States  (table)

| State | Trigger | CLI behavior |
|-------|---------|--------------|
| Success (adopted) | No baseline exists; repo has violations | Writes baseline; prints adoption report (per-gate counts, total, ratchet-to-zero framing); exit 0 |
| Success (empty) | No baseline exists; repo has no violations | Writes zero-finding baseline; report says "0 accepted findings"; exit 0 |
| Success (force) | Baseline exists; `--force` given | Overwrites baseline with current violations; prints report; exit 0 |
| Error (skip-if-exists) | Baseline exists; no `--force` | No write (file byte-unchanged); prints "baseline already exists … use --force to overwrite"; exit non-zero |
| Machine (json) | `--json` given | Emits adoption verdict JSON `{adopted, skipped, path, total, per_gate}`; exit 0 on adopt, non-zero on skip; no human prose |
| Loading | n/a | Not applicable — a synchronous one-shot CLI command |

#### Out-of-Scope

- Auto-running adoption inside `centinela init` / brownfield without explicit consent — adoption
  is a deliberate, visible act, never silent.
- Per-gate selective adoption (choosing which gates' violations to accept).
- Ratchet-down automation / re-baselining as debt shrinks — that is `centinela audit baseline`'s
  job and stays there.
- Editing or merging existing baseline entries — adopt fully replaces (via `Save`) on `--force`,
  it never selectively edits.
- Any change to gate logic, fingerprinting, or the baseline file format; no second on-disk
  artifact (same `.workflow/audit-baseline.json` the ratchet reads).

#### Deferred Findings

none

#### Handoff — Next role: senior-engineer

Implement per `docs/plans/adoption-baseline.md`. Confirm/wire:
- `audit.Adopt(cfg, force) (Outcome, error)` in `internal/audit` owns the skip-if-exists decision
  (G7: not in `cmd/`): `Load` the configured `BaselinePath`; if it exists and `!force`, return an
  `Outcome{Skipped:true}` with no write; else `Record` + `Save` and return the recorded `Baseline`
  + path for the report. The `Skipped` (no-force-over-existing) path must surface as a non-zero
  exit at the cmd boundary with a clear "use --force" message.
- `Outcome` shape carries: `Skipped bool`, `Path string`, recorded `Baseline` (for per-gate counts
  + total). A per-gate count helper can mirror `countFingerprints` in `audit_baseline.go`.
- `internal/ui/render_adopt.go` `RenderAdoption(Outcome) string` over `audit.Baseline` —
  no new import edge (`internal/ui` already imports `internal/audit`).
- `cmd/centinela/adopt.go` (thin Cobra cmd) + `adopt_render.go` (`--json` verdict
  `{adopted, skipped, path, total, per_gate}`, mirroring `audit_render.go`). Register `adoptCmd`
  on `rootCmd` in `init()`.
- No new config; reuse `cfg.Gates.AuditBaseline.BaselinePath`. Zero G2 import-graph change.
- Docs (docs step): name `centinela adopt` as the step between `roadmap brownfield` and first
  `centinela start` in the brownfield onboarding guide.

qa-senior: trace each spec Scenario with `// Scenario: <name>`; assert exit codes, the file written
at the configured path, the report lines, the byte-unchanged file on skip, and byte-identity vs
`audit.Record`+`Save`.
