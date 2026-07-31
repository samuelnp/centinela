# docs-step-markdown-first — gatekeeper

### Adversarial Verifier Report: docs-step-markdown-first
**Date:** 2026-07-30
**Status:** WARNING

#### Inputs Read

- `git diff main...HEAD` (stat + per-file diffs) and working-tree status (clean before my own artifacts)
- docs/features/docs-step-markdown-first.md (contract)
- specs/docs-step-markdown-first.feature (acceptance criteria)
- docs/plans/docs-step-markdown-first.md (locked design)
- centinela.toml (validate commands, gates config)
- internal/orchestration/output_rules.go, output_helpers.go
- internal/workflow/validate.go, validate_docs.go
- internal/docsctx/context.go, render.go; cmd/centinela/docs_context.go, docs.go
- cmd/centinela/merge.go, hook_context.go, hook_statusline_rules.go (diffs)
- docs/architecture/documentation-generator-prompt.md (diff)
- tests/acceptance/docs_step_markdown_first_{evidence,pipeline}_test.go (full), scenario comments of all 4 new acceptance files
- specs/right-size-docs-step.feature, docs-knowledge-base-pages.feature, generate-html-project-docs.feature, improve-docs-llm-hybrid-ui.feature, merge-truthful-delivery.feature (grep for contradicted scenarios)
- Output of the commands I ran (below)

Delegation contamination: the orchestrator's prompt contained NO narrative summary of the implementation — only the feature slug and input-contract paths. Clean delegation.

## Analyzed Specs

- specs/docs-step-markdown-first.feature — all 10 scenarios individually attacked (see below); 10/10 have `// Scenario:` acceptance coverage (confirmed by grep and by the spec-traceability gate).
- Legacy specs listed above — checked for contradictions with shipped behavior (Finding 1).

#### Refutation Attempts

- **Claim attacked:** the real-file docs gate actually runs at the docs step for user-facing features (not just in unit tests).
  **How:** traced `ValidateArtifacts(f,"docs",cfg)` → `validateDocsOutput` + `validateOrchestration` → `orchestration.RequiredRolesForFeature("uf","docs")` includes RoleDocsSpecialist → `validateActionableOutputs` → `dispatchRoleOutputs` RoleDocsSpecialist case; the acceptance tests drive this exact entry point.
  **Result:** could not refute — the wiring is real and exercised.
- **Claim attacked:** `hasDocsOutput` can be gamed (directories, missing files, un-normalized paths).
  **How:** read existingOutputFiles/missingOutputFiles — paths are normalized (`normalizeOutputPath`) and directories rejected (`!info.IsDir()`); missing paths fail earlier with "actionable outputs must be real files". Listing `docs/plans/<f>.md` passes by design (plan decision 3, documented).
  **Result:** could not refute.
- **Claim attacked:** `docs generate`/`docs validate` might still exit 0 (cobra prints help for unknown subcommands).
  **How:** dogfooded `/tmp/centinela-verify docs generate` and `docs validate` — both exit 1 with "unknown command"; docs.go adds an explicit RunE guard; the pipeline acceptance test pins non-zero exit.
  **Result:** could not refute.
- **Claim attacked:** merge still regenerates the portal.
  **How:** diff of cmd/centinela/merge.go (seam, const, regen call, docgen import all deleted); grep for docgen/project-docs in non-test Go code — only the deliberate `hasImplementationOutput` legacy exclusion remains (kept per plan); `internal/docgen/` and `docs/project-docs/` are gone from the tree.
  **Result:** could not refute.
- **Claim attacked:** `docs context` misbehaves on missing/optional inputs.
  **How:** dogfooded happy path (this feature), missing-everything (single aggregated error naming brief+plan+spec paths, exit 1), bad slug (exit 1), and a scratch fixture with no changelog (exit 0 + "no changelog draft yet — run: centinela artifact new f changelog" hint).
  **Result:** could not refute — matches all three context scenarios verbatim.
