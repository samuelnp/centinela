package docstring

import (
	"go/ast"
	"os"
	"path/filepath"
	"testing"
)

func TestScan_ImportsAndEmbeddedInterfacesAreNotIdentifiers(t *testing.T) {
	src := "package a\n\nimport \"io\"\n\n// I is an interface.\ntype I interface {\n\tio.Reader\n}\n"
	rep := scan(t, anyRoot(), writeGo(t, "a.go", src))
	if !rep.OK() {
		t.Fatalf("violations=%v", names(rep))
	}
	if rep.Inspected != 1 {
		t.Fatalf("only the type itself is an identifier, inspected=%d", rep.Inspected)
	}
}

func TestScan_QualifiedReceiverIsNotAnExportedLocalType(t *testing.T) {
	src := "package a\n\nimport \"io\"\n\nfunc (x io.Reader) M() {}\n"
	rep := scan(t, anyRoot(), writeGo(t, "a.go", src))
	if rep.Inspected != 0 || !rep.OK() {
		t.Fatalf("inspected=%d violations=%v", rep.Inspected, names(rep))
	}
}

func TestExportedReceiver_RejectsMissingAndEmptyReceiverLists(t *testing.T) {
	if exportedReceiver(nil) {
		t.Fatal("nil receiver list is not an exported receiver")
	}
	if exportedReceiver(&ast.FieldList{}) {
		t.Fatal("empty receiver list is not an exported receiver")
	}
	if baseTypeName(&ast.BasicLit{}) != "" {
		t.Fatal("an unrecognized receiver expression has no base type name")
	}
}

func TestScan_UnreadableFileIsReportedAsAParseError(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "d.go")
	if err := os.Mkdir(p, 0o755); err != nil {
		t.Fatal(err)
	}
	rep := scan(t, anyRoot(), filepath.ToSlash(p))
	if len(rep.ParseErrors) != 1 || rep.OK() {
		t.Fatalf("parse errors=%+v", rep.ParseErrors)
	}
}

func TestFirstLine_TrimsToTheFirstLineOnly(t *testing.T) {
	if got := firstLine("one\ntwo\n"); got != "one" {
		t.Fatalf("got %q", got)
	}
	if got := firstLine("  only  "); got != "only" {
		t.Fatalf("got %q", got)
	}
}

func TestSortReport_OrdersByPathThenLine(t *testing.T) {
	r := Report{
		Violations: []Violation{
			{Path: "b.go", Line: 1}, {Path: "a.go", Line: 9}, {Path: "a.go", Line: 2},
		},
		Exemptions: []Exemption{
			{Path: "b.go", Line: 1}, {Path: "a.go", Line: 9}, {Path: "a.go", Line: 2},
		},
	}
	sortReport(&r)
	want := []struct {
		path string
		line int
	}{{"a.go", 2}, {"a.go", 9}, {"b.go", 1}}
	for i, w := range want {
		if r.Violations[i].Path != w.path || r.Violations[i].Line != w.line {
			t.Fatalf("violations[%d] = %+v", i, r.Violations[i])
		}
		if r.Exemptions[i].Path != w.path || r.Exemptions[i].Line != w.line {
			t.Fatalf("exemptions[%d] = %+v", i, r.Exemptions[i])
		}
	}
}
