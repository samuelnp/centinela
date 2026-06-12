### Feature-Specialist Report: deferred-findings-roadmap-capture
**Date:** 2026-06-12

#### Behavior Summary

`centinela roadmap defer <slug> --summary <text> [--source <feature>/<role>]` appends a finding to a `Backlog` phase in `.workflow/roadmap.json` (creating the phase if absent, always as the last phase) via raw-preserving read-modify-write. The `Feature` struct gains three `omitempty` fields — `summary`, `source` (`{feature, role}`), and `deferredAt` (RFC3339) — that serialize as absent on non-Backlog entries, producing zero diff churn on existing content. The `Backlog` phase name is matched case-insensitively and trimmed by an `isBacklogPhaseName` helper that mirrors the existing `isBootstrapPhaseName` predicate. Slug validation (kebab-case rule) is duplicated in `internal/roadmap` with a `// mirrors worktree.ValidateFeatureSlug` comment, keeping the G2 import graph edge-free.

`roadmap validate` (ValidateAnalysis + ValidateQuality) exempts Backlog features via `NonBacklogFeatureSet` replacing the current `roadmapFeatureSet` call in `analysis.go` and `quality.go`. `DeriveReadiness` (and therefore `ReadySet`, `RenderRoadmap`, and unmet-dep enumeration) skips Backlog features. `workflowOrderForFeature` in `cmd/centinela/start_guard.go` refuses a Backlog slug with a "promote it first" error before any step-order logic runs.

`centinela roadmap promote <slug> --phase <name> [--scores ac,uv,dc,dep,ee,overall]`: without `--scores`, prints the roadmap-quality-evaluator context block (finding name/summary/source, target phase, threshold 9, six-dimension schema, the literal re-invocation line, one-line instruction to run a quality-evaluator pass) and exits 0 with zero writes. With `--scores` (exactly six comma-separated ints, each 1-10, overall >= 9), validates before any write, then: removes the entry from the Backlog array; appends `{name, dependsOn:[]}` (metadata stripped) to the target phase; appends name-only entry to `roadmap-analysis.json`; appends scored+summary entry to `roadmap-quality.json`; appends provenance bullets to `roadmap-analysis.md` and `roadmap-quality.md`; runs validate last. All writes via temp-file+rename. Prints ROADMAP.md sync reminder on success.

`centinela roadmap` renders a Backlog section (slug + summary, one entry per line) after the phase overview, only when the Backlog phase is present and non-empty. The phase overview loop skips Backlog entries (they are excluded from `DeriveReadiness`).

All eight role prompts (big-thinker, feature-specialist, senior-engineer, qa-senior, edge-case-tester, ux-ui-specialist, validation-specialist, gatekeeper) and their byte-identical scaffold mirrors under `internal/scaffold/assets/docs/architecture/` gain a required `#### Deferred Findings` section anchored near each role's existing deferred-prose section.

#### Gherkin Scenarios

See `specs/deferred-findings-roadmap-capture.feature`. Summary by slice:

- **Slice 1 — defer**: 8 scenarios (happy path creates Backlog with correct fields; appends to existing Backlog without disturbing prior entries; rejects empty summary; rejects slug collision in Backlog phase; rejects slug collision in non-Backlog phase; rejects invalid slug; auto-resolves `--source` from worktree CWD; omits source field when outside a worktree).
- **Slice 2 — rendering + exemptions**: 7 scenarios (Backlog section present; Backlog section absent when phase missing; Backlog section absent when phase empty; Backlog features absent from `roadmap ready`; validate passes with Backlog entries having no analysis/quality; validate still fails for uncovered non-Backlog feature; phase named "Pre-Backlog Work" is not exempt from validate).
- **Slice 2 — start guard**: 1 scenario (start refuses Backlog slug with promote-first error).
- **Slice 3 — promote evaluator path**: 1 scenario (no `--scores` prints evaluator context and writes nothing).
- **Slice 3 — promote scored path**: 2 scenarios (happy path moves entry + appends artifacts + provenance bullets + validate green; raw-preserving I/O preserves unknown fields).
- **Slice 3 — promote rejections**: 5 scenarios (overall < 9; dimension outside 1-10; unknown phase; slug not in Backlog; malformed CSV wrong count).
- **Slice 4 — prompt contract**: 1 scenario (all eight pairs byte-identical; all sixteen files contain Deferred Findings section + `centinela roadmap defer` text).

