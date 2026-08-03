package gates

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

func TestRunWithFilter_RegistersTheDocstringGateOnlyWhenEnabled(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n\nfunc Exported() {}\n")
	filter := gitdiff.NewSet([]string{p})

	off := &config.Config{}
	for _, r := range RunWithFilter(off, filter) {
		if r.Name == docstringGate {
			t.Fatal("a new gate must never turn itself on in an existing project")
		}
	}

	found := false
	for _, r := range RunWithFilter(docstringCfg("fail"), filter) {
		if r.Name == docstringGate {
			found = true
			if r.Status != Fail {
				t.Fatalf("status = %v", r.Status)
			}
		}
	}
	if !found {
		t.Fatal("an enabled docstring gate must appear in the result set")
	}
}

func TestDocstringOptions_MapTheConfigKnobsOntoTheScanner(t *testing.T) {
	cfg := docstringCfg("fail")
	cfg.Gates.Docstring.CheckFields = true
	cfg.Gates.Docstring.RequireNamePrefix = true
	off := false
	cfg.Gates.Docstring.IncludeInternal = &off

	opts := DocstringOptions(cfg)
	if !opts.CheckFields || !opts.RequireNamePrefix || opts.IncludeInternal {
		t.Fatalf("options = %+v", opts)
	}
	if len(opts.Roots) != 1 || opts.Roots[0] != "." {
		t.Fatalf("roots = %v", opts.Roots)
	}
}

func TestCheckDocstring_CheckFieldsExtendsTheReportThroughTheGate(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n\n// S is a struct.\ntype S struct {\n\tName string\n}\n")
	cfg := docstringCfg("fail")
	if r := checkDocstring(cfg, gitdiff.NewSet([]string{p})); r.Status != Pass {
		t.Fatalf("check_fields off must pass: %v %v", r.Status, r.Details)
	}
	cfg.Gates.Docstring.CheckFields = true
	r := checkDocstring(cfg, gitdiff.NewSet([]string{p}))
	if r.Status != Fail || !strings.Contains(r.Details[0], "field Name") {
		t.Fatalf("status=%v details=%v", r.Status, r.Details)
	}
}

func TestCheckDocstring_ExposesTheScanReportToTheCLI(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n\nfunc Exported() {}\n")
	r, rep := CheckDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{p}))
	if r.Status != Fail {
		t.Fatalf("status = %v", r.Status)
	}
	if rep.Inspected != 1 || len(rep.Violations) != 1 || rep.Files != 1 {
		t.Fatalf("report = %+v", rep)
	}
}

func TestReportDocstring_MissingScannerAndScannerErrorAreHonest(t *testing.T) {
	rep := reportDocstring(docstringReportFixture(), "fail")
	if rep.Status != Fail || len(rep.Details) != 2 {
		t.Fatalf("result = %+v", rep)
	}
	warn := reportDocstring(docstringReportFixture(), "warn")
	if warn.Status != Warn {
		t.Fatalf("status = %v", warn.Status)
	}
}
