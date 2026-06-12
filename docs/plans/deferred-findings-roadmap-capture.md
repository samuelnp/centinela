# Plan: deferred-findings-roadmap-capture

- Feature brief: `docs/features/deferred-findings-roadmap-capture.md`
- Roadmap: Phase 5 — Operability & DX
- Author role: big-thinker
- Date: 2026-06-12

## 1. Problem framing

Four roles already produce deferred knowledge as a *mandatory* part of their
reports — big-thinker's "Out" bullets, feature-specialist's `#### Out-of-Scope`,
senior-engineer's "Outstanding TODOs", qa-senior/edge-case-tester's
`#### Residual Risks` — and all of it dies in prose under `.workflow/`. The
roadmap (`.workflow/roadmap.json`) is the single planning source of truth, yet
no mechanism connects these capture points to it. The operator re-discovers
(or ships) the same gaps later. The hurt is concrete and current: this very
roadmap has drifted by hand twice (see Phase 5 preamble in ROADMAP.md), and the
397-entry legacy memory corpus is full of findings that never became features.

Why now: Phase 5 is "Operability & DX — keep Centinela's own artifacts honest
and self-healing." This is the cheapest remaining leak of already-paid-for
information, and later phases (Phase 7 instrumentation, Phase 8 continuous
governance) assume findings are machine-captured, not buried in markdown.

## 2. Decision: capture mechanism — design (b), ledger + explicit promote

**Chosen: (b) a deferred-findings ledger written by `centinela roadmap defer`,
surfaced by `centinela roadmap`, and promoted into a real roadmap phase by an
explicit `centinela roadmap promote` at triage time.** With one refinement:
the ledger is **one file per finding** (`.workflow/deferred/<slug>.json`), not
a single shared JSON.

### Why not (a) — atomic triple-write into roadmap.json at defer time

1. **It games the quality gate.** `roadmap validate` requires every
   roadmap.json feature to appear in `roadmap-quality.json` with
   `overall ≥ 9` and a non-empty summary (`internal/roadmap/quality.go`).
   A raw mid-step finding has *not* been through senior-PM analysis or
   quality evaluation. Design (a) keeps validate green only by fabricating
   a ≥ 9 score for an untriaged one-liner — exactly the "lower the gate to
   pass it" anti-pattern this project forbids. The score would be a lie the
   moment it is written.
2. **It corrupts derived planning state.** `DeriveReadiness`/`ReadySet`
   (`internal/roadmap/readiness.go`) and the `start` dependency guard
   (`cmd/centinela/start_guard.go`) treat every roadmap.json feature as a
   schedulable, started-able unit. Untriaged findings would show up in
   `centinela roadmap ready` and be startable, despite having no brief, no
   phase decision, and no dependency analysis. The brief's non-goal ("no
   auto-prioritization") is violated structurally by (a).
3. **It guarantees merge conflicts.** `.workflow/` is git-tracked and each
   worktree carries its own committed copy. `worktree.Merge` is a plain
   `git merge --no-ff` (`internal/worktree/merger.go`). Two concurrent
   worktrees each appending to the *same three* JSON arrays
   (roadmap.json + analysis + quality) conflict at merge time and summon
   the Merge Steward for what should be a trivial append.

### Why (b) works

1. **Validate stays green by construction** because `roadmap validate` never
   reads the ledger. No fake scores, no gate games.
2. **Worktree-safe by construction.** Each finding is a *new file* under
   `.workflow/deferred/`. Git merges file-adds cleanly; the only possible
   conflict is two worktrees choosing the same slug (add/add), which is
   precisely the collision we *want* surfaced, and the existing Merge
   Steward path already handles it. This answers "how does a finding inside
   `.worktrees/<feature>/` reach the root roadmap": it rides the feature
   branch and lands in root `.workflow/deferred/` at `centinela merge`
   time, like every other `.workflow` artifact.
3. **Promote is where honesty is possible.** At triage (on root main, post-
   merge), a human or triage agent decides the phase and supplies real
   quality scores. `roadmap promote` then performs the (a)-style atomic
   triple-write — roadmap.json + analysis + quality — in the one place it
   is legitimate, keeping validate green and the greenfield start-guard
   satisfied (a known failure mode: growing roadmap.json without
   regenerating analysis/quality blocks ALL starts).

## 3. Scope boundaries

**In (v1):**
- `centinela roadmap defer <slug> --summary <text> [--source <feature>/<role>]`
  — writes `.workflow/deferred/<slug>.json`; rejects slug collisions against
  the ledger and against roadmap.json feature names.
- Deferred findings rendered in `centinela roadmap` output (count + open
  list) so they are seen at every roadmap glance, plus machine-readable
  listing via `centinela roadmap defer --list` (or `defer list`).
- `centinela roadmap promote <slug> --phase <name> --summary <text>
  --scores <ac,uv,dc,dep,ee,overall>` — atomically appends the feature to
  the named phase in roadmap.json, an entry to roadmap-analysis.json, and a
  scored entry to roadmap-quality.json (validating ≥ 9 before writing),
  appends a bullet to both companion .md files, marks the ledger entry
  `promoted`, and finishes by running the same checks as `roadmap validate`.
