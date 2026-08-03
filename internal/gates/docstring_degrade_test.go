package gates

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/docstring"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// errScanner always fails, standing in for a scanner that cannot run.
type errScanner struct{}

func (errScanner) Scan([]string, docstring.Options) (docstring.Report, error) {
	return docstring.Report{}, errBoom{}
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }

// swapGoScanner installs s as the Go scanner for the duration of the test.
func swapGoScanner(t *testing.T, s docstring.Scanner) {
	t.Helper()
	prev, _ := docstring.For(docstring.GoLang)
	if s == nil {
		docstring.Unregister(docstring.GoLang)
	} else {
		docstring.Register(docstring.GoLang, s)
	}
	t.Cleanup(func() { docstring.Register(docstring.GoLang, prev) })
}

func TestCheckDocstring_ScannerErrorFailsAndNamesTheCause(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n")
	swapGoScanner(t, errScanner{})
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{p}))
	if r.Status != Fail || !strings.Contains(r.Message, "boom") {
		t.Fatalf("status=%v message=%q", r.Status, r.Message)
	}
}

func TestCheckDocstring_MissingScannerSkipsInsteadOfPassing(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n")
	swapGoScanner(t, nil)
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{p}))
	if r.Status != Skip || !strings.Contains(r.Message, "No docstring scanner registered") {
		t.Fatalf("status=%v message=%q", r.Status, r.Message)
	}
}
