### Gatekeeper Report: claim-verification
**Date:** 2026-05-29
**Status:** WARNING

Feature "claim-verification" adds a new `internal/verify` domain package and a
`centinela verify` command, and wires independent claim verification into the
shared `centinela complete` validate gate (`cmd/centinela/complete.go`). It also
adds one additive, optional field — `coverage` — to the evidence schema
(`internal/evidence`) and to `docs/architecture/evidence-contract.md` (mirror in
sync). All findings below are non-blocking; the single WARNING is a
documentation/governance follow-up (PROJECT.md G2 does not yet name
`internal/verify`).

#### Analyzed Specs
Full conflict scan focused on specs touching the shared surfaces this feature
changes (the complete gate, the evidence schema, coverage, orchestration):

- claim-verification.feature (the feature under review)
- enforce-coverage-in-validate.feature
- enforce-actionable-orchestration-evidence.feature
- add-agent-evidence-contract.feature
- evidence-cli.feature
- enforce-step-subagent-orchestration.feature
- enforce-plan-snapshot-inputs.feature
- refine-ux-specialist-evidence.feature
- raise-test-coverage-90.feature / reach-100-coverage.feature
- merge-steward-auto-dispatch.feature
- parallel-feature-worktrees.feature
- diff-aware-gatekeeper.feature
- (plus a survey of the remaining ~60 .feature files in specs/ for references
  to `complete`, evidence fields, coverage, and orchestration roles — none
  contradicted by this change)

#### Findings

- **Affected spec:** PROJECT.md → G2 rule (layer boundaries) — governance, not a
  `.feature` scenario.
  - **Risk:** `internal/verify` is a NEW domain package that imports
    `internal/config`, `internal/evidence`, and `internal/orchestration`
    (confirmed by import scan; it never imports `cmd/` or `ui/`). The current G2
    rule only names `internal/workflow`, `internal/gates`, `internal/ui`,
    `internal/config`, and `internal/roadmap`. A package not named in the rule
    is technically un-governed, so a future contributor cannot tell from
    PROJECT.md what `internal/verify` is allowed to import.
  - **Suggestion:** Update PROJECT.md → G2 to explicitly name `internal/verify`
    as a domain service that may import `internal/config`,
    `internal/evidence`, `internal/orchestration`, and `internal/worktree`
    (read-only), and must not import `cmd/` or `internal/ui`. See verdict below.

- **Affected spec:** enforce-coverage-in-validate.feature
  - **Risk (assessed, NOT a conflict):** Both features deal with "coverage."
    enforce-coverage operates a *total* coverage gate in the validate *pipeline*
    (`./scripts/check-coverage.sh`, fail-below-threshold). claim-verification's
    coverage check is a *per-package, per-claim* re-derivation in the *complete*
    gate that compares a claimed figure to measured coverage within
    `coverage_tolerance`. Different mechanism, different trigger, no shared
    state.
  - **Suggestion:** None. The two coexist; both passed in `centinela validate`.

- **Affected spec:** add-agent-evidence-contract.feature, evidence-cli.feature,
  enforce-actionable-orchestration-evidence.feature
  - **Risk (assessed, NOT a conflict):** The new `coverage` field is added to the
    shared evidence schema consumed by ALL roles. The orchestration structural
    validator (`internal/orchestration/evidence.go`) unmarshals into its own
    struct with plain `json.Unmarshal` (no `DisallowUnknownFields`), so the new
    field is silently ignored by the structural gate and cannot break existing
    evidence parsing or the actionable-output rules. The field is a
    `*float64` (`omitempty`), so absent claims serialize unchanged and older
    files round-trip cleanly via the `Extra` map. Coverage enforcement is
    opt-in: a nil value SKIPs the check.
  - **Suggestion:** None. Backward compatible.

- **Affected spec:** parallel-feature-worktrees.feature
  - **Risk (assessed, NOT a conflict):** `centinela verify` and the complete
    gate resolve their working root via `verifyRoot()` →
    `worktree.DetectFeatureFromCwd`, honoring the worktree operational model.
    Consistent with existing worktree behavior.
  - **Suggestion:** None.

#### Gate Keepers Checklist

- [x] **File size (G1):** PASS. All non-test AND `_test.go` files in `internal/`
  and `cmd/` are <=100 lines (full-tree scan: zero files over 100; largest verify
  file is `claim_stubs.go` at 89). `centinela validate` G1 gate also PASS.
- [~] **Cross-layer imports (G2/G7):** PASS for code, WARNING for docs.
  `internal/verify` imports only inner/leaf domain packages
  (config/evidence/orchestration/worktree); no `cmd/` or `ui/` import. This is an
  acceptable domain-service dependency shape for n-tier. The WARNING is that
  PROJECT.md G2 does not yet name the package (see finding + verdict).
- [x] **`centinela validate` passes:** PASS (exit 0). G1 clean, `go test ./...`
  green, `./scripts/check-coverage.sh` green. `go build ./...` and `go vet ./...`
  also clean.
- [x] **No business logic in outer layer (G7):** PASS. `cmd/centinela/verify.go`
  and `complete.go` are thin: they load config/workflow, call `verify.Verify`,
  render via `ui.RenderVerification`, and branch only on `HasFailures()` /
  `HasWarnings()`. All classification logic lives in `internal/verify`.
- [n/a] **i18n:** N/A. Project is English-only; `gates.i18n = false`.
- [x] **Scaffold-mirror parity (this feature's doc):**
  `docs/architecture/evidence-contract.md` is byte-identical to its mirror
  `internal/scaffold/assets/docs/architecture/evidence-contract.md` (qa-senior
  sync stuck). Note: `diff -r` over the whole arch dir shows PRE-EXISTING drift
  in gatekeepers.md, new-project-guide.md, testing-strategy.md,
  workflow-enforcement.md, and production-readiness-prompt.md — all explicitly
  allowlisted in `tests/acceptance/scaffold_arch_parity_acceptance_test.go` and
  unrelated to this feature (the only arch-doc change vs main is
  evidence-contract.md). The parity acceptance test PASSES.

#### G2 Verdict on internal/verify

**Recommended: YES — update PROJECT.md → G2 to name `internal/verify`.**

The dependency shape is correct and idiomatic for n-tier: `internal/verify` is a
domain service that orchestrates lower domain/leaf packages
(config + evidence + orchestration + worktree) and is consumed by the outer
`cmd/` layer; it imports nothing from `cmd/` or `internal/ui`, so it introduces
no cycle and no upward dependency. The only gap is documentation: the rule as
written enumerates the allowed importers exhaustively and silently omits the new
package, leaving its boundary un-stated. This is a low-risk, non-blocking
governance follow-up — fold it into the docs step (step 5) of this feature so
the rule stays the source of truth. It does not require reworking any code.

#### Recommendation
- WARNING: No definite conflicts with existing specs; the shared changes (complete
  gate, evidence schema `coverage` field) are additive and backward compatible,
  and all gates pass. Proceed. Before/at the docs step, amend PROJECT.md G2 to
  name `internal/verify` and its allowed imports so the layer rule remains
  authoritative.
