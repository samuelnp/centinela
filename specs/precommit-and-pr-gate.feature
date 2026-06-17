Feature: precommit and pr-gate — fire the mechanical gates at commit-time and PR-time
  As a developer who wants gate feedback before a violation lands, a governance owner, or a
  CI gate author
  I want `centinela precommit` to run the mechanical gates scoped to the STAGED changes and
  block the commit on a fail-severity violation, an installer that wires `.git/hooks/pre-commit`
  to it without clobbering an existing hook, and `centinela pr-gate` to render a deterministic
  Markdown verdict over the PR's changed files for posting as a single PR comment
  So that a violation never enters history, reviewers see gate verdicts inline, and the existing
  gate / verdict machinery is reused at the two moments developers actually want feedback

  # `centinela precommit` resolves the STAGED changeset
  # (`git diff --cached --name-only --diff-filter=ACMR`, NOT the working tree) and runs the
  # diff-aware mechanical gates scoped to it.
  #   all fail-severity gates pass → exit 0 (the commit proceeds).
  #   a fail-severity gate fails    → exit non-zero, naming the gate (the commit is BLOCKED).
  #   a warn-severity gate fails    → reported but NON-blocking → exit 0.
  # Precommit is FAST: it does not run the cross-compile/build gate by default. Outside a git
  # repo, or with nothing staged, there is nothing to gate → exit 0, no error/stack trace.
  # The installer writes `centinela precommit` into `.git/hooks/pre-commit` inside a
  # marker-delimited block, makes it executable, is idempotent (re-install = no-op, a single
  # block), preserves any pre-existing hook content, and uninstall removes ONLY its block.
  # `centinela pr-gate` runs the same gates scoped to the PR's changed files and emits a
  # deterministic Markdown verdict (each gate listed with a pass/fail marker + details) with a
  # pass/fail exit code; outside a PR context it prints the verdict to stdout and never posts.
  # `[pr_gate] fail_on_warning = true` makes a failing warn gate fail the PR gate; the default
  # does not. Custom gates and the audit-baseline gate participate exactly as in `validate`.
  # Output is deterministic: two runs over the same staged content are byte-identical with the
  # same exit code. Scenario titles map 1:1 to Go acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid centinela.toml
    And it is a git repository with an initial commit
    And the G1 oversized-file gate has a 100-line limit and severity "fail"
    And `[precommit] enabled` and `[pr_gate] enabled` are true unless a scenario states otherwise
    And the staged changeset is resolved with `git diff --cached --name-only --diff-filter=ACMR`

  # ---------------------------------------------------------------------------
  # precommit — staged scope, exit codes
  # ---------------------------------------------------------------------------

  Scenario: Staging a change that violates a fail gate blocks the commit and names the failing gate
    Given a 140-line file "internal/oversized.go" that violates the G1 fail gate
    And the file has been added with `git add internal/oversized.go`
    When the operator runs:
      centinela precommit
    Then the command exits with a non-zero code
    And the output names the failing G1 oversized-file gate
    And the output does not contain a runtime panic or stack trace

  Scenario: Staging only clean changes passes precommit and exits 0
    Given a 20-line file "internal/clean.go" that violates no fail gate
    And the file has been added with `git add internal/clean.go`
    When the operator runs:
      centinela precommit
    Then the command exits with code 0
    And the output reports no blocking gate failure

  Scenario: Unstaged working-tree changes are ignored by precommit
    Given a 20-line staged file "internal/clean.go" added with `git add internal/clean.go`
    And an additional 140-line file "internal/unstaged.go" that exists in the working tree but is NOT staged
    When the operator runs:
      centinela precommit
    Then the command exits with code 0
    And the unstaged "internal/unstaged.go" violation is not reported
    And only the staged content was gated

  Scenario: Outside a git repo or with nothing staged precommit exits 0 cleanly
    Given there is nothing in the staging index
    When the operator runs:
      centinela precommit
    Then the command exits with code 0
    And the output does not contain an error message or stack trace
    And the same clean exit 0 occurs when run outside a git repository

  Scenario: Precommit does not run the cross-compile build gate by default
    Given the cross-compile/build gate is enabled in [gates] as it is for validate
    And a clean 20-line file "internal/clean.go" has been staged
    When the operator runs:
      centinela precommit
    Then the command exits with code 0
    And the cross-compile/build gate did not run
    And precommit completes on the fast diff-aware path

  Scenario: A failing warn-severity gate under precommit is reported but does not block the commit
    Given a gate configured with severity "warn" that fails on the staged changes
    And a clean file plus the warn-triggering change have been staged
    When the operator runs:
      centinela precommit
    Then the command exits with code 0
    And the warn gate is reported as a warning rather than a blocking failure

  # ---------------------------------------------------------------------------
  # installer — idempotent, non-clobbering
  # ---------------------------------------------------------------------------

  Scenario: The installer writes an executable pre-commit hook that calls centinela precommit
    Given the repository has no ".git/hooks/pre-commit" file
    When the operator runs the centinela pre-commit hook installer
    Then the command exits with code 0
    And ".git/hooks/pre-commit" exists and is executable
    And ".git/hooks/pre-commit" invokes "centinela precommit"

  Scenario: Installing the pre-commit hook twice leaves a single centinela block
    Given the centinela pre-commit hook has already been installed
    When the operator runs the centinela pre-commit hook installer a second time
    Then the command exits with code 0
    And ".git/hooks/pre-commit" contains exactly one centinela marker block
    And the second install made no further change to the file

  Scenario: The installer preserves a pre-existing pre-commit hook and uninstall removes only its own block
    Given ".git/hooks/pre-commit" already exists with the line "echo pre-existing-hook"
    When the operator runs the centinela pre-commit hook installer
    Then the command exits with code 0
    And ".git/hooks/pre-commit" still contains "echo pre-existing-hook"
    And the centinela block is appended inside its markers
    When the operator runs the centinela pre-commit hook uninstaller
    Then the command exits with code 0
    And ".git/hooks/pre-commit" still contains "echo pre-existing-hook"
    And the centinela marker block has been removed
    And no other line of the original hook was modified

  # ---------------------------------------------------------------------------
  # pr-gate — Markdown verdict, exit codes
  # ---------------------------------------------------------------------------

  Scenario: pr-gate emits a Markdown verdict listing each gate with a pass/fail marker and details
    Given the PR's changed files include a 140-line file that violates the G1 fail gate
    When the operator runs:
      centinela pr-gate
    Then the command exits with a non-zero code
    And the output is Markdown listing each participating gate with a pass or fail marker
    And the failing G1 gate is shown as failing with its violation details
    And the output does not contain a runtime panic or stack trace

  Scenario: pr-gate over an all-passing changeset exits 0 with a Markdown all-pass verdict
    Given the PR's changed files are all clean and violate no fail gate
    When the operator runs:
      centinela pr-gate
    Then the command exits with code 0
    And the Markdown verdict reports every participating gate as passing

  Scenario: pr-gate run outside a PR context prints the verdict to stdout and does not post or error
    Given no GitHub PR environment variables are set
    When the operator runs:
      centinela pr-gate
    Then the command exits with code 0 when the changeset is clean
    And the Markdown verdict is printed to stdout
    And nothing is posted to GitHub
    And the output does not contain an error message or stack trace

  Scenario: fail_on_warning makes a failing warn gate fail the PR gate while the default does not
    Given a gate configured with severity "warn" that fails on the PR's changed files
    When the operator runs centinela pr-gate with "[pr_gate] fail_on_warning" false
    Then the command exits with code 0
    And the warn gate is reported as a warning in the Markdown verdict
    When the operator runs centinela pr-gate with "[pr_gate] fail_on_warning" true
    Then the command exits with a non-zero code
    And the warn gate failure is reported as blocking the PR gate

  # ---------------------------------------------------------------------------
  # Shared gate participation — custom + audit-baseline gates
  # ---------------------------------------------------------------------------

  Scenario: Custom gates and the audit-baseline gate participate in precommit and pr-gate like in validate
    Given a `[[gates.custom]]` gate named "no-todo" with command "false" and severity "fail"
    And the audit-baseline gate is enabled
    And a change has been staged
    When the operator runs:
      centinela precommit
    Then the command exits with a non-zero code
    And the gate report contains a gate named "no-todo" reported as failing
    When the operator runs:
      centinela pr-gate
    Then the "no-todo" custom gate appears in the Markdown verdict
    And the audit-baseline gate participates through the same Result path as in validate

  # ---------------------------------------------------------------------------
  # Determinism
  # ---------------------------------------------------------------------------

  Scenario: Two runs over the same staged content produce identical verdict output and exit code
    Given a fixed staged changeset containing a 140-line G1-violating file
    When the operator runs centinela pr-gate twice in succession over the same content
    Then both runs produce byte-identical Markdown verdict output
    And both runs exit with the same non-zero code
    And the gates are listed in the same deterministic order in both runs
