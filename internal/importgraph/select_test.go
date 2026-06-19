package importgraph

import (
	"errors"
	"testing"
)

func TestSelect_Explicit(t *testing.T) {
	for _, k := range []string{"go", "node", "python"} {
		p, err := Select(".", k, "", nil, okRunner(""))
		if err != nil || p.Name() != k {
			t.Fatalf("explicit %s: %v %v", k, p, err)
		}
	}
}

func TestSelect_Script(t *testing.T) {
	p, err := Select(".", "script", "", []string{"x"}, okRunner(""))
	if err != nil || p.Name() != "script" {
		t.Fatalf("script: %v %v", p, err)
	}
	if _, err := Select(".", "script", "", nil, nil); err == nil {
		t.Fatal("empty script_command must error")
	}
}

func TestSelect_AutoDetectsGo(t *testing.T) {
	// The package dir resolves to this repo's go.mod via walk-up; nil runner
	// exercises the default-runner branch.
	p, err := Select(".", "", "", nil, nil)
	if err != nil || p.Name() != "go" {
		t.Fatalf("auto-detect go: %v %v", p, err)
	}
}

func TestSelect_NoProvider(t *testing.T) {
	if _, err := Select(t.TempDir(), "", "", nil, nil); !errors.Is(err, ErrNoProvider) {
		t.Fatalf("want ErrNoProvider, got %v", err)
	}
}

func TestSelect_UnknownProvider(t *testing.T) {
	if _, err := Select(".", "cobol", "", nil, nil); err == nil {
		t.Fatal("unknown provider must error")
	}
}
