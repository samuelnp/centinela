package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gitdiff"
)

// nested writes body at a nested path under the test's CWD and returns it.
func nested(t *testing.T, rel, body string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(rel), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rel, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(rel)
}

// F1 regression, on the GATE path: a vendored tree under a configured root is
// a legacy backlog the ratchet was never asked to open.
func TestCheckDocstring_VendoredCodeIsNotInspected(t *testing.T) {
	inDir(t)
	v := nested(t, "vendor/x/v.go", "package v\n\nfunc VendoredExport() {}\n")
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{v}))
	if r.Status != Skip {
		t.Fatalf("status = %v (%s) details=%v", r.Status, r.Message, r.Details)
	}
}

// F1 regression, on the GATE path: a deliberately-invalid testdata fixture
// cannot carry //centinela:nodoc and stay invalid, so scanning it would be a
// hard failure with no opt-out.
func TestCheckDocstring_TestdataFixturesAreNotInspected(t *testing.T) {
	inDir(t)
	bad := nested(t, "testdata/bad.go", "package t\n\nfunc (((\n")
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{bad}))
	if r.Status != Skip {
		t.Fatalf("status = %v (%s) details=%v", r.Status, r.Message, r.Details)
	}
	for _, d := range r.Details {
		if strings.Contains(d, "unparseable") {
			t.Fatalf("testdata fixture reported as unparseable: %s", d)
		}
	}
}

// F1 regression: the gate and the --full report must agree on what is source.
func TestCheckDocstring_GateAndReportShareOneExclusionSet(t *testing.T) {
	inDir(t)
	good := nested(t, "src/ok.go", "package a\n\n// Ok is documented.\nfunc Ok() {}\n")
	excluded := []string{
		nested(t, "vendor/v.go", "package v\n\nfunc V() {}\n"),
		nested(t, "testdata/t.go", "package t\n\nfunc T() {}\n"),
		nested(t, "node_modules/n.go", "package n\n\nfunc N() {}\n"),
		nested(t, ".worktrees/f/w.go", "package w\n\nfunc W() {}\n"),
		nested(t, "dist/d.go", "package d\n\nfunc D() {}\n"),
	}
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet(append(excluded, good)))
	if r.Status != Pass {
		t.Fatalf("status = %v (%s) details=%v", r.Status, r.Message, r.Details)
	}
	if !strings.Contains(r.Message, "All 1 exported identifiers") {
		t.Fatalf("only src/ok.go should have been inspected: %q", r.Message)
	}
}
