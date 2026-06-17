package unit_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// customGateConfig chdirs into a temp repo whose centinela.toml declares the
// given [[gates.custom]] body and returns the loaded config.
func customGateConfig(t *testing.T, customBody string) *config.Config {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	toml := "[gates]\nfile_size = false\ni18n = true\n\n" + customBody
	if err := os.WriteFile(filepath.Join(dir, "centinela.toml"), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
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

func resultByName(rs []gates.Result, name string) (gates.Result, bool) {
	for _, r := range rs {
		if r.Name == name {
			return r, true
		}
	}
	return gates.Result{}, false
}

// TestRunWithFilterIncludesCustomResults: a passing and a failing custom gate
// both appear in RunWithFilter output with the correct status.
func TestRunWithFilterIncludesCustomResults(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh assumed")
	}
	body := strings.Join([]string{
		"[[gates.custom]]",
		"enabled = true",
		"name = \"passes\"",
		"command = \"true\"",
		"severity = \"fail\"",
		"",
		"[[gates.custom]]",
		"enabled = true",
		"name = \"fails\"",
		"command = \"false\"",
		"severity = \"fail\"",
	}, "\n")
	cfg := customGateConfig(t, body)
	rs := gates.RunWithFilter(cfg, nil)

	p, ok := resultByName(rs, "passes")
	if !ok || p.Status != gates.Pass {
		t.Fatalf("passes gate wrong: %+v ok=%v", p, ok)
	}
	f, ok := resultByName(rs, "fails")
	if !ok || f.Status != gates.Fail {
		t.Fatalf("fails gate wrong: %+v ok=%v", f, ok)
	}
}

// TestWarnSeverityNonBlocking: a non-zero severity=warn custom gate reports Warn,
// which never sets AllPassed to false.
func TestWarnSeverityNonBlocking(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh assumed")
	}
	body := "[[gates.custom]]\nenabled = true\nname = \"nit\"\ncommand = \"false\"\nseverity = \"warn\"\n"
	cfg := customGateConfig(t, body)
	rs := gates.RunWithFilter(cfg, nil)
	r, ok := resultByName(rs, "nit")
	if !ok || r.Status != gates.Warn {
		t.Fatalf("warn gate wrong: %+v ok=%v", r, ok)
	}
	if !gates.AllPassed(rs) {
		t.Fatal("a warn custom gate must not fail the suite")
	}
}
