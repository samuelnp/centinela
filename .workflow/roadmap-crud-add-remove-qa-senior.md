# roadmap-crud-add-remove — qa-senior

Colocated coverage tests (move the per-package gate) plus the required tests/
tier trio, all ≤100 lines, no network / no installed-binary dependence. Full
suite green; repo coverage gate PASS at 97.3% (floor 95%).

## Test Inventory

### Colocated — internal/roadmap (moves coverage; run without -coverpkg)
| File | Lines | Focus |
|------|------|-------|
| crud_helpers_test.go | 56 | shared `crudBody`, temp/chdir/bytes helpers |
| add_test.go | 81 | Add sets draft; description/deps/archetype; validate stays PASS + untouched round-trip |
| add_reject_test.go | 69 | 7 rejections byte-identical; owning-phase dup; empty roadmap; missing file |
| remove_test.go | 57 | success; last-feature → `"features": []`; not-found byte-identical |
| remove_guard_test.go | 64 | in-progress/done refusal; dependent refusal; **draft** dependent refusal |
| promote_inplace_test.go | 86 | in-place finalize clears draft + writes artifacts + revalidates; non-draft error; summary fallback |
| promote_inplace_error_test.go | 51 | slug-not-found; missing-artifacts preflight abort (draft intact); Backlog `--phase`/unknown-phase guards |
| draft_test.go | 51 | IsDraftFeature/DraftFeatures; coverage-set vs dependency-set split |
| readiness_draft_test.go | 28 | classifyFeature draft; ReadySet & Summary exclude draft |
| view_draft_test.go | 30 | BuildView draft:true+readiness:"draft"; counts exclude draft |
| dependency_draft_test.go | 41 | depend-on-a-draft validates; self-dep is a cycle; unknown dep errors |
| rawfeature_find_test.go | 66 | findFeature/featurePhase/toRoadmap/featureDependents (incl draft) |
| rawfeature_mutate_test.go | 79 | appendFeatureToPhase byte-stable render + rejections; removeFeatureAt; replaceFeatureAt |
| rawfeature_error_test.go | 77 | malformed feature/phase decode-error propagation; Add/Remove read errors |
| mutate_validate_test.go | 41 | requirePlannedStatus; requireNoDependents; joinNames |
| artifacts_shared_test.go | 34 | appendScoreArtifacts writes both JSON+MD; missing-file error |
| mdgen_feature_draft_test.go | 24 | deterministic ` *(draft)*` marker; non-draft omits it |

### Colocated — cmd/centinela (moves coverage)
| File | Lines | Focus |
|------|------|-------|
| roadmap_add_test.go | 57 | runRoadmapAdd success; `--phase` required; error propagation |
| roadmap_remove_test.go | 38 | runRoadmapRemove success; not-found error |
| start_guard_draft_test.go | 39 | draftStartError; resolveArchetypeOrder refuses draft; non-draft flag path |
| roadmap_promote_inplace_test.go | 71 | promoteScored in-place finalize; below-9 refusal byte-identical; promoteResultMessage branches |

### tests/ tier trio (required by the tests-step gate; does NOT move coverage)
| File | Lines | Focus |
|------|------|-------|
| tests/unit/roadmap_crud_add_remove_unit_test.go | 99 | Add draft+validate; reject byte-identical; four-reader invariant |
| tests/integration/roadmap_crud_add_remove_integration_test.go | 74 | add→promote-in-place→validate crossing; draft-dependent remove refusal |
| tests/acceptance/roadmap_crud_add_remove_test.go | 100 | binary: add → validate PASS → --json draft readers → ready excludes; reject rows byte-identical |
| tests/acceptance/roadmap_crud_promote_remove_test.go | 79 | binary: remove + rm alias; not-found byte-identical; in-place promote finalize + validate PASS |

## Coverage Gaps

Repo-wide gate (the enforced one, `scripts/check-coverage.sh`): **97.3%** total —
PASS, ~2.3% above the 95% floor.

Per-package (statement-weighted, from the gate run):
- `internal/roadmap` **96.7%**
- `cmd/centinela` **96.3%** (every NEW file — roadmap_add/remove, start_guard_draft,
  the promote in-place additions — is at 100%; the residual is pre-existing untested
  code in unrelated cmd files, out of scope for this feature)

Residual uncovered in internal/roadmap is defensive filesystem-error branches only
(`writeAtomic` temp-create/rename failures, `Save`, `indentValue` re-indent of already-
valid JSON, post-preflight artifact-append failures). Triggering them requires forcing
OS-level I/O errors — flaky/platform-specific — so they are intentionally left; the
write path is atomic and shared with the already-exercised defer/promote commands.

## Acceptance Wiring

- `centinela.toml` already wires acceptance via `go test ./tests/acceptance/...` (feature 1);
  not modified.
- New acceptance tests reuse feature 1's `rmcBin`/`rmcProject`/`rmcRun` helpers — the
  binary is temp-built once per run from the repo (`go build ./cmd/centinela`); no network,
  no git push, no installed-binary dependence.
- `// Acceptance:`/`// Scenario:` traceability tags map to the spec across all four tiers.
- **Deviation (noted):** the full `centinela start <draft>` refusal is asserted at the
  guard level (`resolveArchetypeOrder`/`IsDraftFeature`, cmd/centinela) rather than by
  driving `start` through the acceptance binary — `start` needs a valid PROJECT.md +
  centinela.toml the temp acceptance project does not carry. The acceptance tier instead
  asserts the draft's exemption through the `roadmap --json`/`ready`/`validate` readers.

## Deferred Findings

None. No production code was modified; no real bug found. Senior-engineer's
`dependencyTargetSet` correction (drafts are valid dependency targets) is verified:
a non-draft depending on a draft validates, a self-dep is reported as a cycle.

## Verification

- `go test ./...` — PASS (full suite, exit 0). `go vet ./...` — clean. `check-fmt.sh` — clean.
- New acceptance tests: 5/5 PASS (binary temp-built, 1.4s).
- Coverage gate: repo total 97.3% (PASS).

## Handoff

**→ validation-specialist.** All three test tiers exist and pass; edge-cases artifact
filled; coverage gate green at 97.3%. Run the gatekeeper + `centinela validate` on the
merged tree (per the parallel-merge caution, run the real gate and read its output —
`validate` is not a required check). No open QA risks beyond the noted defensive
fs-error branches and the guard-level (vs binary-level) draft-start assertion.
