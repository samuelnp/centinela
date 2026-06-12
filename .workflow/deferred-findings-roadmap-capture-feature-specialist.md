### Feature-Specialist Report: deferred-findings-roadmap-capture
**Date:** 2026-06-12

#### Behavior Summary

`centinela roadmap defer <slug>` writes a one-file-per-finding ledger entry under `.workflow/deferred/<slug>.json` — a JSON record containing the slug, a non-empty summary, an optional source (feature + role), status `open`, and a `createdAt` timestamp. When run inside a worktree, the active feature is auto-detected from the shell CWD via `worktree.DetectFeatureFromCwd`; outside a worktree the source is optional. Defer validates the slug against the ledger (no duplicate files) and against `roadmap.json` feature names (no shadowing), then writes atomically. `centinela roadmap` renders a "Deferred findings" section — count plus open slug list — whenever the ledger has open entries; the section is hidden when the count is zero. `centinela roadmap defer --list` prints a machine-readable listing of all open entries. `centinela roadmap promote <slug>` performs a four-stage triage operation at the root checkout: score validation (each dimension 1–10, overall ≥ 9) before any disk write, re-check that the slug is not already in `roadmap.json`, raw-JSON-preserving append to all three roadmap artifacts (roadmap.json + roadmap-analysis.json + roadmap-quality.json) via temp-file + rename, ledger status update to `promoted` with a `promotedAt` field, and a post-write `roadmap validate` run that surfaces any artifact inconsistency before the next `centinela start`. Finally, the four role prompts (big-thinker, feature-specialist, senior-engineer, qa-senior) and their byte-identical scaffold mirrors each gain a mandatory "Deferred Findings" section that requires agents to run `centinela roadmap defer` for every out-of-scope or not-fixed-now finding and to list the resulting slugs (or "none") in the report.

#### Gherkin Scenarios

All scenarios are in `specs/deferred-findings-roadmap-capture.feature`.

- **Happy-path defer creates a ledger entry with required fields** — Given active workflow and clean ledger/roadmap; When defer runs with slug + summary + source; Then `.workflow/deferred/<slug>.json` is created with all five required fields (slug, summary, source, status `open`, createdAt) and exit 0.
- **Defer rejects a slug that already exists in the ledger** — Given an existing open ledger entry; When defer is re-run for the same slug; Then non-zero exit, collision message, file unchanged.
- **Defer rejects a slug that matches an existing roadmap feature name** — Given roadmap.json has the named feature; When defer uses the same slug; Then non-zero exit, "already a roadmap feature" message, no file created.
- **Defer with an empty summary is rejected before any file is written** — Given clean ledger; When `--summary ""` is passed; Then non-zero exit, no file created.
- **Defer with an invalid slug is rejected** — Given a slug containing spaces or special chars; When defer runs; Then non-zero exit with slug format error.
- **Source flag is optional; when inside a worktree the feature is auto-detected** — Given CWD inside `.worktrees/auto-source-feat`; When defer runs without `--source`; Then exit 0 and `source.feature` populated from CWD detection.
- **Source flag is required when run outside any worktree and CWD detection yields nothing** — Given CWD is repo root; When defer runs without `--source`; Then exit 0 with source omitted.
- **Deferred findings count and list are shown in centinela roadmap output** — Given two open ledger entries; When `centinela roadmap` runs; Then both slugs and count are present in output.
- **Deferred findings section is hidden when there are no open entries** — Given empty ledger; When `centinela roadmap` runs; Then no deferred section in output.
- **defer --list prints all open ledger entries as machine-readable JSON** — Given two open and one promoted entry; When `--list` runs; Then open-only entries appear, promoted entry absent, exit 0.
- **Happy-path promote appends to all three roadmap artifacts** — Given open ledger entry and valid scores ≥ 9; When promote runs; Then all three artifacts updated, ROADMAP.md reminder printed, `roadmap validate` passes.
- **Promote preserves unknown JSON fields in analysis and quality artifacts** — Given analysis has legacy `dependsOn` field; When promote appends a new entry; Then existing `dependsOn` fields are byte-stable.
- **Promote marks the ledger entry status as promoted** — Given successful promote; Then ledger file has `status: promoted` and non-empty `promotedAt`.
- **Promote with overall score below 9 is rejected before any write** — Given scores summing to overall 7; When promote runs; Then non-zero exit, score error, all three artifacts unchanged.
- **Promote into a non-existent phase is rejected with known phases listed** — Given unknown `--phase`; When promote runs; Then non-zero exit, known-phases enumerated in output, roadmap.json unchanged.
- **Promote a slug already present as a roadmap feature is rejected cleanly** — Given stale open ledger entry whose slug root roadmap already carries; When promote runs; Then non-zero exit, artifacts unchanged.
- **centinela roadmap validate passes after a successful promotion** — Given successful promotion; When `roadmap validate` runs; Then exit 0.
- **Four role prompts and their scaffold mirrors contain the Deferred Findings obligation byte-identically** — Given the four source prompt files and four scaffold mirrors; When each pair is compared; Then all pairs are byte-identical and every file contains `centinela roadmap defer` and a "Deferred Findings" heading.

#### UX States

