### Feature-Specialist Report: adversarial-validate-verifier
**Date:** 2026-07-28

#### Behavior Summary
The validate step's gatekeeper stops being a compliance-checklist narrator and
becomes a fresh-context adversary whose verdict is only admissible when it is
grounded. Operationally: `docs/architecture/gatekeeper-prompt.md` (and its
scaffold mirror) is rewritten to a refutation stance with a paths-only input
contract and an explicit prohibition on treating orchestrator summaries or
other roles' narratives as evidence. The verifier itself runs `centinela
validate` and the project test suite, records every command (argv, exit code,
duration) in a fenced ` ```json centinela:verification ` block inside
`.workflow/<feature>-gatekeeper.md` (D2), and finishes by stamping that block
with the HEAD revision and a `.workflow/`-excluded working-tree digest via
`centinela artifact stamp <feature>` (D3/D3a). `centinela complete` at the
validate step now enforces, mechanically and fail-closed: (1) the `**Status:**`
first-token verdict is `SAFE`/`WARNING` (advance) or `CRITICAL`/legacy aliases
`BLOCKING`/`UNSAFE` (block), never a prose scan (D5); (2) the commands-run
record is non-empty and includes a passing `centinela validate` invocation —
a report lacking this is refused as inadmissible, closing the dead-subagent-
stub hole; (3) the stamped revision and tree digest match the tree at
`complete` time — a HEAD-only check would be theater because `complete`
auto-commits *after* the gate, so uncommitted in-place fixes leave HEAD
unchanged (D3). `RequiredRolesForFeature(f, "validate")` drops
`validation-specialist` and requires only `gatekeeper`, but this is
state-dated via a `ValidateContract` field pinned at workflow start (D4):
workflows already in flight when this ships have an empty `ValidateContract`
and keep today's existence-only, two-role behavior verbatim; only features
started after this lands are gated under the new contract, and they cannot
dodge it by hand-authoring legacy-named evidence files. The validate-step hook
directive names the verifier, its reasoning-tier model, and the paths-only
delegation contract (AC6).

#### Gherkin Scenarios
File: `specs/adversarial-validate-verifier.feature` (18 `Scenario`/`Scenario
Outline` declarations; the alias outline expands to 2 executed rows, 19
scenarios total).

- **Happy path** — "A grounded SAFE verdict with a matching revision advances
  validate" and "A WARNING verdict advances validate but the finding is
  recorded" (AC4, AC7; brief's WARNING-preserved edge case).
- **Verdict semantics (D5)** — "A CRITICAL verdict blocks complete with the
  finding echoed"; Outline "Legacy severity aliases normalize to CRITICAL and
  block complete" (BLOCKING, UNSAFE); "A missing Status line blocks complete";
  "An unparseable Status line blocks complete, never fails open" (AC4).
- **Grounding / commands-run record (AC2)** — "A report without a non-empty
  commands-run record fails evidence validation"; "centinela artifact new
  followed immediately by centinela complete FAILS" (the plan's named
  dead-subagent regression test in one line); "A report whose commands never
  include a passing centinela validate run is refused"; "A verifier that
  cannot execute commands in its harness fails closed" (brief's harness-
  without-Bash edge case).
- **Freshness (D3, D3a)** — "Revision skew after fixes landed on top of the
  verified commit demands fresh verification"; "Uncommitted in-place fixes
  leave HEAD unchanged but stale the tree digest" (the scenario that proves a
  HEAD-only check would be insufficient); "Mutating only files under
  .workflow/ does not stale a fresh verification" (D3a self-invalidation
  regression, both directions covered across these two scenarios); "centinela
  artifact stamp records the verified revision and tree digest".
- **Legacy back-compat (D4, AC3)** — "A legacy in-flight workflow still
  completes with an old-format validation-specialist report"; "A workflow
  pinned to the new contract refuses a hand-authored legacy-format report"
  (state-dated, not file-presence either-set); "The validate step no longer
  requires validation-specialist evidence".
- **Hook directive (AC6)** — "The validate-step hook directive names the
  adversarial verifier and its reasoning tier".

#### UX States
This is a CLI governance feature; there is no graphical UI surface. Surfaces
are `centinela complete`/`centinela artifact stamp` stdout/stderr lines and
the hook-injected directive text.

| State    | Trigger | Surface |
|----------|---------|---------|
| loading  | n/a — no long-running foreground UI state; `centinela validate`/test-suite runs happen inside the verifier subagent, not the CLI the operator watches | n/a |
| empty    | `centinela artifact new <f> gatekeeper` stub with an empty `commands` array and no Status line | stdout of the subsequent blocked `centinela complete <f>`: "gatekeeper report has no commands-run record — a narrated verdict is not evidence. Re-run the verifier (see docs/architecture/gatekeeper-prompt.md)." |
| error    | CRITICAL verdict, missing/unparseable Status, or stale revision/digest | stderr of `centinela complete <f>`, one of the three house-style messages named in the plan (verdict/commands/staleness), always naming the remedy |
| success  | SAFE or WARNING verdict, non-empty grounded commands record, fresh revision+digest | stdout of `centinela complete <f>`: step advances; WARNING additionally surfaces the finding text and lands it in the memory ledger |

#### Out-of-Scope
- The plan-role merge (`unified-plan-specialist`) — independent feature.
- Changes to `centinela verify`'s claim-verification mechanics — it remains
  the mechanical layer the verifier complements, unchanged by this feature.
- The production-readiness gate and merge-steward — untouched.
- Mechanical detection of contaminated delegation (orchestrator pasting a
  narrative into the verifier's prompt anyway) — prompt-flagged only
  (WARNING-level smell the prompt instructs the verifier to self-report under
  "Inputs Read"); no mechanical enforcement in this feature.
- `centinela verify` cross-checking the verifier's recorded argv/exit codes
  against its own claim-verification ground truth — v1 enforces the record's
  shape and non-emptiness, not the truthfulness of its contents.
- Typed gate error codes replacing `hook_statusline_rules.go`'s substring
  classification.
- Reusing `verify.Deps.PriorTestRun` in the `complete` path to avoid a
  redundant suite run.
- Closing the scaffold `mirrorParityAllowlist` for `workflow-enforcement.md`.

#### Deferred Findings
No new gaps found beyond what big-thinker already deferred
(`verify-crosscheck-verifier-commands`, `typed-gate-error-codes`,
`reuse-prior-test-run-in-complete-verify`, `mirror-workflow-enforcement-doc`,
all recorded with `--source adversarial-validate-verifier/big-thinker`).
None.

#### Handoff
- Next role: senior-engineer
- Open clarifications: none — the big-thinker report's three "outstanding
  questions for the Gherkin" are resolved in this spec: (1) the gate requires
  only a passing `centinela validate` entry in the commands record, not a
  separately-matched test-suite argv (plan §3 D2/§4 slice 1 `check.go`); (2)
  the digest input set is pinned as `git status --porcelain=v1` + `git diff
  HEAD`, both filtered of `.workflow/` paths (plan §3 D3a, §4 slice 2
  `digest.go`) — spec'd via the "excluding .workflow/" and "workflow-only
  churn" scenarios so a later refactor cannot quietly loosen it; (3) legacy
  acceptance is state-dated via `ValidateContract` (D4), not the
  file-presence either-set rule — spec'd explicitly in the "hand-authored
  legacy-format report" negative scenario.
