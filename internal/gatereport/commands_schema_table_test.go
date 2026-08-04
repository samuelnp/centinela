package gatereport

import (
	"encoding/json"
	"strings"
	"testing"
)

// The full accept/reject matrix for the one place that knows the shape.
func TestValidateCommandsSchemaTable(t *testing.T) {
	runSchemaCases(t, []schemaCase{
		{"commands as object", `{"argv":["c","validate"]}`, false},
		{"commands as string", `"centinela validate"`, false},
		{"commands as number", `7`, false},
		{"entry as null", `[null]`, false},
		{"entry as array", `[["centinela","validate"]]`, false},
		{"entry as string", `["centinela validate"]`, false},
		{"argv absent", `[{"exitCode":0}]`, false},
		{"argv empty", `[{"argv":[],"exitCode":0}]`, false},
		{"argv null", `[{"argv":null,"exitCode":0}]`, false},
		{"argv string", `[{"argv":"centinela validate","exitCode":0}]`, false},
		{"argv non-string element", `[{"argv":["c",1],"exitCode":0}]`, false},
		{"argv nested", `[{"argv":[["c"]],"exitCode":0}]`, false},
		{"durationMs string", `[{"argv":["c","validate"],"exitCode":0,"durationMs":"12"}]`, false},
		{"second entry malformed", `[{"argv":["c","validate"],"exitCode":0},{"argv":["c"]}]`, false},
		{"unknown key ignored", `[{"argv":["c","validate"],"exitCode":0,"note":"x"}]`, true},
		{"empty-string argv element", `[{"argv":[""],"exitCode":0}]`, true},
		{"duplicate entries", `[{"argv":["c","validate"],"exitCode":0},{"argv":["c","validate"],"exitCode":0}]`, true},
		{"duplicate key last wins", `[{"argv":["c","validate"],"exitCode":1,"exitCode":0}]`, true},
		{"durationMs integer", `[{"argv":["c","validate"],"exitCode":0,"durationMs":84210}]`, true},
	})
}

// The index makes a long record navigable; without it an operator must guess
// which entry the gate refused.
func TestValidateCommandsSchemaNamesTheOffendingIndex(t *testing.T) {
	err := ValidateCommandsSchema(json.RawMessage(
		`[{"argv":["c","validate"],"exitCode":0},{"argv":["c","validate"],"exitCode":null}]`))
	if err == nil || !strings.Contains(err.Error(), "commands[1]") {
		t.Fatalf("want commands[1] named, got %v", err)
	}
}
