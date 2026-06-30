package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/selfupdate"
)

var updateCheck bool

// newSelfUpdater is an overridable seam so the command paths are testable
// offline (tests inject an Updater pointed at an httptest.Server).
var newSelfUpdater = func(version string) *selfupdate.Updater { return selfupdate.New(version) }

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update centinela to the latest release (--check reports only, read-only)",
	Long: "Resolves the latest GitHub release for your platform, verifies the " +
		"download against the release SHA256SUMS, and atomically replaces the " +
		"running binary. --check is read-only: it reports whether a newer version " +
		"exists and exits non-zero when you are behind, writing nothing to the " +
		"binary. A development build is an informational no-op.",
	RunE:          runUpdate,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false,
		"Report update availability without installing (read-only)")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(_ *cobra.Command, _ []string) error {
	u := newSelfUpdater(Version)
	if updateCheck {
		res, err := u.Check()
		if err != nil {
			return err
		}
		fmt.Println(res.Message)
		if res.Behind {
			exitMain(1)
		}
		return nil
	}
	msg, err := u.Update()
	if err != nil {
		return err
	}
	fmt.Println(msg)
	return nil
}
