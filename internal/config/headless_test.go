package config

import "testing"

// IsHeadless precedence truth table. CENTINELA_HEADLESS, when non-empty, is
// authoritative; otherwise [headless] enabled; otherwise detect_ci AND CI.
func TestIsHeadless_TruthTable(t *testing.T) {
	cases := []struct {
		name        string
		headlessEnv string
		ciEnv       string
		enabled     bool
		detectCI    bool
		want        bool
	}{
		{"env 1 wins over disabled config", "1", "", false, false, true},
		{"env true wins", "true", "", false, false, true},
		{"env TRUE case-insensitive", "TRUE", "", false, false, true},
		{"env 0 forces off even with enabled", "0", "", true, true, false},
		{"env false forces off", "false", "", true, false, false},
		{"empty env falls through to disabled config", "", "true", false, false, false},
		{"config enabled true", "", "", true, false, true},
		{"detect_ci with CI true", "", "true", false, true, true},
		{"detect_ci with CI 1", "", "1", false, true, true},
		{"detect_ci off with CI present is off", "", "true", false, false, false},
		{"detect_ci on but CI absent is off", "", "", false, true, false},
		{"zero-config default off", "", "", false, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CENTINELA_HEADLESS", tc.headlessEnv)
			t.Setenv("CI", tc.ciEnv)
			cfg := &Config{}
			cfg.Headless = HeadlessConfig{Enabled: tc.enabled, DetectCI: tc.detectCI}
			if got := IsHeadless(cfg); got != tc.want {
				t.Fatalf("IsHeadless = %v, want %v", got, tc.want)
			}
		})
	}
}

// A nil config is never headless once the env signal is absent.
func TestIsHeadless_NilConfig(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	t.Setenv("CI", "true")
	if IsHeadless(nil) {
		t.Fatal("nil config must resolve to not-headless")
	}
	t.Setenv("CENTINELA_HEADLESS", "1")
	if !IsHeadless(nil) {
		t.Fatal("env signal must win even with nil config")
	}
}

func TestEnvTrue(t *testing.T) {
	t.Setenv("X", "1")
	if !envTrue("X") {
		t.Fatal(`"1" must be true`)
	}
	t.Setenv("X", "True")
	if !envTrue("X") {
		t.Fatal(`"True" must be true`)
	}
	t.Setenv("X", "no")
	if envTrue("X") {
		t.Fatal(`"no" must be false`)
	}
}
