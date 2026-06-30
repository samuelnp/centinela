package selfupdate

// CheckResult is the read-only verdict returned by Check.
type CheckResult struct {
	Current string
	Latest  string
	Behind  bool
	Message string
}

// Check resolves the latest version (honoring the TTL cache) and reports whether
// the running binary is behind. It performs ZERO writes to the binary; it may
// refresh the cache. A dev build is uncomparable and reported exit-zero.
func (u *Updater) Check() (CheckResult, error) {
	if u.isDev() {
		return CheckResult{Current: u.Version, Message: devMessage}, nil
	}
	latest, err := u.latestTag()
	if err != nil {
		return CheckResult{}, err
	}
	behind := u.behind(latest)
	msg := "centinela v" + normalize(u.Version) + " is up to date"
	if behind {
		msg = availableMsg(u.Version, latest)
	}
	return CheckResult{Current: u.Version, Latest: latest, Behind: behind, Message: msg}, nil
}

// Update resolves the latest release, downloads + verifies the host asset, and
// atomically installs it, returning the message to print. A dev build is an
// informational no-op; an already-current binary is a no-op. It does not write
// the cache: the explicit action always resolves fresh.
func (u *Updater) Update() (string, error) {
	if u.isDev() {
		return devMessage, nil
	}
	rel, err := u.resolveLatest()
	if err != nil {
		return "", err
	}
	if !u.behind(rel.Tag) {
		return "already up to date", nil
	}
	if err := u.install(rel); err != nil {
		return "", err
	}
	return normalize(u.Version) + " -> " + normalize(rel.Tag), nil
}

// latestTag returns the latest release tag, serving a fresh cache entry without
// any network call and otherwise fetching from the API and refreshing the cache.
func (u *Updater) latestTag() (string, error) {
	if tag, ok := u.readCache(); ok {
		return tag, nil
	}
	rel, err := u.resolveLatest()
	if err != nil {
		return "", err
	}
	u.writeCache(rel.Tag)
	return rel.Tag, nil
}
