package selfupdate

// install downloads the host asset and SHA256SUMS, verifies the checksum, and
// atomically replaces the running binary. Any failure (missing asset, checksum
// mismatch, unwritable dir) returns a typed error and leaves the binary intact.
func (u *Updater) install(rel *Release) error {
	name := u.assetName(rel.Tag)
	assetURL, ok := rel.assetURL(name)
	if !ok {
		return newErr(KindPlatform, "no release asset for "+u.GOOS+"/"+u.GOARCH+" ("+name+")", nil)
	}
	sumsURL, ok := rel.assetURL(sumsAsset)
	if !ok {
		return newErr(KindPlatform, "release is missing "+sumsAsset, nil)
	}
	bin, err := u.fetchBytes(assetURL)
	if err != nil {
		return err
	}
	sums, err := u.fetchBytes(sumsURL)
	if err != nil {
		return err
	}
	want, ok := sumFor(sums, name)
	if !ok {
		return newErr(KindChecksum, "no SHA256SUMS entry for "+name, nil)
	}
	if !verify(bin, want) {
		return newErr(KindChecksum, "checksum verification failed for "+name, nil)
	}
	target, err := u.Target()
	if err != nil {
		return err
	}
	return replaceBinary(target, bin)
}
