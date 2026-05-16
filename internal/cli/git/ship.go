package git

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/EtienneMR/tbx/internal/process"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var shipCmd = &cobra.Command{
	Use:   "ship",
	Short: "Commit, push, and finish a WIP-to-base workflow",
	Long: `Commits the staged changes on the base branch, pushes them, and then
asks whether to switch back to the WIP branch or delete it.`,
	Run: runShipCmd,
}

func runShipCmd(cmd *cobra.Command, args []string) {
	tui.Header("Shipping staged changes")

	tui.Check(git.Resolve(), "ship")

	branch := git.Output("rev-parse", "--abbrev-ref", "HEAD")
	wip_branch := WIP_BRANCH_PREFIX + branch

	currentTag := currentVersion()
	major, minor, patch := parseSemver(currentTag)

	nextMajor := fmt.Sprintf("v%d.0.0", major+1)
	nextMinor := fmt.Sprintf("v%d.%d.0", major, minor+1)
	nextPatch := fmt.Sprintf("v%d.%d.%d", major, minor, patch+1)

	var message string
	var next_action string
	var bump string

	var fields []huh.Field
	fields = append(fields, huh.NewText().
		Title("Commit message").
		Description("Describe the changes being shipped to "+branch+".").
		Value(&message))

	if git.Test(1, "show-ref", "--verify", "--quiet", wip_branch) {
		fields = append(fields, huh.NewSelect[string]().
			Title("What next?").
			Description("Choose what to do with "+wip_branch+" after the ship commit.").
			Options(
				huh.NewOption("Switch", "switch"),
				huh.NewOption("Reset", "reset"),
				huh.NewOption("Delete", "delete"),
				huh.NewOption("Keep", "keep"),
			).
			Value(&next_action))
	}

	fields = append(fields, huh.NewSelect[string]().
		Title("Version bump").
		Description("Tag the shipped commit. Current: "+currentTag).
		Options(
			huh.NewOption("major  "+nextMajor, nextMajor),
			huh.NewOption("minor  "+nextMinor, nextMinor),
			huh.NewOption("patch  "+nextPatch, nextPatch).Selected(true),
			huh.NewOption("custom", "custom"),
			huh.NewOption("none", "none"),
		).
		Value(&bump))

	err := huh.NewForm(huh.NewGroup(fields...)).Run()
	tui.Check(err, "ship: form")

	if bump == "custom" {
		bump = ""
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Version tag").
					Placeholder("v0.0.0").
					Value(&bump),
			),
		).Run()
		tui.Check(err, "ship: version form")
	}

	tui.Step("Committing")
	git.Must("add", "--all")
	git.Must("commit", "--message", message)

	if bump != "none" {
		tui.Step("Tagging %s", bump)
		git.Must("tag", "-a", bump, "-m", bump)

		tui.Success("Tagged %s", bump)
	}

	tui.Step("Pushing")
	git.Must("push", "--follow-tags")

	switch next_action {
	case "switch", "reset":
		tui.Step("Switching back to %q", wip_branch)
		git.Must("switch", wip_branch)

		tui.Step("Advancing %q to new %q HEAD", wip_branch, branch)
		// Files not present in `branch` become untracked
		git.Must("reset", "--soft", branch)

		if next_action == "reset" {
			tui.Step("Reseting to %q", branch)
			// Untracked files are kept
			git.Must("reset", "--hard", branch)
		}

		snapshot_if_dirty("post-ship snapshot", true)

	case "delete":
		tui.Step("Deleting local branch %q", wip_branch)
		git.Must("branch", "--delete", wip_branch)

		tui.Step("Deleting remote branch %q", wip_branch)
		result, err := getRemoteOf(wip_branch)
		process.Check(result, err)

		res, err := git.Run("push", result.Stdout, "--delete", wip_branch)
		if res.Code != 1 {
			process.Check(res, err)
		}
	}

	tui.Blank()
	tui.Success("Committed changes to %q", branch)
}

// currentVersion returns the most recent semver tag reachable from HEAD,
// or "none" when the repository has no tags yet.
func currentVersion() string {
	res, err := git.Run("describe", "--tags", "--abbrev=0", "--match", "v*")
	if err != nil || res.Stdout == "" {
		return "none"
	}
	return res.Stdout
}

// parseSemver parses a vMAJOR.MINOR.PATCH string.
func parseSemver(tag string) (int, int, int) {
	var ma, mi, pa int
	fmt.Sscanf(strings.TrimPrefix(tag, "v"), "%d.%d.%d", &ma, &mi, &pa)
	return ma, mi, pa
}
