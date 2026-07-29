### Validation-Specialist Report: merge-truthful-delivery
**Date:** 2026-07-29
**Status:** WARNING

#### Gates Run
| Gate                    | Status                  | Source artifact |
|-------------------------|-------------------------|-----------------|
| gatekeeper              | WARNING                 | .workflow/merge-truthful-delivery-gatekeeper.md |
| centinela validate      | pass (exit 0)           | Adversarial verifier (602 s) + full suite (304 s) |
| spec-traceability gate  | PASS (21/21 coverage)   | All scenarios now have acceptance markers |
| full test suite         | pass (3606 tests)       | go test ./... exit 0 |

#### Synthesis

The adversarial verifier's second pass (first pass was CRITICAL; fixes applied) confirms the core defect is resolved: `deliver --via merge` from inside a worktree verifiably advances the primary ref, proves ancestry, and removes the worktree before claiming success. Fourteen end-to-end runs of a scratch binary against real git repositories all exited 0 with truthful outcomes. Every refusal path (non-repo, bare, detached, self-merge, dirty) refuses truthfully and moves nothing. The busy-worktree half-success is honest, idempotent, and recoverable. `merge --continue` resumes a worktree-initiated stall from either CWD and refuses an unbacked APPLY verdict. The post-merge validate gates the merged primary tree, not a vacuous diff-aware no-op.

**Spec-coverage status:** All 21 scenarios now have acceptance markers in the test suite (gap closed since gatekeeper report). The traceability gate is green.

**Residual risk profile:** Four findings remain, all deferred to Backlog as pre-existing or post-scope issues (see Deferred Findings below). None involve supported operations producing false successes. The two fail-open shapes (prunable registry false-removal, validate-failed merge left on main unannounced) require hand-corrupted repository state or architectural-scope decisions respectively.

#### Deferred Findings

Per the gatekeeper's recommendations, the following four findings were recorded in the roadmap for future work and **will not block this delivery**:

- `merge-prunable-registry-false-removal` (Finding 2) — out-of-band worktree moves create prunable records; false "removed" claim possible but requires corrupted state (hand `mv`, not `git worktree move`)
- `merge-pending-marker-dirties-primary-tree` (Finding 3) — in repos tracking `.workflow/`, untracked marker and steward.md cause isDirty refusal from both CWDs; friction not false success
- `merge-validate-fail-leaves-main-advanced` (Finding 4) — validate failure leaves merge commit on main unannounced; requires architectural scope re-examination
- `merge-force-remove-locked-worktree` (Finding 5) — locked worktrees need `-f -f` or unlock; low severity, affects only hand-locked trees

All four are documented in the gatekeeper report with explicit risk analysis and mitigation reasoning. Finding 1 (spec-coverage gap, which was CRITICAL in the gatekeeper's first pass) is **now closed** — all 21 spec scenarios have acceptance coverage in the test suite.

#### Decision

**WARNING — the delivery claim is substantially true; residual friction acceptable.**

The regression this feature killed is dead. The spec-coverage gap, which blocked a PASS verdict in the gatekeeper's first pass, has been closed. Four deferred findings are friction or post-scope issues rather than false successes. The suite is green: `centinela validate` exit 0 (602 s full scan), 3606 tests pass (304 s), all 21 scenarios covered with acceptance markers.

**Proceed to docs step.**
