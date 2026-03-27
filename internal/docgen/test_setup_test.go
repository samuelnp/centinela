package docgen

import (
	"os"
	"testing"
)

func writeFixture(t *testing.T) {
	t.Helper()
	os.MkdirAll(".workflow", 0755)                                                                                                                                                        //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                                                                                                                    //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                                                                                                                                                       //nolint:errcheck
	os.MkdirAll("specs", 0755)                                                                                                                                                            //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("# P"), 0644)                                                                                                                                       //nolint:errcheck
	os.WriteFile("ROADMAP.md", []byte("# R"), 0644)                                                                                                                                       //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("# Feature"), 0644)                                                                                                                         //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("# Plan"), 0644)                                                                                                                               //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: x\n  Scenario: s"), 0644)                                                                                                            //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P1","features":[{"name":"f"}]}]}`), 0644)                                                                          //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.md", []byte("# A"), 0644)                                                                                                                    //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.json", []byte(`{"role":"senior-product-manager","features":[{"name":"f","dependsOn":[]}]}`), 0644)                                           //nolint:errcheck
	os.WriteFile(".workflow/f.json", []byte(`{"feature":"f","currentStep":"code","steps":{"code":{"status":"in-progress"}}}`), 0644)                                                      //nolint:errcheck
	os.WriteFile(".workflow/f-senior-engineer.json", []byte(`{"feature":"f","step":"code","role":"senior-engineer","handoffTo":"qa-senior","outputs":["cmd/centinela/start.go"]}`), 0644) //nolint:errcheck
}
