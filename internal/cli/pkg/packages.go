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
		orphansCmd,
		explicitCmd,
	)
}

var pacman = process.New("pacman")

var pm = process.New("paru", "yay", "pacman")

func writePm() *process.Process {
	if err := pm.Resolve(); err != nil {
		tui.Fatal("no package manager found: %v", err)
	}
	if pm.Resolved.Name == "pacman" {
		spm, err := pm.Sudo()
		if err != nil {
			tui.Fatal("sudo not available: %v", err)
		}
		return spm
	}
	return pm
}
