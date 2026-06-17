# Edge Cases: security-gate

Each edge case from the plan's "Edge Cases" section, mapped to the test(s) that
assert it. Tier key: U = colocated unit (`internal/gates`, `internal/config`),
A = acceptance (`tests/acceptance`).

## Covered

- **Malformed JSON (secrets) → Warn, never false Pass.** A scanner that ran but
  emitted non-JSON output classifies as Warn.
  - U `internal/gates/security_classify_test.go::TestClassifySecrets_MalformedJSONYieldsWarn`
  - U `internal/gates/security_fake_bin_test.go::TestCheckSecrets_FakeGitleaksMalformedJSON`
  - U `internal/gates/security_secrets_parse_test.go::TestReadSecretsReport_MalformedJSONIsError`

- **Malformed JSON (vuln) → Warn, never false Pass.** A vuln tool that emits
  unparseable output surfaces as a Warn note, not a Pass.
  - U `internal/gates/security_vuln_parse_test.go::TestParseGovulncheck_MalformedOutputIsError`
  - U `internal/gates/security_vuln_parse_test.go::TestParseOSVScanner_MalformedOutputIsError`
  - U `internal/gates/security_fold_test.go::TestFoldVuln_ToolWarnWithNoFindingsYieldsWarn`

- **Empty JSON / nothing scanned → clean (never a false Fail).** Empty or missing
  report is treated as "no findings" → Pass, never Fail.
  - U `internal/gates/security_secrets_parse_test.go::TestReadSecretsReport_EmptyFileIsClean`
  - U `internal/gates/security_secrets_parse_test.go::TestReadSecretsReport_MissingFileIsClean`
  - U `internal/gates/security_vuln_parse_test.go::TestParseGovulncheck_EmptyOutputIsClean`
  - U `internal/gates/security_vuln_parse_test.go::TestParseOSVScanner_EmptyResultsIsClean`

- **No dependency manifest for a vuln tool → not a Fail.** A present vuln tool
  that finds nothing to scan emits empty output, which folds to a clean,
  non-failing result (Pass), never a Fail. Tool *absence* (no scanner at all) is
  the distinct Skip path below.
  - U `internal/gates/security_vuln_parse_test.go::TestParseOSVScanner_EmptyResultsIsClean`
  - U `internal/gates/security_fold_test.go::TestFoldVuln_NoFindingsYieldsPass`

- **Same CVE from both vuln tools → de-duplicated by (package, vuln-id).** The
  shared finding appears exactly once in Details.
  - U `internal/gates/security_fold_test.go::TestFoldVuln_DedupByPackageAndID`
  - U `internal/gates/security_fold_test.go::TestVulnKey_DedupAcrossTwoTools`
  - A `tests/acceptance/security_gate_more_test.go::TestAccept_Security_DuplicateCVE_Deduped`
    (drives real govulncheck + osv-scanner fakes reporting the same id)

- **Allowlist entry that matches nothing → silently ignored, not an error.**
  - U `internal/gates/security_retain_test.go::TestRetainFindings_UnmatchedAllowlistIsIgnored`

- **Both scanner families absent → two Skips + "no scanners available" signal**
  (must not be mistaken for a verified clean scan); AllPassed stays true.
  - U `internal/gates/security_skip_test.go::TestCheckSecurity_BothAbsentYieldsTwoSkipsWithNoScannersNote`
  - U `internal/gates/security_skip_test.go::TestRunWithFilter_SecurityEnabledBothAbsentAppendsTwo`
  - A `tests/acceptance/security_gate_test.go::TestAccept_Security_GitleaksAbsent_Skips`

- **Diff-aware vs full scan.** Locally a non-nil gitdiff.Set scopes secret
  findings to changed files (out-of-diff secret → Pass); a nil filter (CI
  full-scan) still catches it.
  - U `internal/gates/security_secrets_parse_test.go::TestRetainFindings_DiffFilterDropsOutOfDiff`
  - U `internal/gates/security_secrets_parse_test.go::TestRetainFindings_DiffFilterIncludesInDiff`
  - A `tests/acceptance/security_gate_more_test.go::TestAccept_Security_DiffAware_FiltersUnchangedFile`

- **Scan timeout → Warn (never wedge validate, never false Pass).** A scanner that
  exceeds `scanTimeout` (security_exec.go) is mapped to `errScanTimeout`; the
  sentinel is classified as a non-Pass (Warn) outcome.
  - U `internal/gates/security_classify_test.go::TestClassifySecrets_LaunchFailureIsNotPass`
    (asserts an `errScanTimeout`-class run never yields Pass)

## Residual Risks

- The live 120s context-deadline path inside `runScanner` is not exercised by a
  unit test (it would require a real long-running fake binary). The
  timeout→Warn *mapping* is asserted via the `errScanTimeout` sentinel; the
  deadline mechanism itself relies on Go's `context.WithTimeout`.
- Acceptance fakes emit fixed JSON; they validate the gate's parsing/severity
  wiring, not the real gitleaks/govulncheck/osv-scanner output formats. Tool
  output-format drift is mitigated by defensive parsers (malformed → Warn) and
  pinned tool versions documented in the feature plan.
- Full git-history secret scanning and vulnerability remediation are explicitly
  out of scope for v1 (see plan "out of scope").
