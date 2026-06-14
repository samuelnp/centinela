# Edge Cases: roadmap-doc-sync

**Date:** 2026-06-14

## Covered

Every case below has at least one executable assertion (colocated unit,
acceptance, or integration) and PASSES on the committed tree.

- **Backlog deferred-finding rendering (full provenance).** Backlog phase renders
  `- **name** — summary *(deferred <at> · <feature>/<role>)*`, not the
  description/fixes bullet shape. — `mdgen_feature_test.go::TestRenderBacklogFeatureFull`,
  acceptance `TestRds_BacklogDeferredFinding`. **Pass.**
- **Backlog with empty source/deferredAt.** Parenthetical omitted entirely; never
  emits `()`, `*(`, or a bare `· /`. — `TestRenderBacklogFeatureNoSource`,
  acceptance `TestRds_BacklogEmptySource`. **Pass.**
- **Backlog provenance halves.** feature-only, role-only, and nil Source each
  render without a dangling `/`. — `TestBacklogProvenanceHalves`. **Pass.**
- **Feature with description AND fixes.** Em-dash clause plus an indented
  `*Fixes: …*` line. — `TestRenderFeatureBothFields`,
  `TestRds_FeatureDescriptionAndFixes`. **Pass.**
- **Description only (no fixes).** Em-dash clause, no `*Fixes:*` line. —
  `TestRenderFeatureDescriptionOnly`, `TestRds_FeatureDescriptionOnly`. **Pass.**
- **Fixes only (no description).** Bare bullet (no dangling ` — `) followed by the
  `*Fixes:*` line. — `TestRenderFeatureFixesOnly`, `TestRds_FeatureFixesOnly`.
  **Pass.**
- **No description and no fixes.** Renders exactly `- **name**`; no em-dash, no
  Fixes line. — `TestRenderFeatureBareBullet`, `TestRds_FeatureBareBullet`. **Pass.**
- **dependsOn in declared slice order with description.** ` (depends on a, b)` in
  order, after the description. — `TestRenderFeatureDependsOnWithDescription`,
  `TestRds_FeatureDependsOnOrder`. **Pass.**
- **dependsOn with no description.** Clause attaches directly to the bullet, no
  em-dash. — `TestRenderFeatureDependsOnNoDescription`,
  `TestRds_FeatureDependsOnNoDescription`. **Pass.**
- **Empty dependsOn (`[]`).** No annotation emitted. — `TestRenderFeatureEmptyDependsOn`,
  `TestRds_FeatureEmptyDependsOn`. **Pass.**
- **Intro blockquote with a blank inner line.** Each line `> `, blank inner line
  → bare `>`. — `TestRenderMarkdownIntroBlockquote`, `TestRds_IntroBlockquote`. **Pass.**
- **Multi-paragraph phase note.** `\n\n` separator renders as a bare `>`, keeping
  the blockquote unbroken. — `TestRenderPhaseMultiParagraphNote`,
  `TestRds_PhaseNoteBlockquote`. **Pass.**
- **Phase with no note.** Heading immediately followed by features, no blockquote.
  — `TestRenderPhaseNoNote`, `TestRds_PhaseNoNote`. **Pass.**
- **Phase with zero features.** Renders heading (and optional note) only, no stray
  blank line, file still ends with one newline. — `TestRenderPhaseZeroFeatures`,
  `TestRenderPhaseEmptyWithNote`, `TestRds_PhaseZeroFeatures`. **Pass.**
- **Authored status glyph in phase name.** `✅ Phase 0: Bootstrap` preserved
  verbatim in `## …`. — `TestRenderPhaseHeadingGlyph`,
  `TestRds_PhaseHeadingGlyphPreserved`. **Pass.**
- **No live-status glyph on feature bullets.** No `✓`/`✅` on `- ` lines. —
  `TestRenderNoFeatureStatusGlyph`, `TestRds_NoFeatureStatusGlyph`. **Pass.**
- **Non-ASCII passthrough.** Em-dashes, curly quotes, accents emitted
  byte-for-byte. — `TestRenderMarkdownNonASCII`, `TestRds_NonASCIIPassthrough`. **Pass.**
- **Exactly one trailing newline, no trailing whitespace, LF only.** —
  `TestRenderMarkdownTrailingNewline`, `TestRds_TrailingNewlineAndNoTrailingWS`,
  `TestRds_LFLineEndings`. **Pass.**
