package main

import "os"

func writeDocsFixture() {
	os.MkdirAll(".workflow", 0755)                                                                                                              //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                                                                          //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                                                                                                             //nolint:errcheck
	os.MkdirAll("specs", 0755)                                                                                                                  //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("# P"), 0644)                                                                                             //nolint:errcheck
	os.WriteFile("ROADMAP.md", []byte("# R"), 0644)                                                                                             //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("# Feature"), 0644)                                                                               //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("# Plan"), 0644)                                                                                     //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: x\n  Scenario: s"), 0644)                                                                  //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P1","features":[{"name":"f"}]}]}`), 0644)                                //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.md", []byte("# A"), 0644)                                                                          //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.json", []byte(`{"role":"senior-product-manager","features":[{"name":"f","dependsOn":[]}]}`), 0644) //nolint:errcheck
	os.WriteFile(".workflow/f.json", []byte(`{"feature":"f","currentStep":"plan","steps":{"plan":{"status":"in-progress"}}}`), 0644)            //nolint:errcheck
}
