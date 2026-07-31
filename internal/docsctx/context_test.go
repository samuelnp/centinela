package docsctx

import (
	"os"
	"strings"
	"testing"
)

// seed writes the docs-step inputs present=true entries under a temp CWD.
func seed(t *testing.T, files map[string]string) {
	t.Helper()
	t.Chdir(t.TempDir())
	for path, body := range files {
		dir := path[:strings.LastIndex(path, "/")]
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func allInputs() map[string]string {
	return map[string]string{
		"docs/features/f.md":       "# brief\n",
		"docs/plans/f.md":          "# plan\n",
		"specs/f.feature":          "Feature: f\n",
		".workflow/f-changelog.md": "- feat: f\n",
	}
}

func TestLoadHappyPath(t *testing.T) {
	seed(t, allInputs())
	ctx, err := Load("f")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for name, s := range map[string]Section{"brief": ctx.Brief, "plan": ctx.Plan, "spec": ctx.Spec, "changelog": ctx.Changelog} {
		if !s.Present || s.Body == "" || s.Source == "" {
			t.Fatalf("%s section incomplete: %+v", name, s)
		}
	}
}

// Every missing required input is aggregated into ONE error naming all paths;
// the optional changelog never appears in it.
func TestLoadAggregatesAllMissingRequired(t *testing.T) {
	required := []string{"docs/features/f.md", "docs/plans/f.md", "specs/f.feature"}
	for _, omit := range append([]string{"ALL"}, required...) {
		files := allInputs()
		if omit == "ALL" {
			files = map[string]string{".workflow/f-changelog.md": "- x\n"}
		} else {
			delete(files, omit)
		}
		seed(t, files)
		_, err := Load("f")
		if err == nil {
			t.Fatalf("omit=%s: expected error", omit)
		}
		for _, path := range required {
			_, present := files[path]
			if !present && !strings.Contains(err.Error(), path) {
				t.Fatalf("omit=%s: error must name %s, got %v", omit, path, err)
			}
			if present && strings.Contains(err.Error(), path) {
				t.Fatalf("omit=%s: error must not name present %s, got %v", omit, path, err)
			}
		}
		if strings.Contains(err.Error(), "changelog") {
			t.Fatalf("omit=%s: optional changelog must not be in the error, got %v", omit, err)
		}
	}
}
