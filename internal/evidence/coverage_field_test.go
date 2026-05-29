package evidence

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSetCoverageField(t *testing.T) {
	r := &RoleEvidence{}
	if err := SetField(r, "coverage", "85.5%"); err != nil {
		t.Fatalf("set coverage: %v", err)
	}
	if r.Coverage == nil || *r.Coverage != 85.5 {
		t.Fatalf("coverage = %v, want 85.5", r.Coverage)
	}
	if err := SetField(r, "coverage", "not-a-number"); err == nil {
		t.Fatal("expected parse error for non-numeric coverage")
	}
}

func TestParseCoverage(t *testing.T) {
	cases := []struct {
		in      string
		want    float64
		wantErr bool
	}{
		{"85", 85, false},
		{"85.0", 85, false},
		{"85%", 85, false},
		{" 0 ", 0, false},
		{"100", 100, false},
		{"-1", 0, true},
		{"101", 0, true},
		{"abc", 0, true},
	}
	for _, tc := range cases {
		got, err := parseCoverage(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseCoverage(%q) expected error", tc.in)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("parseCoverage(%q) = %v, %v; want %v", tc.in, got, err, tc.want)
		}
	}
}

func TestCoverageMarshalsWhenSetOmittedWhenNil(t *testing.T) {
	withCov := &RoleEvidence{Feature: "f", Coverage: func() *float64 { v := 90.0; return &v }()}
	data, err := json.Marshal(withCov)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"coverage":90`) {
		t.Fatalf("coverage missing from marshal: %s", data)
	}
	noCov, _ := json.Marshal(&RoleEvidence{Feature: "f"})
	if strings.Contains(string(noCov), "coverage") {
		t.Fatalf("nil coverage should be omitted: %s", noCov)
	}
}

func TestReadCoverageField(t *testing.T) {
	v := 77.0
	got, err := ReadField(&RoleEvidence{Coverage: &v}, "coverage")
	if err != nil {
		t.Fatal(err)
	}
	if p, ok := got.(*float64); !ok || p == nil || *p != 77.0 {
		t.Fatalf("ReadField coverage = %v", got)
	}
}
