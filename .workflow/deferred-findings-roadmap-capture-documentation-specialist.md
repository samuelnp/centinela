# deferred-findings-roadmap-capture — documentation-specialist

## Summary

This step produced the KB entry for the `deferred-findings-roadmap-capture` feature and regenerated the project documentation HTML. The KB entry is a plain-language reference covering all user-visible behavior introduced by this feature: the `centinela roadmap defer` and `centinela roadmap promote` commands, Backlog rendering, validate exemption, start-guard refusal, score validation, and the prompt contract across all eight roles.

## KB Entry Written

**Path:** `docs/project-docs/kb/deferred-findings-roadmap-capture.md`

**Sections:**
- Frontmatter: `feature`, `summary`, `audience: end-user`, `status: done`
- `## What it does` — four-sentence plain-language description of the end-to-end flow (capture → Backlog → triage → promote)
- `## When you'd use it` — operator-facing context: when to run `defer`, what the mandatory prompt contract means, when to run `promote`
- `## How it behaves` — 13 user-meaningful bullets derived from the 25 spec scenarios, grouped by concern: capture mechanics, validation rejections, Backlog visibility, readiness and start-guard exclusions, validate exemption boundary, promote evaluator path, promote scored path, and prompt contract parity
- `## Examples` — concrete walkthrough: defer with `--source`, roadmap overview, promote without scores (evaluator context), promote with scores (scored path)

## Scenario Grouping (25 scenarios → 13 behavior bullets)

| Spec scenarios | Behavior bullet |
|---|---|
| Happy-path defer; append to existing Backlog | Capturing adds entry; existing entries byte-identical |
| Source auto-detect from worktree CWD | Source auto-detection from worktree |
| Defer outside worktree with no --source | Outside worktree: no source field, still valid |
| Empty summary, duplicate slug (Backlog), duplicate slug (non-Backlog), invalid slug | Validation rejections before any write |
| Backlog shown in roadmap output; absent when Backlog missing; absent when Backlog empty | Backlog section visible when non-empty, absent otherwise |
| Backlog features absent from roadmap ready | Backlog never appears as ready |
| Start refuses a Backlog feature | Start guard: promote-first error |
| Validate passes when Backlog has no analysis/quality; validate still fails for real missing features; similarly-named phase not exempt | Validate exemption: Backlog exempt; non-Backlog not |
| Promote without --scores: prints context, writes nothing | Evaluator path: prints context only |
| Promote with valid scores: moves entry, appends artifacts, validate passes | Scored path: full atomic move |
| Preserve unknown fields (raw-preserving I/O) | Raw-preserving: no fields dropped |
| Score rejections: overall < 9, dimension out of range, malformed CSV, unknown phase, slug not in Backlog | Score validation before any write |
| Prompt contract parity: all 8 pairs byte-identical, Deferred Findings section present | All eight role prompts carry mandatory Deferred Findings section |

## Roadmap Dependencies

None. This is a Phase 5 Operability & DX feature with no declared `dependsOn` in `roadmap.json`.

## Workflow Status Matrix

| Step | Status |
|------|--------|
| 1 — plan | done |
| 2 — code | done |
| 3 — tests | done |
| 4 — validate | done |
| 5 — docs | done |

## Documentation Generation

`centinela docs generate --out docs/project-docs/index.html` emitted:
- `docs/project-docs/kb/deferred-findings-roadmap-capture.html` (9.7 KB)
- `docs/project-docs/kb/index.html` (regenerated)
- `docs/project-docs/index.html` (19.6 KB, regenerated)

## Plain-Language Constraint Handled

The spec uses Given/When/Then DSL and internal engineering terms (`roadmapFeatureSet`, `NonBacklogFeatureSet`, `isBacklogPhaseName`, raw-preserving I/O). All of these were translated into user-visible behavior language. Scenario grouping reduced 25 atomic spec scenarios into 13 user-meaningful bullets by clustering scenarios that test the same behavior from different angles (e.g., the six promote-rejection scenarios became one bullet: "Score validation happens before any write").

#### Deferred Findings

None. No documentation gaps or out-of-scope findings were identified during this docs step that warrant a new Backlog entry.
