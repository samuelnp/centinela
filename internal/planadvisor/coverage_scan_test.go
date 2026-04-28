package planadvisor

import (
	"os"
	"testing"
)

func TestScanReadsCurrentFeatureArtifacts(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                                                                                    //nolint:errcheck
	os.Chdir(d)                                                                                                                                                          //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                                                                                                   //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                                                                                                                                      //nolint:errcheck
	os.MkdirAll("specs", 0755)                                                                                                                                           //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                                                                       //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n## Problem\ntext\n## Scope\nout of scope\n## Constraints\nsecurity\n## Risks\nregression\n"), 0644) //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("## Acceptance Criteria\ntext\nmobile-first\n"), 0644)                                                                        //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: f\nScenario: x\nGiven a\nWhen b\nThen c\n"), 0644)                                                                  //nolint:errcheck
	os.WriteFile(".workflow/f-edge-cases.md", []byte("## Edge Cases\ninvalid input\nloading state\nempty state\nerror state\n"), 0644)                                   //nolint:errcheck
	c := scan("f")
	if !c.UserFacing || !c.Problem || !c.Scope || !c.Constraints || !c.Risks || !c.Acceptance || !c.EdgeCases || !c.MobileFirst || !c.UXStates {
		t.Fatalf("expected full coverage from feature artifacts, got %+v", c)
	}
}
