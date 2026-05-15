package pkg

import (
	"strings"

	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:               "install <package...>",
	Short:             "Install one or more packages",
	Long:              `Install packages from official repositories or the AUR.`,
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: completeLocal,
	Run: func(cmd *cobra.Command, args []string) {
		tui.Step("Installing: %s", strings.Join(args, "  "))

		argv := append([]string{"-S", "--needed"}, args...)
		writePm().Live(argv...)
	},
}
