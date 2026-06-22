package reconstruct

import "strings"

// slugify derives a deterministic, filesystem-safe filename stem from a package
// path: lowercased, non-alphanumeric runs collapsed to single hyphens, leading/
// trailing hyphens trimmed. Empty input yields "module".
func slugify(pkg string) string {
	var b strings.Builder
	prevHyphen := false
	for _, r := range strings.ToLower(pkg) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
			continue
		}
		if !prevHyphen {
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	s := strings.Trim(b.String(), "-")
	if s == "" {
		return "module"
	}
	return s
}

// disambiguate makes slug unique against the already-used set by appending a
// numeric suffix deterministically. The first claimant keeps the bare slug.
func disambiguate(slug string, used map[string]bool) string {
	if !used[slug] {
		used[slug] = true
		return slug
	}
	for i := 2; ; i++ {
		cand := slug + "-" + itoa(i)
		if !used[cand] {
			used[cand] = true
			return cand
		}
	}
}

// itoa avoids importing strconv for a single small positive integer.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
