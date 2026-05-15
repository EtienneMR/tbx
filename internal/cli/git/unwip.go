package git

import (
	"strings"

	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var unwipCmd = &cobra.Command{
	Use:   "unwip",
	Short: "Cherry-pick changes from a WIP branch into a proper commit on the base branch",
	Long: `Selects files that differ between the current WIP branch and its base branch,
restores the chosen files onto the base branch, and commits them with a message.`,
	Run: runUnwipCmd,
}

func runUnwipCmd(cmd *cobra.Command, args []string) {
	tui.Header("Committing changes from WIP branch")

	tui.Check(git.Resolve(), "unwip")

	branch := git.Output("rev-parse", "--abbrev-ref", "HEAD")

	base_branch, found := strings.CutPrefix(branch, WIP_BRANCH_PREFIX)
	if !found {
		tui.Fatal("not on a WIP branch (expected prefix %q)", WIP_BRANCH_PREFIX)
	}

	snapshot_if_dirty("pre-unwip snapshot", false)

	tui.Step("Switching to %q", base_branch)
	git.Must("switch", base_branch)

	tui.Step("Restoring selected files from %q", branch)
	git.Must("restore", "--source", branch, "--", ".")

}
