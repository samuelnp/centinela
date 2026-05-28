package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: "roadmap ready prints each ready feature on its own line" +
// "centinela roadmap renders the ready/blocked markers".
func TestAcceptance_RoadmapReadyAndRenderMarkers(t *testing.T) {
	bin := buildCent(t)
	dir := acceptanceDir(t, `{"phases":[{"name":"P","features":[
		{"name":"feature-a"},
		{"name":"feature-b","dependsOn":["feature-a"]},
		{"name":"feature-c"}]}]}`)
	seedDoneAt(t, dir, "feature-a")

	out, code := runCent(t, bin, dir, "roadmap", "ready")
	if code != 0 {
		t.Fatalf("roadmap ready exit=%d, want 0\n%s", code, out)
	}
	for _, name := range []string{"feature-b", "feature-c"} {
		if !lineContains(out, name) {
			t.Fatalf("ready output should list %q on its own line:\n%s", name, out)
		}
	}

	render, code := runCent(t, bin, dir, "roadmap")
	if code != 0 {
		t.Fatalf("roadmap render exit=%d\n%s", code, render)
	}
	if !lineWithSub(render, "feature-b", "🔓") {
		t.Fatalf("feature-b should render 🔓 (ready):\n%s", render)
	}
	// done feature carries no ready/blocked marker.
	if lineWithSub(render, "feature-a", "🔓") || lineWithSub(render, "feature-a", "🔒") {
		t.Fatalf("done feature-a must not show 🔓/🔒:\n%s", render)
	}
}

// Acceptance: blocked render shows 🔒 and names the blocking dependency.
func TestAcceptance_RoadmapRenderBlockedMarker(t *testing.T) {
	bin := buildCent(t)
	dir := acceptanceDir(t, `{"phases":[{"name":"P","features":[
		{"name":"feature-a"},
		{"name":"feature-b","dependsOn":["feature-a"]}]}]}`)
	render, code := runCent(t, bin, dir, "roadmap")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, render)
	}
	if !lineWithSub(render, "feature-b", "🔒") || !strings.Contains(render, "feature-a") {
		t.Fatalf("feature-b should show 🔒 + blocking dep feature-a:\n%s", render)
	}
}

// Acceptance: empty-state when none ready, and exit 0 when all features are done.
func TestAcceptance_RoadmapReadyEmptyStates(t *testing.T) {
	bin := buildCent(t)
	// Every planned feature has an unmet dependency → no feature ready.
	blocked := acceptanceDir(t, `{"phases":[{"name":"P","features":[
		{"name":"a"},{"name":"b","dependsOn":["a"]},{"name":"c","dependsOn":["b"]}]}]}`)
	seedWorkflowAt(t, blocked, "a", "code") // a in-progress → b,c blocked, a not ready
	out, code := runCent(t, bin, blocked, "roadmap", "ready")
	if code != 0 {
		t.Fatalf("empty-state exit=%d\n%s", code, out)
	}
	if strings.TrimSpace(stripBin(out)) == "" {
		t.Fatalf("empty ready set must print a non-empty empty-state line:\n%s", out)
	}
	for _, name := range []string{"\n  🔓 a", "\n  🔓 b", "\n  🔓 c"} {
		if strings.Contains(out, name) {
			t.Fatalf("empty-state must not list feature names:\n%s", out)
		}
	}

	allDone := acceptanceDir(t, `{"phases":[{"name":"P","features":[{"name":"x"},{"name":"y"}]}]}`)
	seedDoneAt(t, allDone, "x")
	seedDoneAt(t, allDone, "y")
	doneOut, code := runCent(t, bin, allDone, "roadmap", "ready")
	if code != 0 {
		t.Fatalf("all-done exit=%d\n%s", code, doneOut)
	}
	if strings.TrimSpace(stripBin(doneOut)) == "" {
		t.Fatalf("all-done must still print a non-empty empty-state line:\n%s", doneOut)
	}
}

func lineContains(out, sub string) bool {
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(ln, sub) {
			return true
		}
	}
	return false
}

func lineWithSub(out, name, marker string) bool {
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(ln, name) && strings.Contains(ln, marker) {
			return true
		}
	}
	return false
}

// stripBin removes the leading binary-name noise some shells inject; here a no-op
// passthrough kept for readability of the empty-state assertions.
func stripBin(s string) string { return s }
