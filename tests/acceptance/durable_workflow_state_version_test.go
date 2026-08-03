// Acceptance: specs/durable-workflow-state.feature
//
// Binary-driven proof for the three scenarios with no colocated counterpart.
// The binary is built from ./cmd/centinela into a temp dir (dmrBuildBin) —
// never the installed one, which predates the schema version. Every fixture is
// a local temp directory; no test here touches the network.
package acceptance_test

import (
	"strings"
	"testing"
)

// dwsFuture is a state file from a Centinela that this binary does not have:
// schema version 99, plus a field it cannot model.
const dwsFuture = `{"schemaVersion":99,"feature":"delta","startedAt":"2026-01-01T00:00:00Z",` +
	`"currentStep":"code","stepOrder":["plan","code","tests","validate","docs"],` +
	`"steps":{},"futureFieldThisBinaryCannotModel":{"a":[1,2]}}`

// Scenario: A future-version state file is refused on save with an actionable message
func TestAccFutureVersionSaveIsRefused(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "delta", dwsFuture)

	out := dmrRefused(t, bin, dir, "route", "set", "delta", "senior-engineer",
		"balanced", "--reason", "config-only change")
	for _, want := range []string{
		".workflow/delta.json", // the file
		"schema version 99",    // what it carries
		"schema version 1",     // what this binary understands
		"centinela update",     // the fix
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("refusal must name %q, got:\n%s", want, out)
		}
	}
	if got := dmrState(t, dir, "delta"); got != dwsFuture {
		t.Fatalf("a refused file must be byte-for-byte unchanged, got:\n%s", got)
	}
}

// Scenario: A future-version state file does not block file writes
func TestAccFutureVersionDoesNotBlockWrites(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "delta", dwsFuture)

	dmrOK(t, bin, dir, "status", "delta")
	out, code := dwsPrewrite(t, bin, dir, "internal/foo.go")
	if code != 0 {
		t.Fatalf("a governed write must be allowed, got exit %d:\n%s", code, out)
	}
	if strings.Contains(out, "no workflow") || strings.Contains(out, "BLOCKED") {
		t.Fatalf("the hook must not report that no workflow has been started:\n%s", out)
	}
}

// The same guarantee for a future file this binary cannot even unmarshal — the
// case that used to fail Load, empty the active set and block every write.
func TestAccUnmodellableFutureVersionDoesNotBlockWrites(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "delta", `{"schemaVersion":"2.0","feature":"delta",`+
		`"currentStep":"plan","steps":[{"name":"plan","status":["done"]}]}`)

	out, code := dwsPrewrite(t, bin, dir, "internal/foo.go")
	if code != 0 {
		t.Fatalf("a write must not be blocked by a file this binary cannot model, "+
			"got exit %d:\n%s", code, out)
	}
	if strings.Contains(out, "no workflow") || strings.Contains(out, "BLOCKED") {
		t.Fatalf("the hook must not report that no workflow has been started:\n%s", out)
	}
}
