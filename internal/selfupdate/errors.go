package selfupdate

import "fmt"

// Kind classifies a self-update failure so callers (and tests) can branch on the
// cause without string matching.
type Kind string

const (
	KindNetwork    Kind = "network"
	KindAPI        Kind = "api"
	KindPlatform   Kind = "unsupported-platform"
	KindChecksum   Kind = "checksum"
	KindReplace    Kind = "replace"
	KindPermission Kind = "permission"
)

// Error is the typed error returned by every self-update failure path. It never
// panics on bad input; callers print Error() and exit non-zero.
type Error struct {
	Kind Kind
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("centinela update: %s: %s: %v", e.Kind, e.Msg, e.Err)
	}
	return fmt.Sprintf("centinela update: %s: %s", e.Kind, e.Msg)
}

// Unwrap exposes the wrapped cause for errors.Is/As.
func (e *Error) Unwrap() error { return e.Err }

func newErr(kind Kind, msg string, err error) *Error {
	return &Error{Kind: kind, Msg: msg, Err: err}
}
