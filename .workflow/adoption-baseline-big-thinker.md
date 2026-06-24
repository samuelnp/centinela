### Big-Thinker Report: adoption-baseline
**Date:** 2026-06-24

#### Problem

A team adopts Centinela on a mature repo. They run the brownfield flow
(`analyze` → `synthesize` → `reconstruct` → `roadmap brownfield`), then
`centinela start` their first real gap. Day one, `centinela validate` reports
thousands of pre-existing G1 file-size / G2 layer / secret / spec violations that
predate adoption. The gates are unusable; the team turns them off.

The cure already ships: `audit-baseline-ratchet` (DONE) records a baseline so
only *new* work is gated. But it is **undiscoverable and out of sequence**.
Nothing in the brownfield flow tells the adopter to record a baseline, and
`centinela audit baseline` is framed as a ratchet *re-baseline* (a deliberate
overwrite re-run as debt shrinks), not the one-time *adoption* act day one needs.
So the adopter either never records a baseline and drowns, or stumbles onto
`audit baseline` with no guidance, no visibility into the debt they just
accepted, and a command that silently overwrites if run again. **Who's hurting:**
the brownfield adopter (team lead/staff eng) and the onboarding agent. **Why
now:** both deps (`deep-codebase-analysis`, `audit-baseline-ratchet`) are DONE;
only sequencing and a deliberate, visible first-adoption act remain.

#### Scope (In / Out)

**In (v1):**
- New `centinela adopt` command (thin `cmd/`) composing existing `audit.Load`
  (existence check) + `audit.Record` + `audit.Save`, via a small
  `audit.Adopt(cfg, force)` orchestrator that owns the skip-if-exists rule.
- **Skip-if-exists default** with `--force` to re-adopt; `--json` verdict.
- **Adoption report** (per-gate counts, total, "ratchet to zero" framing) via a
  new `internal/ui` renderer over the existing `audit.Baseline` type.
- Docs sequencing: brownfield guide names `adopt` as the step before first
  `centinela start`.

**Explicitly OUT (and why):**
- **Auto-running adoption inside `init`/brownfield without consent** — adoption
  is deliberate, never silent.
- **Ratchet-down automation, per-gate selective adoption, re-baselining as debt
  shrinks** — already `centinela audit baseline`'s job; stays there.
- **Any gate/fingerprint/file-format change.** `adopt` writes the *same*
  `.workflow/audit-baseline.json` the ratchet reads — no second artifact.

**What `adopt` adds OVER `centinela audit baseline`:** (1) skip-if-exists vs
always-overwrite, (2) a human-facing adoption report vs a one-line confirmation,
(3) named/documented flow placement vs being buried under `audit`. Strip those
three and it is a pure alias — so all three are in scope and must be asserted.

#### Dependencies & Assumptions

- Reuses `internal/audit.Record`/`Save`/`Load` verbatim. **A new
  `audit.Adopt(cfg, force) (Outcome, error)` IS warranted** (not pure cmd
  composition): the skip-if-exists decision is business logic and G7 forbids it
  in `cmd/`. `Adopt` keeps that rule in the aggregator and returns the recorded
  `Baseline` for the report.
- **Layer/import-graph: NO new edges.** Verified against `centinela.toml`:
  `internal/audit` is already in the `aggregator` layer
  (`allow = ["domain","leaf","aggregator"]`) and already imports `gates`+`config`
  same-package; `internal/ui` already imports `internal/audit` (`render_audit.go`);
  `cmd/**` already allows aggregator. Zero G2 matrix change.
- No new config — reuses `cfg.Gates.AuditBaseline.BaselinePath`.
- Sibling precedent: `centinela roadmap brownfield` is its own command, not an
  `init` prompt — confirming the dedicated-subcommand decision.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| Thin/duplicate of `audit baseline` | Med — feature reads as a rename | Med | Three deliberate deltas (skip-if-exists, report, flow placement); spec asserts all three so value is testable, not cosmetic |
| Silently overwriting an established baseline | High — re-tolerates paid-down debt | Med (if alias of `audit baseline`) | Skip-if-exists **default** + explicit `--force`; assert refuse-without-force |
| Baseline hides NEW post-adoption violations | Low — by design | Low | Snapshot is point-in-time; the shipped ratchet fails on anything new after adoption — document, don't "fix" |
| Business logic leaks into `cmd/` (G7) | Med | Low | Skip decision lives in `audit.Adopt`; cmd is a thin wire |
| File >100 lines (G1) | Low | Low | Split `Outcome`/counts into `adopt_outcome.go` if `adopt.go` crowds 100 |

#### Rollout

Smallest correct slice, all in one feature (each piece is tiny):
1. `audit.Adopt` + `Outcome` (the skip-if-exists rule) — the load-bearing core.
2. `cmd/centinela/adopt.go` (+ `adopt_render.go` for `--json`) — thin wire.
3. `internal/ui/render_adopt.go` — the adoption report.
4. Docs sequencing note in the brownfield onboarding guide.

No phased rollout needed — the whole surface is ~5 small files reusing shipped
machinery. The single must-ship behavior is **skip-if-exists default**; without
it `adopt` is unsafe and indistinguishable from `audit baseline`.

#### Deferred Findings

none

#### Handoff — Next role: feature-specialist

Author `specs/adoption-baseline.feature` and the detailed plan. Lock the three
deltas as acceptance scenarios: (1) fresh adopt records the baseline + prints the
per-gate report; (2) re-running `adopt` with a baseline present is refused
(exit 0, no write) unless `--force`; (3) post-adoption an unchanged `validate`
reports all violations as baselined (gates usable). Confirm `audit.Adopt`
signature, the `Outcome` shape, the `--json` verdict fields, and the exact
brownfield-guide insertion point. Re-verify the zero-import-graph-change claim
against `centinela.toml` before code.
