### Big-Thinker Report: deferred-findings-roadmap-capture
**Date:** 2026-06-12

#### Problem
Eight workflow roles (big-thinker, feature-specialist, senior-engineer, qa-senior, edge-case-tester, ux-ui-specialist, validation-specialist, gatekeeper) already produce deferred knowledge as a *mandatory* part of their reports — "Out" bullets, `#### Out-of-Scope`, "Outstanding TODOs", `#### Residual Risks`, and deferred UX/validation/remediation notes — but all of it dies in prose under `.workflow/`. The roadmap (`.workflow/roadmap.json`) is the single planning source of truth, yet nothing connects these capture points to it. The project re-discovers (or ships) the same gaps later; this very roadmap has drifted by hand twice and the 397-entry legacy memory corpus is full of findings that never became features. Phase 5 (Operability & DX) is the place to plug this leak of already-paid-for information before later phases assume findings are machine-captured.

This is revision 2: the operator reviewed revision 1 (a separate one-file-per-finding ledger) and overrode it. Four operator decisions are binding and are implemented faithfully below, including an honest accounting of the residual risks the operator accepted.

#### Scope
- In: `centinela roadmap defer <slug> --summary <text> [--source <feature>/<role>]` appends the finding directly to a dedicated `Backlog` phase in roadmap.json (creating it if absent) via raw-preserving read-modify-write; the `Feature` struct gains optional `omitempty` fields (summary, source{feature,role}, deferredAt). `roadmap validate` (ValidateAnalysis + ValidateQuality), `roadmap ready`/readiness, and the `start` dependency guard EXEMPT Backlog features: no analysis/quality required, never ready/startable, `centinela start <backlog-slug>` refused with "promote it first". Backlog findings render in `centinela roadmap`. `centinela roadmap promote <slug> --phase <name> [--summary <text>] [--scores ac,uv,dc,dep,ee,overall]`: with `--scores` non-interactive (validates each 1–10, overall ≥9 before any write); without `--scores` prints the roadmap-quality-evaluator context (finding name/summary/source, threshold 9, six-dimension schema, re-invocation line) and writes NOTHING. Promote moves the entry out of Backlog into the target phase, appends analysis + quality entries (raw-preserving), appends provenance bullets to the two companion .md files, then runs validate last. A required uniform "Deferred Findings" section added to ALL EIGHT role prompts and their byte-identical scaffold mirrors.
- Out: no auto-prioritization/auto-scheduling; no validator hard-gate on "did the agent defer everything" (unverifiable); no gates/claim-verification or evidence-schema change; no retroactive backfill; no ROADMAP.md human-file auto-sync (v1 prints a reminder); no dedupe/similarity detection (exact slug match only); no `defer dismiss`/lifecycle (a Backlog entry is "present until promoted").

