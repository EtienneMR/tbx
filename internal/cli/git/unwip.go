package git

import (
	"errors"
	"strings"

	"charm.land/huh/v2"
	"github.com/EtienneMR/tbx/internal/process"
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

	tui.Check(git.Resolve())

	branch := git.Output("rev-parse", "--abbrev-ref", "HEAD")

	base_branch, found := strings.CutPrefix(branch, WIP_BRANCH_PREFIX)
	if !found {
		tui.Fatal("Not on a WIP branch (expected prefix %q)", WIP_BRANCH_PREFIX)
	}

	snapshot_if_dirty("pre-unwip snapshot", false)

	wip_head := git.Output("rev-parse", "HEAD")
	changed_files := strings.Split(git.Output("diff", "--name-only", base_branch, wip_head), "\n")

	var selected_files []string
	var message string
	var delete bool

	opts := make([]huh.Option[string], len(changed_files))
	for i, o := range changed_files {
		opts[i] = huh.NewOption(o, o).Selected(true)
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Files to commit").
				Description("Select which files to move out of the WIP branch.").
				Options(opts...).
				Value(&selected_files),

			huh.NewText().
				Title("Commit message").
				Description("Describe the change being committed to "+base_branch+".").
				Value(&message),

			huh.NewConfirm().
				Title("Delete branch ?").
				Description("Should branch "+branch+" be deleted ?").
				Value(&delete).
				Validate(func(val bool) error {
					if val && len(selected_files) != len(opts) {
						return errors.New("Branch can only be deleted if all files are commited.")
					}
					return nil
				}),
		),
	).Run()
	tui.Check(err)

	selected_set := make(map[string]bool, len(selected_files))
	for _, f := range selected_files {
		selected_set[f] = true
	}

	var unselected_files []string
	for _, f := range changed_files {
		if !selected_set[f] {
			unselected_files = append(unselected_files, f)
		}
	}

	tui.Step("Switching to %q", base_branch)
	git.Must("switch", base_branch)

	tui.Step("Restoring selected files from %q", branch)
	restore_args := append(
		[]string{"restore", "--source", wip_head, "--"},
		selected_files...,
	)
	git.Must(restore_args...)

	add_args := append(
		[]string{"add"},
		selected_files...,
	)
	git.Must(add_args...)

	tui.Step("Committing")
	git.Must("commit", "--message", message)

	tui.Step("Pushing")
	git.Must("push")

	if delete {
		tui.Step("Deleting local branch %q", branch)
		git.Must("branch", "--delete", branch)

		tui.Step("Deleting remote branch %q", branch)

		result, err := getRemoteOf(branch)
		process.Check(result, err)

		res, err := git.Run("push", result.Stdout, "--delete", branch)
		if res.Code != 1 {
			process.Check(res, err)
		}
	} else {
		tui.Step("Advancing %q to new %q HEAD", branch, base_branch)
		git.Must("branch", "--force", branch, base_branch)

		tui.Step("Switching back to %q", branch)
		git.Must("switch", branch)

		if len(unselected_files) > 0 {
			tui.Step("Restoring unselected files from previous WIP state")
			restore_unselected := append(
				[]string{"restore", "--source", wip_head, "--"},
				unselected_files...,
			)
			git.Must(restore_unselected...)
		}

		snapshot_if_dirty("post-unwip snapshot", true)

		tui.Blank()
		tui.Success("Committed %d file(s) to %q; %d file(s) left as unstaged changes on %q",
			len(selected_files), base_branch, len(unselected_files), branch)

	}
}
