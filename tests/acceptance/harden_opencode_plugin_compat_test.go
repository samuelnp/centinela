package acceptance_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/harden-opencode-plugin-compat.feature
func TestOpenCodePlugin_FilePathFallbackKeysPresent(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	setup.EnsureOpenCodePlugin()                             //nolint:errcheck
	data, _ := os.ReadFile(".opencode/plugins/centinela.js") //nolint:errcheck
	plugin := string(data)

	checks := []string{
		`args.filePath ||`,
		`args.file_path ||`,
		`args.path ||`,
		`args.filename ||`,
		`args.file ||`,
		`nested.filePath ||`,
		`nested.file_path ||`,
		`nested.path ||`,
	}
	for _, c := range checks {
		if !strings.Contains(plugin, c) {
			t.Fatalf("plugin missing compatibility key %q", c)
		}
	}
}
