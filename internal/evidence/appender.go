package evidence

import (
	"fmt"
)

// AppendField appends value to the list-valued field identified by path. The
// only supported list fields are inputs, outputs, and edgeCases (the three
// contract-defined arrays). Appends are dedup-by-equality so a re-run is a
// safe no-op.
func AppendField(r *RoleEvidence, path, value string) error {
	if r == nil {
		return errNilEvidence
	}
	switch path {
	case "inputs":
		r.Inputs = appendUnique(r.Inputs, value)
	case "outputs":
		r.Outputs = appendUnique(r.Outputs, value)
	case "edgeCases":
		r.EdgeCases = appendUnique(r.EdgeCases, value)
	default:
		return fmt.Errorf("field %q is not appendable (allowed: inputs, outputs, edgeCases)", path)
	}
	return nil
}

func appendUnique(list []string, value string) []string {
	for _, item := range list {
		if item == value {
			return list
		}
	}
	return append(list, value)
}

// ReadField returns the field value as the smallest stable Go type so the
// CLI can JSON-encode it for stdout.
func ReadField(r *RoleEvidence, field string) (any, error) {
	if r == nil {
		return nil, errNilEvidence
	}
	switch field {
	case "feature":
		return r.Feature, nil
	case "step":
		return r.Step, nil
	case "role":
		return r.Role, nil
	case "status":
		return r.Status, nil
	case "generatedAt":
		return r.GeneratedAt, nil
	case "handoffTo":
		return r.HandoffTo, nil
	case "mobileFirst":
		return r.MobileFirst, nil
	case "inputs":
		return r.Inputs, nil
	case "outputs":
		return r.Outputs, nil
	case "edgeCases":
		return r.EdgeCases, nil
	case "_meta":
		return r.Meta, nil
	}
	if len(field) > 6 && field[:6] == "extra." {
		if v, ok := r.Extra[field[6:]]; ok {
			return v, nil
		}
		return nil, fmt.Errorf("extra key %q not set", field[6:])
	}
	return nil, fmt.Errorf("unknown field %q", field)
}
