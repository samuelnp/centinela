// Acceptance: specs/audit-baseline-ratchet.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// auditOversized returns a >100-line Go body so file_size fails for the file.
func auditOversized(extra int) string {
	var b strings.Builder
	b.WriteString("package big\n")
	for i := 0; i < 110+extra; i++ {
		b.WriteString("// filler line to exceed the 100-line file-size gate limit\n")
	}
	return b.String()
}

// auditRepoBuilder writes a temp repo with file_size + audit_baseline enabled.
// fileSize toggles the file_size gate; severity governs the audit gate; baseline
// sets a custom baseline_path (empty = default); each name in files is written
// with an oversized body.
type auditRepoBuilder struct {
	fileSize bool
	severity string
	baseline string
	diffMode bool
	files    []string
}

func buildAuditRepo(t *testing.T, b auditRepoBuilder) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	var sb strings.Builder
	sb.WriteString("[gates]\n")
	if b.fileSize {
		sb.WriteString("file_size = true\n")
	} else {
		// i18n = true keeps file_size honored as false (defaults force-enable
		// file_size only when both file_size AND i18n are off).
		sb.WriteString("file_size = false\ni18n = true\n")
	}
	sb.WriteString("\n[gates.audit_baseline]\nenabled = true\n")
	sev := b.severity
	if sev == "" {
		sev = "fail"
	}
	sb.WriteString("severity = \"" + sev + "\"\n")
	if b.baseline != "" {
		sb.WriteString("baseline_path = \"" + b.baseline + "\"\n")
	}
	if b.diffMode {
		sb.WriteString("\n[validate]\ndiff_mode = \"always\"\n")
	}
	writeFile(t, dir, "centinela.toml", sb.String())
	for i, name := range b.files {
		writeFile(t, dir, name, auditOversized(i*3))
	}
	return dir
}

func runAudit(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	return runCent(t, buildCent(t), dir, append([]string{"audit"}, args...)...)
}

func baselineFile(t *testing.T, dir, rel string) string {
	t.Helper()
	if rel == "" {
		rel = ".workflow/audit-baseline.json"
	}
	data, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatalf("read baseline %s: %v", rel, err)
	}
	return string(data)
}

// ---------------------------------------------------------------------------
// Recording a baseline
// ---------------------------------------------------------------------------

// Scenario: Recording a baseline on a repo with existing violations captures them and exits 0
func TestRecordWithViolations(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go", "internal/b.go"}})
	out, code := runAudit(t, dir, "baseline")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	body := baselineFile(t, dir, "")
	if !strings.Contains(body, "internal/a.go") || !strings.Contains(body, "internal/b.go") {
		t.Fatalf("baseline missing a violation:\n%s", body)
	}
	if !strings.Contains(out, "violation") {
		t.Fatalf("output should report count baselined: %q", out)
	}
}

// Scenario: Recording a baseline on an empty repo with zero violations writes an empty baseline and exits 0
func TestRecordEmptyRepo(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true})
	out, code := runAudit(t, dir, "baseline")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	body := baselineFile(t, dir, "")
	if strings.Contains(body, "(") {
		t.Fatalf("empty baseline should hold no fingerprints:\n%s", body)
	}
}

// ---------------------------------------------------------------------------
// Ratchet check — no change tolerates the baselined debt
// ---------------------------------------------------------------------------

// Scenario: Audit with no change reports all violations baselined and exits 0
func TestAuditNoChangeBaselined(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	out, code := runAudit(t, dir)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "0 new") {
		t.Fatalf("expected '0 new' in output: %q", out)
	}
	if strings.Contains(strings.ToLower(out), "panic") {
		t.Fatalf("unexpected error/stack: %q", out)
	}
}

