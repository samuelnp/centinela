package config

import "testing"

// DriverModelFrom local candidate is the LOWEST precedence tier: flag > env >
// [orchestration] driver_model > [orchestration.local] model. It is engaged only
// when every higher tier is empty, and an empty local model is a no-op.
func TestDriverModelFromLocalCandidate(t *testing.T) {
	withLocal := func(driver, localModel string) *Config {
		c := &Config{}
		c.Orchestration.DriverModel = driver
		c.Orchestration.Local = LocalConfig{Provider: "ollama", Endpoint: "http://x/v1", Model: localModel}
		return c
	}
	cases := []struct {
		name, flag, env string
		cfg             *Config
		want            string
	}{
		{"local model used when nothing else", "", "", withLocal("", "local-m"), "local-m"},
		{"driver_model outranks local", "", "", withLocal("drv-m", "local-m"), "drv-m"},
		{"env outranks local", "", "env-m", withLocal("", "local-m"), "env-m"},
		{"flag outranks local", "flag-m", "", withLocal("", "local-m"), "flag-m"},
		{"empty local leaves zero-config", "", "", withLocal("", ""), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CENTINELA_MODEL", tc.env)
			if got := DriverModelFrom(tc.flag, tc.cfg); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
