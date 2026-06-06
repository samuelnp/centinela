package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// AC3: a vulnerable dependency (fake govulncheck emitting a finding) ->
// G-Vuln Warn AND AllPassed stays true (a Warn must not block validate).
func TestAccept_Security_VulnFound_WarnsNotBlocks(t *testing.T) {
	dir := secPath(t)
	// govulncheck -json NDJSON: a finding record pins osv id + trace module.
	secBin(t, dir, "govulncheck",
		`printf '{"finding":{"osv":"GO-2024-0001","trace":[{"module":"example.com/vulnerable"}]}}'
exit 3`)
	results := gates.RunWithFilter(secCfg(), nil)
	r, ok := secResult(t, results, "G-Vuln")
	if !ok || r.Status != gates.Warn {
		t.Fatalf("AC3: G-Vuln must Warn, got ok=%v %v: %q", ok, r.Status, r.Message)
	}
	if !gates.AllPassed(results) {
		t.Fatal("AC3: a vuln Warn must NOT flip AllPassed to false")
	}
	if !strings.Contains(strings.Join(r.Details, "\n"), "GO-2024-0001") {
		t.Fatalf("AC3: details must name the vuln id, got %v", r.Details)
	}
}

// AC6: the only secret finding's rule ID is in secrets.allowlist -> excluded
// -> G-Secrets Pass.
func TestAccept_Security_AllowlistedSecret_Passes(t *testing.T) {
	dir := secPath(t)
	secBin(t, dir, "gitleaks", `printf '[{"RuleID":"known-fp","File":"app.go"}]' > "$6"
exit 1`)
	cfg := secCfg()
	cfg.Gates.Security.Secrets.Allowlist = []string{"known-fp"}
	r, ok := secResult(t, gates.RunWithFilter(cfg, nil), "G-Secrets")
	if !ok || r.Status != gates.Pass {
		t.Fatalf("AC6: allowlisted finding must yield Pass, got %v: %q", r.Status, r.Message)
	}
}

// AC7: diff-aware filter active. When the only secret is in an unchanged file
// (outside the diff set) -> G-Secrets Pass; when the file is in the diff set
// -> G-Secrets Fail. Drives the real filter via gitdiff.NewSet.
func TestAccept_Security_DiffAware_FiltersUnchangedFile(t *testing.T) {
	dir := secPath(t)
	secBin(t, dir, "gitleaks", `printf '[{"RuleID":"generic-api-key","File":"secret.go"}]' > "$6"
exit 1`)

	outOfDiff := gitdiff.NewSet([]string{"other.go"})
	if r, _ := secResult(t, gates.RunWithFilter(secCfg(), outOfDiff), "G-Secrets"); r.Status != gates.Pass {
		t.Fatalf("AC7: secret in unchanged file must be filtered -> Pass, got %v: %q", r.Status, r.Message)
	}

	inDiff := gitdiff.NewSet([]string{"secret.go"})
	if r, _ := secResult(t, gates.RunWithFilter(secCfg(), inDiff), "G-Secrets"); r.Status != gates.Fail {
		t.Fatalf("AC7: secret in changed file must Fail, got %v: %q", r.Status, r.Message)
	}
}

// Edge: both govulncheck and osv-scanner report the same (pkg, id) -> deduped
// to a single G-Vuln detail entry.
func TestAccept_Security_DuplicateCVE_Deduped(t *testing.T) {
	dir := secPath(t)
	secBin(t, dir, "govulncheck",
		`printf '{"finding":{"osv":"GO-2024-0001","trace":[{"module":"example.com/dup"}]}}'`)
	secBin(t, dir, "osv-scanner",
		`printf '{"results":[{"packages":[{"package":{"name":"example.com/dup"},"vulnerabilities":[{"id":"GO-2024-0001"}]}]}]}'`)
	r, _ := secResult(t, gates.RunWithFilter(secCfg(), nil), "G-Vuln")
	var n int
	for _, d := range r.Details {
		if strings.Contains(d, "GO-2024-0001") {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("dedup: expected exactly 1 entry for the shared CVE, got %d in %v", n, r.Details)
	}
}
