package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// Fingerprint is a stable per-violation identity used by the ratchet. Identity
// is location + rule, never the volatile payload (line count, drift line),
// so a baselined violation stays matched across edits that only move counts.
type Fingerprint struct {
	Gate string `json:"gate"` // Result.Name
	Key  string `json:"key"`  // normalized stable identity (path, edge, …)
	Hash string `json:"hash"` // sha256(scheme \x00 gate \x00 key), hex
	Raw  string `json:"raw"`  // last-seen raw Detail, for human PR review only
}

// fingerprintScheme versions the normalization. Bump on any extractor change so
// an old baseline never silently mis-matches a new Detail format.
const fingerprintScheme = "v1"

// Compute builds the deduplicated, Hash-sorted fingerprint set for one gate's
// details. Identical details within a gate collapse to a single entry.
func Compute(gate string, details []string) []Fingerprint {
	seen := make(map[string]Fingerprint)
	for _, d := range details {
		key := identityKey(gate, d)
		fp := Fingerprint{Gate: gate, Key: key, Hash: hashIdentity(gate, key), Raw: d}
		seen[fp.Hash] = fp
	}
	out := make([]Fingerprint, 0, len(seen))
	for _, fp := range seen {
		out = append(out, fp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Hash < out[j].Hash })
	return out
}

// hashIdentity folds the scheme into the hash so a scheme bump changes every
// hash, invalidating a stale baseline rather than silently matching it.
func hashIdentity(gate, key string) string {
	sum := sha256.Sum256([]byte(fingerprintScheme + "\x00" + gate + "\x00" + key))
	return hex.EncodeToString(sum[:])
}

// identityKey reduces one raw Detail to its stable key for a given gate Name.
// Each detail-emitting gate has a reducer; unknown gates use genericKey.
func identityKey(gate, detail string) string {
	switch gate {
	case "G1: File Size":
		return beforeParen(detail)
	case "import_graph":
		return beforeParen(detail)
	case "spec-traceability-gate", "G-Secrets: Secret Scan", "G11: i18n":
		return strings.TrimSpace(detail)
	default:
		return genericKey(detail)
	}
}

// beforeParen returns the substring before the first " (", trimmed. This is the
// stable location for "path (N lines)" and "pkg -> imp (reason)" details.
func beforeParen(detail string) string {
	if i := strings.Index(detail, " ("); i >= 0 {
		return strings.TrimSpace(detail[:i])
	}
	return strings.TrimSpace(detail)
}

// genericKey strips a trailing "(…)" group and any trailing run of digits and
// whitespace so volatile counts don't change a violation's identity. It is a
// no-op on already-stable keys.
func genericKey(detail string) string {
	s := strings.TrimSpace(detail)
	if i := strings.LastIndex(s, " ("); i >= 0 && strings.HasSuffix(s, ")") {
		s = strings.TrimSpace(s[:i])
	}
	s = strings.TrimRight(s, "0123456789 \t")
	return strings.TrimSpace(s)
}
