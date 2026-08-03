package gates

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gitdiff"
)

func TestCheckDocstring_UndocumentedExportInAChangedFileFails(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n\nfunc Exported() {}\n")
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{p}))
	if r.Status != Fail {
		t.Fatalf("status = %v, want Fail", r.Status)
	}
	if len(r.Details) != 1 || !strings.Contains(r.Details[0], "a.go:3: func Exported") {
		t.Fatalf("details = %v", r.Details)
	}
	if AllPassed([]Result{r}) {
		t.Fatal("a Fail must not pass the run")
	}
}

func TestCheckDocstring_DocumentedExportsPassAndNameTheCount(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n\n// Exported does a thing.\nfunc Exported() {}\n")
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{p}))
	if r.Status != Pass {
		t.Fatalf("status = %v (%s)", r.Status, r.Message)
	}
	if !strings.Contains(r.Message, "All 1 exported identifiers across 1 changed Go file(s)") {
		t.Fatalf("message = %q", r.Message)
	}
}

func TestCheckDocstring_WarnSeverityReportsWithoutBlocking(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n\nfunc Exported() {}\n")
	r := checkDocstring(docstringCfg("warn"), gitdiff.NewSet([]string{p}))
	if r.Status != Warn {
		t.Fatalf("status = %v, want Warn", r.Status)
	}
	if len(r.Details) != 1 || !strings.Contains(r.Details[0], "Exported") {
		t.Fatalf("details = %v", r.Details)
	}
	if !AllPassed([]Result{r}) {
		t.Fatal("Warn must not fail the run")
	}
}

func TestCheckDocstring_LegacyUnchangedFileIsNeverScanned(t *testing.T) {
	inDir(t)
	fixture(t, "legacy.go", "package a\n\nfunc Legacy() {}\nfunc AlsoLegacy() {}\n")
	changed := fixture(t, "new.go", "package a\n\n// New is documented.\nfunc New() {}\n")
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{changed}))
	if r.Status != Pass {
		t.Fatalf("status = %v (%s) details=%v", r.Status, r.Message, r.Details)
	}
	for _, d := range r.Details {
		if strings.Contains(d, "legacy.go") {
			t.Fatalf("legacy file leaked into the report: %s", d)
		}
	}
}

func TestCheckDocstring_ExemptionIsListedOnAPassingRun(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n\n//centinela:nodoc\nfunc Exported() {}\n")
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{p}))
	if r.Status != Pass {
		t.Fatalf("status = %v (%s)", r.Status, r.Message)
	}
	if len(r.Details) != 1 || !strings.Contains(r.Details[0], "exempt via //centinela:nodoc") {
		t.Fatalf("an invisible exemption is a hole: %v", r.Details)
	}
}

func TestCheckDocstring_UnparseableFileIsReportedNotSkipped(t *testing.T) {
	inDir(t)
	p := fixture(t, "broken.go", "package a\n\nfunc (((\n")
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{p}))
	if r.Status == Pass || r.Status == Skip {
		t.Fatalf("status = %v, want a reported failure", r.Status)
	}
	if len(r.Details) == 0 || !strings.Contains(r.Details[0], "broken.go") {
		t.Fatalf("details = %v", r.Details)
	}
}
