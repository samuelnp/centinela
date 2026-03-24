package projectstage

import "testing"

func TestParseProjectStage(t *testing.T) {
	if Parse("x") != Greenfield {
		t.Fatal("missing stage should default greenfield")
	}
	if Parse("Project Stage: existing") != Existing {
		t.Fatal("existing should parse")
	}
	if Parse("**Project Stage:** greenfield") != Greenfield {
		t.Fatal("markdown stage should parse")
	}
	if Parse("Project Stage: weird") != Greenfield {
		t.Fatal("unknown stage should default greenfield")
	}
}
