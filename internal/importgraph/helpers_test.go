package importgraph

import "testing"

// swapOnPath replaces the onPath lookup for the duration of a test so backend
// tool-detection branches are exercised without the real tools installed.
func swapOnPath(t *testing.T, f func(string) bool) {
	t.Helper()
	orig := onPath
	onPath = f
	t.Cleanup(func() { onPath = orig })
}

// okRunner returns a Runner that always yields out with no error.
func okRunner(out string) Runner {
	return func(string, ...string) ([]byte, error) { return []byte(out), nil }
}

// errRunner returns a Runner that always fails with e.
func errRunner(e error) Runner {
	return func(string, ...string) ([]byte, error) { return nil, e }
}
