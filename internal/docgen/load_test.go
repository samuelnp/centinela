package docgen

import (
	"os"
	"testing"
)

func TestLoadData(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	writeFixture(t)
	x, err := LoadData("T")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if x.Scenarios != 1 || len(x.RoadmapNodes) != 1 || len(x.Evidence) != 1 || len(x.States) != 1 {
		t.Fatalf("unexpected aggregates: %#v", x)
	}
}
