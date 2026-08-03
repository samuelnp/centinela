package docstring

import "testing"

func TestScan_GroupedBlockDocCoversEverySpec(t *testing.T) {
	src := "package a\n\n// Limits are the caps.\nconst (\n\tA = 1\n\tB = 2\n)\n\n" +
		"// Vars hold state.\nvar (\n\tC = 1\n)\n"
	rep := scan(t, anyRoot(), writeGo(t, "a.go", src))
	if !rep.OK() || rep.Inspected != 3 {
		t.Fatalf("inspected=%d violations=%v", rep.Inspected, names(rep))
	}
}

func TestScan_UndocumentedSpecInsideUndocumentedBlockIsAViolation(t *testing.T) {
	src := "package a\n\nconst (\n\t// A is first.\n\tA = 1\n\tB = 2\n)\n"
	rep := scan(t, anyRoot(), writeGo(t, "a.go", src))
	if len(rep.Violations) != 1 || rep.Violations[0].Name != "B" {
		t.Fatalf("violations=%v", names(rep))
	}
	if rep.Violations[0].Kind != "const" {
		t.Fatalf("kind=%s", rep.Violations[0].Kind)
	}
}

func TestScan_InterfaceMethodsAreTheContractAndAreChecked(t *testing.T) {
	src := "package a\n\n// I is an interface.\ntype I interface {\n\tDo()\n\t// Done is documented.\n\tDone()\n}\n"
	rep := scan(t, anyRoot(), writeGo(t, "a.go", src))
	if len(rep.Violations) != 1 || rep.Violations[0].Name != "Do" {
		t.Fatalf("violations=%v", names(rep))
	}
	if rep.Violations[0].Kind != "interface method" {
		t.Fatalf("kind=%s", rep.Violations[0].Kind)
	}
}

func TestScan_StructFieldsFollowCheckFields(t *testing.T) {
	src := "package a\n\n// S is a struct.\ntype S struct {\n\tName string\n\thidden int\n}\n"
	p := writeGo(t, "a.go", src)
	if rep := scan(t, anyRoot(), p); !rep.OK() {
		t.Fatalf("check_fields off must pass, got %v", names(rep))
	}
	opts := anyRoot()
	opts.CheckFields = true
	rep := scan(t, opts, p)
	if len(rep.Violations) != 1 || rep.Violations[0].Name != "Name" ||
		rep.Violations[0].Kind != "field" {
		t.Fatalf("violations=%+v", rep.Violations)
	}
}

func TestScan_MethodsOnExportedTypesAreReported(t *testing.T) {
	src := "package a\n\n// G is generic.\ntype G[T any] struct{}\n\nfunc (*G[T]) M() {}\n"
	rep := scan(t, anyRoot(), writeGo(t, "a.go", src))
	if len(rep.Violations) != 1 || rep.Violations[0].Name != "M" ||
		rep.Violations[0].Kind != "method" {
		t.Fatalf("violations=%+v", rep.Violations)
	}
}
