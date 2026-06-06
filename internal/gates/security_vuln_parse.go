package gates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// parseVuln dispatches raw scanner output to the matching defensive parser.
func parseVuln(tool string, out []byte) ([]vulnKey, error) {
	if tool == "osv-scanner" {
		return parseOSVScanner(out)
	}
	return parseGovulncheck(out)
}

// govulncheckMsg mirrors the subset of a govulncheck -json NDJSON message the
// gate consumes: an OSV record (vuln definition) and finding records carry the
// module path and the OSV id we pair into a vulnKey.
type govulncheckMsg struct {
	OSV     *struct{ ID string } `json:"osv"`
	Finding *struct {
		OSV   string                    `json:"osv"`
		Trace []struct{ Module string } `json:"trace"`
	} `json:"finding"`
}

// parseGovulncheck decodes the streamed NDJSON output, pairing each finding's
// OSV id with its top-of-trace module. Empty output is "no findings"; malformed
// non-empty output is a parse error so the caller emits a Warn.
func parseGovulncheck(out []byte) ([]vulnKey, error) {
	if len(bytes.TrimSpace(out)) == 0 {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(out))
	var keys []vulnKey
	for {
		var m govulncheckMsg
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("parsing govulncheck output: %w", err)
		}
		if m.Finding == nil || m.Finding.OSV == "" {
			continue
		}
		pkg := "unknown"
		if len(m.Finding.Trace) > 0 && m.Finding.Trace[0].Module != "" {
			pkg = m.Finding.Trace[0].Module
		}
		keys = append(keys, vulnKey{Pkg: pkg, ID: m.Finding.OSV})
	}
	return keys, nil
}

// osvReport mirrors the osv-scanner --format json document shape.
type osvReport struct {
	Results []struct {
		Packages []struct {
			Package struct {
				Name string `json:"name"`
			} `json:"package"`
			Vulnerabilities []struct {
				ID string `json:"id"`
			} `json:"vulnerabilities"`
		} `json:"packages"`
	} `json:"results"`
}

// parseOSVScanner decodes the single JSON document into (package, id) pairs.
// Empty output is "no findings / nothing to scan"; malformed non-empty output
// is a parse error.
func parseOSVScanner(out []byte) ([]vulnKey, error) {
	if len(bytes.TrimSpace(out)) == 0 {
		return nil, nil
	}
	var doc osvReport
	if err := json.Unmarshal(out, &doc); err != nil {
		return nil, fmt.Errorf("parsing osv-scanner output: %w", err)
	}
	var keys []vulnKey
	for _, res := range doc.Results {
		for _, p := range res.Packages {
			for _, v := range p.Vulnerabilities {
				keys = append(keys, vulnKey{Pkg: strings.TrimSpace(p.Package.Name), ID: v.ID})
			}
		}
	}
	return keys, nil
}
