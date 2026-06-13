package verdict

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verify"
	"github.com/samuelnp/centinela/internal/workflow"
)

const goldenJSON = `{
  "schema": "centinela.verdict/v1",
  "run": {
    "feature": "headless-governance",
    "step": "validate",
    "profile": "strict",
    "archetype": "canonical",
    "headless": false,
    "generatedAt": "2026-06-12T00:00:00Z"
  },
  "summary": {
    "verdict": "pass",
    "exitCode": 0,
    "gates": {
      "pass": 1,
      "fail": 0,
      "warn": 0,
      "skip": 0
    },
    "verify": {
      "pass": 1,
      "fail": 0,
      "warn": 0,
      "skip": 0
    }
  },
  "gates": [
    {
      "name": "G1: File Size",
      "status": "pass",
      "message": "ok"
    }
  ],
  "verify": [
    {
      "role": "qa-senior",
      "claim": "tests-pass",
      "status": "PASS"
    }
  ],
  "evidence": [
    {
      "role": "big-thinker",
      "step": "plan",
      "status": "done",
      "path": ".workflow/headless-governance-big-thinker.json"
    }
  ]
}`

// Two identical runs produce byte-identical JSON matching the golden file, with
// no maps and evidence pre-sorted by role.
func TestAssembleVerdict_DeterministicGolden(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	checks := vr(verify.Check{Role: "qa-senior", Claim: "tests-pass", Status: verify.StatusPass})
	ev := []EvidLine{{Role: "big-thinker", Step: "plan", Status: "done", Path: ".workflow/headless-governance-big-thinker.json"}}
	deps := fakeDeps([]gates.Result{passGate()}, checks, ev)
	wf := &workflow.Workflow{Feature: "headless-governance", CurrentStep: "validate"}

	first := marshal(t, AssembleVerdict("headless-governance", &config.Config{}, wf, deps))
	second := marshal(t, AssembleVerdict("headless-governance", &config.Config{}, wf, deps))
	if first != second {
		t.Fatal("two runs must be byte-identical")
	}
	if strings.TrimSpace(first) != strings.TrimSpace(goldenJSON) {
		t.Fatalf("golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", first, goldenJSON)
	}
}

func marshal(t *testing.T, pkt *Packet) string {
	t.Helper()
	b, err := json.MarshalIndent(pkt, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
