Feature: deep codebase analysis — deterministic, read-only repo inventory
  As a team adopting Centinela on a mature (brownfield) codebase, a downstream
  Phase 9 feature author, or a CI integrator
  I want `centinela analyze` to mechanically scan the current repo and emit a
  machine-readable Inventory (languages, manifests/build-test signals, i18n
  locales, package layout, dependency graph) to a stable well-known path plus a
  concise human summary
  So that brownfield onboarding has a single, trustworthy, deterministic
  description of "what this repo actually is" without any LLM call or interview

  # `centinela analyze` walks the project root (read-only) with a fixed skip set
  # (vendor/, node_modules/, .git/, .workflow/, dist/, build/, gitignored paths),
  # counts source files by extension → languages + primaryLanguage, detects
  # manifests (go.mod, package.json+scripts, Gemfile, Cargo.toml,
  # pyproject.toml/requirements.txt, Makefile) and extracts build/test/framework
  # signals + declared deps, detects i18n locales, builds a depth-bounded package
  # layout, and assembles a dependency graph (Go: real `go list -json` package
  # edges via internal/golist; other ecosystems: declared manifest dep names).
  # The result is a schemaVersion-tagged Inventory written deterministically
  # (every list sorted, MarshalIndent + trailing newline → byte-stable re-runs)
  # to .workflow/analysis.json (overridable with --out), and a summary is printed
  # to stdout. Analyze is DIAGNOSTIC, never a gate: any sub-detector failure
  # degrades to a best-effort/empty result with a recorded reason and the command
  # still exits 0. The ONLY hard error is an un-writable output path (or being run
  # outside a usable directory). Scenario titles map 1:1 to Go acceptance tests
  # (// Scenario: <name>) under tests/acceptance/.

  Background:
    Given a project directory that is the analysis root
    And the well-known output path is ".workflow/analysis.json"
    And the analyze command performs a read-only scan that never mutates source files

  # ---------------------------------------------------------------------------
  # AC-1/2/4 — Happy path: analyze a Go repo
  # ---------------------------------------------------------------------------

  Scenario: Analyzing a Go module writes a complete inventory and prints a summary
    Given a Go module whose go.mod declares a module path
    And the module contains Go source files and a Makefile with a "go test" target
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the file ".workflow/analysis.json" is written
    And the inventory "schemaVersion" equals 1
    And the inventory "primaryLanguage" equals "Go"
    And the "languages" list contains "Go" with a positive "fileCount"
    And the "manifests" list contains a manifest of kind "go-mod" whose path is "go.mod"
    And that go-mod manifest records the declared module path
    And the "manifests" list contains a manifest of kind "make" with a "test" signal
    And the "graph" has kind "go-packages" with a non-empty "edges" list
    And the "graph" records the module path
    And stdout reports the primary language, the build/test signal, the locale count, the package count, and the graph edge count

  # ---------------------------------------------------------------------------
  # AC-3 — Determinism: byte-identical re-run
  # ---------------------------------------------------------------------------

  Scenario: Re-running analyze on an unchanged repo produces a byte-identical inventory
    Given analyze has been run once on an unchanged repo and ".workflow/analysis.json" exists
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the newly written ".workflow/analysis.json" is byte-identical to the previous run
    And every list in the inventory (languages, manifests, locales, packages, graph edges) is in a stable sorted order

  # ---------------------------------------------------------------------------
  # Polyglot / non-Go manifest detection (package.json)
  # ---------------------------------------------------------------------------

  Scenario: Analyzing a Node project detects the npm manifest with build and test scripts and declared deps
    Given a project containing a valid package.json with "build" and "test" scripts and declared dependencies
    And the project also contains JavaScript source files
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the "manifests" list contains a manifest of kind "npm" whose path is "package.json"
    And that npm manifest records the "build" script as its build signal
    And that npm manifest records the "test" script as its test signal
    And that npm manifest lists the declared dependency names, sorted
    And the "languages" list counts the JavaScript source files

  Scenario: A polyglot repo counts every language and picks the highest-count primary deterministically
    Given a repo containing Go, JavaScript, and Ruby source files in differing counts
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the "languages" list contains an entry for Go, JavaScript, and Ruby
    And "primaryLanguage" equals the language with the highest file count
    And languages with equal counts are ordered alphabetically as a deterministic tiebreak
    And every detected manifest across ecosystems is listed even though only one language is the headline

  # ---------------------------------------------------------------------------
  # i18n locale detection
  # ---------------------------------------------------------------------------

  Scenario: Analyzing a repo with locale files lists the detected locale codes
    Given a project with a "locales/" directory containing "en" and "es" locale files
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the "locales" list contains "en" and "es" in sorted order
    And stdout reports a locale count of at least 2

  Scenario: A repo with no i18n reports an empty locale list and exit 0
    Given a project that contains no locale files or i18n directories
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the "locales" list is empty
    And stdout reports a locale count of 0
    And the absence of locales is not treated as an error

  # ---------------------------------------------------------------------------
  # AC-5/6 — Skip set and read-only guarantee
  # ---------------------------------------------------------------------------

  Scenario: The scan skips dependency and build directories so counts reflect real source
    Given a repo whose vendor/, node_modules/, .git/, and .workflow/ directories contain many files
    And a path listed in .gitignore
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the language file counts exclude every file under vendor/, node_modules/, .git/, and .workflow/
    And the language file counts exclude the gitignored path
    And no file other than ".workflow/analysis.json" is created or modified on disk

  # ---------------------------------------------------------------------------
  # AC-7 — Best-effort: unfamiliar / no-manifest repo still produces a valid inventory
  # ---------------------------------------------------------------------------

  Scenario: A repo with no recognized manifest still produces a valid inventory and exits 0
    Given a project that contains source files but no recognized manifest file
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the file ".workflow/analysis.json" is written
    And the "languages" and "packages" sections are populated
    And the "manifests" list is empty
    And the "graph" kind is "none" or an empty best-effort graph
    And analyze does not hard-fail on the unfamiliar repo

  Scenario: A malformed package.json is recorded as detected-but-unparsable and the scan continues
    Given a project containing a package.json whose contents are not valid JSON
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the file ".workflow/analysis.json" is written
    And the npm manifest is recorded as detected without parsed build/test/deps signals
    And the rest of the inventory (languages, locales, layout) is still populated

  Scenario: When go list fails the Go graph is recorded as best-effort empty with a note and the rest still emits
    Given a repo whose Go code does not compile or has no working Go toolchain
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the file ".workflow/analysis.json" is written
    And the "graph" has an empty "edges" list
    And the "graph" carries a "note" describing why the Go graph is empty
    And the languages and manifests sections are still populated

  # ---------------------------------------------------------------------------
  # NEGATIVE / edge — unusable directory is a hard error, no partial artifact
  # ---------------------------------------------------------------------------

  Scenario: Running analyze with an un-writable output path fails clearly with a non-zero exit and writes no partial inventory
    Given a repo whose ".workflow/analysis.json" destination cannot be written (the output directory is not writable)
    When the operator runs:
      centinela analyze
    Then the command exits with a non-zero code
    And stderr contains a clear error message about the un-writable output path
    And no partial or corrupt ".workflow/analysis.json" is left on disk

  Scenario: Running analyze against a non-existent or unreadable root fails clearly and writes no inventory
    Given the operator points analyze at a root path that does not exist or cannot be read
    When the operator runs:
      centinela analyze --out .workflow/analysis.json
    Then the command exits with a non-zero code
    And stderr contains a clear error message naming the unreadable root
    And no ".workflow/analysis.json" is written

  # ---------------------------------------------------------------------------
  # --out override and overwrite behavior
  # ---------------------------------------------------------------------------

  Scenario: The --out flag redirects the inventory to a custom path
    Given a Go module at the analysis root
    When the operator runs:
      centinela analyze --out build/inventory.json
    Then the command exits with code 0
    And the inventory is written to "build/inventory.json"
    And the default ".workflow/analysis.json" is not created by this run

  Scenario: An empty or docs-only repo yields a valid empty inventory and exits 0
    Given a project containing only documentation files and no source code or manifests
    When the operator runs:
      centinela analyze
    Then the command exits with code 0
    And the file ".workflow/analysis.json" is written
    And "primaryLanguage" is the empty string
    And the "manifests" and "graph" sections are empty
    And analyze does not treat the empty repo as an error
