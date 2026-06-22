package acceptance_test

import "testing"

// runReconstructBin runs `centinela reconstruct [args...]` in dir against the
// shared real binary, returning stdout and the exit code.
func runReconstructBin(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	return runCent(t, buildAnalyzeBin(t), dir, append([]string{"reconstruct"}, args...)...)
}

// goNtierReconInventory promotes three targets: cmd/app (command),
// internal/handler (endpoint), internal/service (module).
const goNtierReconInventory = `{"schemaVersion":1,"primaryLanguage":"Go",
"packages":["cmd/app","internal/handler","internal/service"],
"graph":{"kind":"go-packages","edges":[]}}`

// docOnlyReconInventory has no behavioral packages, so zero targets are selected.
const docOnlyReconInventory = `{"schemaVersion":1,"primaryLanguage":"Markdown",
"packages":["docs","readme"],"graph":{"kind":"none","edges":[]}}`

// polyglotReconInventory has an empty Go graph; targets come from the express
// manifest + the api package, exercising the non-Go path.
const polyglotReconInventory = `{"schemaVersion":1,"primaryLanguage":"JavaScript",
"packages":["src/api/users","src/util"],
"manifests":[{"kind":"npm","path":"package.json","framework":"express","deps":["express"]}],
"graph":{"kind":"none","edges":[]}}`
