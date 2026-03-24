package projectstage

import (
	"os"
	"testing"
)

func TestParseProjectStage(t *testing.T) {
	if Parse("x") != Greenfield {
		t.Fatal("missing stage should default greenfield")
	}
	if Parse("Project Stage: existing") != Existing {
		t.Fatal("existing should parse")
	}
	if Parse("**Project Stage:** greenfield") != Greenfield {
		t.Fatal("markdown stage should parse")
	}
	if Parse("Project Stage: weird") != Greenfield {
		t.Fatal("unknown stage should default greenfield")
	}
}

func TestLoadProjectStage(t *testing.T) {
	d := t.TempDir()
	path := d + "/PROJECT.md"
	if _, err := Load(path + ".missing"); err == nil {
		t.Fatal("expected missing file error")
	}
	os.WriteFile(path, []byte("Project Stage: existing\n"), 0644) //nolint:errcheck
	stage, err := Load(path)
	if err != nil || stage != Existing {
		t.Fatalf("expected existing stage from file: %q %v", stage, err)
	}
}