- **Claim attacked:** scaffold mirror parity broken.
  **How:** byte-compared (cmp) all 4 changed docs/architecture files against internal/scaffold/assets/docs/architecture; confirmed scaffold CLAUDE.md and both harness golden AGENTS.md fixtures changed in lockstep.
  **Result:** could not refute — all identical.
- **Claim attacked:** spec scenarios lack acceptance coverage.
  **How:** counted 10 scenarios vs `// Scenario:` comments across the 4 new acceptance files (all 10 present); spec-traceability gate independently reports all 10 covered.
  **Result:** could not refute.
- **Claim attacked:** suite/gates fail on the real tree.
  **How:** ran `centinela validate` (scratch binary built from this worktree) — full `go test ./... -coverprofile` + coverage gate + fmt gate + G1 + cross-compile + import-graph + traceability + roadmap-drift.
  **Result:** could not refute — exit 0, all gates pass (two pre-configured warn-mode warnings, Findings 3).
- **Claim attacked:** plan commitments fully delivered.
  **How:** checked the plan's test-strategy commitment to rewrite contradicted legacy spec scenarios — `git diff main...HEAD -- specs/` touches ONLY the new spec.
  **Result:** REFUTED — see Finding 1.

#### Commands Run

- `go build -o /tmp/centinela-verify ./cmd/centinela` — exit 0, ~5s
- `/tmp/centinela-verify artifact new docs-step-markdown-first gatekeeper` — exit 0, <1s
- `/tmp/centinela-verify evidence init docs-step-markdown-first gatekeeper` — exit 0, <1s
- `/tmp/centinela-verify validate` — **exit 0, 378s** (foreground run; an earlier background invocation of the same argv was lost mid-suite when its shell died and is NOT counted as evidence). This single run executed the full `[validate] commands`: `go test ./... -coverprofile=coverage.out`, `COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh`, `./scripts/check-fmt.sh` — all ✓. Per centinela.toml the profiled `go test ./...` inside validate IS the full suite (acceptance tier included under ./...), so no separate suite run was performed, per the single-run design.
- Dogfood runs, each <1s: `docs context docs-step-markdown-first` (exit 0), `docs generate` (exit 1), `docs validate` (exit 1), `docs context no-such-feature` (exit 1, aggregated error), `docs context 'Bad Slug!'` (exit 1), scratch-fixture `docs context f` without changelog (exit 0 + hint)
- `/tmp/centinela-verify roadmap defer legacy-docs-specs-contradiction …` — exit 0, <1s
- `/tmp/centinela-verify roadmap defer changelog-template-placeholder-passes-gate …` — exit 0, <1s
- `/tmp/centinela-verify artifact stamp docs-step-markdown-first` — run last, after this report body

#### Findings

1. **Affected spec:** plan test-strategy commitment (docs/plans/docs-step-markdown-first.md, "Legacy specs whose scenarios now contradict behavior")
   **Affected scenario:** e.g. specs/merge-truthful-delivery.feature "the documentation portal is regenerated in the primary tree"; scenarios in specs/docs-knowledge-base-pages.feature, generate-html-project-docs.feature, improve-docs-llm-hybrid-ui.feature, right-size-docs-step.feature pinning `docs generate`/`docs validate`/portal
   **Risk:** the spec corpus now asserts behavior this feature deleted; the diff-aware warn-mode traceability gate does not surface it, so the contradiction persists silently and misleads future planners/verifiers. Severity: WARNING.
   **Suggestion:** rewrite/remove the contradicted scenarios (deferred: legacy-docs-specs-contradiction).
2. **Affected spec:** specs/docs-step-markdown-first.feature "User-facing docs step fails without a changelog entry" (gate strength)
   **Affected scenario:** changelog non-empty check
   **Risk:** `validateChangelog` accepts an unfilled `- <FILL: type>: <FILL: …>` template line as non-empty (pre-existing behavior carried over, not introduced here); this feature's own changelog artifact is currently still the template and would pass. Severity: minor.
   **Suggestion:** require a filled entry (deferred: changelog-template-placeholder-passes-gate); fill this feature's changelog in the docs step.
