package selfupdate

// Notice returns a one-line "update available" string for the SessionStart hook,
// or "" when there is nothing to say. It is fail-silent: any network, parse, or
// cache error yields "" so a session is never blocked or errored. A dev build is
// suppressed with no network call. It honors the TTL cache and never installs.
func (u *Updater) Notice() string {
	if u.isDev() {
		return ""
	}
	latest, err := u.latestTag()
	if err != nil || latest == "" {
		return ""
	}
	if !u.behind(latest) {
		return ""
	}
	return availableMsg(u.Version, latest)
}
