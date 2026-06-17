### Edge-Case Report: deferred-findings-roadmap-capture
**Date:** 2026-06-12

#### Risk Matrix

- **Case:** Empty summary (defer) / **Impact:** Low / **Likelihood:** Medium / **Why:** Agents might pass empty flag; validated before any write, correct exit 1.
- **Case:** Whitespace-only summary (defer) / **Impact:** Low / **Likelihood:** Medium / **Why:** `strings.TrimSpace` check fires; no write occurs.
- **Case:** Invalid slug — uppercase, path-traversal, unicode, spaces (defer) / **Impact:** Low / **Likelihood:** Medium / **Why:** Regex gate blocks all variants before read; all produce exit 1, no write.
- **Case:** Missing roadmap.json (defer or promote) / **Impact:** Low / **Likelihood:** Low / **Why:** Clean error; no write occurs.
- **Case:** Corrupt/truncated roadmap.json (defer or promote) / **Impact:** Low / **Likelihood:** Low / **Why:** `json.Unmarshal` fails with a clear error; no write occurs.
- **Case:** Empty phases array in roadmap.json (defer) / **Impact:** Low / **Likelihood:** Low / **Why:** Creates Backlog phase as only phase; valid output.
- **Case:** roadmap.json with no "phases" key at all (defer) / **Impact:** Low / **Likelihood:** Low / **Why:** Creates "phases" key with Backlog; preserves other top-level keys.
- **Case:** Backlog as only phase (defer appends to it) / **Impact:** Low / **Likelihood:** Low / **Why:** Appends correctly; validate exempt.
- **Case:** Slug collision — same slug deferred twice / **Impact:** Low / **Likelihood:** Medium / **Why:** `phaseFeatureNames` detects it; clear error, no write.
- **Case:** Slug collision — slug exists in non-Backlog phase (defer) / **Impact:** Low / **Likelihood:** Medium / **Why:** Same detection; reports the phase that holds the slug.
- **Case:** Duplicate feature names across two non-Backlog phases / **Impact:** Medium / **Likelihood:** Low / **Why:** Map is last-writer-wins for the dup; defer of dup slug still correctly blocked. No corruption.
- **Case:** Promote before any defer / **Impact:** Low / **Likelihood:** Low / **Why:** "not a Backlog finding" error, exit 1, no write.
- **Case:** Double-promote same slug / **Impact:** Low / **Likelihood:** Low / **Why:** First promote removes slug from Backlog; second attempt correctly errors.
- **Case:** Defer after promote (slug now in real phase) / **Impact:** Low / **Likelihood:** Low / **Why:** Collision check spans all phases; correctly blocked.
- **Case:** Promote with missing analysis.json / **Impact:** High / **Likelihood:** Medium / **Why:** roadmap.json IS written before `appendFeatureEntry` tries to open analysis.json; missing file leaves roadmap half-promoted; every subsequent `centinela start` blocked. **Confirmed broken.**
- **Case:** Promote with missing quality.json (analysis.json present) / **Impact:** High / **Likelihood:** Medium / **Why:** roadmap.json AND analysis.json written before quality.json attempted; same blocked state. **Confirmed broken.**
- **Case:** Validate-after-write fails (wrong role in analysis.json) / **Impact:** Medium / **Likelihood:** Low / **Why:** All files written; "promote wrote files but validate failed" reported at exit 1. By design per plan §4.3.
- **Case:** promote-to-Backlog as --phase / **Impact:** Low / **Likelihood:** Low / **Why:** `appendToPhase` skips Backlog; "unknown phase" error lists only non-Backlog phases. Correct.
- **Case:** case-variant --phase "backlog" / " BACKLOG " / **Impact:** Low / **Likelihood:** Low / **Why:** appendToPhase does exact-string match on target; both produce "unknown phase" error. Correct.
- **Case:** Phase named "backlog" (lowercase) in roadmap.json / **Impact:** Low / **Likelihood:** Low / **Why:** `isBacklogPhaseName` uses EqualFold; lowercase "backlog" correctly treated as Backlog (validate-exempt, defer appends to it, ready excludes it).
- **Case:** Phase named "Backlog Phase Items" (partial-match) / **Impact:** Medium / **Likelihood:** Low / **Why:** EqualFold exact-match only; this phase is NOT exempt. validate correctly demands coverage. Correct.
- **Case:** --scores "" (empty string) / **Impact:** Low / **Likelihood:** Medium / **Why:** cobra treats "" as unset; falls through to evaluator path (prints context, exit 0). Surprising but no data corruption.
- **Case:** --scores with 5 / 7 values / non-numeric / negative / 0 / 11 / **Impact:** Low / **Likelihood:** Low / **Why:** All rejected before any write with clear error messages.
- **Case:** Score boundary values 1 and 10 / **Impact:** Low / **Likelihood:** Low / **Why:** Both accepted as valid; promote succeeds.
- **Case:** Overall score exactly 9 (minimum) / **Impact:** Low / **Likelihood:** Low / **Why:** Passes correctly (>= 9 threshold).
- **Case:** JSON special chars / newlines / ANSI escapes in summary / **Impact:** Low / **Likelihood:** Medium / **Why:** Proper `json.Encoder` usage; quotes/braces escaped; newlines become `\n`; ANSI codes stored verbatim. No injection.
- **Case:** Very long summary (200+ chars) in roadmap render / **Impact:** Low / **Likelihood:** Low / **Why:** Panel expands horizontally without truncation; ugly but not a crash. UX gap only.
- **Case:** Slug exists in Backlog AND target phase (hand-edited roadmap) / **Impact:** Medium / **Likelihood:** Low / **Why:** No pre-flight check in `appendToPhase` for existing slug in target; promote creates a second duplicate entry. validate passes (set deduplication). **Confirmed broken.**
- **Case:** Source auto-resolution from worktree CWD / **Impact:** Low / **Likelihood:** Low / **Why:** `DetectFeatureFromCwd` correctly extracts feature name from `.worktrees/<name>/` path.
- **Case:** Defer from worktree writes worktree's roadmap.json, not root / **Impact:** Medium / **Likelihood:** Medium / **Why:** By design. Risk: stale worktree copy causes avoidable merge conflicts. Accepted by operator.
- **Case:** Concurrent defer from two worktrees → git merge conflict / **Impact:** Low / **Likelihood:** Medium / **Why:** Merge produces conflict markers. Trivial union resolution. Accepted by operator.
- **Case:** Untouched phases reformatted on first write / **Impact:** Low / **Likelihood:** High / **Why:** `indentValue` normalizes all untouched phases; large spurious diff on first operation. Known trade-off per senior-engineer report.
- **Case:** Non-deterministic key order in artifact JSON writes / **Impact:** Low / **Likelihood:** Medium / **Why:** `writeArtifact` iterates a Go map; key order is randomized between runs. Custom fields preserved but order varies; causes spurious diffs.

