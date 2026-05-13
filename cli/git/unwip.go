package git

import (
	"strings"

	"charm.land/huh/v2"
	"github.com/EtienneMR/tbx/tlog"
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
	tlog.Header("Committing changes from WIP branch")

	tlog.Check(git.Resolve())

	branch := git.Output("rev-parse", "--abbrev-ref", "HEAD")

	base_branch, found := strings.CutPrefix(branch, WIP_BRANCH_PREFIX)
	if !found {
		tlog.Fatal("Not on a WIP branch (expected prefix %q)", WIP_BRANCH_PREFIX)
	}

	snapshot_if_dirty("pre unwip snapshot", false)

	wip_head := git.Output("rev-parse", "HEAD")
	changed_files := strings.Split(git.Output("diff", "--name-only", base_branch, wip_head), "\n")

	var selected_files []string
	var message string

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
		),
	).Run()
	tlog.Check(err)

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

	tlog.Step("Switching to %q", base_branch)
	git.Must("switch", base_branch)

	tlog.Step("Restoring selected files from %q", branch)
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

	tlog.Step("Committing")
	git.Must("commit", "-m", message)

	tlog.Step("Pushing")
	//git.Must("push")

	tlog.Step("Advancing %q to new %q HEAD", branch, base_branch)
	git.Must("branch", "-f", branch, base_branch)

	tlog.Step("Switching back to %q", branch)
	git.Must("switch", branch)

	if len(unselected_files) > 0 {
		tlog.Step("Restoring unselected files from previous WIP state")
		restore_unselected := append(
			[]string{"restore", "--source", wip_head, "--"},
			unselected_files...,
		)
		git.Must(restore_unselected...)
	}

	tlog.Blank()
	tlog.Success("Committed %d file(s) to %q; %d file(s) left as unstaged changes on %q",
		len(selected_files), base_branch, len(unselected_files), branch)
}
