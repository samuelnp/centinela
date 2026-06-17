Feature: Security Gate
  As a maintainer or agent working in a Centinela-governed project
  I want `centinela validate` to mechanically enforce secret-scanning and
  dependency-vulnerability auditing via a `[gates.security]` gate
  So that leaked secrets fail validation (hard block) and vulnerable
  dependencies surface as warnings, catching both at validate-time instead
  of at audit-time or not at all

  Background:
    Given the project has a `centinela.toml` at its root
    And the file contains `[gates.security]` with `enabled = true`

  # ---------------------------------------------------------------------------
  # AC1 — Secret detected → G-Secrets Fail, blocks validate
  # ---------------------------------------------------------------------------

  Scenario: A detectable secret is present in a tracked file — secrets gate fails
    Given `gitleaks` is installed and reachable via PATH
    And the project contains a file with a string that gitleaks identifies as a secret
      (e.g. an API key matching a built-in rule)
    And the diff-aware filter includes that file (or is nil / CI full-scan mode)
    When `centinela validate` runs
    Then the gate result named `G-Secrets` has status `Fail`
    And the `G-Secrets` details list names the file and the matched rule ID
    And `AllPassed` is false
    And the exit code is 1

  # ---------------------------------------------------------------------------
  # AC2 — No secrets found → G-Secrets Pass
  # ---------------------------------------------------------------------------

  Scenario: No secrets detected — secrets gate passes
    Given `gitleaks` is installed and reachable via PATH
    And none of the scanned files contain a string matching any gitleaks rule
    When `centinela validate` runs
    Then the gate result named `G-Secrets` has status `Pass`
    And `AllPassed` is not set to false by the secrets gate
    And the exit code reflects only other gates' results

  # ---------------------------------------------------------------------------
  # AC3 — Vulnerable dependency found → G-Vuln Warn, does NOT block validate
  # ---------------------------------------------------------------------------

  Scenario: A dependency with a known CVE is present — vuln gate warns but does not block
    Given at least one vuln tool from `gates.security.vuln.tools` is installed
    And the project dependency set contains a package with a known CVE
    When `centinela validate` runs
    Then the gate result named `G-Vuln` has status `Warn`
    And the `G-Vuln` details list names the affected package and vulnerability ID
    And `AllPassed` is true (the Warn does not block validate)
    And the exit code is 0 (assuming no other gates fail)

  # ---------------------------------------------------------------------------
  # AC4 — gitleaks absent → G-Secrets Skip, no crash, not Fail
  # ---------------------------------------------------------------------------

  Scenario: gitleaks is not installed — secrets gate skips with a clear message
    Given `gitleaks` is NOT installed (exec.LookPath returns ErrNotFound)
    And `gates.security.secrets.tool` is set to `"gitleaks"`
    When `centinela validate` runs
    Then the gate result named `G-Secrets` has status `Skip`
    And the `G-Secrets` message names "gitleaks" as the missing tool
    And `centinela validate` does not crash or exit with an unexpected error
    And the `G-Secrets` status is NOT `Fail`

  # ---------------------------------------------------------------------------
  # AC5 — Gate disabled or absent → no security result produced (zero-config-safe)
  # ---------------------------------------------------------------------------

  Scenario: Security gate is disabled — no security results emitted
    Given `centinela.toml` contains `[gates.security]` with `enabled = false`
    When `centinela validate` runs
    Then no gate result named `G-Secrets` appears in the output
    And no gate result named `G-Vuln` appears in the output
    And existing non-security gates are unaffected

  Scenario: No `[gates.security]` block present — no security results emitted
    Given `centinela.toml` does not contain a `[gates.security]` block
    When `centinela validate` runs
    Then no gate result named `G-Secrets` appears in the output
    And no gate result named `G-Vuln` appears in the output
    And the exit code reflects only the other gates' results

  # ---------------------------------------------------------------------------
  # AC6 — Secret match is in secrets.allowlist → excluded → Pass
  # ---------------------------------------------------------------------------

  Scenario: The only secret finding matches an allowlist entry — secrets gate passes
    Given `gitleaks` is installed and reachable via PATH
    And the project file contains a string that gitleaks identifies as a secret
    And `gates.security.secrets.allowlist` contains the matching gitleaks rule ID
      (or a path glob that matches the file containing the secret)
    When `centinela validate` runs
    Then the allowlisted finding is excluded from the results
    And the gate result named `G-Secrets` has status `Pass`
    And `AllPassed` is not set to false by the secrets gate

  # ---------------------------------------------------------------------------
  # AC7 — Diff-aware filter: secret in unchanged file → locally filtered, not scanned
  # ---------------------------------------------------------------------------

  Scenario: Diff-aware filter is active and the secret is in an unchanged file — locally filtered
    Given `gitleaks` is installed and reachable via PATH
    And the diff-aware filter is active (local validate, non-nil gitdiff.Set)
    And the only secret-bearing file is NOT in the current git diff set
    When `centinela validate` runs locally
    Then the secrets scan is scoped to the diff set only (matching G1/G11 behavior)
    And the gate result named `G-Secrets` has status `Pass` (no diff-scoped findings)
    And the out-of-scope file is not scanned in the local run
    And the documentation notes that CI full-scan (nil filter) would still catch it

  # ---------------------------------------------------------------------------
  # Edge Case — Both scanner families absent → two Skips + "no scanners" signal
  # ---------------------------------------------------------------------------

  Scenario: Both gitleaks and all vuln tools are absent — two Skips with distinct signal
    Given `gitleaks` is NOT installed
    And none of the tools listed in `gates.security.vuln.tools` are installed
    When `centinela validate` runs
    Then the gate result named `G-Secrets` has status `Skip`
    And the gate result named `G-Vuln` has status `Skip`
    And the combined output contains a message indicating "no scanners available"
      (distinct from a clean pass — the signal must not be read as a verified clean scan)
    And `AllPassed` is true (Skips do not block validate)

  # ---------------------------------------------------------------------------
  # Edge Case — Malformed / empty JSON output → Warn or Skip, never false Pass
  # ---------------------------------------------------------------------------

  Scenario: gitleaks returns malformed JSON — gate reports parse error as Warn, not Pass
    Given `gitleaks` is installed and reachable via PATH
    And `gitleaks` exits with a non-zero code and produces output that is not valid JSON
    When `centinela validate` runs
    Then the gate result named `G-Secrets` has status `Warn`
    And the `G-Secrets` message describes a parse or tool error
    And the status is NOT `Pass` (parse failure must never yield a false clean result)

  Scenario: A vuln tool returns malformed JSON — gate reports parse error as Warn, not Pass
    Given at least one tool from `gates.security.vuln.tools` is installed
    And that tool exits and produces output that is not valid JSON
    When `centinela validate` runs
    Then the gate result named `G-Vuln` has status `Warn`
    And the `G-Vuln` message describes the parse or tool error
    And the status is NOT `Pass`

  # ---------------------------------------------------------------------------
  # Edge Case — Duplicate CVE from both vuln tools → de-duplicated in Details
  # ---------------------------------------------------------------------------

  Scenario: Both govulncheck and osv-scanner report the same CVE for the same package — de-duped
    Given both `govulncheck` and `osv-scanner` are installed
    And both tools independently report the same vulnerability ID for the same package
    When `centinela validate` runs
    Then the gate result named `G-Vuln` has status `Warn`
    And the `G-Vuln` details list contains exactly one entry for that (package, vuln-id) pair
    And the duplicate is not surfaced as a second entry

  # ---------------------------------------------------------------------------
  # Edge Case — No dependency manifest for a vuln tool → Skip not Fail
  # ---------------------------------------------------------------------------

  Scenario: osv-scanner finds no lockfile or manifest — skips without failing
    Given `osv-scanner` is installed and reachable via PATH
    And the project contains no lockfile or manifest that osv-scanner can scan
    When `centinela validate` runs
    Then the gate result named `G-Vuln` has status `Skip` for the osv-scanner check
    And the skip message indicates no scannable manifest was found
    And the status is NOT `Fail`

  # ---------------------------------------------------------------------------
  # Edge Case — Allowlist entry that matches nothing is silently ignored
  # ---------------------------------------------------------------------------

  Scenario: An allowlist entry matches no finding — ignored, not an error
    Given `gitleaks` is installed and reachable via PATH
    And `gates.security.secrets.allowlist` contains a rule ID that matches no finding
    And the project contains no real secrets
    When `centinela validate` runs
    Then the gate result named `G-Secrets` has status `Pass`
    And no error or warning is emitted about the unmatched allowlist entry
