---
id: 0cd0b9d5a3adf24f
feature: security-gate
step: tests
type: lesson
title: - **Malformed JSON (secrets) → Warn, never false Pass.** A scanner that ran but
tags: edge-cases, lesson
sourceArtifact: .workflow/security-gate-edge-cases.md
createdAt: 2026-06-06T18:53:48Z
---

- **Malformed JSON (secrets) → Warn, never false Pass.** A scanner that ran but
- U `internal/gates/security_classify_test.go::TestClassifySecrets_MalformedJSONYieldsWarn`
- U `internal/gates/security_fake_bin_test.go::TestCheckSecrets_FakeGitleaksMalformedJSON`
- U `internal/gates/security_secrets_parse_test.go::TestReadSecretsReport_MalformedJSONIsError`
- **Malformed JSON (vuln) → Warn, never false Pass.** A vuln tool that emits
- U `internal/gates/security_vuln_parse_test.go::TestParseGovulncheck_MalformedOutputIsError`
- U `internal/gates/security_vuln_parse_test.go::TestParseOSVScanner_MalformedOutputIsError`
- U `internal/gates/security_fold_test.go::TestFoldVuln_ToolWarnWithNoFindingsYieldsWarn`
- **Empty JSON / nothing scanned → clean (never a false Fail).** Empty or missing
- U `internal/gates/security_secrets_parse_test.go::TestReadSecretsReport_EmptyFileIsClean`
- U `internal/gates/security_secrets_parse_test.go::TestReadSecretsReport_MissingFileIsClean`
- U `internal/gates/security_vuln_parse_test.go::TestParseGovulncheck_EmptyOutputIsClean`
- U `internal/gates/security_vuln_parse_test.go::TestParseOSVScanner_EmptyResultsIsClean`
- **No dependency manifest for a vuln tool → not a Fail.** A present vuln tool
- U `internal/gates/security_vuln_parse_test.go::TestParseOSVScanner_EmptyResultsIsClean`
- U `internal/gates/security_fold_test.go::TestFoldVuln_NoFindingsYieldsPass`
- **Same CVE from both vuln tools → de-duplicated by (package, vuln-id).** The
- U `internal/gates/security_fold_test.go::TestFoldVuln_DedupByPackageAndID`
- U `internal/gates/security_fold_test.go::TestVulnKey_DedupAcrossTwoTools`
- A `tests/acceptance/security_gate_more_test.go::TestAccept_Security_DuplicateCVE_Deduped`
- **Allowlist entry that matches nothing → silently ignored, not an error.**
- U `internal/gates/security_retain_test.go::TestRetainFindings_UnmatchedAllowlistIsIgnored`
- **Both scanner families absent → two Skips + "no scanners available" signal**
- U `internal/gates/security_skip_test.go::TestCheckSecurity_BothAbsentYieldsTwoSkipsWithNoScannersNote`
- U `internal/gates/security_skip_test.go::TestRunWithFilter_SecurityEnabledBothAbsentAppendsTwo`
- A `tests/acceptance/security_gate_test.go::TestAccept_Security_GitleaksAbsent_Skips`
- **Diff-aware vs full scan.** Locally a non-nil gitdiff.Set scopes secret
- U `internal/gates/security_secrets_parse_test.go::TestRetainFindings_DiffFilterDropsOutOfDiff`
- U `internal/gates/security_secrets_parse_test.go::TestRetainFindings_DiffFilterIncludesInDiff`
- A `tests/acceptance/security_gate_more_test.go::TestAccept_Security_DiffAware_FiltersUnchangedFile`
- **Scan timeout → Warn (never wedge validate, never false Pass).** A scanner that
- U `internal/gates/security_classify_test.go::TestClassifySecrets_LaunchFailureIsNotPass`
- The live 120s context-deadline path inside `runScanner` is not exercised by a
- Acceptance fakes emit fixed JSON; they validate the gate's parsing/severity
- Full git-history secret scanning and vulnerability remediation are explicitly
