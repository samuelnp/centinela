package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// oversizedGo returns a >100-line Go source body so the file_size gate fails.
func oversizedGo(extra int) string {
	var b strings.Builder
	b.WriteString("package big\n")
	for i := 0; i < 110+extra; i++ {
		b.WriteString("// filler line to exceed the 100-line file-size limit\n")
	}
	return b.String()
}

// tempRepo chdirs into a fresh repo with file_size enabled and one oversized
// .go file under internal/, returning the loaded cfg. Chdir is reverted on
// cleanup. severity governs the audit_baseline gate.
func tempRepo(t *testing.T, severity string, files map[string]string) *config.Config {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	toml := "[gates]\nfile_size = true\n\n[gates.audit_baseline]\nenabled = true\nseverity = \"" +
		severity + "\"\nbaseline_path = \".workflow/audit-baseline.json\"\n"
	write(t, dir, "centinela.toml", toml)
	for name, body := range files {
		write(t, dir, name, body)
	}
	wd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load: %v", err)
	}
	return cfg
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