#### Missing or Weak Scenarios

1. **Pre-flight artifact existence check for promote** — no scenario covers "promote when analysis.json or quality.json does not exist." A test should assert roadmap.json UNCHANGED when artifact files are absent (currently the command breaks this invariant).
2. **Slug-already-in-target-phase check for promote** — no scenario covers "promote when slug already exists as a non-Backlog feature." A test should assert exit 1 before any write.
3. **`--scores ""` (empty string) disambiguation** — no scenario verifies this falls through to the evaluator path rather than producing a validation error.
4. **Non-deterministic key order in artifact JSON** — no golden-file test locks down key order in `writeArtifact`; a refactoring that changes iteration order would silently generate spurious diffs.
5. **Very long summary truncation / wrap in render** — no scenario verifies the render panel handles summaries exceeding terminal width gracefully.
6. **Retry promote after partial-write failure** — no test covers the "fix the broken artifact, re-run promote" recovery path; the provenance bullet would be re-appended on the retry.
7. **Backlog-only roadmap + promote** — promote on a Backlog-only roadmap should produce "unknown phase" (no non-Backlog phases to list); untested.
8. **`deferredAt` RFC3339 round-trip** — no test asserts the stored `deferredAt` value survives promote and appears correctly in the provenance bullet.
9. **Provenance bullet format for source-less finding** — `provenanceBullet` uses `"unknown"` when `f.Source` is nil; no test covers this exact output.

#### Proposed/Added Tests

