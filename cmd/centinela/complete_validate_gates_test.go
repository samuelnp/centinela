package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// gatesRepo builds a REAL temp git repo (the freshness check shells out to
// git) at the validate step under the given contract.
func gatesRepo(t *testing.T, contract, validateCmd, report string) *config.Config {
	t.Helper()
	stampRepo(t, report)
	os.WriteFile("centinela.toml", []byte("[validate]\ncommands=[\""+validateCmd+"\"]\n"), 0o644) //nolint:errcheck
	wf := workflow.New("f")
	wf.ValidateContract = contract
	wf.CurrentStep = "validate"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

// Freshness runs FIRST so the operator is told to re-verify before paying for
// a full suite run they would only have to repeat.
func TestRunValidateGatesStaleVerificationShortCircuits(t *testing.T) {
	stale := "**Status:** SAFE\n\n```json centinela:verification\n" +
		`{"revision":"deadbeef","treeDigest":"sha256:nope",` +
		`"commands":[{"argv":["centinela","validate"],"exitCode":0}]}` + "\n```\n"
	cfg := gatesRepo(t, workflow.ValidateContractAdversarial, "true", stale)
	err := runValidateGates("f", "validate", "", cfg)
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("want the stale-verification block first, got %v", err)
	}
}

func TestRunValidateGatesFailingCommandBlocks(t *testing.T) {
	cfg := gatesRepo(t, "", "false", "**Status:** SAFE\n")
	if err := runValidateGates("f", "validate", "", cfg); err == nil {
		t.Fatal("a failing validate command must block completion")
	}
}

// Regression: claim verification resolved roles through the contract-blind
// policy layer, so a LEGACY workflow's validate claims were looked up under
// "gatekeeper", found nothing, and went entirely unverified — the same root
// cause as the directive/gate divergence, one layer down.
func TestRunValidateGatesVerifiesTheLegacyRolesClaims(t *testing.T) {
	cfg := gatesRepo(t, "", "true", "**Status:** SAFE\n")
	evidence := `{"feature":"f","step":"validate","role":"validation-specialist","status":"done",` +
		`"generatedAt":"2026-07-28T00:00:00Z","inputs":["i"],"outputs":[],` +
		`"edgeCases":[],"handoffTo":"documentation-specialist"}`
	os.WriteFile(".workflow/f-validation-specialist.json", []byte(evidence), 0o644) //nolint:errcheck
	out := captureStdout(t, func() {
		if err := runValidateGates("f", "validate", "", cfg); err != nil {
			t.Fatalf("honest legacy evidence must not block: %v", err)
		}
	})
	if !strings.Contains(out, "validation-specialist") {
		t.Fatalf("legacy claims went unverified — no role named in: %s", out)
	}
	if strings.Contains(out, "no claims to verify") {
		t.Fatalf("legacy claims were skipped entirely: %s", out)
	}
}
