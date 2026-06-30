Feature: centinela update — self-update and passive startup notice
  As a Centinela user running an installed release binary
  I want `centinela update` to fetch, verify, and install the latest release for my platform
  And `centinela update --check` to tell me whether I am behind without changing anything
  And a quiet throttled startup notice when a newer version exists
  So that I stay current without manual download steps and without surprises

  # Asset name contract (verified against v0.40.2):
  #   centinela-<tag>-<goos>-<goarch>[.exe]
  #   where <tag> carries the leading v (e.g. centinela-v0.40.2-darwin-arm64)
  # SHA256SUMS format: <64-hex><two-spaces><filename> per line (coreutils sha256sum)
  # Running version: ldflag-injected without the v prefix (e.g. "0.40.2")
  # Release tag_name: v-prefixed (e.g. "v0.40.2")
  # Comparison: strip leading v from both sides before comparing
  # Dev build sentinel: "dev" — treated as uncomparable, not upgradeable
  # HTTP is behind an injected Doer interface for deterministic tests
  # All tests use httptest.Server + temp HOME/XDG dir — no real network, no real GitHub

  Background:
    Given an httptest.Server serving the GitHub Releases API and asset downloads
    And a temp HOME/XDG dir used for all cache reads and writes
    And the running binary version is controlled by the test harness

  # ---------------------------------------------------------------------------
  # AC1 — Update happy path
  # ---------------------------------------------------------------------------

  Scenario: Update installs a newer release and prints old and new versions
    Given the running binary version is "0.37.0"
    And the fake GitHub API returns tag "v0.40.2" with an asset matching the host platform
    And the fake SHA256SUMS contains the correct checksum for the asset
    And the install directory is writable
    When the user runs:
      centinela update
    Then the asset matching the host GOOS/GOARCH is downloaded
    And the downloaded asset's SHA256 matches the SHA256SUMS entry
    And the binary is replaced atomically via a temp file in the same directory
    And the output contains "0.37.0 -> 0.40.2"
    And the command exits with code 0

  Scenario: Update is a no-op when already on the latest version
    Given the running binary version is "0.40.2"
    And the fake GitHub API returns tag "v0.40.2"
    When the user runs:
      centinela update
    Then no binary download occurs
    And the output contains "already up to date"
    And no file on disk is modified
    And the command exits with code 0

  # ---------------------------------------------------------------------------
  # AC2 — --check is read-only
  # ---------------------------------------------------------------------------

  Scenario: --check reports a newer version and exits non-zero with zero writes
    Given the running binary version is "0.37.0"
    And the fake GitHub API returns tag "v0.40.2"
    When the user runs:
      centinela update --check
    Then the output contains a message indicating a newer version is available
    And the command exits with code 1
    And no binary file is written or modified
    And no temp file is created
    And the cache file may be written but the binary path is untouched

  Scenario: --check reports already current and exits zero with zero writes
    Given the running binary version is "0.40.2"
    And the fake GitHub API returns tag "v0.40.2"
    When the user runs:
      centinela update --check
    Then the output contains a message indicating the binary is up to date
    And the command exits with code 0
    And no binary file is written or modified
    And no temp file is created

  Scenario: --check honors the TTL cache and makes no network call within TTL
    Given the running binary version is "0.37.0"
    And a valid cache file exists with tag "v0.40.2" written within the last 24 hours
    When the user runs:
      centinela update --check
    Then no HTTP request is made to the GitHub API
    And the output contains a message indicating a newer version is available
    And the command exits with code 1
    And no binary file is written or modified

  # ---------------------------------------------------------------------------
  # AC3 — Checksum mismatch fails safe
  # ---------------------------------------------------------------------------

  Scenario: Checksum mismatch aborts without touching the installed binary
    Given the running binary version is "0.37.0"
    And the fake GitHub API returns tag "v0.40.2" with an asset matching the host platform
    And the fake SHA256SUMS contains an INCORRECT checksum for the asset
    And the install directory is writable
    When the user runs:
      centinela update
    Then the command exits with a non-zero code
    And the output contains an error message mentioning checksum or verification failure
    And the installed binary is byte-identical to its state before the command ran
    And any temp file created during the download is removed

  # ---------------------------------------------------------------------------
  # AC4 — Unsupported platform / missing asset
  # ---------------------------------------------------------------------------

  Scenario: Missing asset for the host platform returns a typed error with no partial write
    Given the running binary version is "0.37.0"
    And the fake GitHub API returns tag "v0.40.2" with NO asset matching the host platform
    When the user runs:
      centinela update
    Then the command exits with a non-zero code
    And the output contains a clear error message identifying the unsupported platform
    And no binary file is written or modified
    And no temp file is created

  # ---------------------------------------------------------------------------
  # AC5 — Permission denied fails safe
  # ---------------------------------------------------------------------------

  Scenario: Unwritable install directory returns a typed error and leaves binary untouched
    Given the running binary version is "0.37.0"
    And the fake GitHub API returns tag "v0.40.2" with an asset matching the host platform
    And the fake SHA256SUMS contains the correct checksum for the asset
    And the install directory is NOT writable
    When the user runs:
      centinela update
    Then the command exits with a non-zero code
    And the output contains a clear error message about the permission failure
    And the installed binary is byte-identical to its state before the command ran
    And any temp file created during the download is removed

  # ---------------------------------------------------------------------------
  # AC6 — Startup notice: throttled and fail-silent
  # ---------------------------------------------------------------------------

  Scenario: Startup notice appears when running an older version and cache is stale
    Given the running binary version is "0.37.0"
    And no valid cache file exists within the TTL
    And the fake GitHub API returns tag "v0.40.2"
    When the hook "centinela hook session" runs
    Then the output contains an update-available notice mentioning "v0.40.2"
    And no binary is installed or modified
    And the cache file is written with the latest tag and current timestamp

  Scenario: Startup notice is suppressed when the cache is within the TTL
    Given the running binary version is "0.37.0"
    And a valid cache file exists with tag "v0.40.2" written within the last 24 hours
    When the hook "centinela hook session" runs
    Then no HTTP request is made to the GitHub API
    And the output does not contain an update-available notice
    And the cache file is NOT rewritten

  Scenario: Startup notice is suppressed when already on the latest version
    Given the running binary version is "0.40.2"
    And the fake GitHub API returns tag "v0.40.2"
    When the hook "centinela hook session" runs
    Then the output does not contain an update-available notice
    And no binary is installed or modified

  Scenario: Startup notice fails silently when the GitHub API is unreachable
    Given the running binary version is "0.37.0"
    And no valid cache file exists within the TTL
    And the fake GitHub API returns a network error or HTTP 5xx
    When the hook "centinela hook session" runs
    Then the command exits with code 0
    And no error message or stack trace is printed to stdout
    And no binary is installed or modified

  Scenario: Startup notice never auto-installs
    Given the running binary version is "0.37.0"
    And the fake GitHub API returns tag "v0.40.2"
    When the hook "centinela hook session" runs
    Then the installed binary is byte-identical to its state before the hook ran

  # ---------------------------------------------------------------------------
  # AC7 — Deterministic tests (infrastructure contract)
  # ---------------------------------------------------------------------------

  Scenario: All network calls target the httptest.Server and not the real GitHub API
    Given an httptest.Server is configured to serve releases and asset downloads
    And a temp HOME/XDG dir is set as the cache home
    When "centinela update" and "centinela update --check" run under the test harness
    Then no DNS lookup for api.github.com or github.com occurs
    And the temp HOME/XDG dir contains all written cache files
    And the test can assert exact HTTP request counts against the test server

  # ---------------------------------------------------------------------------
  # Edge case — version-string normalization
  # ---------------------------------------------------------------------------

  Scenario: Version comparison strips leading v from the release tag
    Given the running binary version is "0.40.2"
    And the fake GitHub API returns tag_name "v0.40.2"
    When "centinela update --check" runs
    Then the versions are treated as equal
    And the output contains a message indicating the binary is up to date
    And the command exits with code 0

  Scenario: Asset name is constructed with the leading v from the tag
    Given the running binary version is "0.37.0"
    And the fake GitHub API returns tag_name "v0.40.2"
    And the host platform is "linux" / "amd64"
    When "centinela update" resolves the asset
    Then the resolved asset name is "centinela-v0.40.2-linux-amd64"

  Scenario: Asset name for Windows carries the .exe suffix
    Given the host platform is "windows" / "amd64"
    And the release tag is "v0.40.2"
    When the asset name is computed
    Then the resolved asset name is "centinela-v0.40.2-windows-amd64.exe"

  # ---------------------------------------------------------------------------
  # Edge case — dev build sentinel
  # ---------------------------------------------------------------------------

  Scenario: dev build prints an informational message and skips the update
    Given the running binary version is "dev"
    And the fake GitHub API returns tag "v0.40.2"
    When the user runs:
      centinela update
    Then the command exits with code 0
    And the output contains an informational message indicating this is a development build
    And no binary is downloaded or modified
    And no temp file is created

  Scenario: dev build suppresses the startup notice
    Given the running binary version is "dev"
    And the fake GitHub API returns tag "v0.40.2"
    When the hook "centinela hook session" runs
    Then the output does not contain an update-available notice
    And no HTTP request is made to the GitHub API

  # ---------------------------------------------------------------------------
  # Edge case — symlinked binary
  # ---------------------------------------------------------------------------

  Scenario: Update resolves a symlinked binary to its real path before replacing
    Given the running binary version is "0.37.0"
    And the binary on PATH is a symlink pointing to the real binary in a writable directory
    And the fake GitHub API returns tag "v0.40.2" with an asset matching the host platform
    And the fake SHA256SUMS contains the correct checksum for the asset
    When the user runs:
      centinela update
    Then os.Executable and EvalSymlinks are used to resolve the real target path
    And the real binary file is replaced atomically
    And the symlink still points to the now-updated real binary

  # ---------------------------------------------------------------------------
  # Edge case — offline / API error on explicit update
  # ---------------------------------------------------------------------------

  Scenario: Explicit update returns a clear error when GitHub API is unreachable
    Given the running binary version is "0.37.0"
    And the fake GitHub API returns a network error
    When the user runs:
      centinela update
    Then the command exits with a non-zero code
    And the output contains a clear error message about the network failure
    And no binary is modified

  # ---------------------------------------------------------------------------
  # Edge case — stale or corrupt cache
  # ---------------------------------------------------------------------------

  Scenario: Stale cache older than the TTL triggers a fresh network check
    Given the running binary version is "0.37.0"
    And a cache file exists with tag "v0.39.0" written more than 24 hours ago
    And the fake GitHub API now returns tag "v0.40.2"
    When the hook "centinela hook session" runs
    Then one HTTP request is made to the GitHub API
    And the cache file is rewritten with tag "v0.40.2" and the current timestamp
    And the output contains an update-available notice mentioning "v0.40.2"

  Scenario: Corrupt or empty cache file triggers a fresh network check without panic
    Given the running binary version is "0.37.0"
    And the cache file at the XDG path contains invalid JSON
    And the fake GitHub API returns tag "v0.40.2"
    When the hook "centinela hook session" runs
    Then the command does not panic or exit non-zero due to the corrupt cache
    And one HTTP request is made to the GitHub API
    And the cache file is rewritten with valid JSON

  # ---------------------------------------------------------------------------
  # Edge case — GitHub API rate limit
  # ---------------------------------------------------------------------------

  Scenario: GitHub API 429 during startup notice fails silently
    Given the running binary version is "0.37.0"
    And no valid cache file exists within the TTL
    And the fake GitHub API returns HTTP 429
    When the hook "centinela hook session" runs
    Then the command exits with code 0
    And no error is printed
    And no binary is modified

  Scenario: GitHub API 403 during explicit update returns a clear typed error
    Given the running binary version is "0.37.0"
    And the fake GitHub API returns HTTP 403
    When the user runs:
      centinela update
    Then the command exits with a non-zero code
    And the output contains a clear error message about the API failure
    And no binary is modified
