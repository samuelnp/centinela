package evidence

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// skelFields renders the skeleton and returns its feature and handoffTo.
func skelFields(t *testing.T, feature string, role Role) (string, string) {
	t.Helper()
	data, err := SchemaSkeleton(feature, role, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Feature   string `json:"feature"`
		HandoffTo string `json:"handoffTo"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("skeleton is not valid JSON: %v\n%s", err, data)
	}
	return got.Feature, got.HandoffTo
}

// With no feature every role prints both unfilled slots, so nothing plausible
// is handed to an author who must decide. merge-steward is the one exception:
// its "complete" is a fixed literal, not a derivation, so a placeholder there
// would replace a correct answer with one the gate then refuses.
func TestSchemaSkeletonNoFeaturePlaceholdersEveryRole(t *testing.T) {
	chdirToTemp(t)
	for _, role := range AllRoles() {
		f, h := skelFields(t, "", role)
		want := unfilledHandoffSlot
		if role == orchestration.RoleMergeSteward {
			want = workflow.TerminalHandoff
		}
		if f != placeholderFeature || h != want {
			t.Fatalf("%s: got %q/%q, want %q/%q", role, f, h, placeholderFeature, want)
		}
	}
}

// The E1/E2 regression at unit level: a feature slug the caller resolved (from
// a `.worktrees/<x>` segment) whose workflow state is NOT readable must degrade
// to the placeholder. Deriving anyway falls through to legacyHandoffForRole and
// prints a stale successor beside a correct-looking slug — more convincing than
// the placeholder it replaced, and the exact value this feature removes.
func TestSchemaSkeletonStatelessFeatureNeverFallsBackToLegacy(t *testing.T) {
	chdirToTemp(t)
	for _, role := range []Role{Role("gatekeeper"), orchestration.RoleQASeniorEngineer, orchestration.RolePlanner} {
		f, h := skelFields(t, "never-started", role)
		if f != placeholderFeature || h != unfilledHandoffSlot {
			t.Fatalf("%s: stateless feature derived %q/%q", role, f, h)
		}
		if h == legacyHandoffForRole(role) {
			t.Fatalf("%s: legacy chain %q leaked into the skeleton", role, h)
		}
	}
}

// One derivation, three callers: with state on disk the schema skeleton, the
// `evidence init` prefill and the completion gate must all name the same
// successor. Any divergence is the bug, whichever of the three is wrong.
func TestSchemaSkeletonAgreesWithInitAndGate(t *testing.T) {
	chdirToTemp(t)
	if err := workflow.Save(workflow.New("demo-internal")); err != nil {
		t.Fatal(err)
	}
	for _, role := range []Role{Role("gatekeeper"), orchestration.RoleSeniorEngineer, orchestration.RolePlanner} {
		f, h := skelFields(t, "demo-internal", role)
		if f != "demo-internal" {
			t.Fatalf("%s: feature = %q", role, f)
		}
		if want := Skeleton("demo-internal", role, "1.0.0").HandoffTo; h != want {
			t.Fatalf("%s: schema %q != init prefill %q", role, h, want)
		}
		if err := workflow.CheckHandoffTo("demo-internal", stepForRole(role), role, h); err != nil {
			t.Fatalf("%s: gate refuses the value the CLI printed: %v", role, err)
		}
	}
}

// The skeleton is embedded in prompts, so it must stay parseable and must not
// carry any pre-filled authoring content.
func TestSchemaSkeletonStaysEmptyAndParseable(t *testing.T) {
	chdirToTemp(t)
	data, err := SchemaSkeleton("", orchestration.RoleBigThinker, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) || strings.Contains(string(data), "docs/plans/") {
		t.Fatalf("skeleton poisoned or unparseable: %s", data)
	}
}