3. **Affected spec:** none (repo hygiene)
   **Affected scenario:** n/a
   **Risk:** warn-mode `roadmap_drift` reports ROADMAP.md drifted at line 259 (deferred-findings section; my own two deferrals add more drift) — non-blocking by configured severity. `import_graph` warns on unmapped packages (pre-existing, conservative matrix policy). Severity: minor.
   **Suggestion:** run `centinela roadmap generate` before delivery.

#### Deferred Findings

- `legacy-docs-specs-contradiction` — recorded via `centinela roadmap defer … --source docs-step-markdown-first/gatekeeper`
- `changelog-template-placeholder-passes-gate` — recorded via `centinela roadmap defer … --source docs-step-markdown-first/gatekeeper`

#### Recommendation

- **WARNING** — I personally attempted to break every spec scenario and the gate-swap invariants and could not: the real-file docs gate is enforced on the real code path, the HTML pipeline is verifiably gone (commands exit non-zero, no merge-time regen path, package and portal tree deleted), `docs context` matches all scenarios live, mirrors/goldens are in lockstep, and one full validate run (suite + all gates) passed at the verified revision. WARNING rather than SAFE because a locked plan commitment (rewriting the contradicted legacy spec scenarios) was not delivered, leaving five spec files asserting deleted behavior; deferred, not blocking.

```json centinela:verification
{
  "revision": "277d6cbf70b1116a4b0fa034256828bb95475ef0",
  "treeDigest": "sha256:96a296d224f285c67bee93c30f8a309157f0daa35dc5b87e410b78630a09cfc7",
  "commands": [
    {"argv": ["go", "build", "-o", "/tmp/centinela-verify", "./cmd/centinela"], "exitCode": 0, "durationMs": 5000},
    {"argv": ["/tmp/centinela-verify", "artifact", "new", "docs-step-markdown-first", "gatekeeper"], "exitCode": 0, "durationMs": 200},
    {"argv": ["/tmp/centinela-verify", "evidence", "init", "docs-step-markdown-first", "gatekeeper"], "exitCode": 0, "durationMs": 200},
    {"argv": ["/tmp/centinela-verify", "validate"], "exitCode": 0, "durationMs": 378000},
    {"argv": ["/tmp/centinela-verify", "docs", "context", "docs-step-markdown-first"], "exitCode": 0, "durationMs": 100},
    {"argv": ["/tmp/centinela-verify", "docs", "generate"], "exitCode": 1, "durationMs": 100},
    {"argv": ["/tmp/centinela-verify", "docs", "validate"], "exitCode": 1, "durationMs": 100},
    {"argv": ["/tmp/centinela-verify", "docs", "context", "no-such-feature"], "exitCode": 1, "durationMs": 100},
    {"argv": ["/tmp/centinela-verify", "docs", "context", "Bad Slug!"], "exitCode": 1, "durationMs": 100},
    {"argv": ["/tmp/centinela-verify", "roadmap", "defer", "legacy-docs-specs-contradiction", "--summary", "Rewrite legacy spec scenarios still pinning deleted docs generate/validate + portal regen (right-size-docs-step, docs-knowledge-base-pages, generate-html-project-docs, improve-docs-llm-hybrid-ui, merge-truthful-delivery)", "--source", "docs-step-markdown-first/gatekeeper"], "exitCode": 0, "durationMs": 200},
    {"argv": ["/tmp/centinela-verify", "roadmap", "defer", "changelog-template-placeholder-passes-gate", "--summary", "validateChangelog accepts an unfilled FILL template line as a non-empty changelog; require a filled entry", "--source", "docs-step-markdown-first/gatekeeper"], "exitCode": 0, "durationMs": 200}
  ]
}
```
