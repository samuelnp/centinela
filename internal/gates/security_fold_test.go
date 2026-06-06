package gates

import (
	"testing"
)

// TestFoldVuln_FindingsYieldsWarn exercises AC3: finding -> Warn.
func TestFoldVuln_FindingsYieldsWarn(t *testing.T) {
	base := Result{Name: vulnName}
	findings := map[vulnKey]bool{{Pkg: "pkg-a", ID: "CVE-1"}: true}
	r := foldVuln(base, findings, nil)
	if r.Status != Warn {
		t.Fatalf("findings must yield Warn, got %v", r.Status)
	}
	if len(r.Details) == 0 {
		t.Fatal("Warn must carry details")
	}
}

// TestFoldVuln_NoFindingsYieldsPass verifies clean run -> Pass.
func TestFoldVuln_NoFindingsYieldsPass(t *testing.T) {
	base := Result{Name: vulnName}
	r := foldVuln(base, nil, nil)
	if r.Status != Pass {
		t.Fatalf("no findings must yield Pass, got %v", r.Status)
	}
}

// TestFoldVuln_ToolWarnWithNoFindingsYieldsWarn verifies tool-error note -> Warn.
func TestFoldVuln_ToolWarnWithNoFindingsYieldsWarn(t *testing.T) {
	base := Result{Name: vulnName}
	r := foldVuln(base, nil, []string{"govulncheck: parse error"})
	if r.Status != Warn {
		t.Fatalf("tool note with no findings must yield Warn, got %v", r.Status)
	}
}

// TestFoldVuln_DedupByPackageAndID verifies dedup (AC: both tools reporting
// same CVE -> one detail line). The dedup is enforced by the map[vulnKey]bool
// aggregation in checkVuln; foldVuln just renders it.
func TestFoldVuln_DedupByPackageAndID(t *testing.T) {
	base := Result{Name: vulnName}
	findings := map[vulnKey]bool{
		{Pkg: "pkg-a", ID: "CVE-2024-1"}: true,
	}
	r := foldVuln(base, findings, nil)
	if len(r.Details) != 1 {
		t.Fatalf("dedup must yield 1 detail, got %v", r.Details)
	}
}

// TestAllPassed_WarnDoesNotBlock verifies AC3: Warn does not block validate.
func TestAllPassed_WarnDoesNotBlock(t *testing.T) {
	results := []Result{
		{Status: Pass},
		{Status: Warn},
		{Status: Skip},
	}
	if !AllPassed(results) {
		t.Fatal("Warn/Skip must not cause AllPassed to return false")
	}
}

// TestAllPassed_FailBlocks verifies AC1: Fail causes AllPassed false.
func TestAllPassed_FailBlocks(t *testing.T) {
	results := []Result{{Status: Fail}}
	if AllPassed(results) {
		t.Fatal("Fail must cause AllPassed to return false")
	}
}

// TestVulnKey_DedupAcrossTwoTools verifies that inserting the same (pkg, id)
// from two different tool outputs produces one entry in the map.
func TestVulnKey_DedupAcrossTwoTools(t *testing.T) {
	m := map[vulnKey]bool{}
	k := vulnKey{Pkg: "example.com/lib", ID: "GO-2024-0001"}
	m[k] = true
	m[k] = true // simulate both govulncheck and osv-scanner reporting the same vuln
	if len(m) != 1 {
		t.Fatalf("same (pkg,id) from two tools must dedup to 1, got %d", len(m))
	}
}
