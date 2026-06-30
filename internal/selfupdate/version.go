package selfupdate

import "strings"

// isDev reports whether the running version is the uncomparable dev sentinel.
func (u *Updater) isDev() bool { return u.Version == devVersion }

// normalize strips a single leading "v" so the v-prefixed release tag
// ("v0.40.2") and the v-stripped ldflag version ("0.40.2") compare equal.
func normalize(v string) string { return strings.TrimPrefix(v, "v") }

// behind reports whether the running binary should update to latest. The
// release tag is the authoritative "latest", so any normalized difference means
// a new release is available; equality means the binary is current.
func (u *Updater) behind(latest string) bool {
	return normalize(u.Version) != normalize(latest)
}

// availableMsg is the shared one-line "update available" string used by both
// the --check verdict and the SessionStart notice.
func availableMsg(current, latest string) string {
	return "update available: v" + normalize(current) + " -> v" + normalize(latest) +
		" (run centinela update)"
}
