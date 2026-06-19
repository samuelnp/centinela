package importgraph

import (
	"os"
	"path/filepath"
	"testing"
)

func mkManifest(t *testing.T, files ...string) string {
	t.Helper()
	d := t.TempDir()
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(d, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

func TestDetectKind(t *testing.T) {
	cases := []struct {
		files []string
		want  string
	}{
		{[]string{"go.mod"}, "go"},
		{[]string{"package.json"}, "node"},
		{[]string{"pyproject.toml"}, "python"},
		{[]string{"requirements.txt"}, "python"},
		{[]string{"setup.py"}, "python"},
		{[]string{"go.mod", "package.json"}, "go"}, // precedence go > node
		{nil, ""},
	}
	for _, c := range cases {
		if got := detectKind(mkManifest(t, c.files...)); got != c.want {
			t.Errorf("detectKind(%v)=%q want %q", c.files, got, c.want)
		}
	}
}

func TestDetectKind_WalksUp(t *testing.T) {
	d := mkManifest(t, "go.mod")
	sub := filepath.Join(d, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if detectKind(sub) != "go" {
		t.Fatal("must walk up to the nearest ancestor manifest")
	}
}

func TestHasFile_DirIsNotAManifest(t *testing.T) {
	d := t.TempDir()
	if err := os.Mkdir(filepath.Join(d, "go.mod"), 0o755); err != nil {
		t.Fatal(err)
	}
	if hasFile(d, "go.mod") {
		t.Fatal("a directory named go.mod must not count as a manifest file")
	}
}
