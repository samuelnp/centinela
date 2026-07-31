package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunDocsLint_JSONCarriesTheVerdictAndCounts(t *testing.T) {
	docsLintRepo(t, "fail")
	p := srcFile(t, "a.go", "package a\n\nfunc Exported() {}\n")
	stubDocsLintDiff(t, p)
	docsLintJSON = true
	var err error
	out := captureStdout(t, func() { err = runDocsLint(nil, nil) })
	if err == nil {
		t.Fatal("json mode must still exit non-zero on Fail")
	}
	for _, want := range []string{`"scope": "changed"`, `"status": "fail"`,
		`"undocumented": 1`, `"inspected": 1`, "src/a.go:3: func Exported"} {
		if !strings.Contains(out, want) {
			t.Fatalf("json missing %q:\n%s", want, out)
		}
	}
}

func TestRunDocsLint_FullReportsTheBacklogAndAlwaysExitsZero(t *testing.T) {
	docsLintRepo(t, "fail")
	srcFile(t, "legacy.go", "package a\n\nfunc Legacy() {}\nfunc AlsoLegacy() {}\n")
	docsLintFull = true
	var err error
	out := captureStdout(t, func() { err = runDocsLint(nil, nil) })
	if err != nil {
		t.Fatalf("a full scan is a report, never a gate: %v", err)
	}
	if !strings.Contains(out, "2 undocumented of 2 exported identifiers") {
		t.Fatalf("backlog count missing:\n%s", out)
	}
	if !strings.Contains(out, "src/legacy.go:3: func Legacy") {
		t.Fatalf("backlog detail missing:\n%s", out)
	}
}

func TestRunDocsLint_FullListsExemptionsAndStatusIsReport(t *testing.T) {
	docsLintRepo(t, "fail")
	srcFile(t, "a.go", "package a\n\n//centinela:nodoc\nfunc Exempted() {}\n")
	docsLintFull = true
	var err error
	out := captureStdout(t, func() { err = runDocsLint(nil, nil) })
	if err != nil || !strings.Contains(out, "exempt via //centinela:nodoc") {
		t.Fatalf("err=%v out=\n%s", err, out)
	}

	docsLintJSON = true
	out = captureStdout(t, func() { err = runDocsLint(nil, nil) })
	if err != nil {
		t.Fatalf("full json must exit 0: %v", err)
	}
	for _, want := range []string{`"scope": "full"`, `"status": "report"`, `"exempt": 1`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json missing %q:\n%s", want, out)
		}
	}
}

func TestRunDocsLint_ConfigErrorIsSurfaced(t *testing.T) {
	chdir(t, t.TempDir())
	if err := os.WriteFile("centinela.toml",
		[]byte("[gates.docstring]\nenabled = true\nseverity = \"nope\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resetDocsLintFlags()
	if err := runDocsLint(nil, nil); err == nil ||
		!strings.Contains(err.Error(), "gates.docstring.severity") {
		t.Fatalf("err = %v", err)
	}
}
