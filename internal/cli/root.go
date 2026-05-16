package cli

import (
	"fmt"

	"github.com/EtienneMR/tbx/internal/cli/edit"
	"github.com/EtienneMR/tbx/internal/cli/git"
	"github.com/EtienneMR/tbx/internal/cli/pkg"
	"github.com/EtienneMR/tbx/internal/cli/self"
	"github.com/EtienneMR/tbx/internal/cli/web"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

// Version is injected at build time via -ldflags.
var Version = "dev"

func Execute() {
	cmd := &cobra.Command{
		Use:   "tbx",
		Short: "Personal toolbox CLI",
		Long:  "tbx — a personal toolbox for everyday dev tasks.",
	}
	cmd.PersistentFlags().CountVarP(
		&tui.Verbosity, "verbose", "v",
		"increase verbosity (-v shows commands, -vv adds debug output)",
	)

	cmd.Version = Version
	cmd.SetVersionTemplate(fmt.Sprintf("tbx %s\n", Version))

	cmd.AddCommand(git.Cmd, web.Cmd, pkg.Cmd, edit.Cmd, self.Cmd)

	if err := cmd.Execute(); err != nil {
		tui.Fatal("%s", err.Error())
	}
}
