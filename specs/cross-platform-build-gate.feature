Feature: Cross-platform build gate
  As a maintainer
  I want centinela validate to cross-compile every configured release target
  So that platform-specific build breaks are caught locally before the release pipeline

  Background:
    Given a centinela.toml with "[gates] build = true"
    And "[gates.build] command = \"go build ./cmd/centinela\""
    And "[gates.build] targets" lists the six release targets:
      linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64

  # ── Happy path ────────────────────────────────────────────────────────────

  Scenario: All targets compile — gate passes and validate proceeds
    Given the codebase compiles cleanly on all six configured targets
    When the user runs "centinela validate"
    Then the G-Build gate should report Pass
    And the pass message should be "All 6 release targets compile."
    And centinela validate should exit with status 0

  # ── Negative paths ────────────────────────────────────────────────────────

  Scenario: One target fails to build — gate fails naming the broken GOOS/GOARCH
    Given the codebase contains a Unix-only syscall (e.g. syscall.Flock)
    When the user runs "centinela validate"
    Then the G-Build gate should report Fail
    And the failure message should begin with "These release targets failed to build:"
    And the details list should include "windows/amd64"
    And the details list should include "windows/arm64"
    And the details list should NOT include "linux/amd64"
    And the details list should NOT include "linux/arm64"
    And the details list should NOT include "darwin/amd64"
    And the details list should NOT include "darwin/arm64"
    And centinela validate should exit with a non-zero status

  Scenario: Gate disabled — build check is skipped
    Given centinela.toml sets "[gates] build = false"
    When the user runs "centinela validate"
    Then the G-Build gate should NOT appear in the gate results
    And centinela validate should exit with status 0

  Scenario: Target list in centinela.toml drifts from release.yml matrix — parity test fails
    Given the centinela.toml targets list contains only 4 targets (two removed)
    And ".github/workflows/release.yml" still declares the full {linux,darwin,windows} x {amd64,arm64} matrix
    When "go test ./..." is executed (as part of centinela validate)
    Then the parity test "TestBuildMatrixParity" should fail
    And the failure message should name the targets present in release.yml but absent in centinela.toml

  # ── Edge cases ────────────────────────────────────────────────────────────

  Scenario: CGO disabled — cross-compile succeeds without a C toolchain
    Given the build gate is enabled with the six release targets
    And no C compiler is available on the host
    When the gate runs each cross-compile
    Then each "go build" invocation is executed with "CGO_ENABLED=0"
    And the gate completes without a C-toolchain error

  Scenario: Unknown or garbage target reported cleanly without panic
    Given "[gates.build] targets" includes an invalid entry { goos = "obscureos", goarch = "noarch" }
    When the user runs "centinela validate"
    Then the G-Build gate should report Fail
    And the details list should include "obscureos/noarch" with a descriptive error message
    And centinela should NOT panic or produce a stack trace

  Scenario: Empty targets list — gate no-ops with a clear message
    Given "[gates.build] targets" is an empty array
    And "[gates] build = true"
    When the user runs "centinela validate"
    Then the G-Build gate should report Pass
    And the pass message should indicate that no targets are configured (e.g. "No targets configured; skipping cross-compile.")

  Scenario: Build cache reused on second run — gate is fast
    Given centinela validate was already run successfully (warm build cache)
    When the user runs "centinela validate" a second time without changing any source file
    Then each cross-compile should complete in well under 1 second per target
    And the gate should still report Pass

  Scenario: Acceptance — simulate broken target via unbuildable command
    Given the build gate is configured with a custom command that exits non-zero for "linux/amd64"
      (e.g. command = "sh -c 'if [ \"$GOOS\" = linux ] && [ \"$GOARCH\" = amd64 ]; then exit 1; fi; go build ./cmd/centinela'")
    And all other five targets would succeed under the real go build command
    When the user runs "centinela validate"
    Then the G-Build gate should report Fail
    And the details list should include "linux/amd64"
    And the details list should NOT include the other five targets
    And centinela validate should exit with a non-zero status