- **Determinism (render/generate twice → byte-identical).** No Go map iterated. —
  `TestRenderMarkdownDeterministic`, `TestRds_GenerateDeterministic`. **Pass.**
- **Golden full roadmap.** intro + glyph phase + note + both-field feature +
  Backlog → exact canonical bytes. — `TestRenderMarkdownGolden`. **Pass.**
- **generate creates ROADMAP.md from scratch when absent.** —
  `TestRunRoadmapGenerateCreatesFile`, `TestRds_GenerateCreatesFromScratch`. **Pass.**
- **generate write failure (ROADMAP.md is a directory).** Surfaced as an error,
  not a panic. — `TestRunRoadmapGenerateWriteError`. **Pass.**
- **generate load failure (roadmap.json absent).** Surfaced as a command error. —
  `TestRunRoadmapGenerateLoadError`, `TestRds_GenerateLoadError` via gate. **Pass.**
- **Drift gate in sync → Pass** ("ROADMAP.md is in sync."). —
  `TestCheckRoadmapDriftInSync`, `TestRds_DriftPassInSync`, integration round-trip.
  **Pass.**
- **Drift under severity=fail → Fail** with first differing line number +
  `centinela roadmap generate`. — `TestCheckRoadmapDriftFail`,
  `TestRds_DriftFailUnderFail`. **Pass.**
- **Drift under severity=warn → Warn** (non-blocking, exit 0). —
  `TestCheckRoadmapDriftWarn`, `TestRds_DriftWarnUnderWarn`. **Pass.**
- **Regenerate after drift → Pass.** — `TestRds_RegenerateThenPasses`,
  integration round-trip. **Pass.**
- **Missing ROADMAP.md → Fail/Warn** per severity with a clear "missing" message,
  never a panic or raw I/O error. — `TestCheckRoadmapDriftMissingFile`,
  `TestRds_MissingUnderFail`, `TestRds_MissingUnderWarn`. **Pass.**
- **Non-missing read error (ROADMAP.md is a directory).** Fails with "cannot read",
  not a panic. — `TestCheckRoadmapDriftReadError`. **Pass.**
- **roadmap.json load error in the gate.** Fails naming the file. —
  `TestCheckRoadmapDriftLoadError`. **Pass.**
- **firstDifferingLine helper.** identical (0), first/middle line, missing trailing
  line, extra trailing line, both-empty. — `TestFirstDifferingLine`. **Pass.**
- **Config: default severity = warn.** Unset severity normalizes to warn;
  whitespace trimmed. — `TestNormalizeRoadmapDriftDefaultsToWarn`,
  `TestNormalizeRoadmapDriftTrims`. **Pass.**
- **Config: unknown severity rejected at validate.** Error names the field and the
  valid values `fail`/`warn`. — `TestValidateRoadmapDriftRejectsUnknown`,
  `TestRds_UnknownSeverityRejected`. **Pass.**
- **Config: bad severity is a no-op when disabled.** config.Load succeeds. —
  `TestValidateRoadmapDriftNoopWhenDisabled`, `TestRds_UnknownSeverityNoopWhenDisabled`.
  **Pass.**
- **Gate disabled → absent from results / Skip.** No `roadmap_drift` line, exit 0
  even when ROADMAP.md is absent. — `TestRds_GateDisabledSkips`, integration
  disabled-gate assertion. **Pass.**
- **Ships enabled=true, severity=warn** in Centinela's own centinela.toml. —
  `TestRds_ShipsEnabledWarn`. **Pass.**

## Residual Risks

- **`RenderMarkdown` error branch is unreachable.** The function signature returns
  `([]byte, error)` for forward-compatibility but never returns a non-nil error
  today, so the `render failed` branches in `checkRoadmapDrift` and
  `runRoadmapGenerate` are not coverable without a source change. Left as
  defensive code; flagged here rather than gamed with an artificial test.
- **CRLF-on-disk vs LF generator output.** The generator only ever emits LF, and
  the gate uses `bytes.Equal` (no normalization), so a CRLF-terminated on-disk file
  is correctly treated as drift. We assert LF-only output directly
  (`TestRds_LFLineEndings`); we do not author a CRLF fixture because the generator
  cannot produce one — the byte-compare path is the same as any other content
  mismatch and is covered by the drift-Fail tests.
- **Prose-migration fidelity (one-time).** The accuracy of the hand-transcribed
  ROADMAP.md prose in roadmap.json is a content judgment, not a mechanical
  property; the gate only guarantees the committed file matches generator output
  (verified in-sync on this tree), not that the prose matches original intent.
