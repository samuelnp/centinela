# Plan: deferred-findings-roadmap-capture

- Feature brief: `docs/features/deferred-findings-roadmap-capture.md`
- Roadmap: Phase 5 — Operability & DX
- Author role: big-thinker
- Date: 2026-06-12
- Revision: 2 (post operator review — design changed from "separate ledger"
  to "validate-exempt backlog phase in roadmap.json"; four operator decisions
  are binding and recorded verbatim in §2.1).

## 1. Problem framing

Eight workflow roles produce deferred knowledge as a *mandatory* part of their
reports — big-thinker's "Out" bullets, feature-specialist's `#### Out-of-Scope`,
senior-engineer's "Outstanding TODOs", qa-senior + edge-case-tester's
`#### Residual Risks`, and (newly in scope this revision) ux-ui-specialist,
validation-specialist, and gatekeeper findings — and all of it dies in prose
under `.workflow/`. The roadmap (`.workflow/roadmap.json`) is the single
planning source of truth, yet no mechanism connects these capture points to it.
The operator re-discovers (or ships) the same gaps later. The hurt is concrete:
this very roadmap has drifted by hand twice (Phase 5 preamble in ROADMAP.md),
and the 397-entry legacy memory corpus is full of findings that never became
features.

Why now: Phase 5 is "Operability & DX — keep Centinela's own artifacts honest
and self-healing." This is the cheapest remaining leak of already-paid-for
information, and later phases (Phase 7 instrumentation, Phase 8 continuous
governance) assume findings are machine-captured, not buried in markdown.

## 2. Decision: capture mechanism

### 2.1 Operator decisions (binding — do not relitigate)

1. **Capture = a dedicated Backlog phase in `roadmap.json`, validate-exempt.**
   `centinela roadmap defer <slug> --summary <text> [--source <feature>/<role>]`
   appends the finding directly to a Backlog phase in `roadmap.json` (no
   separate ledger directory). `roadmap validate` (ValidateAnalysis +
   ValidateQuality coverage), `roadmap ready`/readiness, and the
   `start` dependency guard MUST EXEMPT backlog-phase features: no
   analysis/quality entries are required for them, they never surface as
   ready/startable, and `centinela start <backlog-slug>` is refused with a
   "promote it first" error. Finding metadata (summary, source, deferredAt)
   lives as optional `omitempty` fields on the `Feature` entry itself; load /
   write must preserve them and any unknown fields on *other* entries
   (raw-preserving read-modify-write). The operator ACCEPTED the worktree
   merge-conflict risk of concurrent appends to one array; it is mitigated
   honestly (one-entry-per-line array formatting → conflicts are a trivial
   union) and recorded as an accepted risk (§7).

2. **Promote scoring = quality-evaluator agent path.**
   `centinela roadmap promote <slug> --phase <name> [--summary <text>]
   [--scores ac,uv,dc,dep,ee,overall]`. WITH `--scores`: non-interactive,
   validates each dimension 1–10 and overall ≥ 9 BEFORE any write. WITHOUT
   `--scores`: promote prints the roadmap-quality-evaluator prompt context
   (the finding's name/summary/source, the threshold, the expected scores
   schema, and instructions) and writes NOTHING, so the orchestrator can run
   an honest scoring agent and re-invoke promote with `--scores`. Promote
   moves the entry OUT of the Backlog phase into the target phase, appends
   analysis + quality entries (raw-preserving), appends bullets to the two
   companion `.md` files, then runs validate as its last step.

3. **Prompt contract = ALL role prompts in v1**: big-thinker,
   feature-specialist, senior-engineer, qa-senior, ux-ui-specialist,
   validation-specialist, gatekeeper, and edge-case-tester. Uniform required
   "Deferred Findings" section, wording anchored to each role's existing
   deferred-prose section. Scaffold mirrors updated byte-identically.

4. **Naming stays** `centinela roadmap defer` / `centinela roadmap promote`.

