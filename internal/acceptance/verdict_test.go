package acceptance

import (
	"strings"
	"testing"
)

const cmdName = "npx cucumber-js"

func TestJudge_PolicyMatrix(t *testing.T) {
	skips := "3 scenarios (1 skipped, 2 passed)\n"
	cases := []struct {
		policy string
		want   Verdict
	}{
		{PolicyFail, VerdictFail},
		{"", VerdictFail}, // an absent key behaves as fail
		{PolicyWarn, VerdictWarn},
		{PolicyOff, VerdictPass},
	}
	for _, c := range cases {
		t.Run("policy="+c.policy, func(t *testing.T) {
			v, detail := Judge(cmdName, skips, c.policy)
			if v != c.want {
				t.Fatalf("verdict = %v, want %v (detail %q)", v, c.want, detail)
			}
			if c.want == VerdictPass {
				return
			}
			if !strings.Contains(detail, "1 skipped") || !strings.Contains(detail, cmdName) {
				t.Fatalf("detail must name the counts and the command, got %q", detail)
			}
		})
	}
}

func TestJudge_ExecutedAndCleanPasses(t *testing.T) {
	v, detail := Judge(cmdName, "5 scenarios (5 passed)\n", PolicyFail)
	if v != VerdictPass || detail != "" {
		t.Fatalf("a clean acceptance run must pass silently, got %v / %q", v, detail)
	}
}

func TestJudge_ZeroScenariosFails(t *testing.T) {
	v, detail := Judge(cmdName, "0 scenarios\n", PolicyFail)
	if v != VerdictFail {
		t.Fatalf("a run that executed nothing must fail, got %v", v)
	}
	if !strings.Contains(detail, "no scenarios") {
		t.Fatalf("detail must state that nothing was executed, got %q", detail)
	}
}

// AC3: unparseable is a warning naming the limitation AND a remedy.
func TestJudge_UnparseableWarnsWithARemedy(t *testing.T) {
	v, detail := Judge(cmdName, "Ran 12 examples, 0 failures\n", PolicyFail)
	if v != VerdictWarn {
		t.Fatalf("unparseable output must warn, never pass or fail, got %v", v)
	}
	if !strings.Contains(detail, "could not be parsed") || !strings.Contains(detail, "-json") {
		t.Fatalf("warning must name the limitation and a remedy, got %q", detail)
	}
}

// R4: a recognized shape that structurally carries no skip data is a quiet
// pass with a note — not a permanent warning.
func TestJudge_NoSkipDataIsAQuietPass(t *testing.T) {
	out := "ok  \tgithub.com/x/a\t0.30s\tcoverage: 97.1% of statements\n"
	v, detail := Judge("go test ./... -coverprofile=coverage.out", out, PolicyFail)
	if v != VerdictPass {
		t.Fatalf("a recognized shape without skip data must pass, got %v (%q)", v, detail)
	}
	if !strings.Contains(detail, "no skip data") {
		t.Fatalf("the pass note must say skip data was unavailable, got %q", detail)
	}
	if strings.Contains(detail, "could not be parsed") {
		t.Fatalf("this must not be reported as an unparseable report: %q", detail)
	}
}
