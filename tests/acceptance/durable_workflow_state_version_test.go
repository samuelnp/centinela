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

// Scenario: A future-version state file this binary can still model keeps governing
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

// Scenario: A future-version state file this binary cannot model refuses the write
//
// A future file this binary cannot even unmarshal REFUSES the write and names
// the upgrade. Passing would be a self-service bypass: `.workflow/*.json` is an
// ungoverned write target, so any agent could write a future version over its
// own state file and open the gate. What the degraded path still buys is the
// rest of the contract — Load succeeds, so the hook never claims no workflow
// was started and autostart never forks a duplicate.
func TestAccUnmodellableFutureVersionRefusesAndNamesTheUpgrade(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "delta", `{"schemaVersion":"2.0","feature":"delta",`+
		`"currentStep":"plan","steps":[{"name":"plan","status":["done"]}]}`)

	out, code := dwsPrewrite(t, bin, dir, "internal/foo.go")
	if code == 0 {
		t.Fatalf("an unreadable state file must not open the gate:\n%s", out)
	}
	if !strings.Contains(out, "centinela update") {
		t.Fatalf("the refusal must name the upgrade, got:\n%s", out)
	}
	if strings.Contains(out, "no workflow") || strings.Contains(out, "centinela start") {
		t.Fatalf("the hook must not claim no workflow has been started:\n%s", out)
	}
}
