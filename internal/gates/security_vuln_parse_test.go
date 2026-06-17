package gates

import (
	"strings"
	"testing"
)

// govulncheck NDJSON fixtures.
const govulnHappy = `{"finding":{"osv":"GO-2024-0001","trace":[{"module":"example.com/pkg"}]}}` + "\n"
const govulnMalformed = `{"finding": BROKEN`

// TestParseGovulncheck_EmptyOutputIsClean verifies empty input = no findings.
func TestParseGovulncheck_EmptyOutputIsClean(t *testing.T) {
	keys, err := parseGovulncheck(nil)
	if err != nil || len(keys) != 0 {
		t.Fatalf("empty output must yield no findings, got %v / %v", keys, err)
	}
}

// TestParseGovulncheck_HappyFinding decodes one finding correctly.
func TestParseGovulncheck_HappyFinding(t *testing.T) {
	keys, err := parseGovulncheck([]byte(govulnHappy))
	if err != nil || len(keys) != 1 {
		t.Fatalf("expected 1 finding, got %v / %v", keys, err)
	}
	if keys[0].ID != "GO-2024-0001" || keys[0].Pkg != "example.com/pkg" {
		t.Fatalf("wrong key: %+v", keys[0])
	}
}

// TestParseGovulncheck_MalformedOutputIsError verifies non-NDJSON -> error.
func TestParseGovulncheck_MalformedOutputIsError(t *testing.T) {
	_, err := parseGovulncheck([]byte(govulnMalformed))
	if err == nil {
		t.Fatal("expected error for malformed govulncheck output")
	}
}

// osv-scanner JSON fixtures.
const osvHappy = `{"results":[{"packages":[{"package":{"name":"pkg-a"},"vulnerabilities":[{"id":"CVE-2024-1234"}]}]}]}`
const osvEmpty = `{"results":[]}`
const osvMalformed = `{"results": NOT_JSON}`

// TestParseOSVScanner_HappyFinding decodes one vuln correctly.
func TestParseOSVScanner_HappyFinding(t *testing.T) {
	keys, err := parseOSVScanner([]byte(osvHappy))
	if err != nil || len(keys) != 1 {
		t.Fatalf("expected 1 finding, got %v / %v", keys, err)
	}
	if keys[0].Pkg != "pkg-a" || keys[0].ID != "CVE-2024-1234" {
		t.Fatalf("wrong key: %+v", keys[0])
	}
}

// TestParseOSVScanner_EmptyResultsIsClean verifies empty results = no findings.
func TestParseOSVScanner_EmptyResultsIsClean(t *testing.T) {
	keys, err := parseOSVScanner([]byte(osvEmpty))
	if err != nil || len(keys) != 0 {
		t.Fatalf("empty results must yield no findings, got %v / %v", keys, err)
	}
}

// TestParseOSVScanner_MalformedOutputIsError verifies bad JSON -> error.
func TestParseOSVScanner_MalformedOutputIsError(t *testing.T) {
	_, err := parseOSVScanner([]byte(osvMalformed))
	if err == nil {
		t.Fatal("expected error for malformed osv-scanner output")
	}
}

// TestParseVuln_DispatchesToGovulncheck confirms routing for govulncheck.
func TestParseVuln_DispatchesToGovulncheck(t *testing.T) {
	keys, err := parseVuln("govulncheck", []byte(govulnHappy))
	if err != nil || len(keys) != 1 {
		t.Fatalf("dispatch to govulncheck failed, got %v / %v", keys, err)
	}
}

// TestParseVuln_DispatchesToOSVScanner confirms routing for osv-scanner.
func TestParseVuln_DispatchesToOSVScanner(t *testing.T) {
	keys, err := parseVuln("osv-scanner", []byte(osvHappy))
	if err != nil || len(keys) != 1 {
		t.Fatalf("dispatch to osv-scanner failed, got %v / %v", keys, err)
	}
}

// TestVulnDetails_SortedOutput verifies details are emitted in sorted order.
func TestVulnDetails_SortedOutput(t *testing.T) {
	findings := map[vulnKey]bool{
		{Pkg: "z-pkg", ID: "CVE-2"}: true,
		{Pkg: "a-pkg", ID: "CVE-1"}: true,
	}
	lines := vulnDetails(findings)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %v", lines)
	}
	if !strings.HasPrefix(lines[0], "a-pkg") {
		t.Fatalf("expected sorted output, first line: %q", lines[0])
	}
}
