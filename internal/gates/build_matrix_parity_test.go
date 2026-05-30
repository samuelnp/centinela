package gates

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/samuelnp/centinela/internal/config"
)

// repoRoot resolves the worktree root from internal/gates.
func repoRoot(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..")
}

// tomlTargets reads [gates.build].targets from centinela.toml.
func tomlTargets(t *testing.T, root string) map[string]bool {
	t.Helper()
	var cfg config.Config
	if _, err := toml.DecodeFile(filepath.Join(root, config.Filename), &cfg); err != nil {
		t.Fatalf("decode toml: %v", err)
	}
	set := map[string]bool{}
	for _, tg := range cfg.Gates.Build.Targets {
		set[tg.GOOS+"/"+tg.GOARCH] = true
	}
	return set
}

var listRe = regexp.MustCompile(`(?m)\[([a-z0-9, ]+)\]`)

// matrixTargets extracts the {goos,goarch} cross-product from release.yml.
func matrixTargets(t *testing.T, root string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatalf("read release.yml: %v", err)
	}
	var goos, goarch []string
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.Contains(line, "matrix:") {
			continue
		}
		lists := listRe.FindAllStringSubmatch(line, -1)
		if len(lists) != 2 {
			t.Fatalf("expected goos+goarch lists, got %d in %q", len(lists), line)
		}
		goos = splitList(lists[0][1])
		goarch = splitList(lists[1][1])
	}
	if len(goos) == 0 || len(goarch) == 0 {
		t.Fatal("no matrix found in release.yml")
	}
	set := map[string]bool{}
	for _, o := range goos {
		for _, a := range goarch {
			set[o+"/"+a] = true
		}
	}
	return set
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func TestBuildMatrixParity(t *testing.T) {
	root := repoRoot(t)
	got := tomlTargets(t, root)
	want := matrixTargets(t, root)
	missing := diffKeys(want, got)
	extra := diffKeys(got, want)
	if len(missing) > 0 || len(extra) > 0 {
		t.Fatalf("target drift: in release.yml not centinela.toml=%v; in centinela.toml not release.yml=%v", missing, extra)
	}
}

func diffKeys(a, b map[string]bool) []string {
	var out []string
	for k := range a {
		if !b[k] {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