#### Dependencies & Assumptions
- Internal modules: `internal/roadmap` (Load/Save, `roadmapFeatureSet`, ValidateAnalysis/ValidateQuality, DeriveReadiness, bootstrap-predicate pattern), `internal/ui` (render helpers, panel styles, i18n routing), `internal/worktree` (`DetectFeatureFromCwd`; slug rule referenced by comment, NOT imported — keeps G2 import-graph edge-free), `cmd/centinela` cobra wiring + `start_guard`.
- Canonical backlog phase name: `Backlog` (case-insensitive trim match via an `isBacklogPhaseName` helper mirroring `isBootstrapPhaseName`). `defer` creates the phase if absent, appended as the last phase. A Backlog feature is any `Feature` in a phase whose name matches that predicate — the single source of the exemption.
- Verified this round: `roadmapFeatureSet` is the shared coverage gate for BOTH ValidateAnalysis and ValidateQuality (exempt once → both exempt); `DeriveReadiness` is the single source for `ReadySet`, `RenderRoadmap`, and unmet-dependency enumeration (skip Backlog there → ready+render covered); `workflowOrderForFeature` is the only path `centinela start` takes to a roadmap feature (backlog refusal belongs there).
- `omitempty` on the new fields means existing non-Backlog entries serialize byte-identically (no diff churn). Raw-preserving I/O (`map[string]any`/`json.RawMessage`) is mandatory because the live analysis JSON carries fields the structs drop. Quality threshold stays 9; role-string constants (`senior-product-manager`, `roadmap-quality-evaluator`) reused, not re-declared. `.workflow/` stays git-tracked and merges via plain `git merge --no-ff`.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Concurrent defer appends to the single `Backlog.features` array → git merge conflict (ACCEPTED by operator) | Low — Merge Steward resolves; conflict is a trivial union | Medium | One-entry-per-line array formatting so both lines survive; raw write touches only the Backlog region |
| Validate-exemption regression — exempting Backlog weakens coverage for non-Backlog features | High — silently lowers the quality gate | Medium | Exemption in ONE predicate; tests assert real features still REQUIRE analysis+quality, Backlog does not, and a same-named feature in a normal phase is unaffected |
| Start-guard bypass — a backlog slug becomes startable | Medium — un-triaged finding runs a full workflow with no brief/scores | Low | `workflowOrderForFeature` refuses `IsBacklogFeature` ("promote it first"); test asserts refusal; readiness never lists backlog features |
| Raw read-modify-write drops unknown fields / re-keys existing entries | High — corrupts roadmap/analysis/quality | Medium | `map[string]any`/`RawMessage` raw I/O; `omitempty` leaves existing entries untouched; golden-file byte-stability test |
| Promote scored path writes a fabricated ≥9 score | Medium — gate-gaming if agents guess | Medium | Default (no `--scores`) prints evaluator context and writes nothing, steering to an honest agent pass; scores validated before write |
| Promote partial write mutates roadmap.json but not analysis/quality | High — blocks every `centinela start` | Low | Validate scores BEFORE any write; temp-file+rename per file; validate runs last and reports loudly |
| Prompt mirror drift across eight pairs | Medium — scaffolded projects get a stale contract | Low | Parity test already covers all eight pairs byte-for-byte (verified); edit both sides in one commit |
| `centinela roadmap` render regression | Medium — regresses prior UX/tests | Low | Backlog section renders only when present; phase render untouched (Backlog skipped in the normal loop) |

#### Rollout
- Step 1 — Slice 1 (roadmap struct + backlog exemptions): extend `Feature`/`Source`; add `backlog.go` predicate; wire exemption into `roadmapFeatureSet` (analysis+quality), `DeriveReadiness`, and `workflowOrderForFeature` (start refusal). The safety net everything else depends on — ships first. No new command yet.
- Step 2 — Slice 2 (`roadmap defer` + rendering): `rawio.go`, `defer.go`, `defer_validate.go`, `cmd/centinela/roadmap_defer.go`, `internal/ui/render_backlog.go`, wire into `runRoadmap`. Agents can capture; findings visible; validate stays green via Slice 1.
- Step 3 — Slice 3 (`roadmap promote` incl. evaluator path): `promote.go`, `promote_artifacts.go`, `cmd/centinela/roadmap_promote.go`, evaluator-context ui block. No-`--scores` prints context and writes nothing; scored path moves+appends raw-preserving, validate passes; below-threshold + unknown-phase reject with zero writes; golden-file preservation.
- Step 4 — Slice 4 (prompt contract everywhere): edit all eight source prompts + eight mirrors byte-identically. Done last so the obligation never points at a missing command. Parity test green by construction (no extension needed).
- Post-v1: `defer dismiss`/lifecycle, dedupe heuristics, evidence-schema field for deferred slugs, ROADMAP.md auto-sync.

#### Handoff
- Next role: feature-specialist
- Outstanding questions:
  - Exact promote evaluator-context stdout format (fields, ordering, literal re-invocation line, i18n key per Hard Rule 7).
  - Backlog rendering shape (panel vs inline; summary inline or count+slug; placement vs phase overview).
  - Confirm promote metadata-stripping (pinned: keep `summary` as quality-entry summary, drop `source`/`deferredAt` from the roadmap feature but record them in the analysis/quality `.md` provenance bullet) and the bullet wording.
  - Confirm the raw-writer emits `Backlog.features` one object per line (merge-union friendly) without perturbing other phases' formatting, verified by a round-trip golden test.
  - Confirm appending `Backlog` as the last phase does not perturb `HasBootstrapPhase`/bootstrap ordering or readiness.
