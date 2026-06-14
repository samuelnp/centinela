package doctor

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/config"
)

// verifyTimeoutFloor is the minimum recommended verify_timeout (seconds). The
// real suite runs ~75s; a value below this floor produces spurious
// claim-verifier timeouts.
const verifyTimeoutFloor = 180

// configCheck performs centinela.toml sanity checks (WARN, report-only):
// verify_timeout below the floor, a gate referencing a missing directory, and
// unknown TOML keys. An unparseable centinela.toml degrades this check to ERROR
// with a clear message; the other checks still run.
type configCheck struct{}

func (configCheck) Name() string { return "config" }

func (configCheck) Run(ctx Context) Diagnosis {
	d := Diagnosis{Name: "config"}
	if ctx.CfgErr != nil {
		d.Status = Error
		d.Message = "centinela.toml could not be parsed: " + ctx.CfgErr.Error()
		return d
	}
	var details []string
	if t := ctx.Config.Verify.TimeoutSeconds; t > 0 && t < verifyTimeoutFloor {
		details = append(details, fmt.Sprintf(
			"verify_timeout is %ds — below the recommended minimum of %ds",
			t, verifyTimeoutFloor))
	}
	details = append(details, missingGateDirs(ctx.Config)...)
	details = append(details, unknownConfigKeys()...)
	if len(details) == 0 {
		d.Status = OK
		d.Message = "centinela.toml passes sanity checks"
		return d
	}
	d.Status = Warn
	d.Message = "centinela.toml has advisory issues (report-only)"
	d.Details = details
	return d
}

// missingGateDirs reports gate-referenced directories that do not exist.
func missingGateDirs(cfg *config.Config) []string {
	var out []string
	check := func(label, dir string) {
		if dir == "" {
			return
		}
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			out = append(out, label+" references missing directory: "+dir)
		}
	}
	check("i18n.dir", cfg.I18n.Dir)
	check("spec_traceability.spec_dir", cfg.Gates.SpecTraceability.SpecDir)
	check("spec_traceability.test_dir", cfg.Gates.SpecTraceability.TestDir)
	return out
}
