package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/samuelnp/centinela/internal/verdict"
)

func stubDeps(p *verdict.Packet, err error) Deps {
	return Deps{
		Verdict: func(string) (*verdict.Packet, error) { return p, err },
		Rules:   func() RulesOutput { return RulesOutput{Profile: "strict"} },
	}
}

func TestHandlersStampSchemaAndCoalesce(t *testing.T) {
	p := pkt(1, 0, 0, 0) // gate fail, nil Gates/Verify/Evidence slices
	d := stubDeps(p, nil)
	ctx := context.Background()

	_, g, err := d.handleGates(ctx, nil, FeatureInput{})
	if err != nil || g.Schema != SchemaVersion || g.Decision != Block || g.Gates == nil {
		t.Fatalf("gates: %+v err=%v", g, err)
	}
	_, v, _ := d.handleVerify(ctx, nil, VerifyInput{})
	if v.Schema != SchemaVersion || v.Decision != Allow || v.Checks == nil {
		t.Fatalf("verify: %+v", v)
	}
	_, s, _ := d.handleState(ctx, nil, FeatureInput{})
	if s.Schema != SchemaVersion || s.Evidence == nil {
		t.Fatalf("state: %+v", s)
	}
	_, r, _ := d.handleRules(ctx, nil, RulesInput{})
	if r.Schema != SchemaVersion || r.Gates == nil || r.Locales == nil {
		t.Fatalf("rules: %+v", r)
	}
}

func TestHandlersPropagateVerdictError(t *testing.T) {
	d := stubDeps(nil, errors.New("boom"))
	ctx := context.Background()
	if _, _, err := d.handleGates(ctx, nil, FeatureInput{}); err == nil {
		t.Error("handleGates should propagate the verdict error")
	}
	if _, _, err := d.handleVerify(ctx, nil, VerifyInput{}); err == nil {
		t.Error("handleVerify should propagate the verdict error")
	}
	if _, _, err := d.handleState(ctx, nil, FeatureInput{}); err == nil {
		t.Error("handleState should propagate the verdict error")
	}
}

func TestNewServerBuilds(t *testing.T) {
	if NewServer(stubDeps(&verdict.Packet{}, nil)) == nil {
		t.Fatal("NewServer returned nil")
	}
}
