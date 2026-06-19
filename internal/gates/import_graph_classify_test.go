package gates

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/importgraph"
)

func TestClassifyLoadError_NoProviderWarns(t *testing.T) {
	r := classifyLoadError(importgraph.ErrNoProvider)
	if r.Status != Warn || !strings.Contains(r.Message, "no provider matched") {
		t.Fatalf("no-provider must self-skip with Warn: %v %q", r.Status, r.Message)
	}
}

func TestClassifyLoadError_ToolMissingWarns(t *testing.T) {
	r := classifyLoadError(&importgraph.ToolMissingError{Tool: "depcruise"})
	if r.Status != Warn || !strings.Contains(r.Message, "depcruise") {
		t.Fatalf("tool-missing must Warn and name the tool: %v %q", r.Status, r.Message)
	}
}

func TestClassifyLoadError_OtherFails(t *testing.T) {
	r := classifyLoadError(errors.New("boom"))
	if r.Status != Fail || !strings.HasPrefix(r.Message, "import_graph: ") {
		t.Fatalf("a real load error must Fail: %v %q", r.Status, r.Message)
	}
}

// TestCheckImportGraph_NoManifestSelfSkips drives the full gate over a
// manifest-less directory: detection finds no provider, so the gate self-skips
// with a non-failing Warn (the fix that stops non-Go projects hard-failing).
func TestCheckImportGraph_NoManifestSelfSkips(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	layers := []config.Layer{{Name: "all", Paths: []string{"**"}, Allow: nil}}
	r := checkImportGraph(igCfg("", layers), nil)
	if r.Status != Warn || !strings.Contains(r.Message, "no provider matched") {
		t.Fatalf("no-manifest dir should self-skip Warn, got %v: %q", r.Status, r.Message)
	}
}