Total: **25 scenarios**.

#### UX States Table

| Surface | Input / condition | Output | Exit code |
|---------|-------------------|--------|-----------|
| `roadmap defer <slug> --summary <text>` | Valid slug, no collision, valid summary | Backlog entry appended; roadmap.json updated | 0 |
| `roadmap defer <slug>` (in worktree) | No `--source`; CWD inside `.worktrees/<feat>` | Entry appended; `source.feature` auto-populated from CWD | 0 |
| `roadmap defer <slug>` (outside worktree) | No `--source`; CWD not in `.worktrees/` | Entry appended; `source` field absent | 0 |
| `roadmap defer <slug> --summary ""` | Empty summary | Error: summary required/empty; zero writes | non-zero |
| `roadmap defer <slug>` | Slug collision in Backlog or any other phase | Error: slug collision; zero writes | non-zero |
| `roadmap defer "bad slug!"` | Invalid slug format | Error: invalid slug + required format; zero writes | non-zero |
| `centinela roadmap` | Backlog phase present and non-empty | Phase overview + Backlog section (slug + summary per entry) | 0 |
| `centinela roadmap` | Backlog phase absent or empty | Phase overview only; no Backlog section | 0 |
| `centinela roadmap ready` | Backlog features present | Backlog features excluded from ready list | 0 |
| `centinela roadmap validate` | Backlog entries with no analysis/quality coverage | Passes (exempted) | 0 |
| `centinela roadmap validate` | Non-Backlog feature missing analysis/quality | Fails; names the missing feature | non-zero |
| `centinela start <backlog-slug>` | Slug is in Backlog phase | Error: promote it first; zero workflow created | non-zero |
| `roadmap promote <slug> --phase <name>` | No `--scores`; slug in Backlog | Evaluator context printed (name/summary/source/phase/threshold/schema/re-invocation line); zero writes | 0 |
| `roadmap promote <slug> --phase <name> --scores ...` | All scores 1-10, overall >= 9; phase exists | Entry moved; analysis/quality/provenance appended; validate passes; ROADMAP.md reminder printed | 0 |
| `roadmap promote <slug> --phase <name> --scores ...` | overall < 9 | Error: overall must be >= 9; zero writes | non-zero |
| `roadmap promote <slug> --phase <name> --scores ...` | Any dimension outside 1-10 | Error: each score must be 1-10; zero writes | non-zero |
| `roadmap promote <slug> --phase "nonexistent" --scores ...` | Unknown phase | Error: unknown phase + list of known non-Backlog phases; zero writes | non-zero |
| `roadmap promote <slug> --phase <name> --scores ...` | Slug not in Backlog | Error: not a Backlog finding; zero writes | non-zero |
| `roadmap promote <slug> --phase <name> --scores 9,9,9` | Wrong count (not six) | Error: --scores requires exactly six comma-separated integers; zero writes | non-zero |

#### Out-of-Scope

