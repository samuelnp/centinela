package docstring

import (
	"strings"
	"testing"
)

func TestScan_NodocDirectiveExemptsAndIsReported(t *testing.T) {
	src := "package a\n\n//centinela:nodoc\nfunc Exported() {}\n"
	rep := scan(t, anyRoot(), writeGo(t, "a.go", src))
	if !rep.OK() {
		t.Fatalf("nodoc must not be a violation, got %v", names(rep))
	}
	if len(rep.Exemptions) != 1 || rep.Exemptions[0].Name != "Exported" {
		t.Fatalf("exemptions=%+v", rep.Exemptions)
	}
	if got := rep.ExemptionLines()[0]; !strings.Contains(got, "exempt via //centinela:nodoc") {
		t.Fatalf("exemption line: %s", got)
	}
}

func TestScan_NodocOnATrailingLineCommentAlsoExempts(t *testing.T) {
	src := "package a\n\nvar Exported = 1 //centinela:nodoc\n"
	rep := scan(t, anyRoot(), writeGo(t, "a.go", src))
	if !rep.OK() || len(rep.Exemptions) != 1 {
		t.Fatalf("violations=%v exemptions=%+v", names(rep), rep.Exemptions)
	}
}

func TestScan_DocumentedIdentifierIsNeverListedAsAnExemption(t *testing.T) {
	src := "package a\n\n// Exported is documented.\n//centinela:nodoc\nfunc Exported() {}\n"
	rep := scan(t, anyRoot(), writeGo(t, "a.go", src))
	if !rep.OK() || len(rep.Exemptions) != 0 {
		t.Fatalf("documented must win over nodoc: %+v", rep.Exemptions)
	}
}

func TestDocumented_RequireNamePrefixIsOffByDefault(t *testing.T) {
	src := "package a\n\n// does a thing.\nfunc Exported() {}\n"
	p := writeGo(t, "a.go", src)
	if rep := scan(t, anyRoot(), p); !rep.OK() {
		t.Fatalf("prefix off must pass, got %v", names(rep))
	}
	opts := anyRoot()
	opts.RequireNamePrefix = true
	if rep := scan(t, opts, p); len(rep.Violations) != 1 {
		t.Fatalf("prefix on must flag, got %v", names(rep))
	}
}

func TestDocumented_RequireNamePrefixAcceptsTheNameAlone(t *testing.T) {
	src := "package a\n\n// Exported\nfunc Exported() {}\n"
	opts := anyRoot()
	opts.RequireNamePrefix = true
	if rep := scan(t, opts, writeGo(t, "a.go", src)); !rep.OK() {
		t.Fatalf("bare name must satisfy the prefix rule, got %v", names(rep))
	}
}

func TestScan_EmptyDocCommentIsNotDocumentation(t *testing.T) {
	src := "package a\n\n//\nfunc Exported() {}\n"
	if rep := scan(t, anyRoot(), writeGo(t, "a.go", src)); rep.OK() {
		t.Fatal("an empty comment must not count as documentation")
	}
}

// F2 regression at the unit tier: go/doc accepts a trailing line comment.
func TestScan_TrailingLineCommentCountsAsDocumentation(t *testing.T) {
	src := "package a\n\nconst Trailing = 1 // Trailing is the answer.\n\n" +
		"var VarTrailing = 2 // VarTrailing is a var.\n"
	rep := scan(t, anyRoot(), writeGo(t, "a.go", src))
	if !rep.OK() || len(rep.Exemptions) != 0 {
		t.Fatalf("violations=%v exemptions=%+v", names(rep), rep.Exemptions)
	}
	if rep.Inspected != 2 {
		t.Fatalf("inspected=%d", rep.Inspected)
	}
}

func TestScan_TrailingCommentOnAStructFieldCountsWhenFieldsAreChecked(t *testing.T) {
	src := "package a\n\n// S is a struct.\ntype S struct {\n\tName string // Name is the name.\n\tAge int\n}\n"
	opts := anyRoot()
	opts.CheckFields = true
	rep := scan(t, opts, writeGo(t, "a.go", src))
	if len(rep.Violations) != 1 || rep.Violations[0].Name != "Age" {
		t.Fatalf("violations=%+v", rep.Violations)
	}
}