- Prompt contract: a required "Deferred findings" obligation in the four
  role prompts (big-thinker, feature-specialist, senior-engineer, qa-senior)
  and their byte-identical scaffold mirrors under
  `internal/scaffold/assets/docs/architecture/`.

**Out (v1)** — per the brief's non-goals, plus plan-level exclusions:
- No auto-prioritization/auto-scheduling; promote requires explicit phase
  and scores from the operator/triage agent.
- No validator hard-gate on "did the agent defer everything" (unverifiable).
- No change to gates or claim verification; no evidence-contract schema
  change (deferred slugs are referenced in report prose, not evidence JSON).
- No retroactive backfill of legacy Residual Risks / memory corpus.
- No ROADMAP.md (human file) sync — that is `roadmap-doc-sync`'s job; v1
  prints a reminder after promote to update ROADMAP.md by hand.
- No dedupe/similarity detection between findings (exact slug match only).
- No `defer` from ux-ui-specialist / validation-specialist / gatekeeper
  prompts (can follow trivially once the pattern is proven on four roles).

## 4. CLI surface and data shapes

### Ledger entry — `.workflow/deferred/<slug>.json`

```json
{
  "slug": "hook-timeout-config",
  "summary": "Prewrite hook timeout is hardcoded; should be configurable",
  "source": { "feature": "deferred-findings-roadmap-capture", "role": "senior-engineer" },
  "status": "open",
  "createdAt": "2026-06-12T09:00:00Z"
}
```

