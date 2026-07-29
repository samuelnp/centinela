### Big-Thinker Report: adversarial-validate-verifier
**Date:** 2026-07-28

## Problem

The `validate` step spends two subagents and produces zero artifacts that
re-derive anything. `validation-specialist` (tier `fast`) composes the other
reports into a narrative; `gatekeeper` (tier `fast`) runs a spec-conflict
compliance checklist. The only mechanical teeth are `validateGatekeeper` in
`internal/workflow/validate.go`, which is a bare `os.Stat` on
`.workflow/<feature>-gatekeeper.md`. Consequently a dead subagent's 200-byte
stub passes, a report asserting `Status: SAFE` without having executed
anything passes, and a verdict produced three fix-rounds ago passes. This
project's own history supplies each failure mode: false "build ok / tests
pass" claims from code-step agents, a qa-senior dying mid-write, an
acceptance suite silently skipping every scenario behind a green gate. The
centinela-v2 port-back is the positive evidence: a fresh context primed to
*refute* found a real defect in 20/20 rounds, where self-review by the
implementing context found none. Operators need validate to be the step where
an independent adversary tries to break the completion claim and fails —
before `complete` advances — and they need that verdict to be inadmissible
unless it is grounded in commands actually run against the tree actually seen.

## Scope

- **In:** refutation-stance rewrite of `gatekeeper-prompt.md` + scaffold
  mirror; a machine-readable commands-run record and verified-revision
  binding inside the gatekeeper report; `complete`-gate enforcement of
  verdict + record + freshness; `RequiredRolesForFeature(f,"validate")` drops
  `validation-specialist` and requires `gatekeeper`, with legacy in-flight
  workflows still completing; validate-step hook directive and statusline
  wording; agent-config emitter entry for the verifier.
- **Out (pre-agreed in the brief):** the plan-role merge
  (`unified-plan-specialist`); changes to `centinela verify`'s claim-
  verification mechanics; the production-readiness gate; merge-steward;
  mechanical detection of contaminated delegation (prompt-flagged only).
- **Out (decided in this plan, deferred below):** `centinela verify`
  cross-checking recorded argv against ground truth; typed gate error codes;
  reusing `PriorTestRun` in the complete path; closing the parity allowlist.

## Dependencies & Assumptions

- **Depends on** the PR #59 `**Status:**` first-token contract — preserved
  verbatim and consolidated into one parser instead of today's single private
  copy in `internal/delivery/verdict.go`.
- **Depends on** the pinned-at-start workflow-state field pattern already
  used by `EnforcementProfile`, `Archetype` and `DriverModel`; on
  `internal/setup`'s `BuildSyncPlan`/`ApplySync` managed-file seam; and on the
  existing evidence machinery, which already knows `gatekeeper` as a
  validate-step role and artifact kind.
- **Assumes the role slug stays `gatekeeper`.** The brief renames the
  *stance*, and AC2 names `.workflow/<feature>-gatekeeper.md` explicitly. The
  slug is load-bearing in ~10 places (memory capture, delivery artifacts,
  statusline, evidence roles/kinds, config override keys, the scaffold parity
  allowlist, three acceptance tests); renaming buys nothing and costs a wide,
  risky diff.
- **Assumes `centinela complete` auto-commits *after* the gate**, so HEAD does
  not move during the validate step. This is precisely what makes a HEAD-only
  freshness check nearly worthless and forces a working-tree digest.
- **Assumes** the coverage gate is per-package with no `-coverpkg`, so every
  new package needs colocated `_test.go`; G1's 100-line cap applies to test
  files under `internal/` and `cmd/`.
- **Independent of** `unified-plan-specialist`; the only textual overlap is
  the CLAUDE.md quick-reference table and the `opencode_agent_config.go` map.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Verifier theater — a refutation prompt that still rubber-stamps | High | Medium | Commands record + revision/digest enforced mechanically in the gate, not by convention; missing or unparseable ⇒ blocked, never fail-open |
