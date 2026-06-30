package selfupdate

// AssetName builds the release asset filename for a platform. The tag keeps its
// leading "v" (mirroring the release workflow byte-for-byte) and ".exe" is
// appended only on windows. Examples: centinela-v0.40.2-linux-amd64,
// centinela-v0.40.2-windows-amd64.exe.
func AssetName(goos, goarch, tag string) string {
	name := "centinela-" + tag + "-" + goos + "-" + goarch
	if goos == "windows" {
		name += ".exe"
	}
	return name
}

// assetName builds the asset filename for this Updater's host platform.
func (u *Updater) assetName(tag string) string {
	return AssetName(u.GOOS, u.GOARCH, tag)
}
