package evidence

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SetField mutates r in place, assigning value to the field identified by
// dotted-path. Supported paths: top-level scalar/list fields, extra.<key>,
// and _meta.<cli_version|written_at>. Booleans coerce from "true"/"false".
func SetField(r *RoleEvidence, path, value string) error {
	if r == nil {
		return errNilEvidence
	}
	if strings.HasPrefix(path, "extra.") {
		return setExtra(r, strings.TrimPrefix(path, "extra."), value)
	}
	if strings.HasPrefix(path, "_meta.") {
		return setMeta(r, strings.TrimPrefix(path, "_meta."), value)
	}
	return setTopLevel(r, path, value)
}

func setTopLevel(r *RoleEvidence, field, value string) error {
	switch field {
	case "feature":
		r.Feature = value
	case "step":
		r.Step = value
	case "role":
		r.Role = value
	case "status":
		r.Status = value
	case "generatedAt":
		r.GeneratedAt = value
	case "handoffTo":
		r.HandoffTo = value
	case "mobileFirst":
		b, err := parseBool(value)
		if err != nil {
			return err
		}
		r.MobileFirst = &b
	case "inputs", "outputs", "edgeCases":
		return fmt.Errorf("field %q is a list — use `centinela evidence append`", field)
	default:
		return fmt.Errorf("unknown field %q (try extra.%s for free-form)", field, field)
	}
	return nil
}

func setMeta(r *RoleEvidence, key, value string) error {
	if r.Meta == nil {
		r.Meta = &Meta{}
	}
	switch key {
	case "cli_version":
		r.Meta.CLIVersion = value
	case "written_at":
		r.Meta.WrittenAt = value
	default:
		return fmt.Errorf("unknown _meta field %q", key)
	}
	return nil
}

func setExtra(r *RoleEvidence, key, value string) error {
	if r.Extra == nil {
		r.Extra = map[string]json.RawMessage{}
	}
	// marshalNoEscape on a string can never fail at runtime.
	raw, _ := marshalNoEscape(value)
	r.Extra[key] = raw
	return nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	}
	return false, fmt.Errorf("cannot parse %q as bool", s)
}
