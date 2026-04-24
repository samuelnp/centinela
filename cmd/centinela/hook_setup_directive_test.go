package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunHookSetupRoadmapWhenTemplateRenamed(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                             //nolint:errcheck
	os.Chdir(d)                                   //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("x"), 0644) //nolint:errcheck

	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.Contains(out, "CENTINELA DIRECTIVE: roadmap required") {
			t.Fatalf("expected roadmap directive, got %q", out)
		}
		if !strings.Contains(out, "ROADMAP.md") {
			t.Fatalf("expected roadmap panel content, got %q", out)
		}
	})
}

func TestRunHookSetupDirectiveBeforePanel(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                 //nolint:errcheck
	os.Chdir(d)                                       //nolint:errcheck
	os.WriteFile("centinela.toml", []byte("x"), 0644) //nolint:errcheck

	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.HasPrefix(out, "CENTINELA DIRECTIVE: setup required") {
			t.Fatalf("expected plain directive prefix, got %q", out)
		}
		if !strings.Contains(out, "PROJECT.md not found") {
			t.Fatalf("expected setup panel content, got %q", out)
		}
	})
}

func TestRunHookSetupRoadmapAnalysisDirective(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                                                //nolint:errcheck
	os.Chdir(d)                                                                                                                      //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("x"), 0644)                                                                                    //nolint:errcheck
	os.WriteFile("ROADMAP.md", []byte("x"), 0644)                                                                                    //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                                   //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"Phase 0: Bootstrap","features":[{"name":"setup"}]}]}`), 0644) //nolint:errcheck

	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.Contains(out, "roadmap analysis required") {
			t.Fatalf("expected roadmap analysis directive, got %q", out)
		}
		if !strings.Contains(out, "senior PM review required") {
			t.Fatalf("expected roadmap analysis panel content, got %q", out)
		}
	})
}

func TestRunHookSetupRoadmapJSONDirective(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                             //nolint:errcheck
	os.Chdir(d)                                   //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("x"), 0644) //nolint:errcheck
	os.WriteFile("ROADMAP.md", []byte("x"), 0644) //nolint:errcheck

	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.Contains(out, "roadmap json required") {
			t.Fatalf("expected roadmap json directive, got %q", out)
		}
		if !strings.Contains(out, ".workflow/roadmap.json") || !strings.Contains(out, "exact format") {
			t.Fatalf("expected roadmap json panel content, got %q", out)
		}
	})
}

func TestRunHookSetupNoopWhenNotInitialized(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if strings.TrimSpace(out) != "" {
			t.Fatalf("expected no output, got %q", out)
		}
	})
}

func TestRunHookSetupProductionReadinessDirective(t *testing.T) {
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
	os.WriteFile(".workflow/roadmap-quality.md", []byte("x"), 0644)                                                                  //nolint:errcheck
	os.WriteFile(".workflow/roadmap-quality.json", []byte("{}"), 0644)                                                               //nolint:errcheck
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.Contains(out, "production-readiness prompt") {
			t.Fatalf("expected production-readiness directive, got %q", out)
		}
	})
}
