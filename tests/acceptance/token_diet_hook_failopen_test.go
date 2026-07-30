// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"os"
	"runtime"
	"testing"
)

// Scenario: Every uncertainty fails open and renders
func TestTD_EveryUncertaintyFailsOpenAndRenders(t *testing.T) {
	bin := tdBuildBin(t)

	setupState := map[string]func(t *testing.T, dir string){
		"absent": func(t *testing.T, dir string) {},
		"corrupt": func(t *testing.T, dir string) {
			mustWrite(t, tdDigestPath(dir), "{not valid json")
		},
		"unreadable": func(t *testing.T, dir string) {
			if runtime.GOOS == "windows" || os.Geteuid() == 0 {
				t.Skip("permission bits are not enforced for this test's identity")
			}
			mustWrite(t, tdDigestPath(dir), `{"sessionId":"s-1","digest":"x"}`)
			if err := os.Chmod(tdDigestPath(dir), 0o000); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { os.Chmod(tdDigestPath(dir), 0o644) }) //nolint:errcheck
		},
		"current": func(t *testing.T, dir string) {
			digest := tdCurrentDigest(t, dir)
			tdWriteDigestState(t, dir, "s-1", digest)
		},
	}
	payloads := map[string]string{
		"valid JSON with session id": tdSessionPayload("s-1"),
		"empty":                      "",
		"not JSON":                   "not json at all",
		"JSON without session_id":    "{}",
	}
	cases := []struct{ state, payload string }{
		{"absent", "valid JSON with session id"},
		{"corrupt", "valid JSON with session id"},
		{"unreadable", "valid JSON with session id"},
		{"current", "empty"},
		{"current", "not JSON"},
		{"current", "JSON without session_id"},
	}
	for _, tc := range cases {
		dir := tdRepo(t)
		tdWriteRoadmap(t, dir, tdRoadmap1)
		setupState[tc.state](t, dir)

		out, code := tdHookContext(t, bin, dir, payloads[tc.payload])
		if code != 0 {
			t.Fatalf("state=%s payload=%s: hook must exit 0, got %d: %s", tc.state, tc.payload, code, out)
		}
		mustContain(t, out, tdRoadmapLine)
	}
}
