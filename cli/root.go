package cli

import (
	"fmt"

	"github.com/EtienneMR/tbx/cli/git"
	"github.com/EtienneMR/tbx/tlog"
	"github.com/spf13/cobra"
)

// Version is injected at build time via -ldflags.
var Version = "dev"

// Execute is the single entry-point called by main.
func Execute() {
	cmd := &cobra.Command{
		Use:   "tbx",
		Short: "Personal toolbox CLI",
		Long:  "tbx — a personal toolbox for everyday dev tasks.",
	}
	cmd.PersistentFlags().CountVarP(
		&tlog.Verbosity, "verbose", "v",
		"increase verbosity (-v shows commands, -vv adds debug output)",
	)

	cmd.Version = Version
	cmd.SetVersionTemplate(fmt.Sprintf("tbx %s\n", Version))

	cmd.AddCommand(git.Cmd)

	if err := cmd.Execute(); err != nil {
		tlog.Fatal("%s", err.Error())
	}
}
