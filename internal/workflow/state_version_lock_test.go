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
var wantWorkflowFields = []string{
	"schemaVersion", "feature", "startedAt", "currentStep", "steps",
	"stepOrder", "orchestrationMode", "enforcementProfile", "archetype",
	"worktreePath", "driverModel", "validateContract", "planContract",
	"revisions", "modelRoutes",
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
// Unexported fields (loadedDigest, unmodellable) carry no tag and never reach
// the file, so they are skipped.
func workflowJSONFields() []string {
	rt := reflect.TypeOf(Workflow{})
	out := make([]string, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		tag := rt.Field(i).Tag.Get("json")
		name, _, _ := strings.Cut(tag, ",")
		if name == "" || name == "-" {
			continue
		}
		out = append(out, name)
	}
	return out
}