- `slug`: validated with the existing `worktree.ValidateFeatureSlug` rules
  (reuse, don't duplicate).
- `status`: `open` | `promoted` | `dismissed` (v1 writes `open`; `promote`
  sets `promoted`; `dismissed` reserved — `defer dismiss` is a fast-follow,
  not v1).
- `source` optional; when run inside a worktree with an active workflow,
  default it from the workflow state if cheaply available, else require the
  flag. (Feature-specialist to pin down: read `.workflow/<slug>-workflow.json`
  current feature/step the same way hooks resolve it.)

### Promote semantics

1. Load root roadmap.json; fail if slug already a feature (re-check at root —
   the worktree-side check at defer time may have used a stale copy).
2. Append `{name: slug}` to the named phase (create phase if `--phase` names
   a new one? **No — v1 requires an existing phase name**, to avoid silently
   forking phase taxonomy; error lists known phases).
3. Append to `roadmap-analysis.json` and `roadmap-quality.json` using
   raw-JSON-preserving read-modify-write (`map[string]any` /
   `json.RawMessage` for untouched entries) — the live analysis file carries
   `dependsOn` keys the Go struct dropped in Option B; promote must not
   destroy unknown fields.
4. Validate scores (each 1–10, overall ≥ 9) *before* any write; write all
   three files via temp-file + rename; then run ValidateAnalysis +
   ValidateQuality and report.
5. Update ledger entry status → `promoted`, add `promotedAt`.

### Prompt contract addition (all four prompts + mirrors)

A short, uniform section inserted before each prompt's Handoff section:

```
#### Deferred Findings
- For every out-of-scope detection / not-fixed-now finding, run:
  `centinela roadmap defer <slug> --summary "<one line>" --source <feature>/<role>`
- List the recorded slugs here, or state "none".
```

Wording per role references its existing section (Out bullets, Out-of-Scope,
Outstanding TODOs, Residual Risks) so the obligation lands where the prose
already exists. The scaffold mirrors must be updated byte-identically; note
the parity test only covers some docs — update both sides regardless and
extend the parity test to cover these four prompt files if not already
covered (feature-specialist to verify which files
`internal/scaffold` parity-tests).

## 5. Files to touch (all new Go files ≤ 100 lines — G1)

| File | Action | Notes |
|------|--------|-------|
| `internal/roadmap/deferred.go` | new | `Deferred` struct, `DeferredDir` const, Load/Save one entry, ListOpen |
| `internal/roadmap/deferred_validate.go` | new | slug validation (delegating to shared slug rules), collision check vs ledger + roadmapFeatureSet |
| `internal/roadmap/promote.go` | new | promote orchestration: load/append/validate/write roadmap.json |
| `internal/roadmap/promote_artifacts.go` | new | raw-preserving append to analysis/quality JSON + md bullet append |
| `cmd/centinela/roadmap_defer.go` | new | cobra `roadmap defer` (+ `--list`) |
| `cmd/centinela/roadmap_promote.go` | new | cobra `roadmap promote` with flags |
| `internal/ui/render_deferred.go` | new | deferred section for `centinela roadmap` + list rendering |
| `cmd/centinela/roadmap.go` | edit | include deferred summary in `runRoadmap` output |
| `docs/architecture/big-thinker-prompt.md` | edit | Deferred Findings section |
| `docs/architecture/feature-specialist-prompt.md` | edit | same |
| `docs/architecture/senior-engineer-prompt.md` | edit | same |
| `docs/architecture/qa-senior-prompt.md` | edit | same |
| `internal/scaffold/assets/docs/architecture/*-prompt.md` (4) | edit | byte-identical mirrors |
| colocated `_test.go` files | new | per-package coverage ≥ 95%; tests are source — ≤ 100 lines each, split as needed |
| `specs/deferred-findings-roadmap-capture.feature` | new | Gherkin acceptance |

Slug-validation note: `worktree.ValidateFeatureSlug` lives in
`internal/worktree`; importing worktree from roadmap may be an unwanted
edge in the import graph. If so, extract the slug rule into a tiny shared
package (or duplicate the ~10-line check in `deferred_validate.go` with a
comment) — feature-specialist decides after checking G2 import-graph rules.

## 6. Dependencies & assumptions

- **Internal modules:** `internal/roadmap` (Load/Save, roadmapFeatureSet,
  ValidateAnalysis/Quality), `internal/ui` (render helpers, panel styles),
  `internal/worktree` (slug rules, merge behavior — read-only dependency),
  `cmd/centinela` cobra wiring.
- **Prior features this builds on:** roadmap dependencies/readiness (Option B
  shape: deps on roadmap.json, analysis has name-only entries), worktree
  merge + Merge Steward, evidence CLI (for this workflow's own artifacts),
  scaffold mirror parity discipline.
- **Assumptions:**
  - `.workflow/` remains git-tracked in worktrees and merges via plain git —
    verified in `internal/worktree/merger.go`.
  - `roadmap validate` reads only roadmap.json/analysis/quality — verified;
    the ledger is invisible to it.
  - Quality threshold stays 9 and the role strings
    (`senior-product-manager`, `roadmap-quality-evaluator`) are stable
    constants promote must reuse, not re-declare.
  - Findings are append-only at capture time; mutation (promote/dismiss)
    happens only at the root checkout.

## 7. Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Promote's read-modify-write of analysis/quality JSON drops unknown fields (e.g. legacy `dependsOn` in analysis) | High — corrupts source-of-truth artifacts | Medium | Raw-preserving JSON (map[string]any / RawMessage); golden-file test asserting byte-stable untouched entries |
| Two worktrees defer the same slug → add/add merge conflict | Low — Merge Steward already handles; conflict is the desired signal | Low | Document in prompt contract: prefer source-prefixed or specific slugs; collision check at defer vs local ledger + roadmap |
| Worktree's stale roadmap.json lets defer accept a slug that root roadmap already has | Medium — promote would fail later | Medium | Promote re-checks at root and errors cleanly; ledger entry stays `open` and can be re-slugged |
| Agents fabricate or skip `--summary`, producing junk findings | Medium — ledger noise erodes trust | Medium | Non-empty summary enforced by CLI; prompt requires listing slugs in the report so the orchestrator can spot junk; triage (promote) is the human filter |
| Prompt mirrors drift (parity test may not cover all four prompt files) | Medium — scaffolded projects get stale contract | Medium | Update both sides in the same commit; extend parity test coverage to these prompts |
| `centinela roadmap` output regression (existing render + new deferred section) | Medium — regresses an earlier feature's UX/tests | Low | Deferred section renders only when count > 0; existing render tests untouched, new tests added |
| Promote breaks greenfield start-guard expectations (analysis/quality must cover ALL features) | High — blocks every `centinela start` | Low | Promote writes all three artifacts atomically and runs validate as its last step; failing validate rolls back nothing silently — it reports loudly before anyone runs `start` |
| New Go files exceed 100-line G1 limit | Low | Medium | File split already planned (deferred vs promote vs artifacts vs ui); tests split per concern |

## 8. Rollout sequence

1. **Slice 1 — ledger core + `roadmap defer`** (smallest correct slice):
   `internal/roadmap/deferred.go` + `deferred_validate.go` +
   `cmd/centinela/roadmap_defer.go` + tests. An agent can capture; nothing
   reads it yet, validate untouched. Shippable alone.
2. **Slice 2 — visibility:** `internal/ui/render_deferred.go`, wire into
   `runRoadmap`, `defer --list`. Findings are now seen at triage points.
3. **Slice 3 — `roadmap promote`:** promote orchestration + raw-preserving
   artifact append + validate-after-write + tests (including golden-file
   preservation test). Closes the loop into the real roadmap.
4. **Slice 4 — prompt contract:** edit four prompts + four mirrors
   (byte-identical), extend parity test if needed. Done last so the
   obligation never points at a command that doesn't exist yet.
5. **Can wait (post-v1):** `defer dismiss`, dedupe heuristics, defer from
   the other role prompts, evidence-schema field for deferred slugs,
   ROADMAP.md auto-sync (owned by `roadmap-doc-sync`).

## 9. Open questions for feature-specialist

- Exact `--source` default resolution inside a worktree (reuse hook CWD
  resolution vs mandatory flag).
- `--scores` flag shape on promote (one CSV flag vs six flags vs prompt-
  driven quality-evaluator subagent invocation).
- Whether parity tests currently cover the four prompt files; extend if not.
- Where the slug-validation rule should live to satisfy G2 import-graph
  constraints.
