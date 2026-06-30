package setup

import (
	"os"
	"path/filepath"
	"testing"
)

// buildOpenCodeConfig with local=nil must be byte-for-byte identical to the
// committed no-local opencode.json golden — the zero-config regression tripwire,
// mirroring the host-harness golden parity guard.
func TestBuildOpenCodeConfigNilLocalGoldenParity(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, "opencode.json")
	changed, data, err := buildOpenCodeConfig(path, nil)
	if err != nil || !changed {
		t.Fatalf("build: changed=%v err=%v", changed, err)
	}
	root, _ := os.Getwd()
	want, err := os.ReadFile(filepath.Join(root, "testdata", "golden", "opencode", "opencode.json"))
	if err != nil {
		t.Fatalf("golden read: %v", err)
	}
	if string(data) != string(want) {
		t.Fatalf("nil-local output drifted from the no-local golden")
	}
}
