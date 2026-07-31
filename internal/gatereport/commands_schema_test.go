package gatereport

import (
	"encoding/json"
	"strings"
	"testing"
)

// schemaCase is one raw `commands` array and whether the shared shape rule
// must accept it.
type schemaCase struct {
	name string
	raw  string
	ok   bool
}

// runSchemaCases applies the shape rule to every case and reports the verdict
// it should have produced.
func runSchemaCases(t *testing.T, cases []schemaCase) {
	t.Helper()
	for _, tc := range cases {
		err := ValidateCommandsSchema(json.RawMessage(tc.raw))
		if tc.ok && err != nil {
			t.Errorf("%s: want accept, got %v", tc.name, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("%s: want reject, got accept (%s)", tc.name, tc.raw)
		}
	}
}

// TestValidateCommandsSchemaRejectsNullExitCode pins the bypass that defeated
// the present-key rule entirely: `"exitCode": null` satisfied BOTH halves of
// it, because json.Unmarshal of the literal null into an int is a documented
// no-op that returns nil. The entry then decoded to ExitCode 0 — "the gates
// passed" — and Assess called the report grounded. An explicit null is the
// same defect the absent-key check exists to close, four characters wider.
func TestValidateCommandsSchemaRejectsNullExitCode(t *testing.T) {
	runSchemaCases(t, []schemaCase{
		{"null exitCode", `[{"argv":["centinela","validate"],"exitCode":null}]`, false},
		{"absent exitCode", `[{"argv":["centinela","validate"]}]`, false},
		{"string exitCode", `[{"argv":["centinela","validate"],"exitCode":"0"}]`, false},
		{"bool exitCode", `[{"argv":["centinela","validate"],"exitCode":true}]`, false},
		{"fractional exitCode", `[{"argv":["centinela","validate"],"exitCode":1.5}]`, false},
		{"float-zero exitCode", `[{"argv":["centinela","validate"],"exitCode":0.0}]`, false},
		{"exponent exitCode", `[{"argv":["centinela","validate"],"exitCode":1e2}]`, false},
		{"bignum exitCode", `[{"argv":["centinela","validate"],"exitCode":99999999999999999999}]`, false},
		{"null durationMs", `[{"argv":["c","validate"],"exitCode":0,"durationMs":null}]`, false},
		{"integer exitCode", `[{"argv":["centinela","validate"],"exitCode":0}]`, true},
		{"nonzero exitCode", `[{"argv":["centinela","validate"],"exitCode":1}]`, true},
		{"negative exitCode", `[{"argv":["centinela","validate"],"exitCode":-1}]`, true},
	})
}

// The refusal must name the field and say null, or an author reading it will
// re-type the same value assuming a different key was at fault.
func TestNullExitCodeErrorNamesTheField(t *testing.T) {
	err := ValidateCommandsSchema(json.RawMessage(`[{"argv":["c","validate"],"exitCode":null}]`))
	if err == nil {
		t.Fatal("null exitCode must be refused")
	}
	for _, want := range []string{"commands[0]", "exitCode", "null"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error missing %q: %v", want, err)
		}
	}
}

// The shape rule deliberately does NOT judge whether enough was recorded: an
// absent, empty or null ARRAY is an admissibility question Assess answers with
// a remedy. Pinning it keeps the division of labour from drifting.
func TestValidateCommandsSchemaDefersAdmissibility(t *testing.T) {
	runSchemaCases(t, []schemaCase{
		{"absent array", ``, true},
		{"empty array", `[]`, true},
		{"null array", `null`, true},
	})
}
