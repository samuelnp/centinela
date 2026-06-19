package acceptance_test

import (
	"path/filepath"
	"testing"
)

// writeAnalysis writes an analysis.json fixture into dir and returns its path.
func writeAnalysis(t *testing.T, dir, body string) string {
	t.Helper()
	writeFile(t, dir, "analysis.json", body)
	return filepath.Join(dir, "analysis.json")
}

// runSynthesizeBin runs `centinela synthesize [args...]` in dir against the
// shared real binary, returning stdout and the exit code.
func runSynthesizeBin(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	return runCent(t, buildAnalyzeBin(t), dir, append([]string{"synthesize"}, args...)...)
}

const railsInventory = `{"schemaVersion":1,"primaryLanguage":"Ruby",
"manifests":[{"kind":"gem","path":"Gemfile","deps":["rails"]}],
"packages":["app/models","app/controllers","app/views"],
"graph":{"kind":"none","edges":[]}}`

const ecsInventory = `{"schemaVersion":1,"primaryLanguage":"GDScript",
"packages":["src/systems","src/components","src/entities"],
"graph":{"kind":"none","edges":[]}}`

const ambiguousInventory = `{"schemaVersion":1,"primaryLanguage":"Go",
"packages":["service","domain"],"graph":{"kind":"none","edges":[]}}`