- `defer dismiss` / status lifecycle: a Backlog entry is "present until promoted"; removal is by hand-editing roadmap.json.
- Auto-prioritization or auto-scheduling of Backlog entries.
- Validator hard-gate on "did the agent defer everything it should have" (unverifiable; contract is prompt-level).
- ROADMAP.md auto-sync after promote (promote prints a reminder; actual sync is `roadmap-doc-sync`'s job).
- Dedupe/similarity detection (exact slug match only).
- Retroactive backfill of legacy Residual Risks or the 397-entry memory corpus.
- Evidence-schema field for deferred slugs (post-v1).
- `defer --list` as a separate subcommand (Backlog is visible via `centinela roadmap`).

#### Resolved Clarifications

1. **CSV `--scores` shape (carried from round 1).** Six comma-separated integers in the order `ac,uv,dc,dep,ee,overall` matching `QualityScores` field declaration order: `acceptanceCriteria, userValue, definitionClarity, dependencies, effortEstimation, overall`. Exactly six values required; wrong count is a distinct error.

2. **Parity coverage now covers all eight pairs (carried, now confirmed).** `TestExtractAgentSharedBlocks_ScaffoldMirrorParity` in `tests/acceptance/extract_agent_shared_blocks_acceptance_test.go` already lists all eight prompts byte-for-byte. No extension needed.

3. **Slug validation duplicated in `internal/roadmap` (carried from round 1).** `defer_validate.go` duplicates the kebab-case rule with `// mirrors worktree.ValidateFeatureSlug`; no import edge from roadmap → worktree.

4. **`--source` auto-resolution via `worktree.DetectFeatureFromCwd` (carried from round 1).** `DetectFeatureFromCwd(os.Getwd())` walks parents for `.worktrees/<feature>` segment; resolves symlinks. Outside a worktree, `source` is omitted entirely (omitempty). Inside a worktree, `source.feature` is set from the slug; `source.role` remains blank.

5. **Exact evaluator-context print format (NEW — resolved by code inspection).** Rendered via a `ui` helper using `renderSystemPanel("ROADMAP", "QUALITY EVALUATOR CONTEXT", toneInfo, body)` (matching `render_roadmap_checkpoint.go` pattern), routed through an i18n key. Body contains: (a) finding `name`; (b) `summary`; (c) `source` (if present); (d) target `--phase`; (e) threshold `9`; (f) six-dimension schema with field names and range 1-10; (g) literal re-invocation line: `centinela roadmap promote <slug> --phase "<phase>" --scores ac,uv,dc,dep,ee,overall`; (h) one-line instruction to run a quality-evaluator pass then re-invoke with `--scores`. Exits 0, writes nothing.

6. **Backlog rendering shape (NEW — resolved by code inspection).** Rendered inside the same `renderSystemPanel` call as the phase overview in `RenderRoadmap`. After the last normal-phase section, a "Backlog" bold header is appended, then each entry as `  ○ <slug>  <summary>` (using `IconPending` and `StyleMuted` for the summary — no readiness state applies). Section entirely absent when Backlog phase is missing or empty.

7. **Promote metadata-stripping confirmation (NEW — resolved).** `summary` moves to the quality-entry's `summary` field. `source` and `deferredAt` are dropped from the roadmap `Feature` entry but are recorded verbatim in provenance bullets appended to `roadmap-analysis.md` and `roadmap-quality.md` — e.g. "Promoted from Backlog: source=deferred-findings-roadmap-capture/senior-engineer, deferredAt=2026-06-12T09:00:00Z". Analysis/quality `.md` are the correct provenance home.

8. **One-entry-per-line array formatting (NEW — resolved).** `internal/roadmap/rawio.go` emits the `Backlog.features` array with one JSON object per line; other phases' formatting is untouched. A golden-file round-trip test asserts byte-stability of non-Backlog entries and that the Backlog array is one-entry-per-line.

9. **Backlog phase placement does not perturb bootstrap (NEW — resolved).** `HasBootstrapPhase`/`BootstrapFeatures` use `isBootstrapPhaseName` (prefix "phase 0" + "bootstrap") which is disjoint from `isBacklogPhaseName`. Appending Backlog last does not shift Phase 0. `DeriveReadiness` skips Backlog entirely once the predicate is applied.

#### Handoff → senior-engineer

Implementation order follows the four slices in docs/plans/deferred-findings-roadmap-capture.md §8. Cobra wiring note: `roadmapCmd` currently uses `RunE` as a leaf command — senior-engineer should verify whether adding `defer` and `promote` as subcommands of `roadmapCmd` requires converting it to a group command (removing `RunE` and adding `roadmap` as a subcommand alias), or whether cobra supports both `RunE` and child subcommands on the same command.
