package gates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/docstring"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// docstringCfg builds a config with the gate enabled at the given severity and
// roots that accept any path, so fixtures can live in a temp dir.
func docstringCfg(severity string) *config.Config {
	cfg := &config.Config{}
	cfg.Gates.Docstring = config.NormalizeDocstring(config.DocstringConfig{
		Enabled:  true,
		Severity: severity,
		Roots:    []string{"."},
	})
	return cfg
}

// fixture writes body to name under a fresh temp dir the test chdirs into,
// returning the repo-relative slash path.
func fixture(t *testing.T, name, body string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(name)
}

// inDir chdirs into a fresh temp dir for the duration of the test.
func inDir(t *testing.T) { t.Helper(); t.Chdir(t.TempDir()) }

// stubRatchet replaces the gate's own changed-file resolution for one test.
func stubRatchet(t *testing.T, set *gitdiff.Set, degrade string) {
	t.Helper()
	prev := docstringRatchet
	docstringRatchet = func(string) (*gitdiff.Set, string) { return set, degrade }
	t.Cleanup(func() { docstringRatchet = prev })
}

// docstringReportFixture is a report holding one violation and one exemption.
func docstringReportFixture() docstring.Report {
	return docstring.Report{
		Files:      1,
		Inspected:  2,
		Violations: []docstring.Violation{{Path: "a.go", Line: 3, Kind: "func", Name: "F"}},
		Exemptions: []docstring.Exemption{{Path: "a.go", Line: 9, Kind: "var", Name: "V"}},
	}
}
