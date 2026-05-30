// Package packages implements the `tbx packages` command group.
//
// It wraps any AUR helper or plain pacman that is present on the system,
// preferring paru → yay → pacman in that order.  When the resolved manager
// is plain pacman, privilege escalation through sudo is added automatically
// for write operations; AUR helpers handle that themselves.
package pkg

import (
	"github.com/EtienneMR/tbx/internal/process"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

// Cmd is the `tbx package` root command.
var Cmd = &cobra.Command{
	Use:   "package",
	Short: "Manage system packages via paru / yay / pacman",
}

func init() {
	Cmd.AddCommand(
		installCmd,
		tempCmd,
		searchCmd,
		orphanCmd,
		explicitCmd,
	)
}

var pacman = process.New("pacman")

var pm = process.New("paru", "yay", "pacman")

func UpdatePackages() {
	tui.Check(pm.Resolve(), "resolve package manager")

	tui.Header("Updating system packages")
	writePm().Live("-Syu")

	removeOrphans()
}

func writePm() *process.Process {
	if err := pm.Resolve(); err != nil {
		tui.Fatal("no package manager found: %v", err)
	}
	if pm.Resolved.Name == "pacman" {
		return pm.Sudo()
	}
	return pm
}