### 2.2 Canonical backlog phase

- **Phase name: `Backlog`** (exact string, case-sensitive match in code, via
  an `isBacklogPhaseName` helper that lower-cases and trims, mirroring
  `isBootstrapPhaseName` in `internal/roadmap/bootstrap.go`).
- **`defer` creates the Backlog phase if absent**, appending it as the last
  phase; otherwise it appends the finding to the existing Backlog phase.
- A Backlog feature is *any* `Feature` inside a phase whose name matches
  `isBacklogPhaseName`. That single predicate drives every exemption point, so
  the exemption is defined in exactly one place.

### 2.3 Superseded design (recorded for honesty)

Revision 1 chose a **separate one-file-per-finding ledger**
(`.workflow/deferred/<slug>.json`) precisely to keep validate green *by
construction* (validate never read the ledger) and to make worktree merges
conflict-free *by construction* (file-adds merge cleanly; only same-slug
add/add collided). The operator has overridden this in favour of a single
in-roadmap Backlog phase. The honest cost of that override, now accepted:

- **Validate is no longer green by construction — it is green by exemption.**
  Coverage checks must actively skip Backlog features. A regression that drops
  the exemption would either (a) demand analysis/quality for raw findings
  (breaking validate) or (b) over-broaden the skip and stop demanding coverage
  for *real* features (silently weakening the gate). This is a real new risk
  (§7) the ledger design did not have.
- **Merges are no longer conflict-free by construction.** Two worktrees
  appending to the same `Backlog.features` array touch the same JSON region
  and CAN conflict at `git merge` time. Accepted by the operator; mitigated by
  one-entry-per-line formatting so the conflict is a trivial textual union the
  Merge Steward resolves by keeping both lines.

The upside the operator is buying: one source of truth (no second artifact to
discover, render, or keep in sync), and findings are visible in the same
`roadmap.json` everything else already reads.

## 3. Scope boundaries

**In (v1):**
- `centinela roadmap defer <slug> --summary <text> [--source <feature>/<role>]`
  — appends a `Feature{Name, Summary, Source, DeferredAt}` to the `Backlog`
  phase (creating it if absent), via raw-preserving read-modify-write; rejects
  slug collisions against every existing roadmap feature name (Backlog or not)
  and invalid slugs.
- `--source` defaults via `worktree.DetectFeatureFromCwd` when run inside a
  worktree (resolved decision, feature-specialist round 1); flag is an explicit
  override and is optional everywhere.
- Backlog findings rendered in `centinela roadmap` output (a distinct Backlog
  section showing slug + summary), so they are seen at every roadmap glance.
- `centinela roadmap promote <slug> --phase <name> [--summary <text>]
  [--scores ac,uv,dc,dep,ee,overall]` — the evaluator-path / scored-path
  behavior of operator decision #2.
- Required "Deferred Findings" section in **all eight** role prompts and their
  byte-identical scaffold mirrors.

**Out (v1)** — per brief non-goals + plan exclusions:
- No auto-prioritization/auto-scheduling; promote requires an explicit phase
  and (eventually) real scores.
- No validator hard-gate on "did the agent defer everything" (unverifiable;
  contract stays prompt-level).
- No change to gates or claim verification; no evidence-contract schema change
  (deferred slugs are named in report prose, not evidence JSON).
- No retroactive backfill of legacy Residual Risks / memory corpus.
- No ROADMAP.md (human file) sync — that is `roadmap-doc-sync`'s job; v1 prints
  a reminder after promote.
- No dedupe/similarity detection (exact slug match only).
- No `defer dismiss` / status lifecycle on a Backlog entry (a Backlog feature
  is simply "present until promoted"); removing a finding = editing
  roadmap.json by hand for now.

## 4. CLI surface and data shapes

### 4.1 Backlog entry — extension of the `Feature` struct (roadmap.json)

The `Feature` struct gains three optional `omitempty` fields:

