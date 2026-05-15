package pkg

import (
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query...>",
	Short: "Search repositories and AUR for packages",
	Long: `Search for packages matching all provided terms.

Multiple words are passed as separate arguments so the package manager can
apply its own AND logic.`,
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: completeLocal,
	Run: func(cmd *cobra.Command, args []string) {
		argv := append([]string{"-Ss"}, args...)
		pm.Live(argv...)
	},
}
