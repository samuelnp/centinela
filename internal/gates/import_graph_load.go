package gates

import (
	"github.com/samuelnp/centinela/internal/golist"
)

// goListPkg is the gate-local alias for golist.Pkg. The loader lives in the
// shared internal/golist leaf (reused by codebase analysis); the gate delegates
// to it so the streamed-JSON decode + stderr-surfacing logic has one home.
type goListPkg = golist.Pkg

// loadModulePath returns the current module path via the shared golist seam.
func loadModulePath() (string, error) {
	return golist.ModulePath()
}

// loadPackages loads the module's package import graph via the shared golist
// seam. A non-zero exit (e.g. uncompilable code) is surfaced as an error so the
// gate Fails rather than reporting a false Pass.
func loadPackages() ([]goListPkg, error) {
	return golist.Packages()
}