```go
type Feature struct {
    Name       string   `json:"name"`
    DependsOn  []string `json:"dependsOn,omitempty"`
    Summary    string   `json:"summary,omitempty"`    // deferred-finding one-liner
    Source     *Source  `json:"source,omitempty"`     // {feature, role}, both omitempty
    DeferredAt string   `json:"deferredAt,omitempty"` // RFC3339 capture time
}
type Source struct {
    Feature string `json:"feature,omitempty"`
    Role    string `json:"role,omitempty"`
}
```

A Backlog feature in `roadmap.json` then reads:

```json
{ "name": "Backlog", "features": [
  { "name": "hook-timeout-config",
    "summary": "Prewrite hook timeout is hardcoded; should be configurable",
    "source": { "feature": "deferred-findings-roadmap-capture", "role": "senior-engineer" },
    "deferredAt": "2026-06-12T09:00:00Z" }
] }
```

- `omitempty` means non-Backlog features serialize exactly as today (no new
  keys) — **no churn on existing entries**, which keeps the parity/golden
  diffs clean and avoids merge noise on unrelated lines.
- **Raw-preserving I/O is mandatory.** The live `roadmap-analysis.json` already
  carries fields the Go structs drop, and round-tripping the whole roadmap
  through `MarshalIndent` of the typed struct would re-key every entry and
  risk dropping unknown fields. `defer` and `promote` must read roadmap.json as
  `map[string]any` / `json.RawMessage`, mutate only the Backlog array (and, for
  promote, the source + target phase arrays), and write back — preserving every
  untouched entry byte-for-byte. The array is formatted one entry per line so
  concurrent appends conflict as a trivial union.

### 4.2 `defer` semantics

1. Read roadmap.json raw. Build the feature-name set across ALL phases.
2. Validate slug (kebab-case rule duplicated in `internal/roadmap` with a
   `// mirrors worktree.ValidateFeatureSlug` comment — resolved decision, G2
   import-graph: no new edge from roadmap → worktree).
3. Reject empty summary, reject slug collision against any existing feature.
4. Resolve `--source`: explicit flag wins; else
   `worktree.DetectFeatureFromCwd(os.Getwd())` populates `source.feature`;
   role left blank if not supplied; source omitted entirely outside a worktree.
5. Append the new entry to the `Backlog` phase (create the phase if absent),
   stamp `deferredAt = now (RFC3339)`, write raw.

### 4.3 `promote` semantics

1. Load roadmap.json raw (at the root checkout). Locate the slug **inside the
   Backlog phase**; error if absent ("not a backlog finding") or if it already
   exists as a non-Backlog feature.
2. **No `--scores` → evaluator path (writes NOTHING):** print the
   roadmap-quality-evaluator context to stdout and exit 0:
   - the finding's `name`, `summary`, `source`;
   - target `--phase`;
   - the threshold (`9`) and the six-dimension schema
     (`acceptanceCriteria, userValue, definitionClarity, dependencies,
     effortEstimation, overall`, each 1–10);
   - the exact re-invocation line:
     `centinela roadmap promote <slug> --phase <name> --scores ac,uv,dc,dep,ee,overall`;
   - a one-line instruction: run an honest quality-evaluator pass, then
     re-invoke with `--scores`.
   This block is rendered by a `ui` helper so it is testable and i18n-routed.
3. **With `--scores` → scored path:** parse the CSV (exactly six ints),
   validate each 1–10 and overall ≥ 9 **before any write**. On failure: error,
   zero writes.
4. Reject an unknown `--phase` (enumerate known non-Backlog phases). v1 does
   not create the target phase.
