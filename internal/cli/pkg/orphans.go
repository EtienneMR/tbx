package pkg

import (
	"strings"

	"github.com/EtienneMR/tbx/internal/process"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var orphansCmd = func() *cobra.Command {
	var remove bool

	cmd := &cobra.Command{
		Use:   "orphans [package...]",
		Short: "List orphaned packages; optionally remove them",
		Long: `List all installed packages that were pulled in as dependencies but are no
longer required by any explicitly-installed package.

Without flags the list is printed and nothing is changed.

  --remove    Remove every detected orphan`,
		ValidArgsFunction: completeInstalled,
		Run: func(cmd *cobra.Command, _ []string) {
			runOrphans(remove)
		},
	}

	cmd.Flags().BoolVar(&remove, "remove", false, "remove all detected orphans")

	return cmd
}()

func removeOrphans() {
	runOrphans(true)
}

func runOrphans(remove bool) {
	res, err := pm.Run("-Qdttq")
	if process.IsErrorCode(err, 1) {
		tui.Success("No orphaned packages found")
		return
	}
	process.Check(res, err)

	orphans := strings.Fields(res.Stdout)

	if !remove {
		tui.Header("Orphaned packages")
		for _, orphan := range orphans {
			optdeps := getOptDepends(orphan)
			if len(optdeps) > 0 {
				tui.Item("%s (optional dependency of: %s)", orphan, strings.Join(optdeps, ", "))
			} else {
				tui.Item("%s", orphan)
			}
		}
		tui.Blank()
		tui.Info("Run with --remove to remove them")
		return
	}

	tui.Header("Removing orphans")

	argv := append([]string{"-Rns"}, orphans...)
	writePm().Live(argv...)
}

func getOptDepends(target string) []string {
	result := pacman.Output("-Qi", target)

	for line := range strings.SplitSeq(result, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 || !strings.HasPrefix(parts[0], "Optional For") {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 1 || fields[0] == "None" {
			return nil
		}
		return fields
	}

	tui.Fatal(`failed to find "Optional For" field for %q`, target)
	return nil
}
