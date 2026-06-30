package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
)

// fetchBytes downloads url into memory. The asset is held in memory and only
// written to disk after its checksum verifies, so a failed/tampered download
// never produces a temp file.
func (u *Updater) fetchBytes(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, newErr(KindNetwork, "build download request", err)
	}
	resp, err := u.HTTP.Do(req)
	if err != nil {
		return nil, newErr(KindNetwork, "download asset", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, newErr(KindAPI, "asset download returned "+resp.Status, nil)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, newErr(KindNetwork, "read asset body", err)
	}
	return data, nil
}

// sumFor extracts the hex checksum for filename from coreutils-format
// SHA256SUMS: "<64-hex><two spaces><filename>" per line.
func sumFor(sums []byte, filename string) (string, bool) {
	for _, line := range strings.Split(string(sums), "\n") {
		fields := strings.SplitN(strings.TrimSpace(line), "  ", 2)
		if len(fields) == 2 && fields[1] == filename {
			return fields[0], true
		}
	}
	return "", false
}

// verify reports whether data's sha256 matches the expected hex digest.
func verify(data []byte, expectedHex string) bool {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]) == strings.ToLower(strings.TrimSpace(expectedHex))
}
