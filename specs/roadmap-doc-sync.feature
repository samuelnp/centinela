Feature: Roadmap doc sync — generate ROADMAP.md from roadmap.json with drift gate
  As a maintainer or agent working in a Centinela-governed project
  I want `centinela roadmap generate` to write ROADMAP.md deterministically from
  roadmap.json and `centinela validate` to fail when the on-disk file diverges
  So that the two representations can never silently drift apart

  # ROADMAP.md is generated; it is never hand-edited. roadmap.json is the single
  # source of truth. The drift gate byte-compares the on-disk file against what
  # the generator would produce. Severity is configurable: "warn" (non-blocking,
  # safe adoption default) or "fail" (blocking). Scenario titles map 1:1 to Go
  # acceptance tests (// Scenario: <name>).

  Background:
    Given a Centinela-governed project with a valid centinela.toml
    And roadmap.json exists with at least one phase containing at least one feature

  # ---------------------------------------------------------------------------
  # generate command — happy path
  # ---------------------------------------------------------------------------

  Scenario: generate writes ROADMAP.md and exits 0
    Given roadmap.json contains an intro, a phase with a note, and a feature with description and fixes
    When the operator runs:
      centinela roadmap generate
    Then the command exits with code 0
    And ROADMAP.md exists on disk
    And the output reports the path written

  Scenario: Generated output is deterministic — running generate twice yields byte-identical files
    Given roadmap.json is populated with prose fields
    When the operator runs centinela roadmap generate twice in succession
    Then the two output files are byte-identical

  # ---------------------------------------------------------------------------
  # Canonical format — prose fidelity round-trip
  # ---------------------------------------------------------------------------

  Scenario: Top-level intro blockquote round-trips from roadmap.json into ROADMAP.md
    Given roadmap.json has an "intro" field containing a multi-line string
    When the operator runs centinela roadmap generate
    Then ROADMAP.md begins with "# Roadmap" followed by a blank line
    And each line of the intro is prefixed with "> "
    And blank lines within the intro are emitted as a bare ">"

  Scenario: Per-phase note renders as a blockquote preceding the feature list
    Given roadmap.json has a phase whose "note" field contains two paragraphs separated by a blank line
    When the operator runs centinela roadmap generate
    Then the note appears after the phase heading and before the first feature bullet
    And each note line is prefixed with "> "
    And the blank line between paragraphs is emitted as a bare ">" keeping the blockquote unbroken

  Scenario: A phase with no note renders heading and features with no blockquote
    Given roadmap.json has a phase with an empty or absent "note" field
    When the operator runs centinela roadmap generate
    Then the phase heading is immediately followed by the feature list with no blockquote in between

  Scenario: Feature with description and fixes renders both fields
    Given a feature in roadmap.json has a non-empty "description" and a non-empty "fixes" field
    When the operator runs centinela roadmap generate
    Then the feature bullet reads "- **<name>** — <description>"
    And the following line reads "  *Fixes: <fixes>*"

  Scenario: Feature with description only renders no Fixes line
    Given a feature in roadmap.json has a non-empty "description" and no "fixes" field
    When the operator runs centinela roadmap generate
    Then the feature bullet includes the description after the em-dash
    And no "*Fixes:*" line is emitted for that feature

  Scenario: Feature with fixes only renders no em-dash clause on the bullet line
    Given a feature in roadmap.json has no "description" field and a non-empty "fixes" field
    When the operator runs centinela roadmap generate
    Then the feature bullet reads "- **<name>**" with no em-dash clause
    And the following line reads "  *Fixes: <fixes>*"

  Scenario: Feature with no description and no fixes renders as a bare bullet with no dangling em-dash
    Given a feature in roadmap.json has neither "description" nor "fixes" fields
    When the operator runs centinela roadmap generate
    Then the feature bullet reads exactly "- **<name>**"
    And no em-dash is emitted on the bullet line
    And no "*Fixes:*" line is emitted for that feature

  Scenario: Feature with dependsOn renders dependency annotation in declared slice order
    Given a feature in roadmap.json has a non-empty "dependsOn" list ["feat-a", "feat-b"]
    When the operator runs centinela roadmap generate
    Then the feature line contains "(depends on feat-a, feat-b)" in that order
    And the annotation appears after the description when description is present

  Scenario: Feature with dependsOn but no description attaches the annotation directly to the bullet
    Given a feature in roadmap.json has no "description" and "dependsOn" set to ["feat-a"]
    When the operator runs centinela roadmap generate
    Then the feature bullet reads "- **<name>** (depends on feat-a)"
    And no em-dash is emitted

  Scenario: Feature with empty dependsOn emits no dependency annotation
    Given a feature in roadmap.json has "dependsOn" set to an empty list
    When the operator runs centinela roadmap generate
    Then no "(depends on" text appears on the feature line

  # ---------------------------------------------------------------------------
  # Backlog phase rendering
  # ---------------------------------------------------------------------------

  Scenario: Backlog phase features render using deferred-finding format
    Given roadmap.json contains a Backlog phase with one deferred finding carrying summary, source, and deferredAt
    When the operator runs centinela roadmap generate
    Then the Backlog feature bullet reads:
      "- **<name>** — <summary> *(deferred <deferredAt> · <source.feature>/<source.role>)*"
    And the Backlog feature does not emit a "*Fixes:*" line
    And the Backlog feature does not emit a "dependsOn" annotation

  Scenario: Backlog feature with empty source fields omits the empty parenthetical
    Given roadmap.json contains a Backlog phase feature with a summary but no source and no deferredAt
    When the operator runs centinela roadmap generate
    Then the Backlog feature bullet contains the summary
    And no empty "()" or "· /" string appears on that line

  # ---------------------------------------------------------------------------
  # No live status in the generated file
  # ---------------------------------------------------------------------------

  Scenario: Generated ROADMAP.md contains no per-feature live status glyph
    Given roadmap.json contains features across multiple phases
    When the operator runs centinela roadmap generate
    Then no feature bullet line in ROADMAP.md begins with or contains a live status marker such as "✓" or "✅"

  Scenario: Phase heading status glyphs authored in the phase name are preserved verbatim
    Given roadmap.json has a phase whose "name" field reads "✅ Phase 0: Bootstrap"
    When the operator runs centinela roadmap generate
    Then the phase heading in ROADMAP.md reads "## ✅ Phase 0: Bootstrap"

  # ---------------------------------------------------------------------------
  # EOF and whitespace contract
  # ---------------------------------------------------------------------------

  Scenario: Generated file ends with exactly one trailing newline and no trailing whitespace
    Given any valid roadmap.json
    When the operator runs centinela roadmap generate
    Then ROADMAP.md ends with exactly one newline character
    And no line in the generated file has trailing whitespace

  # ---------------------------------------------------------------------------
  # Drift gate — in-sync pass
  # ---------------------------------------------------------------------------

  Scenario: Drift gate passes when ROADMAP.md matches generator output
    Given centinela.toml has "[gates.roadmap_drift]" with "enabled = true" and any valid severity
    And ROADMAP.md on disk is byte-identical to what centinela roadmap generate would produce
    When the operator runs centinela validate
    Then the gate result named "roadmap_drift" is "Pass"
    And the output reports "ROADMAP.md is in sync"

  # ---------------------------------------------------------------------------
  # Drift gate — mismatch detection
  # ---------------------------------------------------------------------------

  Scenario: Drift gate fails when ROADMAP.md is hand-edited under severity fail
    Given centinela.toml has "[gates.roadmap_drift]" with "enabled = true" and "severity = \"fail\""
    And ROADMAP.md has been hand-edited to differ from what centinela roadmap generate would produce
    When the operator runs centinela validate
    Then the gate result named "roadmap_drift" is "Fail"
    And the detail names the first line number where the files diverge
    And the detail instructs the operator to run "centinela roadmap generate"
    And the exit code is non-zero

  Scenario: Drift gate warns but does not block when ROADMAP.md drifts under severity warn
    Given centinela.toml has "[gates.roadmap_drift]" with "enabled = true" and "severity = \"warn\""
    And ROADMAP.md has been hand-edited to differ from generator output
    When the operator runs centinela validate
    Then the gate result named "roadmap_drift" is "Warn"
    And the detail names the first differing line number
    And the detail instructs the operator to run "centinela roadmap generate"
    And the exit code is 0

  Scenario: Running generate after a drift failure then re-validating passes the gate
    Given centinela.toml has "[gates.roadmap_drift]" with "enabled = true" and "severity = \"fail\""
    And ROADMAP.md has drifted from roadmap.json
    And the drift gate previously returned "Fail"
    When the operator runs centinela roadmap generate
    And then runs centinela validate
    Then the gate result named "roadmap_drift" is "Pass"
    And the exit code is 0

  # ---------------------------------------------------------------------------
  # Drift gate — missing ROADMAP.md
  # ---------------------------------------------------------------------------

  Scenario: Missing ROADMAP.md is reported as a clear failure under severity fail
    Given centinela.toml has "[gates.roadmap_drift]" with "enabled = true" and "severity = \"fail\""
    And ROADMAP.md does not exist on disk
    When the operator runs centinela validate
    Then the gate result named "roadmap_drift" is "Fail"
    And the detail states that ROADMAP.md is missing
    And the detail instructs the operator to run "centinela roadmap generate"
    And the process does not panic or emit a raw I/O error

  Scenario: Missing ROADMAP.md produces a warn result under severity warn
    Given centinela.toml has "[gates.roadmap_drift]" with "enabled = true" and "severity = \"warn\""
    And ROADMAP.md does not exist on disk
    When the operator runs centinela validate
    Then the gate result named "roadmap_drift" is "Warn"
    And the detail states that ROADMAP.md is missing
    And the detail instructs the operator to run "centinela roadmap generate"

  Scenario: generate creates ROADMAP.md from scratch when the file is absent
    Given ROADMAP.md does not exist on disk
    And roadmap.json exists and is valid
    When the operator runs centinela roadmap generate
    Then the command exits with code 0
    And ROADMAP.md is created on disk

  # ---------------------------------------------------------------------------
  # Config validation
  # ---------------------------------------------------------------------------

  Scenario: An unknown severity value is rejected at config load
    Given centinela.toml sets "[gates.roadmap_drift]" severity to an unsupported value such as "error"
    When the configuration is loaded
    Then loading fails with an error naming the severity field
    And the error identifies the valid values "fail" and "warn"

  Scenario: Unknown severity is a no-op when the gate is disabled
    Given centinela.toml sets "[gates.roadmap_drift]" with "enabled = false" and "severity = \"bad\""
    When the configuration is loaded
    Then loading succeeds without error

  Scenario: The drift gate is registered and ships with enabled true and severity warn
    Given Centinela's own centinela.toml
    When the configured gates are read
    Then the roadmap_drift gate is enabled
    And its severity is "warn"

  # ---------------------------------------------------------------------------
  # Gate disabled
  # ---------------------------------------------------------------------------

  Scenario: Gate disabled skips the check even when ROADMAP.md is absent
    Given centinela.toml has "[gates.roadmap_drift]" with "enabled = false"
    And ROADMAP.md does not exist on disk
    When the operator runs centinela validate
    Then no "roadmap_drift" result appears in the output
    And the exit code is 0

  # ---------------------------------------------------------------------------
  # Edge cases — phase with zero features
  # ---------------------------------------------------------------------------

  Scenario: A phase with no features renders only its heading and optional note
    Given roadmap.json has a phase whose "features" array is empty
    When the operator runs centinela roadmap generate
    Then the phase heading appears in ROADMAP.md
    And no feature bullet follows the heading for that phase
    And the file is still valid with exactly one trailing newline

  # ---------------------------------------------------------------------------
  # Edge cases — non-ASCII and special characters
  # ---------------------------------------------------------------------------

  Scenario: Non-ASCII characters in prose fields are passed through byte-for-byte
    Given a feature description contains em-dashes, curly quotes, and accented characters
    When the operator runs centinela roadmap generate
    Then those characters appear unchanged in ROADMAP.md

  # ---------------------------------------------------------------------------
  # Edge cases — line ending contract
  # ---------------------------------------------------------------------------

  Scenario: Generated file uses LF line endings on all platforms
    Given a valid roadmap.json
    When the operator runs centinela roadmap generate on any supported platform
    Then every line ending in ROADMAP.md is LF only and no CRLF sequences are present
