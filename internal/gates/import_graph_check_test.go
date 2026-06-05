package gates

import "testing"

func checkMatrix(t *testing.T) matrix {
	t.Helper()
	m, err := buildMatrix(sampleLayers())
	if err != nil {
		t.Fatalf("buildMatrix: %v", err)
	}
	return m
}

func TestCheckEdges_AllowedAndForbidden(t *testing.T) {
	m := checkMatrix(t)
	pkgs := []pkg{
		{Path: "internal/config", Imports: nil},                                  // leaf
		{Path: "internal/gates", Imports: []string{"internal/config"}},           // domain->leaf OK
		{Path: "internal/workflow", Imports: []string{"internal/gates"}},         // intra-domain OK
		{Path: "cmd/centinela", Imports: []string{"internal/gates", "cmd/sub"}},  // cmd->domain OK + intra
		{Path: "cmd/sub"},                                                        //
	}
	v, unmapped := checkEdges(pkgs, m)
	if len(v) != 0 {
		t.Fatalf("expected no violations, got %v", v)
	}
	if len(unmapped) != 0 {
		t.Fatalf("expected none unmapped, got %v", unmapped)
	}
}

func TestCheckEdges_ForbiddenMessageFormat(t *testing.T) {
	m := checkMatrix(t)
	pkgs := []pkg{
		{Path: "internal/config", Imports: []string{"internal/gates"}}, // leaf->domain FORBIDDEN
		{Path: "internal/gates"},
	}
	v, _ := checkEdges(pkgs, m)
	want := "internal/config -> internal/gates (leaf may not import domain)"
	if len(v) != 1 || v[0] != want {
		t.Fatalf("violation format wrong.\n got: %v\nwant: %q", v, want)
	}
}

func TestCheckEdges_UnmappedAndIgnoredEdges(t *testing.T) {
	m := checkMatrix(t)
	pkgs := []pkg{
		{Path: "internal/ui", Imports: []string{"internal/config"}}, // unmapped importer: edge ignored
		{Path: "internal/gates", Imports: []string{"internal/ui"}},  // edge INTO unmapped: ignored
		{Path: "internal/config"},
	}
	v, unmapped := checkEdges(pkgs, m)
	if len(v) != 0 {
		t.Fatalf("edges touching unmapped pkgs must be ignored, got %v", v)
	}
	if len(unmapped) != 1 || unmapped[0] != "internal/ui" {
		t.Fatalf("expected internal/ui unmapped, got %v", unmapped)
	}
}

func TestCheckEdges_SortedAndDeduped(t *testing.T) {
	m := checkMatrix(t)
	pkgs := []pkg{
		{Path: "internal/config", Imports: []string{"internal/workflow", "internal/gates", "internal/gates"}},
		{Path: "internal/gates"}, {Path: "internal/workflow"},
	}
	v, _ := checkEdges(pkgs, m)
	if len(v) != 2 {
		t.Fatalf("expected 2 deduped violations, got %v", v)
	}
	if v[0] >= v[1] {
		t.Fatalf("violations not sorted: %v", v)
	}
}