| Stale-verdict hole: in-place fixes leave HEAD unchanged | High | High | Freshness = HEAD **and** a working-tree digest; a HEAD-only check would be theater in exactly the targeted scenario |
| Digest self-invalidation: the verifier's own report dirties the tree it stamped | High | Certain if naive | Exclude `.workflow/` from the digest (it is verification's output, never its input); assert both directions |
| Legacy in-flight workflow bricked mid-validate | Medium | Medium | `ValidateContract` pinned at start; empty ⇒ today's behavior verbatim. Chosen over a file-presence "either-set" rule, which would let a fresh feature dodge the format by hand-authoring legacy-named files |
| Prompt-doc test coupling — three acceptance tests enumerate `validation-specialist-prompt.md` | Medium | High | Keep the file plus its CLI-mandate and Deferred-Findings blocks; deprecate in prose only |
| Scaffold mirror drift | Medium | High | Source + mirror edited in the same commit; parity test is a gate. `workflow-enforcement.md` is unmirrored by allowlist — do not "fix" it here |
| Wall-clock blowout: the verifier's suite run is additive to `complete`'s runs | Medium | Medium | Prompt mandates `centinela validate` ONCE (it already runs `[validate] commands`); keep the redundant acceptance command out of `validate.commands`; background long completes. `verify.timeout_seconds` bounds one command, not total wall clock |
| `git` latency injected into the per-prompt hook | Low | Medium | Freshness check lives in the `complete` path only; `ValidateArtifacts` (what `hook context` calls) stays pure file I/O |
| Blocking loop — over-zealous CRITICAL on a healthy feature | Medium | Low | WARNING advances and lands in the memory ledger; `centinela revise` is the escape valve; no skip flag exists, by design |
| Coverage dip below the 95% gate on touched packages | Low | Medium | Colocated tests ≤100 lines, aim ≥97% so a parallel merge cannot tip main red |
| Parallel-merge conflict with `unified-plan-specialist` | Low | Medium | Overlap is textual only; run the real gate on the merged tree before merging second |

## Rollout

- **Step 1 — `internal/gatereport` (parser).** Verdict parsing (`SAFE` /
  `WARNING` / `CRITICAL`, with `BLOCKING`/`UNSAFE` normalized as legacy
  aliases), the `centinela:verification` fenced-JSON block, and the
  admissibility check. `internal/delivery/verdict.go` collapses to a
  delegation so there is one parser. Pure; provably inert.
- **Step 2 — `internal/treestate` + `centinela artifact stamp <feature>`.**
  HEAD revision plus a `.workflow/`-excluded working-tree digest, written into
  the report's block as the verifier's last action. New surface only.
- **Step 3 — the gate.** `validate_gatekeeper.go` (verdict + commands record,
  pure) and `validate_freshness.go` (git-backed, called from `runComplete`
  before `executeValidation`), both keyed off a new `ValidateContract` field
  pinned at start. In-flight features untouched.
- **Step 4 — orchestration policy.** `RoleGatekeeper` constant, validate ⇒
  `[gatekeeper]`, reasoning tier, `handoffTo` → documentation-specialist,
  legacy role set selected from the pinned contract. First step that changes
  what validate *requires*.
- **Step 5 — prompt rewrite + contract docs + byte-identical mirrors.** Must
  ship in the same branch as steps 3–4 so the first feature started under the
  new contract has a prompt that produces a conforming report.
- **Step 6 — directive, statusline, agent-config emitter.**
- **Step 7 — spec-driven acceptance e2e.** Not optional: this feature is
  itself started under the *old* contract, so its own `complete` never
  exercises the new gate. The acceptance test is the only place the gate is
  proven end to end. Use a local bare repo as origin — a real network push
  hangs `go test` and times out claim verification.

## Deferred Findings

- `verify-crosscheck-verifier-commands` — have `centinela verify` compare the
  verifier's recorded argv/exit codes against its own claim-verification runs;
  v1 enforces the record's shape, not its truthfulness.
- `typed-gate-error-codes` — `hook_statusline_rules.go` classifies validate
  failures by substring-matching the error text, misreporting any non-BLOCKING
  error as `MISSING_PROD_READINESS`; typed codes are the fix.
- `reuse-prior-test-run-in-complete-verify` — `complete_verify.go` never sets
  `verify.Deps.PriorTestRun`, so `complete` re-runs the suite
  `executeValidation()` just ran; the seam exists and is unused.
- `mirror-workflow-enforcement-doc` — `workflow-enforcement.md` is allowlisted
  as "not mirrored", so scaffolded projects never receive it and edits to it
  (including this feature's) silently never reach them.

All four recorded via `centinela roadmap defer … --source
adversarial-validate-verifier/big-thinker`.

## Handoff

- Next role: feature-specialist
- Plan: `docs/plans/adversarial-validate-verifier.md` (7 slices, file-by-file
  change list, per-slice test strategy).
- Outstanding questions for the Gherkin:
  1. Should the gate require a *test-suite* command in the record in addition
     to `centinela validate` exit 0? The plan says no for v1 — `centinela
     validate` already runs every `[validate] commands` entry, so requiring
     both is either redundant or forces project-specific argv matching.
  2. Pin the exact digest input set: `git status --porcelain=v1` plus
     `git diff HEAD`, both filtered of `.workflow/` paths, with untracked
     non-`.workflow` files counting as dirt. Spec it so a later refactor
     cannot quietly loosen it.
  3. Legacy acceptance is state-dated (`ValidateContract` empty ⇒ legacy),
     deliberately *not* the file-presence "either-set" rule the sibling
     `unified-plan-specialist` brief contemplates. If the two features should
     share one discipline, this is the decision point.
