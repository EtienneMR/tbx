package git

import (
	"strings"

	"github.com/EtienneMR/tbx/tlog"
	"github.com/EtienneMR/tbx/tui"
	"github.com/spf13/cobra"
)

var wipCmd = &cobra.Command{
	Use:   "wip",
	Short: "Stage everything and commit with a quick message in a WIP branch",
	Long:  `Switch to a WIP branch, stages all changes and commits with a message.`,
	Run:   runWipCmd,
}

func init() {}

func runWipCmd(cmd *cobra.Command, args []string) {
	tlog.Header("Creating a WIP snapshot")

	tlog.Check(git.Resolve())

	message, err := tui.Input("Commit message: ", "manual snapshot")
	tlog.Check(err)

	branch := git.Output("rev-parse", "--abbrev-ref", "HEAD")
	wip_branch := WIP_BRANCH_PREFIX + strings.TrimPrefix(branch, WIP_BRANCH_PREFIX)

	if branch != wip_branch {
		tlog.Step("Switching to %q", wip_branch)
		git.Must("switch", "--force-create", wip_branch)
	}

	snapshot_if_dirty(message, true)

	tlog.Blank()
	tlog.Success("Snapshot saved on %q", branch)
}
