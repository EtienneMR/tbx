package pkg

import (
	"strings"

	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var explicitCmd = &cobra.Command{
	Use:   "explicit <package...>",
	Short: "Mark packages as explicitly installed",
	Long: `Change the install reason of packages from 'dependency' to 'explicit'.

This is the inverse of ` + "`tbx packages temp`" + `: it protects a package from being
removed by ` + "`tbx packages orphans --remove`" + ` even when nothing else depends on it.`,
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: completeOrphans,
	Run: func(cmd *cobra.Command, args []string) {
		assertInstalled(args)

		tui.Step("Marking as explicit: %s", strings.Join(args, "  "))

		argv := append([]string{"-D", "--asexplicit"}, args...)
		writePm().Live(argv...)
	},
}

// assertInstalled checks that every name in pkgs is present in the local
// package database.  It returns a descriptive error listing any that are not.
func assertInstalled(pkgs []string) {
	res := pm.Output("-Qq")

	installed := make(map[string]bool, 256)
	for name := range strings.FieldsSeq(res) {
		installed[name] = true
	}

	var missing []string
	for _, pkg := range pkgs {
		if !installed[pkg] {
			missing = append(missing, pkg)
		}
	}
	if len(missing) > 0 {
		tui.Fatal("not installed: %s", strings.Join(missing, ", "))
	}
}
