package gates

import (
	"runtime"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

func cfgWith(gs ...config.CustomGate) *config.Config {
	c := &config.Config{}
	c.Gates.CustomGates = gs
	return c
}

func byName(rs []Result) map[string]Result {
	m := make(map[string]Result, len(rs))
	for _, r := range rs {
		m[r.Name] = r
	}
	return m
}

// TestCustomGatesRunsBoth: two enabled gates produce two Results with the
// correct names and statuses; a failing one does not suppress a passing one.
func TestCustomGatesRunsBoth(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh assumed")
	}
	rs := customGates(cfgWith(
		config.CustomGate{Enabled: true, Name: "ok", Command: "true", Severity: "fail", Output: "blob", TimeoutSeconds: 5},
		config.CustomGate{Enabled: true, Name: "bad", Command: "false", Severity: "fail", Output: "blob", TimeoutSeconds: 5},
	), nil)
	if len(rs) != 2 {
		t.Fatalf("want 2 results, got %d", len(rs))
	}
	m := byName(rs)
	if m["ok"].Status != Pass {
		t.Fatalf("ok status = %v", m["ok"].Status)
	}
	if m["bad"].Status != Fail {
		t.Fatalf("bad status = %v", m["bad"].Status)
	}
}

// TestCustomGatesSkipsDisabled: a disabled entry never runs.
func TestCustomGatesSkipsDisabled(t *testing.T) {
	rs := customGates(cfgWith(
		config.CustomGate{Enabled: false, Name: "off", Command: "false", Severity: "fail", Output: "blob", TimeoutSeconds: 5},
	), nil)
	if len(rs) != 0 {
		t.Fatalf("disabled gate should be skipped, got %d results", len(rs))
	}
}

// TestCustomGatesEmpty: no configured gates returns an empty slice.
func TestCustomGatesEmpty(t *testing.T) {
	if rs := customGates(&config.Config{}, nil); len(rs) != 0 {
		t.Fatalf("want no results, got %d", len(rs))
	}
}

// TestCustomGatesDiffAwareEnv: a diff_aware gate with a non-nil filter receives
// the changed file set via CENTINELA_CHANGED_FILES.
func TestCustomGatesDiffAwareEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh assumed")
	}
	filter := gitdiff.NewSet([]string{"x.go", "y.go"})
	rs := customGates(cfgWith(config.CustomGate{
		Enabled: true, Name: "diff", DiffAware: true, TimeoutSeconds: 5,
		Severity: "fail", Output: "blob",
		Command: "test -n \"$CENTINELA_CHANGED_FILES\" && exit 1 || exit 0",
	}), filter)
	if len(rs) != 1 || rs[0].Status != Fail {
		t.Fatalf("diff-aware env not seen by command: %+v", rs)
	}
	if !strings.Contains(rs[0].Message, "exit 1") {
		t.Fatalf("unexpected message: %q", rs[0].Message)
	}
}