5. **Atomic move + append (raw-preserving):**
   - remove the entry from the Backlog array;
   - append `{name: slug, dependsOn: []}` to the target phase — **strip the
     finding metadata** (`summary` moves into the quality entry's summary,
     `source`/`deferredAt` are provenance the roadmap feature no longer needs);
     decision pinned: **keep `summary` as the quality-entry summary; drop
     `source` and `deferredAt`** (provenance is captured in the
     analysis/quality `.md` bullet appended below, so it is not lost);
   - append a name-only entry to `roadmap-analysis.json`
     (`{name: slug}`) raw-preserving;
   - append a scored entry to `roadmap-quality.json`
     (`{name, scores, summary}`) raw-preserving;
   - append a provenance bullet to `roadmap-analysis.md` and
     `roadmap-quality.md` (records original source + deferredAt);
   - all writes via temp-file + rename.
6. Run `ValidateAnalysis` + `ValidateQuality` as the final step and report.
   If validate fails after the writes, report loudly (the next `centinela
   start` would otherwise be blocked) — promote is the one legitimate place
   for the atomic triple-write, exactly as in the bootstrap-artifacts story.

### 4.4 Prompt contract addition (all eight prompts + mirrors)

A short, uniform section inserted near each prompt's existing deferred-prose
anchor (big-thinker "Out"/Outstanding questions; feature-specialist
`#### Out-of-Scope`; senior-engineer "Outstanding TODOs"; qa-senior +
edge-case-tester `#### Residual Risks`; ux-ui-specialist deferred UX notes;
validation-specialist deferred validation gaps; gatekeeper deferred
remediations):

```
#### Deferred Findings
- For every out-of-scope detection / not-fixed-now finding, run:
  `centinela roadmap defer <slug> --summary "<one line>" --source <feature>/<role>`
- List the recorded slugs here, or state "none".
```

Wording per role references that role's existing section so the obligation
lands where the prose already exists. Mirrors under
`internal/scaffold/assets/docs/architecture/` updated byte-identically.

**Parity coverage (verified):** `TestExtractAgentSharedBlocks_ScaffoldMirrorParity`
in `tests/acceptance/extract_agent_shared_blocks_acceptance_test.go` already
byte-compares all eight role prompts against their mirrors — its
`promptsReferencingInvocation` slice lists gatekeeper, edge-case-tester,
big-thinker, feature-specialist, senior-engineer, qa-senior, ux-ui-specialist,
validation-specialist (plus the production-readiness template and the two
shared reference files). **No parity-test extension is needed.** Every one of
the eight prompts has an existing mirror file. The job is purely: edit each
source prompt and its mirror identically in the same commit.

## 5. Files to touch (all new Go files ≤ 100 lines — G1)

| File | Action | Notes |
|------|--------|-------|
| `internal/roadmap/roadmap.go` | edit | extend `Feature` (Summary/Source/DeferredAt omitempty) + `Source` type; keep typed `Load`/`Save` for read-only callers |
| `internal/roadmap/backlog.go` | new | `BacklogPhaseName` const, `isBacklogPhaseName`, `IsBacklogFeature`, `BacklogFeatures`, `NonBacklogFeatureSet` (mirrors bootstrap.go) |
| `internal/roadmap/rawio.go` | new | raw-preserving read (`map[string]any`/`RawMessage`) + temp-file+rename write of roadmap.json with one-entry-per-line array formatting |
| `internal/roadmap/defer.go` | new | append-to-Backlog orchestration (slug validate, collision, source resolve, stamp) |
| `internal/roadmap/defer_validate.go` | new | slug rule duplicated w/ `// mirrors worktree.ValidateFeatureSlug`; collision vs full feature set |
| `internal/roadmap/promote.go` | new | move-out-of-Backlog + score validate + phase check orchestration |
| `internal/roadmap/promote_artifacts.go` | new | raw-preserving append to analysis/quality JSON + md bullet append |
| `internal/roadmap/analysis.go` | edit | `roadmapFeatureSet` → exempt Backlog (use `NonBacklogFeatureSet`) so ValidateAnalysis ignores backlog findings |
| `internal/roadmap/quality.go` | edit | same exemption — ValidateQuality must not demand scores for backlog findings (reuses `roadmapFeatureSet`) |
| `internal/roadmap/readiness.go` | edit | `DeriveReadiness` skips Backlog-phase features (they are never ready/blocked/startable) |
| `cmd/centinela/start_guard.go` | edit | `workflowOrderForFeature`: if `IsBacklogFeature`, refuse with "promote it first" |
| `cmd/centinela/roadmap_defer.go` | new | cobra `roadmap defer` |
| `cmd/centinela/roadmap_promote.go` | new | cobra `roadmap promote` with `--phase/--summary/--scores` |
| `cmd/centinela/roadmap.go` | edit | render Backlog section in `runRoadmap` |
| `internal/ui/render_backlog.go` | new | Backlog findings section + promote evaluator-context block |
| `docs/architecture/*-prompt.md` (8) | edit | Deferred Findings section (big-thinker, feature-specialist, senior-engineer, qa-senior, ux-ui-specialist, validation-specialist, gatekeeper, edge-case-tester) |
| `internal/scaffold/assets/docs/architecture/*-prompt.md` (8) | edit | byte-identical mirrors |
| colocated `_test.go` files | new | per-package coverage ≥ 95%; tests are source — ≤ 100 lines each |
| `specs/deferred-findings-roadmap-capture.feature` | edit/new | Gherkin acceptance for the revised design |