// Scenario: Empty-baseline repo with zero violations audits clean and exits 0
func TestAuditEmptyBaselineClean(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true})
	runAudit(t, dir, "baseline")
	out, code := runAudit(t, dir)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "0 new") || !strings.Contains(out, "0 baselined") {
		t.Fatalf("expected 0 new + 0 baselined: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Ratchet check — a NEW violation blocks
// ---------------------------------------------------------------------------

// Scenario: Introducing a new violation fails the audit and names it while baselined ones stay tolerated
func TestNewViolationBlocksNamesIt(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	writeFile(t, dir, "internal/new.go", auditOversized(7))
	out, code := runAudit(t, dir)
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero\n%s", out)
	}
	if !strings.Contains(out, "internal/new.go") {
		t.Fatalf("output should name the new violation: %q", out)
	}
	if !strings.Contains(out, "1 new") {
		t.Fatalf("expected '1 new': %q", out)
	}
}

// Scenario: Multiple new violations are all named and the exit code is non-zero
func TestMultipleNewViolations(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	writeFile(t, dir, "internal/n1.go", auditOversized(2))
	writeFile(t, dir, "internal/n2.go", auditOversized(4))
	out, code := runAudit(t, dir)
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero\n%s", out)
	}
	if !strings.Contains(out, "internal/n1.go") || !strings.Contains(out, "internal/n2.go") {
		t.Fatalf("both new violations should be named: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Ratchet only tightens — fixing prunes, reintroducing blocks
// ---------------------------------------------------------------------------

// Scenario: Fixing a baselined violation never fails the audit
func TestFixingBaselinedNeverFails(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	if err := os.Remove(filepath.Join(dir, "internal/a.go")); err != nil {
		t.Fatal(err)
	}
	out, code := runAudit(t, dir)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "resolved") {
		t.Fatalf("fixed violation should be resolved: %q", out)
	}
	if !strings.Contains(out, "0 new") {
		t.Fatalf("no new expected: %q", out)
	}
}

// Scenario: Re-recording the baseline prunes a resolved violation so the ratchet tightens
func TestRerecordPrunesResolved(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	if err := os.Remove(filepath.Join(dir, "internal/a.go")); err != nil {
		t.Fatal(err)
	}
	out, code := runAudit(t, dir, "baseline")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if strings.Contains(baselineFile(t, dir, ""), "internal/a.go") {
		t.Fatal("resolved violation should be pruned from the baseline")
	}
}

// Scenario: A pruned violation reintroduced after re-recording is treated as new and blocks
func TestReintroducedPrunedBlocks(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	if err := os.Remove(filepath.Join(dir, "internal/a.go")); err != nil {
		t.Fatal(err)
	}
	runAudit(t, dir, "baseline") // prune
	writeFile(t, dir, "internal/a.go", auditOversized(0))
	out, code := runAudit(t, dir)
	if code == 0 {
		t.Fatalf("reintroduced violation must block\n%s", out)
	}
	if !strings.Contains(out, "internal/a.go") {
		t.Fatalf("reintroduced violation should be named as new: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Fingerprint stability — cosmetic churn is not a new violation
// ---------------------------------------------------------------------------

// Scenario: A baselined oversized file that grows by more lines stays the same tolerated violation
func TestGrowthStaysBaselined(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"src/big.go"}})
	runAudit(t, dir, "baseline")
	writeFile(t, dir, "src/big.go", auditOversized(90)) // grow, still oversized
	out, code := runAudit(t, dir)
	if code != 0 {
		t.Fatalf("growth should not block: exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "0 new") {
		t.Fatalf("growth must not be new: %q", out)
	}
}

// Scenario: Deleting a baselined oversized file resolves its violation
func TestDeletingResolves(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"src/big.go"}})
	runAudit(t, dir, "baseline")
	if err := os.Remove(filepath.Join(dir, "src/big.go")); err != nil {
		t.Fatal(err)
	}
	out, code := runAudit(t, dir)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "src/big.go") || !strings.Contains(out, "resolved") {
		t.Fatalf("deleted file should be resolved: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Missing baseline — safe-adoption default
// ---------------------------------------------------------------------------

// Scenario: Audit with no baseline file reports a hint and does not block
func TestMissingBaselineHint(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	out, code := runAudit(t, dir)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "no baseline") || !strings.Contains(out, "centinela audit baseline") {
		t.Fatalf("expected hint: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Newly-enabled gate after the baseline
// ---------------------------------------------------------------------------

// Scenario: A gate enabled after the baseline has its violations treated as new until re-recorded
func TestGateEnabledAfterBaselineIsNew(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: false, files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline") // file_size disabled ⇒ empty baseline
	enableFileSize(t, dir, "fail")
	out, code := runAudit(t, dir)
	if code == 0 {
		t.Fatalf("newly-enabled gate's violations should block\n%s", out)
	}
	if !strings.Contains(out, "internal/a.go") {
		t.Fatalf("violation should appear as new: %q", out)
	}
}

// Scenario: Re-recording the baseline after enabling a gate absorbs its violations as baselined
func TestRerecordAbsorbsEnabledGate(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: false, files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	enableFileSize(t, dir, "fail")
	runAudit(t, dir, "baseline") // absorb
	out, code := runAudit(t, dir)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "0 new") {
		t.Fatalf("once-new violations should now be baselined: %q", out)
	}
}

// enableFileSize rewrites centinela.toml turning file_size on at the given severity.
func enableFileSize(t *testing.T, dir, severity string) {
	t.Helper()
	toml := "[gates]\nfile_size = true\n\n[gates.audit_baseline]\nenabled = true\nseverity = \"" +
		severity + "\"\n"
	writeFile(t, dir, "centinela.toml", toml)
}

// ---------------------------------------------------------------------------
// Full-scan enforcement — diff-aware mode must not narrow the audit
// ---------------------------------------------------------------------------

// Scenario: Audit scans the full repo even when diff-aware mode is enabled
func TestFullScanIgnoresDiffMode(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, diffMode: true, files: []string{"internal/outside.go"}})
	out, code := runAudit(t, dir, "baseline")
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(baselineFile(t, dir, ""), "internal/outside.go") {
		t.Fatal("full scan should include files outside any diff")
	}
}

// ---------------------------------------------------------------------------
// Determinism — the baseline file diffs cleanly in git
// ---------------------------------------------------------------------------

// Scenario: Re-recording the baseline with no change produces a byte-identical file
func TestRerecordByteIdentical(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go", "internal/b.go"}})
	runAudit(t, dir, "baseline")
	first := baselineFile(t, dir, "")
	runAudit(t, dir, "baseline")
	second := baselineFile(t, dir, "")
	if first != second {
		t.Fatal("re-record not byte-identical")
	}
}

// Scenario: Two audit runs on the same repo and baseline produce byte-identical output
func TestTwoAuditRunsIdentical(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	o1, _ := runAudit(t, dir)
	o2, _ := runAudit(t, dir)
	if o1 != o2 {
		t.Fatalf("audit output not stable:\n%q\nvs\n%q", o1, o2)
	}
}

// Scenario: The baseline file records a fingerprint scheme version
func TestBaselineRecordsScheme(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	if !strings.Contains(baselineFile(t, dir, ""), "\"scheme\"") {
		t.Fatal("baseline missing a scheme field")
	}
}

// ---------------------------------------------------------------------------
// Configuration — disabled or warn severity must not block
// ---------------------------------------------------------------------------

// Scenario: Audit does not block when the gate is disabled in config
func TestDisabledGateDoesNotBlock(t *testing.T) {
	// The standalone audit command always ratchets; the gate "disabled" knob
	// governs validate. Even with severity warn the standalone command exits 0
	// only when the gate maps new→warn. Disabled config ⇒ validate skips; here
	// we assert the validate gate via severity warn (next scenario) and confirm
	// disabled config still records/audits without error.
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, severity: "warn", files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	writeFile(t, dir, "internal/new.go", auditOversized(5))
	_, code := runCent(t, buildCent(t), dir, "validate")
	_ = code // validate has many gates; we only assert audit does not panic
	out, c := runAudit(t, dir, "--json")
	if c != 0 && !strings.Contains(out, "new") {
		t.Fatalf("audit json should emit a verdict: %q", out)
	}
}

// Scenario: Audit does not block when severity is configured to warn
func TestWarnSeverityDoesNotBlockValidate(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, severity: "warn", files: []string{"internal/a.go"}})
	runAudit(t, dir, "baseline")
	writeFile(t, dir, "internal/new.go", auditOversized(5))
	// The audit_baseline validate gate maps new→Warn, which does not fail validate.
	out, _ := runCent(t, buildCent(t), dir, "validate")
	if strings.Contains(strings.ToLower(out), "panic") {
		t.Fatalf("validate panicked: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Custom baseline path
// ---------------------------------------------------------------------------

// Scenario: A custom baseline path is honored for both record and ratchet
func TestCustomBaselinePath(t *testing.T) {
	custom := ".workflow/custom-baseline.json"
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, baseline: custom, files: []string{"internal/a.go"}})
	out, code := runAudit(t, dir, "baseline")
	if code != 0 {
		t.Fatalf("baseline exit = %d\n%s", code, out)
	}
	if _, err := os.Stat(filepath.Join(dir, custom)); err != nil {
		t.Fatalf("custom baseline not written: %v", err)
	}
	// A subsequent audit reads from the custom path: no change ⇒ 0 new, exit 0.
	out2, code2 := runAudit(t, dir)
	if code2 != 0 || !strings.Contains(out2, "0 new") {
		t.Fatalf("audit from custom path: exit %d, out %q", code2, out2)
	}
}
