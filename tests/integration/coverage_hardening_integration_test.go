package integration_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const repoRoot = "../.."

func TestCoverageGate_FloorUnchanged(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "scripts/check-coverage.sh"))
	if err != nil {
		t.Fatalf("read check-coverage.sh: %v", err)
	}
	if !strings.Contains(string(data), "MIN_COVERAGE:-95.0") {
		t.Fatal("check-coverage.sh default floor must remain MIN_COVERAGE:-95.0")
	}
}

func TestNewTestFiles_WithinG1Limit(t *testing.T) {
	patterns := []string{
		"*_more_test.go",
		"*_edge_test.go",
		"*_cover_test.go",
		"cov2_*_test.go",
		"coverage_*_test.go",
	}
	searchDirs := []string{
		filepath.Join(repoRoot, "cmd"),
		filepath.Join(repoRoot, "internal"),
	}
	found := 0
	for _, dir := range searchDirs {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			for _, pat := range patterns {
				matched, _ := filepath.Match(pat, d.Name())
				if matched {
					found++
					if n := countLines(t, path); n > 100 {
						t.Errorf("%s: %d lines > 100 (G1)", path, n)
					}
					break
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
	if found == 0 {
		t.Skip("no coverage-hardening test files found; pattern may need update")
	}
	t.Logf("checked %d coverage-hardening colocated test files", found)
}

func countLines(t *testing.T, path string) int {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
	}
	return n
}
