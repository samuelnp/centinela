package orchestration

import (
	"regexp"
	"strings"
	"testing"
)

var datedSuffix = regexp.MustCompile(`-\d{8}$`)

// AC19: the property that matters, not the literal — a built-in default may
// never be a dated snapshot, for any tier or runner.
func TestNoBuiltinModelIDIsADatedSnapshot(t *testing.T) {
	for tier, byRunner := range tierModels {
		for runner, id := range byRunner {
			if datedSuffix.MatchString(id) {
				t.Errorf("%s/%s built-in %q is a dated snapshot", tier, runner, id)
			}
		}
	}
}

// AC20: the claude column is bare family aliases with no version digits.
func TestClaudeColumnIsUndatedFamilyAliases(t *testing.T) {
	want := map[Tier]string{TierReasoning: "opus", TierBalanced: "sonnet", TierFast: "haiku"}
	for tier, alias := range want {
		got := tierModels[tier][RunnerClaude]
		if got != alias {
			t.Errorf("tier %s claude default = %q, want %q", tier, got, alias)
		}
		if strings.ContainsAny(got, "0123456789") {
			t.Errorf("tier %s claude default %q contains version digits", tier, got)
		}
	}
}

// AC23: codex stays empty so rule 4 renders the tier name, never another
// runner's concrete id.
func TestCodexColumnStaysEmptyAndRendersTierName(t *testing.T) {
	for _, tier := range AllowedTiers() {
		if id := tierModels[tier][RunnerCodex]; id != "" {
			t.Errorf("codex column for %s must stay empty, got %q", tier, id)
		}
	}
	ref := ModelReference([]Tier{TierReasoning})
	if !strings.Contains(ref, "reasoning (codex)") {
		t.Errorf("codex must render the tier name, got: %s", ref)
	}
	if !strings.Contains(ref, "opus (claude)") {
		t.Errorf("claude column must render the alias, got: %s", ref)
	}
}

func TestTierModelIDsCoversEveryBuiltin(t *testing.T) {
	ids := TierModelIDs()
	want := 0
	for _, byRunner := range tierModels {
		for _, id := range byRunner {
			if id != "" {
				want++
			}
		}
	}
	if len(ids) != want {
		t.Fatalf("TierModelIDs returned %d ids, table holds %d: %v", len(ids), want, ids)
	}
	for _, id := range ids {
		if id == "" {
			t.Fatal("TierModelIDs must never emit an empty id")
		}
	}
	// Stable order: reasoning first (claude then opencode).
	if ids[0] != "opus" || ids[1] != "anthropic/claude-opus-4-7" {
		t.Fatalf("unstable order: %v", ids)
	}
}
