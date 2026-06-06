package gates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/gitdiff"
)

// gitleaksFinding mirrors the subset of a gitleaks v8 JSON report entry the
// gate consumes. The report is a JSON array; an empty file means "clean".
type gitleaksFinding struct {
	RuleID string `json:"RuleID"`
	File   string `json:"File"`
}

// readSecretsReport decodes the gitleaks JSON report file. A missing or empty
// file is treated as "no findings" (gitleaks writes nothing scannable as such).
// Non-empty, non-JSON content is a parse error so the caller emits Warn.
func readSecretsReport(path string) ([]gitleaksFinding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}
	var findings []gitleaksFinding
	if err := json.Unmarshal(data, &findings); err != nil {
		return nil, fmt.Errorf("parsing gitleaks report: %w", err)
	}
	return findings, nil
}

// retainFindings drops findings that are allowlisted or (when a diff filter is
// active) outside the changed-file set, then renders the survivors as sorted,
// deduplicated detail lines naming the file and matched rule ID.
func retainFindings(findings []gitleaksFinding, allowlist []string, filter *gitdiff.Set) []string {
	out := map[string]bool{}
	for _, f := range findings {
		file := filepath.ToSlash(f.File)
		if filter != nil && !filter.Contains(file) {
			continue
		}
		if allowlisted(f.RuleID, file, allowlist) {
			continue
		}
		out[fmt.Sprintf("%s: rule %s", file, f.RuleID)] = true
	}
	return sortedKeys(out)
}

// allowlisted reports whether a finding is suppressed. An entry containing a
// '*' or '/' is treated as a path glob (filepath.Match against the file);
// otherwise it is matched for exact rule-ID equality.
func allowlisted(ruleID, file string, allowlist []string) bool {
	for _, entry := range allowlist {
		if strings.ContainsAny(entry, "*/") {
			if ok, _ := filepath.Match(entry, file); ok {
				return true
			}
			continue
		}
		if entry == ruleID {
			return true
		}
	}
	return false
}