| State | Trigger | Surface |
|-------|---------|---------|
| Ledger entry created | `centinela roadmap defer <slug> --summary <text>` succeeds | stdout confirmation line + exit 0; `.workflow/deferred/<slug>.json` written |
| Slug collision (ledger) | Slug already exists in `.workflow/deferred/` | stderr error "already exists / collision"; exit non-zero; no file change |
| Slug collision (roadmap) | Slug matches a `roadmap.json` feature name | stderr error "already a roadmap feature"; exit non-zero; no file created |
| Empty summary rejected | `--summary ""` passed | stderr "summary required / empty"; exit non-zero; no file written |
| Invalid slug rejected | Slug fails kebab-case validation | stderr names bad slug and required format; exit non-zero |
| Source auto-detected | CWD inside `.worktrees/<feature>`; `--source` omitted | `source.feature` field populated silently; exit 0 |
| Source omitted at root | CWD not inside any worktree; `--source` omitted | `source` field absent/null in ledger file; exit 0 |
| Deferred section visible | `centinela roadmap` run with >= 1 open ledger entry | Deferred findings panel in stdout showing count + slug list |
| Deferred section hidden | `centinela roadmap` run with 0 open entries | No deferred panel in stdout |
| Listing open findings | `centinela roadmap defer --list` | Machine-readable JSON/structured listing of open entries; exit 0 |
| Promote success | All scores valid, phase exists, slug not in roadmap | Three artifacts updated; ledger status -> promoted; ROADMAP.md reminder; `validate` passes; exit 0 |
| Score below threshold | promote `--scores` overall < 9 | stderr score error; zero writes; exit non-zero |
| Unknown phase | promote `--phase` names a phase not in roadmap.json | stderr lists known phases; zero writes; exit non-zero |
| Slug already promoted | promote targets a slug already in roadmap.json | stderr "already a roadmap feature"; zero writes; exit non-zero |
| Prompt contract satisfied | Agent report lists deferred slugs or "none" after running defer | Human-readable evidence in `.workflow/<feature>-<role>.md`; machine-verifiable via prompt text |

#### Out-of-Scope

- No auto-prioritization or auto-scheduling — phase, scores, and summary must be supplied explicitly at promote time.
- No validator hard-gate on "did the agent defer everything it should have" — the contract is prompt-level and unverifiable by the gate.
- No change to gates or claim-verification; no evidence-schema changes (slugs live in report prose, not evidence JSON).
- No retroactive backfill of the 397-entry legacy memory corpus or old Residual Risks sections.
- No ROADMAP.md (human file) auto-sync — that is `roadmap-doc-sync`'s responsibility; v1 prints a reminder.
- No dedupe/similarity detection between findings — exact slug collision only.
- No `defer` obligation wired into ux-ui-specialist, validation-specialist, or gatekeeper prompts — fast-follow.
- No `defer dismiss` command — reserved for v1+; `dismissed` status is defined in the data shape but not written by any v1 command.
- No evidence-schema field for deferred slugs (slugs referenced in report prose only).
- No find-or-create-phase behavior on promote — v1 requires an existing phase name.

#### Handoff

- **Next role:** senior-engineer

- **Open clarifications (resolved):**

  1. **`--source` default inside a worktree:** Reuse existing mechanism. `worktree.DetectFeatureFromCwd(os.Getwd())` already returns the feature slug when the binary is invoked from inside `.worktrees/<feature>/`. `loadActiveWorkflows()` in `cmd/centinela/hook_workflows.go` already calls this to scope the active workflow. The `roadmap defer` command should call `DetectFeatureFromCwd` the same way: if a feature slug is detected, populate `source.feature` from it and derive `source.role` from `--source` if provided or leave the role blank; if no worktree is detected, omit the source field entirely. The `--source` flag remains optional in all cases — requiring it outside a worktree would block agents running at the repo root. Rationale: reuse avoids duplicating CWD-resolution logic; `--source` as an explicit override still works in both contexts.

  2. **`--scores` flag shape on promote:** Use a **single CSV flag** `--scores ac,uv,dc,dep,ee,overall` (six integers, 1–10 each). Rationale: one flag keeps the command concise for agent use (less surface area to mistype); the dimension names are positional and already documented in the quality schema (`accessibility_comprehension`, `uniqueness_value`, `delivery_complexity`, `dependency_risk`, `engineering_excellence`, `overall`). Six separate flags (`--ac`, `--uv`, etc.) would require agents to know six distinct flag names and produce a much longer invocation. The CSV form matches how the quality spec already expresses scores as a vector. A prompt-driven quality-evaluator subagent invocation is out of scope (plan §3 Out: no auto-prioritization).

  3. **Scaffold parity tests for the four prompt files:** `TestExtractAgentSharedBlocks_ScaffoldMirrorParity` in `tests/acceptance/extract_agent_shared_blocks_acceptance_test.go` already covers all four files: `big-thinker-prompt.md`, `feature-specialist-prompt.md`, `senior-engineer-prompt.md`, and `qa-senior-prompt.md` are in the `promptsReferencingInvocation` slice (lines 20-27 of that file), and all are asserted byte-identical against their mirrors in `internal/scaffold/assets/docs/architecture/`. The test does NOT need extending for these four files. The senior-engineer only needs to update both source and mirror in the same commit.

  4. **Where slug validation should live (G2 import-graph gate):** Duplicate the ~10-line check in `internal/roadmap/deferred_validate.go` with a comment pointing to `worktree.ValidateFeatureSlug`. Rationale: the G2 gate defines `internal/roadmap` as an unmapped package (neither `leaf`, `domain`, nor `cmd` in `centinela.toml [gates.import_graph]`). `internal/worktree` is also unmapped. Importing `worktree` from `roadmap` is not formally prohibited today, but it introduces a cross-package dependency for a trivial regexp check — the slug pattern (`^[a-z0-9]+(-[a-z0-9]+)*$`) is self-contained and carries no behavioral logic. Duplicating with a `// mirrors worktree.ValidateFeatureSlug` comment is the lowest-risk option: zero new import edge, G2 stays at Warn, and the comment documents the canonical source so a future "extract to shared/slug" fast-follow is obvious.
