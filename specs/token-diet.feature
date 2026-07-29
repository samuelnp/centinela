Feature: Token diet — four measured leaks closed without weakening any gate
  As an orchestrating agent and as a human operator
  I want plan evidence to snapshot only what the planner read, UX tags to be
    stated once, the roadmap summary to be injected only when it is new, and
    the built-in model defaults to be undated family aliases
  So that Centinela's per-feature and per-prompt token cost stops growing with
    repo age and with session length, while every in-flight workflow and every
    existing operator config keeps working exactly as it does today

  # Design decisions this spec encodes (docs/plans/token-diet.md):
  # D1 RequiredPlanInputs becomes construction-derived (own brief + own plan)
  # and performs no filesystem glob. D2 the snapshot rule is INCLUDE, never
  # ONLY — a superset validates, which is what keeps legacy in-flight evidence
  # (and this feature's own ~122-input plan evidence) valid across the change.
  # D3 path normalization anchors on docs/features/ OR docs/plans/, symmetric
  # for both required entries. D4 normalizeUXTag cuts at the FIRST colon, so a
  # descriptive entry matches; the change only ever ADDS matches. D5 the
  # roadmap summary is digest-deduped per host session, state at the
  # git-ignored .workflow/.roadmap-digest, and every uncertainty fails OPEN
  # (renders). D6 the built-in claude column becomes bare family aliases; the
  # opencode column keeps a provider-qualified id; codex stays empty so the
  # rule-4 tier-name fallback is unchanged. D7 the capability id->class map is
  # a strict SUPERSET — alias keys added, every legacy pinned id retained.
  # D8 out of scope: WS3.2 evidence-lite and WS3.5 scaffold slimming.

  # ── A. Plan snapshot: own brief + own plan, no glob ──────────────────────

  Scenario: The required plan-input set does not grow with the number of briefs
    Given a repository containing 120 files under "docs/features/"
    When the required plan inputs are computed for feature "token-diet"
    Then the set should contain exactly two entries
    And the entries should be "docs/features/token-diet.md" and
      "docs/plans/token-diet.md"
    And no other feature's brief should appear in the set

  Scenario: The required set is identical in an empty repository
    Given a repository containing no files under "docs/features/"
    When the required plan inputs are computed for feature "token-diet"
    Then the set should be identical to the set computed in a repository with
      120 briefs
      # the invariant that fails if anyone reintroduces a filesystem glob

  Scenario: The required set is derived by construction, not by existence
    Given neither "docs/features/ghost.md" nor "docs/plans/ghost.md" exists
    When the required plan inputs are computed for feature "ghost"
    Then both paths should still be required

  Scenario: Planner evidence listing exactly the two required inputs validates
    Given a feature "token-diet" at the plan step pinned to "planner-v1"
    And .workflow/token-diet-planner.json lists inputs
      "docs/features/token-diet.md" and "docs/plans/token-diet.md"
    And it has a non-empty edgeCases list and handoffTo "senior-engineer"
    When the operator runs "centinela evidence validate token-diet"
    Then the command should exit 0
    And no "missing feature-doc snapshot inputs" error should be reported

  Scenario: Legacy evidence enumerating every brief still validates
    Given an in-flight workflow whose plan evidence lists all 120
      "docs/features/*.md" paths plus its own brief and its own plan
    When the plan-snapshot inputs are validated
    Then validation should pass
      # the rule is INCLUDE, not ONLY: a superset is legal, so the change
      # never strands a workflow that was planned under the old rule

  Scenario: Evidence missing the feature's own plan is rejected
    Given plan evidence whose inputs list "docs/features/token-diet.md" but
      not "docs/plans/token-diet.md"
    When the plan-snapshot inputs are validated
    Then validation should fail with "missing feature-doc snapshot inputs"
    And the message should name "docs/plans/token-diet.md"

  Scenario: Evidence with an empty inputs list is rejected naming both paths
    Given plan evidence whose inputs list is empty
    When the plan-snapshot inputs are validated
    Then validation should fail with "missing feature-doc snapshot inputs"
    And the message should name both required paths

  Scenario Outline: Path normalization is symmetric for briefs and plans
    Given plan evidence for feature "token-diet" whose inputs contain
      "<written>" alongside the other required path
    When the plan-snapshot inputs are validated
    Then validation should pass

    Examples:
      | written                                   |
      | docs/features/token-diet.md               |
      | ./docs/features/token-diet.md             |
      | docs\features\token-diet.md               |
      | /home/user/repo/docs/features/token-diet.md |
      | docs/plans/token-diet.md                  |
      | ./docs/plans/token-diet.md                |
      | docs\plans\token-diet.md                  |
      | /home/user/repo/docs/plans/token-diet.md  |

  Scenario: evidence init pre-fills exactly the required inputs
    Given a feature "token-diet" at the plan step pinned to "planner-v1"
    When the operator runs "centinela evidence init token-diet planner"
    Then the written inputs should be exactly the two required paths
    And validating the pre-filled evidence for its snapshot rule should pass
      # init and the validator must keep sharing one function

  Scenario: A retired legacy plan role gets the same shrunken set
    Given an unpinned in-flight workflow "already-in-flight" at the plan step
    When the required plan inputs are computed for role "big-thinker"
    Then the set should contain exactly the feature's own brief and plan
    And the same should hold for role "feature-specialist"

  Scenario: The architecture docs and their scaffold mirrors state the new rule
    Given the files docs/architecture/evidence-contract.md,
      docs/architecture/planner-prompt.md and
      docs/architecture/workflow-enforcement.md
    When each file is read
    Then none of them should require snapshotting every "docs/features/*.md"
    And each should state that additional inputs are allowed
    And each should be byte-identical to its
      internal/scaffold/assets/docs/architecture/ mirror

  # ── B. UX tag matcher: state each fact once ──────────────────────────────

  Scenario: A descriptive UX edge case satisfies its required tag
    Given ux-ui-specialist evidence declaring mobileFirst true
    And its edgeCases contain "mobile-first: renders at 80x24" and a
      descriptive entry for each of the other seven required tags
    When the UX evidence is validated
    Then validation should pass
    And no required tag should be reported missing

  Scenario Outline: Tag normalization cuts at the first colon
    Given an edge case written as "<entry>"
    When the entry is normalized to a UX tag
    Then the resulting tag should be "<tag>"

    Examples:
      | entry                                     | tag           |
      | mobile-first                              | mobile-first  |
      | mobile-first: renders at 80x24            | mobile-first  |
      | loading-state: spinner: with a 3s timeout | loading-state |
      | error-state:                              | error-state   |
      | MOBILE_FIRST : Renders                    | mobile-first  |
      | Empty State: nothing yet                  | empty-state   |

  Scenario: Bare tags keep working exactly as before
    Given ux-ui-specialist evidence whose edgeCases are the eight bare
      required tags and nothing else
    When the UX evidence is validated
    Then validation should pass
      # the change strictly loosens: nothing that passed before may now fail

  Scenario: A tag that is genuinely absent is still reported missing
    Given ux-ui-specialist evidence whose edgeCases cover seven required tags
      descriptively and omit "error-state" entirely
    When the UX evidence is validated
    Then validation should fail
    And the message should name "error-state" as missing
    And it should not name any of the seven covered tags

  Scenario Outline: Degenerate entries match nothing and never panic
    Given an edge case written as "<entry>"
    When the entry is normalized to a UX tag
    Then the resulting tag should be empty
    And it should satisfy none of the required tags

    Examples:
      | entry   |
      | :       |
      | : text  |
      |         |

  Scenario: A non-UX role is unaffected by the matcher
    Given senior-engineer evidence with an arbitrary edgeCases list and no
      mobileFirst field
    When the evidence is validated
    Then the UX tag requirements should not be consulted
    And validation should not fail for missing UX tags

  # ── C. Per-prompt roadmap summary: rendered when new, quiet otherwise ─────

  Scenario: The first hook context of a session renders the roadmap summary
    Given no roadmap digest state exists
    And a hook payload carrying session id "s-1"
    When the user requests hook context
    Then the roadmap summary line should be printed
    And the digest state should record session "s-1" and the current digest

  Scenario: An unchanged roadmap is not re-rendered in the same session
    Given digest state recording session "s-1" and the current roadmap digest
    And a hook payload carrying session id "s-1"
    When the user requests hook context
    Then no roadmap summary line should be printed
    And the active-workflow panel, the feature-brief nudge, the edge-case
      nudge and the changelog nudge should be printed exactly as before

  Scenario: A changed roadmap is rendered again in the same session
    Given digest state recording session "s-1" and a stale roadmap digest
    And a hook payload carrying session id "s-1"
    When the user requests hook context
    Then the roadmap summary line should be printed
    And the digest state should be updated to the current digest

  Scenario: A new session re-renders an unchanged roadmap
    Given digest state recording session "s-1" and the current roadmap digest
    And a hook payload carrying session id "s-2"
    When the user requests hook context
    Then the roadmap summary line should be printed

  Scenario: Roadmap reformatting alone does not force a re-render
    Given digest state recording session "s-1" and the current roadmap digest
    And .workflow/roadmap.json is rewritten with different whitespace but
      identical phases, features and statuses
    And a hook payload carrying session id "s-1"
    When the user requests hook context
    Then no roadmap summary line should be printed
      # the digest hashes the projection, not the raw bytes

  Scenario Outline: Every uncertainty fails open and renders
    Given digest state that is "<state>"
    And a hook payload that is "<payload>"
    When the user requests hook context
    Then the roadmap summary line should be printed
    And the hook should exit 0

    Examples:
      | state       | payload                  |
      | absent      | valid JSON with session id |
      | corrupt     | valid JSON with session id |
      | unreadable  | valid JSON with session id |
      | current     | empty                    |
      | current     | not JSON                 |
      | current     | JSON without session_id  |

  Scenario: An unwritable state directory never breaks the host session
    Given the .workflow directory cannot be written to
    When the user requests hook context
    Then the roadmap summary line should be printed
    And the hook should exit 0
    And no error should be surfaced to the host session

  Scenario: A missing or invalid roadmap prints nothing and writes no state
    Given .workflow/roadmap.json is absent or invalid
    When the user requests hook context
    Then no roadmap summary line should be printed
    And no digest state should be written

  Scenario: The digest state file never dirties the working tree
    Given a clean git working tree
    When the user requests hook context
    Then "git status --porcelain" should report no changes
    And .workflow/.roadmap-digest should be git-ignored

  Scenario: Two worktrees keep independent digest state
    Given the same repository checked out in two worktrees
    And hook context has already rendered in the first worktree
    When the user requests hook context in the second worktree
    Then the roadmap summary line should be printed there

  # ── D. Model family aliases instead of dated pins ─────────────────────────

  Scenario: No built-in model default is a dated snapshot
    Given the built-in tier-to-model table
    When every entry is inspected
    Then no model id should end in a "-YYYYMMDD" date suffix

  Scenario Outline: The claude runner's defaults are bare family aliases
    Given the built-in tier-to-model table
    When the "<tier>" entry for runner "claude" is read
    Then the model id should be "<alias>"
    And it should contain no digits

    Examples:
      | tier      | alias  |
      | reasoning | opus   |
      | balanced  | sonnet |
      | fast      | haiku  |

  Scenario: A tier remap in config overrides the built-in alias
    Given a centinela.toml with "[orchestration.model_map]" mapping tier
      "reasoning" for runner "claude" to a specific model id
    When the model is resolved for a reasoning-tier role on runner "claude"
    Then the resolved id should be the configured override
    And resolution precedence role-override, tier-map, built-in, tier-name
      should otherwise be unchanged

  Scenario: A runner with no built-in entry still falls back to the tier name
    When the model is resolved for a reasoning-tier role on runner "codex"
    Then the resolved value should be "reasoning"
    And resolution should report not-ok
    And it should never return another runner's concrete model id

  Scenario Outline: Capability classification covers aliases and legacy pins
    Given no capability override in centinela.toml
    When the capability class is resolved for model id "<model>"
    Then the class should be "<class>"

    Examples:
      | model                       | class    |
      | opus                        | frontier |
      | sonnet                      | capable  |
      | haiku                       | limited  |
      | claude-opus-4-7             | frontier |
      | anthropic/claude-opus-4-7   | frontier |
      | claude-sonnet-4-6           | capable  |
      | anthropic/claude-sonnet-4-6 | capable  |
      | claude-haiku-4-5-20251001   | limited  |
      | anthropic/claude-haiku-4-5  | limited  |
      # the map is a strict superset: an operator whose driver_model still
      # names a retired pin keeps its class and so keeps its default profile

  Scenario: An operator's retired pin keeps its default enforcement profile
    Given a centinela.toml whose driver_model is "claude-haiku-4-5-20251001"
    When the default enforcement profile is resolved for that model
    Then it should resolve to the same profile as before this feature
    And the capability tier should remain engaged

  Scenario: The orchestration directive prints the alias, not a dated pin
    Given a feature "token-diet" at the plan step in strict orchestration mode
    When the user runs "centinela hook orchestration"
    Then the directive should name the role "planner"
    And the printed claude model should be "opus"
    And the model reference line should contain no dated snapshot id
    And the codex column should render the tier name "reasoning"

  # ── Out of scope ─────────────────────────────────────────────────────────

  Scenario: Evidence shape and scaffold size are unchanged by this feature
    Given the token-diet feature is implemented
    When the evidence contract and the scaffold assets are inspected
    Then per-role .md plus .json evidence pairs should still be required
      under the strict profile
    And internal/scaffold/assets/docs/architecture/ should still ship every
      archetype guide
      # WS3.2 (evidence-lite) and WS3.5 (scaffold slimming) are separate
      # concerns with their own back-compat questions and are not attempted here
