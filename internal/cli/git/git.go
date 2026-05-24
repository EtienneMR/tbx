package git

import (
	"fmt"

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

	if remote, err := getRemoteOf("HEAD"); err != nil {
		tui.Step("Pushing snapshot")
		git.Must("push", remote, "--force-with-lease")
	} else {
		tui.Info("No remotes configured, skipping push")
	}
}

// getRemoteOf gets the upstream remote for a branch
func getRemoteOf(branch string) (string, error) {
	result, err := git.Run("for-each-ref", "--format=%(upstream:remotename)", "refs/heads/"+branch)
	if err != nil {
		return "", err
	}

	if result.Stdout != "" {
		return result.Stdout, nil
	}

	if result, err = git.Run("remote", "get-url", DEFAULT_REMOTE); err == nil && result.Code == 0 {
		return DEFAULT_REMOTE, nil
	}

	return "", fmt.Errorf("no remote for branch %s", branch)
}
