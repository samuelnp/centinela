package gates

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// acceptanceHeader captures the spec slug from a // Acceptance: header. specs?
// tolerates the known spec/ (singular) typo; \S+? stops at .feature so trailing
// annotations like " (AC4, AC5)" after the filename are ignored.
var acceptanceHeader = regexp.MustCompile(`^//\s*Acceptance:\s*specs?/(\S+?)\.feature`)
var scenarioComment = regexp.MustCompile(`^//\s*Scenario:\s*(.+?)\s*$`)

// coveredScenarios scans testDir/*.go and returns coverage keyed by
// [slug][normalizedName]. A // Scenario: comment is only counted under the most
// recent // Acceptance: header seen in the same file; a comment with no header
// above it is ignored.
func coveredScenarios(testDir string) (map[string]map[string]bool, error) {
	entries, err := os.ReadDir(testDir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]map[string]bool{}, nil
		}
		return nil, err
	}
	covered := map[string]map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		scanCoverage(filepath.Join(testDir, e.Name()), covered)
	}
	return covered, nil
}

func scanCoverage(path string, covered map[string]map[string]bool) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	slug := ""
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if m := acceptanceHeader.FindStringSubmatch(line); m != nil {
			slug = m[1]
			continue
		}
		if slug == "" {
			continue
		}
		if m := scenarioComment.FindStringSubmatch(line); m != nil {
			if covered[slug] == nil {
				covered[slug] = map[string]bool{}
			}
			covered[slug][normalizeScenario(m[1])] = true
		}
	}
}

// uncovered returns the scenarios with no matching (slug, normalized-name)
// entry in covered, preserving each scenario's original name for reporting.
func uncovered(scenarios []Scenario, covered map[string]map[string]bool) []Scenario {
	var out []Scenario
	for _, s := range scenarios {
		if !covered[s.Spec][normalizeScenario(s.Name)] {
			out = append(out, s)
		}
	}
	return out
}
