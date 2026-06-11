package gates

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/samuelnp/centinela/internal/gitdiff"
)

// Scenario identifies one Gherkin scenario by its spec slug (the .feature
// basename without extension) and its scenario name.
type Scenario struct {
	Spec string
	Name string
}

var scenarioLine = regexp.MustCompile(`^\s+Scenario(?: Outline)?:\s*(.+?)\s*$`)
var collapseWS = regexp.MustCompile(`\s+`)

// parseScenarios walks specDir/*.feature and returns every scenario found. When
// filter is non-nil only files whose path is in the diff set are parsed (the
// same diff-aware contract G1 uses). Unreadable files are skipped, not fatal.
func parseScenarios(specDir string, filter *gitdiff.Set) ([]Scenario, error) {
	entries, err := os.ReadDir(specDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Scenario
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".feature") {
			continue
		}
		path := filepath.Join(specDir, e.Name())
		if filter != nil && !filter.Contains(path) {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".feature")
		out = append(out, scanScenarios(path, slug)...)
	}
	return out, nil
}

func scanScenarios(path, slug string) []Scenario {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []Scenario
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if m := scenarioLine.FindStringSubmatch(sc.Text()); m != nil {
			out = append(out, Scenario{Spec: slug, Name: m[1]})
		}
	}
	return out
}

// normalizeScenario produces the canonical form used for case-insensitive
// matching: trim, collapse internal whitespace to single spaces, strip one
// trailing period, lowercase. The same form is applied to spec scenario names
// and to acceptance-test // Scenario: comments so authors can fix mismatches
// deterministically.
func normalizeScenario(s string) string {
	s = collapseWS.ReplaceAllString(strings.TrimSpace(s), " ")
	s = strings.TrimSuffix(s, ".")
	return strings.ToLower(strings.TrimSpace(s))
}
