package git

import (
	"github.com/EtienneMR/tbx/internal/process"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

const WIP_BRANCH_PREFIX = "wip/"
const WIP_MESSAGE_PREFIX = "WIP: "
const DEFAULT_REMOTE = "origin"

var git = process.New("git")

var Cmd = &cobra.Command{
	Use:   "git",
	Short: "Git workflow helpers",
	Long:  "Shortcuts for the git operations you run every day.",
}

func init() {
	Cmd.AddCommand(wipCmd, unwipCmd, shipCmd)
}

func snapshot_if_dirty(message string, always_push bool) {
	if git.Output("status", "--porcelain") == "" {
		tui.Info("Working tree is clean, nothing to snapshot")
		if !always_push {
			return
		}
	} else {
		tui.Step("Snapshotting dirty working tree")

		git.Must("add", "--all")

		message = WIP_MESSAGE_PREFIX + message
		tui.Step("Committing %q", message)
		git.Must("commit", "--message", message)
	}

	tui.Step("Pushing snapshot")
	git.Must("push", "--force-with-lease")
}

func getRemoteOf(branch string) (*process.Result, error) {
	result, err := git.Run("for-each-ref", "--format=%(upstream:remotename)", "refs/heads/"+branch)
	if result.Stdout == "" {
		result.Stdout = DEFAULT_REMOTE
	}
	return result, err
}
