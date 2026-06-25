# completion-delivery-prompt — qa-senior

> Authored directly by the orchestrator (the delegated qa-senior subagent was not used for this
> feature). Every gate re-verified independently.

## Test Inventory

| Tier | File | Covers |
|------|------|--------|
| colocated unit | `internal/gitutil/options_test.go` | DeliveryOptions 4 matrix rows; Supports |
| colocated unit | `internal/gitutil/directive_test.go` | DeliveryDirective (both/pr-only/empty); GitHubCLIAvailable |
| colocated unit | `internal/gitutil/remote_test.go` | HasOriginRemote present/absent(ExitError)/empty/real-error |
| colocated unit | `internal/ui/render_delivery_test.go` | RenderDeliveryChoice with options + empty |
| colocated cmd | `cmd/centinela/deliver_test.go` | bad --via; pr-without-origin; merge-without-worktree |
| colocated cmd | `cmd/centinela/deliver_pr_test.go` | pr no-origin (no push); dirty-tree refusal |
| colocated cmd | `cmd/centinela/complete_delivery_test.go` | done-branch emits directive (both options), no delivery |
| tests/unit | `tests/unit/completion_delivery_unit_test.go` | matrix + directive composition |
| tests/integration | `tests/integration/completion_delivery_integration_test.go` | HasOriginRemote on real temp repos |
| tests/acceptance | `tests/acceptance/completion_delivery_complete_test.go` | Scenarios 1–5 (completion directive) |
| tests/acceptance | `tests/acceptance/completion_delivery_deliver_test.go` | Scenarios 6–11 (deliver command) |

All 11 spec scenarios carry exact-match `// Scenario:` comments under `// Acceptance:` headers in
`tests/acceptance/`. Every new `_test.go` is ≤80 lines.

## Coverage Gaps

Repo coverage gate passes (see `centinela validate` / `./scripts/check-coverage.sh`). The only
deliberately-unasserted lines are the live-remote/`gh` side effects in `deliver_pr.go` (real push,
`gh pr create`, `gh`-absent push+manual): they need a live remote/`gh` auth and are not
deterministically reproducible. Acceptance asserts the reachable guarantees (guards passed, no false
PR claim, delegation reached); the real network/`gh` behaviors are exercised by `gh`/`git` and the
existing `centinela merge` command. Documented in `.workflow/completion-delivery-prompt-edge-cases.md`.

## Acceptance Wiring

`centinela.toml` `[validate].commands` already runs `go test ./tests/acceptance/...` and
`./scripts/check-coverage.sh` — wiring exists; not edited.

## Handoff

Next role: validation-specialist. Full suite green, coverage gate green, all 11 scenarios traced,
`.workflow/completion-delivery-prompt-edge-cases.md` filled, evidence JSON validates.
