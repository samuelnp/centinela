package setup

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGoldenParityClaudeOpenCode locks the byte-for-byte managed output of the
// Claude and OpenCode adapters against pre-refactor golden fixtures. It is the
// regression tripwire for the registry refactor.
func TestGoldenParityClaudeOpenCode(t *testing.T) {
	cases := map[string][]string{
		"claude":   {".claude/settings.json"},
		"opencode": {"opencode.json", ".opencode/plugins/centinela.js", "AGENTS.md"},
	}
	root, _ := os.Getwd()
	for agent, files := range cases {
		d := t.TempDir()
		o, _ := os.Getwd()
		os.Chdir(d) //nolint:errcheck
		plan, err := BuildSyncPlan(agent)
		if err != nil {
			os.Chdir(o) //nolint:errcheck
			t.Fatalf("%s build: %v", agent, err)
		}
		if err := ApplySync(plan); err != nil {
			os.Chdir(o) //nolint:errcheck
			t.Fatalf("%s apply: %v", agent, err)
		}
		os.Chdir(o) //nolint:errcheck
		for _, f := range files {
			got, err := os.ReadFile(filepath.Join(d, f))
			if err != nil {
				t.Fatalf("emitted %s/%s: %v", agent, f, err)
			}
			want, err := os.ReadFile(filepath.Join(root, "testdata", "golden", agent, f))
			if err != nil {
				t.Fatalf("golden %s/%s: %v", agent, f, err)
			}
			if string(got) != string(want) {
				t.Fatalf("byte mismatch for %s/%s", agent, f)
			}
		}
	}
}
