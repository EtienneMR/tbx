package git

import (
	"strings"

	"github.com/EtienneMR/tbx/internal/process"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var wipCmd = &cobra.Command{
	Use:   "wip",
	Short: "Stage everything and commit with a quick message in a WIP branch",
	Long:  `Switch to a WIP branch, stages all changes and commits with a message.`,
	Run:   runWipCmd,
}

func runWipCmd(cmd *cobra.Command, args []string) {
	tui.Header("Creating a WIP snapshot")

	tui.Check(git.Resolve())

	message, err := tui.Input("Commit message: ", "manual snapshot")
	tui.Check(err)

	branch := git.Output("rev-parse", "--abbrev-ref", "HEAD")
	wip_branch := WIP_BRANCH_PREFIX + strings.TrimPrefix(branch, WIP_BRANCH_PREFIX)

	if branch != wip_branch {
		tui.Step("Switching to %q", wip_branch)
		res, err := git.Run("switch", wip_branch)

		if res.Code == 128 {
			tui.Step("Creating to %q", wip_branch)

			result, err := getRemoteOf(branch)
			process.Check(result, err)

			git.Must("switch", "--create", wip_branch)
			git.Must("push", "--set-upstream", result.Stdout, wip_branch)

		} else {
			process.Check(res, err)
		}
	}

	snapshot_if_dirty(message, true)

	tui.Blank()
	tui.Success("Snapshot saved on %q", branch)
}