Note: `RenderRoadmap` already calls `DeriveReadiness`; once Backlog features
are skipped there, they won't render in the normal phase loop — the new
`render_backlog.go` section renders them separately and only when present.

## 6. Dependencies & assumptions

- **Internal modules:** `internal/roadmap` (Load/Save, feature-set helpers,
  ValidateAnalysis/Quality, readiness), `internal/ui` (render helpers, panel
  styles, i18n routing), `internal/worktree` (`DetectFeatureFromCwd`,
  slug-rule source-of-truth referenced by comment — no import edge),
  `cmd/centinela` cobra wiring + `start_guard`.
- **Builds on:** roadmap dependencies/readiness (Option B shape — deps on
  roadmap.json, analysis name-only), bootstrap-phase predicate pattern
  (`bootstrap.go`), worktree merge + Merge Steward, evidence CLI, scaffold
  mirror parity discipline.
- **Assumptions (verified in code this round):**
  - `roadmapFeatureSet` is the single coverage gate shared by ValidateAnalysis
    *and* ValidateQuality — exempting Backlog there covers both at once.
  - `DeriveReadiness` is the single source for `ReadySet`, `RenderRoadmap`, and
    UnmetDependencies enumeration — skipping Backlog there covers ready/render.
  - `workflowOrderForFeature` is the only path `centinela start` takes to a
    roadmap feature — the backlog refusal belongs there.
  - Quality threshold stays 9; role-string constants (`senior-product-manager`,
    `roadmap-quality-evaluator`) are reused, not re-declared.
  - `.workflow/` stays git-tracked and merges via plain `git merge --no-ff`.

