package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunHookSetupRoadmapQualityDirective(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                                                //nolint:errcheck
	os.Chdir(d)                                                                                                                      //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("x"), 0644)                                                                                    //nolint:errcheck
	os.WriteFile("ROADMAP.md", []byte("x"), 0644)                                                                                    //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                                   //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"Phase 0: Bootstrap","features":[{"name":"setup"}]}]}`), 0644) //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.md", []byte("x"), 0644)                                                                 //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.json", []byte("{}"), 0644)                                                              //nolint:errcheck
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.Contains(out, "roadmap quality required") {
			t.Fatalf("expected roadmap quality directive, got %q", out)
		}
		if !strings.Contains(out, "overall score") {
			t.Fatalf("expected roadmap quality panel content, got %q", out)
		}
	})
}
