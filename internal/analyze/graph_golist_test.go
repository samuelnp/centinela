package analyze

import (
	"os"
	"path/filepath"
	"testing"
)

// stubGoBin writes a fake `go` onto PATH so golist invokes it instead of the
// real toolchain, giving deterministic control over each subcommand's outcome.
func stubGoBin(t *testing.T, body string) {
	t.Helper()
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "go"), []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", d+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestGoGraph_PackagesErrorBestEffort(t *testing.T) {
	// `go list -m` succeeds (module printed) but `go list -json ./...` fails, so
	// goGraph keeps the module yet records the packages failure as a Note with no
	// edges — the second golist error branch, distinct from the ModulePath one.
	stubGoBin(t, `case "$*" in
  *"-m"*) echo fakemod ;;
  *) echo "boom" 1>&2; exit 7 ;;
esac
`)
	g := goGraph()
	if g.Module != "fakemod" {
		t.Fatalf("expected module from go list -m, got %#v", g)
	}
	if g.Note == "" || len(g.Edges) != 0 {
		t.Fatalf("expected packages failure recorded as Note with no edges: %#v", g)
	}
}
