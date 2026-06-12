### Big-Thinker Report: deferred-findings-roadmap-capture
**Date:** 2026-06-12

#### Problem
Four workflow roles (big-thinker, feature-specialist, senior-engineer, qa-senior/edge-case-tester) already produce deferred knowledge as a *mandatory* part of their reports — big-thinker's "Out" bullets, feature-specialist's `#### Out-of-Scope`, senior-engineer's "Outstanding TODOs", qa-senior's `#### Residual Risks` — but all of it dies in prose under `.workflow/`. The roadmap (`.workflow/roadmap.json`) is the single planning source of truth, yet nothing connects these capture points to it. The project re-discovers (or ships) the same gaps later; this very roadmap has drifted by hand twice and the 397-entry legacy memory corpus is full of findings that never became features. Phase 5 (Operability & DX) is the place to plug this leak of already-paid-for information before later phases assume findings are machine-captured.

#### Scope
- In: `centinela roadmap defer <slug> --summary <text> [--source <feature>/<role>]` writing `.workflow/deferred/<slug>.json` (one file per finding) with collision checks vs the ledger and roadmap.json feature names; deferred findings rendered in `centinela roadmap` output (count + open list) plus `defer --list`; `centinela roadmap promote <slug> --phase <name> --summary <text> --scores <…>` doing the atomic triple-write (roadmap.json + roadmap-analysis.json + roadmap-quality.json) with raw-JSON preservation, md-bullet appends, ledger status → promoted, and a validate-after-write; a required "Deferred Findings" obligation added to the four role prompts and their byte-identical scaffold mirrors.
- Out: no auto-prioritization/auto-scheduling (promote needs explicit phase + scores); no validator hard-gate on "did the agent defer everything" (unverifiable); no gates/claim-verification or evidence-schema change; no retroactive backfill of legacy Residual Risks / memory corpus; no ROADMAP.md human-file auto-sync (owned by roadmap-doc-sync; v1 prints a reminder); no dedupe/similarity detection (exact slug match only); no defer wired into ux-ui-specialist / validation-specialist / gatekeeper prompts (trivial fast-follow).

#### Dependencies & Assumptions
- Internal modules: `internal/roadmap` (Load/Save, roadmapFeatureSet, ValidateAnalysis/ValidateQuality), `internal/ui` (render helpers, panel styles), `internal/worktree` (slug rules, merge behavior — read-only), `cmd/centinela` cobra wiring.
- Builds on prior features: roadmap dependencies/readiness (Option B shape — deps on roadmap.json, analysis carries name-only entries), worktree merge + Merge Steward, evidence CLI, scaffold-mirror parity discipline.
- `.workflow/` stays git-tracked in worktrees and merges via plain `git merge --no-ff` (verified in `internal/worktree/merger.go`); file-adds merge cleanly, only same-slug add/add conflicts.
- `roadmap validate` reads only roadmap.json/analysis/quality — the ledger is invisible to it, so validate stays green by construction.
- Quality threshold stays 9; role-string constants (`senior-product-manager`, `roadmap-quality-evaluator`) are stable and must be reused, not re-declared.
- Findings are append-only at capture time; mutation (promote/dismiss) happens only at the root checkout.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Promote's read-modify-write drops unknown JSON fields (e.g. legacy `dependsOn` in analysis) | High — corrupts source-of-truth artifacts | Medium | Raw-preserving JSON (map[string]any / json.RawMessage); golden-file test asserting byte-stable untouched entries |
| Two worktrees defer the same slug → add/add merge conflict | Low — Merge Steward handles it; the conflict is the desired signal | Low | Prompt contract prefers source-prefixed/specific slugs; collision check at defer vs local ledger + roadmap |
| Worktree's stale roadmap.json lets defer accept a slug root already has | Medium — promote fails later | Medium | Promote re-checks at root and errors cleanly; ledger entry stays `open`, can be re-slugged |
| Agents fabricate or skip `--summary`, producing junk findings | Medium — ledger noise erodes trust | Medium | CLI enforces non-empty summary; prompt requires listing recorded slugs in the report; promote is the human triage filter |
| Prompt mirrors drift (parity test may not cover all four prompt files) | Medium — scaffolded projects get a stale contract | Medium | Update both sides in one commit; extend parity-test coverage to these four prompts |
| `centinela roadmap` output regression (existing render + new deferred section) | Medium — regresses an earlier feature's UX/tests | Low | Deferred section renders only when count > 0; existing render tests untouched, new tests added |
| Promote breaks greenfield start-guard (analysis/quality must cover ALL features) | High — blocks every `centinela start` | Low | Promote writes all three artifacts atomically and runs validate as its last step, reporting loudly before anyone runs `start` |
| New Go files exceed 100-line G1 limit | Low | Medium | Split already planned (deferred / promote / artifacts / ui); tests split per concern |

#### Rollout
- Step 1 — Slice 1 (ledger core + `roadmap defer`): `internal/roadmap/deferred.go` + `deferred_validate.go` + `cmd/centinela/roadmap_defer.go` + tests. Agents can capture; nothing reads it yet; validate untouched. Shippable alone.
- Step 2 — Slice 2 (visibility): `internal/ui/render_deferred.go`, wire into `runRoadmap`, `defer --list`. Findings are now seen at triage points.
- Step 3 — Slice 3 (`roadmap promote`): promote orchestration + raw-preserving artifact append + validate-after-write + tests (incl. golden-file preservation). Closes the loop into the real roadmap.
- Step 4 — Slice 4 (prompt contract): edit four prompts + four mirrors (byte-identical), extend parity test if needed. Done last so the obligation never points at a command that doesn't exist yet.
- Post-v1 (can wait): `defer dismiss`, dedupe heuristics, defer from the other role prompts, evidence-schema field for deferred slugs, ROADMAP.md auto-sync (owned by roadmap-doc-sync).

#### Handoff
- Next role: feature-specialist
- Outstanding questions:
  - Exact `--source` default resolution inside a worktree (reuse hook CWD resolution vs mandatory flag).
  - `--scores` flag shape on promote (one CSV flag vs six flags vs prompt-driven quality-evaluator subagent invocation).
  - Whether parity tests currently cover the four prompt files; extend if not.
  - Where the slug-validation rule should live to satisfy G2 import-graph constraints (reuse `worktree.ValidateFeatureSlug`, extract to a shared package, or duplicate the ~10-line check).
