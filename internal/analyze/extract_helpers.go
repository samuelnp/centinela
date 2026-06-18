package analyze

import "strings"

// tomlSection returns the keys of a single-level TOML table named section as a
// name->"" set. It is a deliberately minimal scanner (no nested tables, no
// value parsing) sufficient to recover declared dependency names best-effort.
func tomlSection(data []byte, section string) map[string]string {
	out := map[string]string{}
	header := "[" + section + "]"
	in := false
	for _, line := range strings.Split(string(data), "\n") {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "[") {
			in = s == header
			continue
		}
		if !in || s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		if key, _, ok := strings.Cut(s, "="); ok {
			if k := strings.TrimSpace(key); k != "" && k != "python" {
				out[k] = ""
			}
		}
	}
	return out
}

// firstQuoted returns the contents of the first single- or double-quoted token
// in s, or "" when none is present.
func firstQuoted(s string) string {
	for _, q := range []byte{'"', '\''} {
		if i := strings.IndexByte(s, q); i >= 0 {
			if j := strings.IndexByte(s[i+1:], q); j >= 0 {
				return s[i+1 : i+1+j]
			}
		}
	}
	return ""
}

// asSet turns a slice of names into a name->"" set.
func asSet(names []string) map[string]string {
	out := map[string]string{}
	for _, n := range names {
		if n != "" {
			out[n] = ""
		}
	}
	return out
}

// splitReqName strips a pip requirement line down to its package name by cutting
// at the first version/extra specifier character.
func splitReqName(s string) string {
	cut := strings.IndexAny(s, "=<>!~;[ ")
	if cut < 0 {
		return s
	}
	return strings.TrimSpace(s[:cut])
}
