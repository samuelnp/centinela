Feature: Adversarial validate verifier — a grounded refutation gate before complete
  As an operator advancing a feature through the validate step
  I want the gatekeeper report to be a fresh-context adversary's grounded verdict,
    not a narrated compliance checklist
  So that a dead subagent's stub, an unexecuted "Status: SAFE" claim, or a
    stale verdict from before the last fix can never let `centinela complete`
    advance

  # Design decisions this spec encodes (docs/plans/adversarial-validate-verifier.md):
  # D1 the role slug stays "gatekeeper"; D2 the machine-readable record is a
  # fenced ```json centinela:verification block inside the .md report; D3/D3a
  # freshness = HEAD revision AND a .workflow/-excluded working-tree digest;
  # D4 legacy back-compat is state-dated via ValidateContract pinned at
  # workflow start (empty => legacy); D5 Status first-token vocabulary is
  # SAFE/WARNING/CRITICAL with BLOCKING/UNSAFE as legacy aliases for CRITICAL.

  # ── Happy path ───────────────────────────────────────────────────────────

  Scenario: A grounded SAFE verdict with a matching revision advances validate
    Given a feature "checkout-fix" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/checkout-fix-gatekeeper.md has "**Status:** SAFE"
    And its centinela:verification block records a non-empty commands array
    And one recorded command's argv joins to "centinela validate" with exitCode 0
    And the block's revision equals HEAD and its treeDigest matches the current tree
    When the user runs "centinela complete checkout-fix"
    Then the validate step should advance without blocking
    And the docs step should become active

  Scenario: A WARNING verdict advances validate but the finding is recorded
    Given a feature "rate-limit-tweak" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/rate-limit-tweak-gatekeeper.md has "**Status:** WARNING"
    And its centinela:verification block records a non-empty commands array
    And one recorded command's argv joins to "centinela validate" with exitCode 0
    And the block's revision equals HEAD and its treeDigest matches the current tree
    When the user runs "centinela complete rate-limit-tweak"
    Then the validate step should advance without blocking
    And the WARNING finding should be surfaced in the complete output
    And the finding should land in the memory ledger

  # ── Verdict semantics — first token of Status, no prose scan (D5) ────────

  Scenario: A CRITICAL verdict blocks complete with the finding echoed
    Given a feature "leaky-session" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/leaky-session-gatekeeper.md has "**Status:** CRITICAL"
    And the report's first finding line is "session tokens are never invalidated on logout"
    When the user runs "centinela complete leaky-session"
    Then completion should be hard-blocked
    And the block message should echo "session tokens are never invalidated on logout"
    And the message should instruct re-verification with a FRESH verifier context

  Scenario Outline: Legacy severity aliases normalize to CRITICAL and block complete
    Given a feature "<feature>" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/<feature>-gatekeeper.md has "**Status:** <alias>"
    When the user runs "centinela complete <feature>"
    Then completion should be hard-blocked as a CRITICAL verdict

    Examples:
      | feature          | alias    |
      | legacy-blocking  | BLOCKING |
      | legacy-unsafe    | UNSAFE   |

  Scenario: A missing Status line blocks complete
    Given a feature "no-status-line" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/no-status-line-gatekeeper.md contains findings but no "**Status:**" line
    When the user runs "centinela complete no-status-line"
    Then completion should be hard-blocked
    And the block message should say the verdict is missing or unparseable

  Scenario: An unparseable Status line blocks complete, never fails open
    Given a feature "garbled-status" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/garbled-status-gatekeeper.md has "**Status:** mostly fine I think"
    When the user runs "centinela complete garbled-status"
    Then completion should be hard-blocked
    And the block message should say the verdict is missing or unparseable

  # ── Grounding: the commands-run record makes a verdict admissible (AC2) ──

  Scenario: A report without a non-empty commands-run record fails evidence validation
    Given a feature "dead-subagent-stub" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/dead-subagent-stub-gatekeeper.md has "**Status:** SAFE"
    And its centinela:verification block has an empty commands array
    When the user runs "centinela complete dead-subagent-stub"
    Then completion should be hard-blocked
    And the block message should say the report has no commands-run record
    And the message should name docs/architecture/gatekeeper-prompt.md as the remedy

  Scenario: centinela artifact new followed immediately by centinela complete FAILS
    Given a feature "just-scaffolded" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And the operator ran "centinela artifact new just-scaffolded gatekeeper" and nothing else
    When the user runs "centinela complete just-scaffolded"
    Then completion should be hard-blocked
    And the block message should say the report has no commands-run record

  Scenario: A report whose commands never include a passing "centinela validate" run is refused
    Given a feature "partial-commands" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/partial-commands-gatekeeper.md has "**Status:** SAFE"
    And its centinela:verification block records only a "go test ./..." command
    When the user runs "centinela complete partial-commands"
    Then completion should be hard-blocked
    And the block message should say the report has no commands-run record

  Scenario: A verifier that cannot execute commands in its harness fails closed
    Given a feature "no-bash-harness" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/no-bash-harness-gatekeeper.md states it could not execute commands
    And its centinela:verification block has an empty commands array
    And its Status line reads "**Status:** CRITICAL"
    When the user runs "centinela complete no-bash-harness"
    Then completion should be hard-blocked
    And no narrated pass should be possible without a commands-run record

  # ── Freshness: HEAD revision AND working-tree digest (D3, D3a) ───────────

  Scenario: Revision skew after fixes landed on top of the verified commit demands fresh verification
    Given a feature "post-fix-drift" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/post-fix-drift-gatekeeper.md has "**Status:** SAFE" with a full commands-run record
    And the block's revision is an ancestor commit, not current HEAD
    When the user runs "centinela complete post-fix-drift"
    Then completion should be hard-blocked
    And the block message should say the verification is stale
    And the message should name both the verified sha and the current HEAD sha
    And the message should instruct spawning a FRESH verifier

  Scenario: Uncommitted in-place fixes leave HEAD unchanged but stale the tree digest
    Given a feature "uncommitted-patch" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/uncommitted-patch-gatekeeper.md has "**Status:** SAFE" with a full commands-run record
    And the block's revision equals the current HEAD
    But a tracked source file was edited after the report was stamped, uncommitted
    When the user runs "centinela complete uncommitted-patch"
    Then completion should be hard-blocked
    And the block message should say the verification is stale
    And a HEAD-only comparison would have wrongly admitted this verdict

  Scenario: Mutating only files under .workflow/ does not stale a fresh verification
    Given a feature "workflow-only-churn" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/workflow-only-churn-gatekeeper.md has "**Status:** SAFE" with a full commands-run record
    And the block's revision and treeDigest were stamped as the verifier's last action
    And after stamping, only files under .workflow/ changed on disk
    When the user runs "centinela complete workflow-only-churn"
    Then the validate step should advance without blocking
      # the verifier's own report write must never self-invalidate its stamp

  Scenario: centinela artifact stamp records the verified revision and tree digest
    Given a feature "stamp-me" at the validate step
    And a gatekeeper report body has already been written with "**Status:** SAFE" and a full commands array
    When the operator runs "centinela artifact stamp stamp-me"
    Then the centinela:verification block's revision should equal "git rev-parse HEAD"
    And the block's treeDigest should reflect the current working tree excluding .workflow/
    And the block's commands array should remain byte-identical to what was already written

  # ── Legacy back-compat is state-dated, not file-presence either-set (D4) ─

  Scenario: A legacy in-flight workflow still completes with an old-format validation-specialist report
    Given a feature "already-in-flight" at the validate step
    And its workflow's ValidateContract is empty because it started before this feature shipped
    And .workflow/already-in-flight-validation-specialist.md exists with status "done"
    And .workflow/already-in-flight-gatekeeper.md exists in the old checklist format with no commands-run record
    When the user runs "centinela complete already-in-flight"
    Then the validate step should advance without blocking
      # empty ValidateContract => today's existence-only gate, verbatim

  Scenario: A workflow pinned to the new contract refuses a hand-authored legacy-format report
    Given a feature "fresh-workflow" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/fresh-workflow-gatekeeper.md exists in the old checklist format with no commands-run record
    And .workflow/fresh-workflow-validation-specialist.md was hand-authored to satisfy the legacy pair
    When the user runs "centinela complete fresh-workflow"
    Then completion should be hard-blocked
    And the block message should say the report has no commands-run record
      # a fresh feature cannot dodge the new format by hand-authoring legacy files

  Scenario: The validate step no longer requires validation-specialist evidence
    Given a feature "no-validation-specialist" at the validate step
    And its workflow was pinned with ValidateContract "adversarial-v1" at start
    And .workflow/no-validation-specialist-gatekeeper.md has "**Status:** SAFE" with a full, fresh commands-run record
    And no validation-specialist evidence file exists for this feature at all
    When the user runs "centinela complete no-validation-specialist"
    Then the validate step should advance without blocking
    And RequiredRolesForFeature for "no-validation-specialist" at step "validate" should be exactly [gatekeeper]

  # ── Hook directive names the verifier, its tier, and the input contract (AC6) ─

  Scenario: The validate-step hook directive names the adversarial verifier and its reasoning tier
    Given a feature "directive-check" at the validate step
    When the user requests hook context for "directive-check"
    Then the directive should name the gatekeeper role as the adversarial verifier
    And the directive should name its model tier as reasoning
    And the directive should instruct passing only the feature slug and file paths
    And the directive should prohibit summarizing the implementation into the verifier's prompt