## 7. Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Concurrent defer appends to the single `Backlog.features` array → git merge conflict** (ACCEPTED by operator) | Low — Merge Steward resolves; conflict is a trivial union | Medium | One-entry-per-line array formatting so both lines survive; document in Merge Steward expectation; raw-preserving write touches only the Backlog region |
| **Validate-exemption regression — exempting Backlog weakens coverage for *non-Backlog* features** | High — silently lowers the quality gate (the anti-pattern this project forbids) | Medium | Exemption defined in ONE predicate (`isBacklogPhaseName`/`NonBacklogFeatureSet`); tests assert: (a) a real non-Backlog feature still REQUIRES analysis+quality, (b) a Backlog feature does NOT, (c) a feature named like Backlog but in a normal phase is unaffected |
| **Start-guard bypass — a backlog slug becomes startable** | Medium — un-triaged finding runs a full workflow with no brief/scores | Low | `workflowOrderForFeature` refuses `IsBacklogFeature` with "promote it first"; test asserts refusal; readiness never lists backlog features |
| **Raw read-modify-write drops unknown fields on other entries / re-keys existing ones** | High — corrupts source-of-truth roadmap/analysis/quality | Medium | `map[string]any`/`RawMessage` raw I/O; `omitempty` so existing entries are untouched; golden-file test asserts byte-stable untouched entries |
| **Promote scored path writes a fabricated ≥9 score** | Medium — gate-gaming if agents guess scores | Medium | Default (no `--scores`) path prints evaluator context and writes nothing, steering toward an honest agent pass; scores validated 1–10 + overall ≥9 before write |
| **Promote partial write leaves roadmap.json mutated but analysis/quality not** | High — blocks every `centinela start` | Low | Validate scores BEFORE any write; temp-file+rename per file; validate runs last and reports loudly |
| **Prompt mirror drift across eight pairs** | Medium — scaffolded projects get a stale contract | Low | Parity test already covers all eight pairs byte-for-byte (verified); edit both sides in one commit |
| **`centinela roadmap` render regression** | Medium — regresses prior UX/tests | Low | Backlog section renders only when present; existing phase render untouched (Backlog skipped in the normal loop) |
| **New Go files exceed 100-line G1 limit** | Low | Medium | Split planned (backlog/rawio/defer/defer_validate/promote/promote_artifacts/ui); tests split per concern |

## 8. Rollout sequence

1. **Slice 1 — roadmap struct + backlog exemptions.** Extend `Feature` +
   `Source`; add `backlog.go` predicate; wire the exemption into
   `roadmapFeatureSet` (analysis + quality), `DeriveReadiness`, and
   `workflowOrderForFeature` (start refusal). Tests: validate still demands
   coverage for real features, exempts Backlog; start refuses a backlog slug;
   readiness omits Backlog. No new command yet — this is the safety net the
   rest depends on, so it ships first.
2. **Slice 2 — `roadmap defer` + rendering.** `rawio.go`, `defer.go`,
   `defer_validate.go`, `cmd/centinela/roadmap_defer.go`,
   `internal/ui/render_backlog.go`, wire into `runRoadmap`. Agents can capture;
   findings are visible; validate stays green (exemption from Slice 1).
3. **Slice 3 — `roadmap promote` incl. evaluator path.** `promote.go`,
   `promote_artifacts.go`, `cmd/centinela/roadmap_promote.go`, evaluator-context
   ui block. Tests: no-`--scores` prints context and writes nothing; scored
   path moves entry, appends three artifacts raw-preserving, validate passes;
   below-threshold + unknown-phase reject with zero writes; golden-file
   preservation.
4. **Slice 4 — prompt contract everywhere.** Edit all eight source prompts +
   eight mirrors byte-identically. Done last so the obligation never points at
   a command that does not exist. Parity test already green by construction.
5. **Post-v1 (can wait):** `defer dismiss`/lifecycle, dedupe heuristics,
   evidence-schema field for deferred slugs, ROADMAP.md auto-sync
   (`roadmap-doc-sync`).

## 9. Open questions for feature-specialist

- **Exact promote evaluator-context format.** Precise stdout block for the
  no-`--scores` path — fields, ordering, the literal re-invocation line, and
  whether it routes through an i18n key (it should, per Hard Rule 7).
- **Backlog rendering shape.** Panel vs inline list; show summary inline or
  count + slug only; placement relative to the phase overview.
- **`promote` metadata-stripping confirmation.** Plan pins: keep `summary` as
  the quality-entry summary, drop `source`/`deferredAt` from the roadmap
  feature but record them in the `.md` provenance bullet. Confirm the bullet
  wording and that analysis/quality `.md` are the right provenance home.
- **One-entry-per-line array formatting.** Confirm the raw-writer emits the
  `Backlog.features` array one object per line (merge-union friendly) while
  leaving other phases' formatting untouched, and that this survives a
  round-trip golden test.
- **Backlog phase placement on creation.** Append as last phase (pinned) — confirm
  this does not perturb `HasBootstrapPhase`/bootstrap ordering or readiness.
