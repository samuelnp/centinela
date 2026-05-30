package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/samuelnp/centinela/internal/evidence"
)

const claimStubs = "outputs-not-stubs"

var (
	emptyTestFunc  = regexp.MustCompile(`func\s+Test\w*\s*\([^)]*\)\s*\{\s*\}`)
	assertionToken = regexp.MustCompile(`\b(t\.(Error|Errorf|Fatal|Fatalf|Fail|FailNow)|assert\.|require\.|want|expect)`)
	testFuncDecl   = regexp.MustCompile(`func\s+Test\w*\s*\(`)
)

// checkStubs confirms each Go outputs file carries substantive content. Test
// files must contain real assertions (no empty `func Test…(){}` bodies);
// non-test files only fail when whitespace/boilerplate-only. Conservative: a
// tiny interface/helper file is never flagged. Non-Go files are not inspected.
func checkStubs(root, role string, ev *evidence.RoleEvidence) Check {
	c := Check{Claim: claimStubs, Role: role}
	if len(ev.Outputs) == 0 {
		c.Status = StatusSkip
		c.Detail = "no outputs to inspect"
		return c
	}
	for _, rel := range ev.Outputs {
		if !strings.HasSuffix(rel, ".go") {
			continue
		}
		if offending := inspectGoFile(filepath.Join(root, rel), rel); offending != "" {
			c.Status = StatusFail
			c.Detail = offending
			return c
		}
	}
	c.Status = StatusPass
	c.Detail = "all inspected outputs carry substantive content"
	return c
}

// inspectGoFile returns a non-empty offending-detail string when path is a
// stub, or "" when it is substantive (or unreadable, treated as not-a-stub to
// stay conservative).
func inspectGoFile(path, rel string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	src := string(data)
	if strings.HasSuffix(rel, "_test.go") {
		return inspectTestFile(src, rel)
	}
	if strings.TrimSpace(stripPackageLine(src)) == "" {
		return fmt.Sprintf("%s is empty beyond its package declaration", rel)
	}
	return ""
}

// inspectTestFile flags test files whose Test funcs are empty-bodied or that
// declare tests yet carry no recognizable assertion.
func inspectTestFile(src, rel string) string {
	if emptyTestFunc.MatchString(src) {
		return fmt.Sprintf("%s contains an empty-bodied test function (no assertions)", rel)
	}
	if testFuncDecl.MatchString(src) && !assertionToken.MatchString(src) {
		return fmt.Sprintf("%s declares tests but contains no assertions", rel)
	}
	return ""
}

// stripPackageLine removes the `package …` clause and comments so an
// otherwise-empty file is recognized as blank.
func stripPackageLine(src string) string {
	var b strings.Builder
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") || strings.HasPrefix(trimmed, "//") {
			continue
		}
		b.WriteString(line)
	}
	return b.String()
}
