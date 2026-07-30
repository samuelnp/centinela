package config

import (
	"os"
	"strings"
	"testing"
)

func TestNormalizeAcceptanceSkipPolicy(t *testing.T) {
	cases := map[string]string{
		"":       AcceptanceSkipFail,
		"fail":   AcceptanceSkipFail,
		" WARN ": AcceptanceSkipWarn,
		"Off":    AcceptanceSkipOff,
		"maybe":  AcceptanceSkipFail,
	}
	for in, want := range cases {
		if got := NormalizeAcceptanceSkipPolicy(in); got != want {
			t.Fatalf("NormalizeAcceptanceSkipPolicy(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidateAcceptanceSkipPolicy(t *testing.T) {
	for _, ok := range []string{"", "fail", "warn", "off", " FAIL "} {
		if err := validateAcceptanceSkipPolicy(ok); err != nil {
			t.Fatalf("%q must be accepted, got %v", ok, err)
		}
	}
	err := validateAcceptanceSkipPolicy("maybe")
	if err == nil || !strings.Contains(err.Error(), "fail, warn, or off") {
		t.Fatalf("expected an error naming the supported values, got %v", err)
	}
}

// AC4 end-to-end through Load: absent key defaults to fail, explicit values
// round-trip, and an unknown value is a LOAD error — not a silent normalization.
func TestLoad_AcceptanceSkipPolicy(t *testing.T) {
	cases := []struct {
		name, body, want string
		wantErr          bool
	}{
		{"absent", "[validate]\ncommands = []\n", AcceptanceSkipFail, false},
		{"fail", "[validate]\nacceptance_skip_policy = \"fail\"\n", AcceptanceSkipFail, false},
		{"warn", "[validate]\nacceptance_skip_policy = \"warn\"\n", AcceptanceSkipWarn, false},
		{"off", "[validate]\nacceptance_skip_policy = \"off\"\n", AcceptanceSkipOff, false},
		{"unknown", "[validate]\nacceptance_skip_policy = \"maybe\"\n", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			if err := os.WriteFile(Filename, []byte(c.body), 0o644); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load()
			if c.wantErr {
				if err == nil || !strings.Contains(err.Error(), "acceptance_skip_policy") {
					t.Fatalf("expected a config error naming the knob, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Validate.AcceptanceSkipPolicy != c.want {
				t.Fatalf("policy = %q, want %q", cfg.Validate.AcceptanceSkipPolicy, c.want)
			}
		})
	}
}
