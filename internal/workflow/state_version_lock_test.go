package workflow

import (
	"reflect"
	"strings"
	"testing"
)

// fieldsLockedAtVersion is the SchemaVersion wantWorkflowFields describes. The
// two are updated together, never separately.
const fieldsLockedAtVersion = 1

// wantWorkflowFields is the golden list of every JSON key Workflow models.
//
// This exists because an equal-version Save re-marshals from the struct: any
// on-disk key the struct does not model is DROPPED, silently. So adding a field
// to Workflow without bumping SchemaVersion ships a binary that looks
// compatible and quietly destroys data written by the newer one — the version
// scheme only works if the constant moves with the shape. This test is the
// mechanism that makes a forgotten bump loud.
// profileContract arrived from guided-by-default while this feature was in
// flight — the exact drift this list exists to make loud.
//
// It stays at version 1 for one reason only: version 1 HAS NOT SHIPPED YET, so
// it can still absorb the field. No released binary carries the version check
// at all, which is also why a bump would buy nothing here — v0.55.6 drops both
// profileContract AND schemaVersion on a load-save round-trip whatever number
// this constant holds. Do not read this as "additive fields are free": once
// version 1 is released, the strict rule stated in schema_version.go and in the
// failure message below applies, and a shape change owes a bump.
var wantWorkflowFields = []string{
	"schemaVersion", "feature", "startedAt", "currentStep", "steps",
	"stepOrder", "orchestrationMode", "enforcementProfile", "archetype",
	"worktreePath", "driverModel", "validateContract", "planContract",
	"profileContract", "revisions", "modelRoutes",
}

func TestWorkflowStructFieldsAreVersionLocked(t *testing.T) {
	if SchemaVersion != fieldsLockedAtVersion {
		t.Fatalf("SchemaVersion is %d but the golden field list describes %d — "+
			"update wantWorkflowFields and fieldsLockedAtVersion in the same commit",
			SchemaVersion, fieldsLockedAtVersion)
	}
	got := workflowJSONFields()
	if strings.Join(got, ",") == strings.Join(wantWorkflowFields, ",") {
		return
	}
	t.Fatalf("Workflow's JSON shape changed at schema version %d.\n got: %v\nwant: %v\n"+
		"An equal-version Save drops every on-disk key this struct does not model, "+
		"so a shape change without a SchemaVersion bump is a silent data-loss "+
		"release. Bump SchemaVersion (and add its migration in schema_version.go), "+
		"then update wantWorkflowFields.", SchemaVersion, got, wantWorkflowFields)
}

// workflowJSONFields lists Workflow's marshalled keys in declaration order.
//
// Skipping is by EXPORTEDNESS, not by tag: an unexported field (loadedDigest,
// unmodellable) never marshals, but an exported one always does — and if its
// tag names nothing, encoding/json falls back to the Go field name. Deciding
// this on the tag alone left a hole: an exported field with no json tag, or
// one whose tag supplies only options and no name, reached the state file
// while passing the lock — so the list could drift from the real marshalled
// shape exactly as the lock promises it cannot.
func workflowJSONFields() []string {
	rt := reflect.TypeOf(Workflow{})
	out := make([]string, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		switch name {
		case "-":
			continue
		case "":
			name = f.Name // encoding/json's own fallback
		}
		out = append(out, name)
	}
	return out
}
