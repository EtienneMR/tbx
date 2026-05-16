package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/EtienneMR/tbx/internal/dirs"
	"github.com/EtienneMR/tbx/internal/process"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show system health: last update time and reboot status",
	Run: func(cmd *cobra.Command, _ []string) {
		quiet, _ := cmd.Flags().GetBool("quiet")
		runStatus(quiet)
	},
}

func init() {
	statusCmd.Flags().BoolP("quiet", "q", false, "suppress informational messages")
}

func runStatus(quiet bool) {
	updatedFile := filepath.Join(dirs.StateHome(), "system", "updated")

	info, err := os.Stat(updatedFile)
	if err != nil {
		if !quiet {
			tui.Warn("System has never been updated using tbx system update")
		}
	} else {
		days := time.Since(info.ModTime()).Hours() / 24
		msg := fmt.Sprintf("System last updated %dd ago", int(days))
		switch {
		case days >= 14:
			tui.Error("%s — run: tbx system update", msg)
		case days >= 7:
			tui.Warn("%s — run: tbx system update", msg)
		case !quiet && days >= 2:
			tui.Info("%s", msg)
		case !quiet:
			tui.Success("%s", msg)
		}
	}

	if rebootRequired() {
		tui.Warn("System reboot required")
	}
}

// rebootRequired returns true when either:
func rebootRequired() bool {
	if _, err := os.Stat("/run/reboot-required"); err == nil {
		return true
	}

	res, err := process.New("pacman").Run("-Q", "linux")
	if err != nil {
		return false
	}

	// Output: "linux 6.9.3.arch1-1"
	parts := strings.Fields(res.Stdout)
	if len(parts) < 2 {
		return false
	}

	installed := strings.ReplaceAll(parts[1], "-", ".")
	running := runningKernel()
	if running == "" {
		return false
	}

	return strings.ReplaceAll(running, "-", ".") != installed
}

// runningKernel returns the kernel release string
func runningKernel() string {
	res, err := process.New("uname").Run("-r")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(res.Stdout)
}
