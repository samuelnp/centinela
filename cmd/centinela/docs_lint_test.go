package main

import (
	"strings"
	"testing"
)

func TestRunDocsLint_RejectsChangedAndFullTogether(t *testing.T) {
	docsLintRepo(t, "fail")
	docsLintChanged, docsLintFull = true, true
	if err := runDocsLint(nil, nil); err == nil ||
		!strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunDocsLint_ExitsNonZeroOnFail(t *testing.T) {
	docsLintRepo(t, "fail")
	p := srcFile(t, "a.go", "package a\n\nfunc Exported() {}\n")
	stubDocsLintDiff(t, p)
	var err error
	out := captureStdout(t, func() { err = runDocsLint(nil, nil) })
	if err == nil {
		t.Fatal("severity fail must exit non-zero on an undocumented export")
	}
	if !strings.Contains(out, "Exported") {
		t.Fatalf("fail output missing the identifier:\n%s", out)
	}
}

func TestRunDocsLint_ExitsZeroOnWarnWhileStillReporting(t *testing.T) {
	docsLintRepo(t, "warn")
	p := srcFile(t, "a.go", "package a\n\nfunc Exported() {}\n")
	stubDocsLintDiff(t, p)
	var err error
	out := captureStdout(t, func() { err = runDocsLint(nil, nil) })
	if err != nil {
		t.Fatalf("severity warn must exit 0: %v", err)
	}
	if !strings.Contains(out, "docstring-gate") {
		t.Fatalf("warn output:\n%s", out)
	}
}

func TestRunDocsLint_PassesWhenTheChangedFileIsDocumented(t *testing.T) {
	docsLintRepo(t, "fail")
	p := srcFile(t, "a.go", "package a\n\n// Exported does a thing.\nfunc Exported() {}\n")
	stubDocsLintDiff(t, p)
	var err error
	out := captureStdout(t, func() { err = runDocsLint(nil, nil) })
	if err != nil {
		t.Fatalf("documented code must pass: %v\n%s", err, out)
	}
}

func TestRunDocsLint_LegacyFileOutsideTheDiffIsNeverInspected(t *testing.T) {
	docsLintRepo(t, "fail")
	srcFile(t, "legacy.go", "package a\n\nfunc Legacy() {}\n")
	p := srcFile(t, "a.go", "package a\n\n// Exported does a thing.\nfunc Exported() {}\n")
	stubDocsLintDiff(t, p)
	var err error
	out := captureStdout(t, func() { err = runDocsLint(nil, nil) })
	if err != nil {
		t.Fatalf("the ratchet must ignore untouched legacy files: %v\n%s", err, out)
	}
	if strings.Contains(out, "legacy.go") {
		t.Fatalf("legacy file leaked into the gate report:\n%s", out)
	}
}
