package pkg

import (
	"strings"

	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var tempCmd = &cobra.Command{
	Use:   "temp <package...>",
	Short: "Install packages as dependencies (eligible for orphan removal)",
	Long: `Install packages marked as dependencies rather than explicit installs.

Because pacman tracks install reason, any package installed with this command
becomes a candidate for automatic removal as soon as nothing else depends on it.`,
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: completeLocal,
	Run: func(cmd *cobra.Command, args []string) {
		tui.Step("Installing as deps: %s", strings.Join(args, "  "))

		argv := append([]string{"-S", "--needed", "--asdeps"}, args...)
		writePm().Live(argv...)
	},
}