- **Unit:**
  - `TestDefer_MissingSummary` and `TestDefer_WhitespaceSummary` — validateSummary edge cases.
  - `TestDefer_InvalidSlugs` — table: uppercase, unicode, path-traversal, spaces, empty.
  - `TestDefer_SlugCollisionInBacklog` and `TestDefer_SlugCollisionInNonBacklog` — validateNoCollision.
  - `TestParseScores_Boundaries` — table: 0, 1, 9, 10, 11, -1; count 5, 6, 7; non-numeric; empty string.
  - `TestParseScores_OverallThreshold` — overall=8 rejected, overall=9 accepted, overall=10 accepted.
  - `TestIsBacklogPhaseName` — exact match, case variants, leading/trailing spaces, partial matches like "Backlog Phase".
  - `TestNonBacklogFeatureSet_ExcludesBacklog` — assert Backlog features absent from set.
  - `TestProvenanceBullet_NoSource` — nil Source produces "unknown" in bullet.

- **Integration:**
  - `TestDefer_RoundTrip_CustomFieldsPreserved` — roadmap.json with `customField` on existing entries; defer; assert field survives.
  - `TestDefer_NoPhasesKey` — roadmap.json with no "phases" key; defer creates the key; other top-level fields preserved.
  - `TestPromote_PreflightMissingAnalysisJSON` — promote when analysis.json absent; assert roadmap.json UNCHANGED (currently fails — exposes partial-write bug).
  - `TestPromote_PreflightMissingQualityJSON` — same for quality.json.
  - `TestPromote_SlugAlreadyInTargetPhase` — promote when slug exists in both Backlog and target; assert exit 1 + no duplicate entry.
  - `TestPromote_WritesNothingOnNoScores` — evaluator path; assert no file modified.
  - `TestPromote_EmptiedBacklogPhaseKept` — after promote, Backlog phase remains with empty features array.
  - `TestDefer_ConcurrentAppend_OneEntryPerLine` — two defers from same base; assert Backlog features are one compact object per line.

- **Acceptance:**
  - Scenario: `promote-partial-write-when-analysis-missing` — assert exit 1 AND roadmap.json byte-identical to before.
  - Scenario: `promote-duplicate-slug-in-target-rejected` — manually inject dup; assert exit 1 + no second entry.
  - Scenario: `defer-then-roadmap-render-long-summary` — 200-char summary renders without panic.

#### Residual Risks

1. **Promote partial-write on missing artifact files (HIGH).** `promote` writes `roadmap.json` first, then reads `analysis.json` and `quality.json`. If either file is absent, `roadmap.json` is mutated and the slug is moved out of Backlog, but analysis/quality are not updated. `centinela validate` then fails, blocking all `centinela start` commands. Mitigation: add a pre-flight `os.Stat` check for both artifact JSON files before writing `roadmap.json`.

2. **Duplicate feature entry when slug exists in target phase (MEDIUM).** `appendToPhase` does not check whether `slug` already names a feature in the target phase. A hand-edited roadmap with the same slug in both Backlog and a real phase causes promote to silently create a duplicate entry. `validate` does not detect this (Go map deduplication). Mitigation: scan the target phase's features in `appendToPhase` and return a collision error if found.

3. **Non-deterministic key order in `writeArtifact` (LOW).** `writeArtifact` iterates a `map[string]json.RawMessage`; Go map iteration is randomized. Custom top-level fields are preserved but may appear in different order after each write, causing spurious git diffs and fragile golden-file tests. Mitigation: sort non-"features" keys before emitting.

4. **Untouched phases reformatted on first write (LOW).** Any defer or promote normalizes ALL phases via `indentValue`, causing a large spurious diff on first operation against a hand-authored roadmap.json. Known trade-off (senior-engineer report). Mitigation: document in Merge Steward notes so operators are not surprised.

5. **Concurrent defer merge conflict is unavoidable (LOW, ACCEPTED).** Two worktrees appending to the same Backlog.features array always produce a `git merge-file` conflict. Conflict is a trivial union. Accepted by operator (plan §2.1). No further mitigation needed beyond Merge Steward documentation.

#### Deferred Findings

Dogfooded via `/tmp/centinela-dfrc roadmap defer` run from the worktree root, writing to `.worktrees/deferred-findings-roadmap-capture/.workflow/roadmap.json`:

- `promote-partial-write-on-missing-artifacts` (source: deferred-findings-roadmap-capture/edge-case-tester)
- `promote-duplicate-entry-on-slug-collision` (source: deferred-findings-roadmap-capture/edge-case-tester)
