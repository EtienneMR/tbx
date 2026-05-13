package git

import (
	"github.com/EtienneMR/tbx/texec"
	"github.com/EtienneMR/tbx/tlog"
	"github.com/spf13/cobra"
)

const WIP_BRANCH_PREFIX = "wip/"
const WIP_MESSAGE_PREFIX = "WIP: "

var git = texec.New("git")

var Cmd = &cobra.Command{
	Use:   "git",
	Short: "Git workflow helpers",
	Long:  "Shortcuts for the git operations you run every day.",
}

func init() {
	Cmd.AddCommand(wipCmd)
	Cmd.AddCommand(unwipCmd)
}

func snapshot_if_dirty(message string, always_push bool) {
	if git.Output("status", "--porcelain") == "" {
		tlog.Info("Working tree is clean, nothing to snapshot")
		if !always_push {
			return
		}
	} else {
		tlog.Step("Snapshotting dirty working tree")

		git.Must("add", "-A")

		message = WIP_MESSAGE_PREFIX + message
		tlog.Step("Committing %q", message)
		git.Must("commit", "--message", message)
	}

	tlog.Step("Pushing snapshot")
	git.Must("push")
}
