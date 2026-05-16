package system

import (
	"os"
	"path/filepath"

	"github.com/EtienneMR/tbx/internal/cli/pkg"
	"github.com/EtienneMR/tbx/internal/cli/self"
	"github.com/EtienneMR/tbx/internal/dirs"
	"github.com/EtienneMR/tbx/internal/process"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update everything: system packages, language toolchains, and firmware",
	Run:   func(_ *cobra.Command, _ []string) { runUpdate() },
}

func runUpdate() {
	self.UpdateTools()
	pkg.UpdatePackages()

	updateIf("flatpak", func(p *process.Process) {
		tui.Step("User flatpaks")
		p.Live("update", "--user")
		tui.Step("System flatpaks")
		p.Sudo().Live("update", "--system")
	})

	updateIf("rustup", func(p *process.Process) { p.Live("update") })
	updateIf("uv", func(p *process.Process) { p.Live("tool", "upgrade", "--all") })
	updateIf("pipx", func(p *process.Process) { p.Live("upgrade-all") })
	updateIf("pnpm", func(p *process.Process) { p.Live("update", "-g") })
	updateIf("npm", func(p *process.Process) { p.Sudo().Live("update", "-g") })

	updateIf("mandb", func(p *process.Process) { p.Sudo().Live("-q") })
	updateIf("updatedb", func(p *process.Process) { p.Sudo().Live() })
	updateIf("journalctl", func(p *process.Process) { p.Sudo().Live("--vacuum-time=30d") })

	updateIf("fwupdmgr", func(p *process.Process) {
		p.Test(2, "refresh")
		p.Test(2, "upgrade")
	})

	tui.Header("Checking system status")

	dir := filepath.Join(dirs.StateHome(), "system")
	tui.Check(os.MkdirAll(dir, 0o755), "system update: create state dir")
	f, err := os.Create(filepath.Join(dir, "updated"))
	tui.Check(err, "system update: write timestamp")
	f.Close()

	if rebootRequired() {
		tui.Warn("Reboot required")
		ok, err := tui.Confirm("Reboot now?", false)
		if ok && err != nil {
			process.New("reboot").Sudo().Live()
		}
	}

	tui.Blank()
	tui.Success("System updated successfully")
}

// updateIf resolves name in PATH and, when found, prints a section header and
// calls fn with the resolved *Process.  When not found it logs at debug level
// and returns immediately.
func updateIf(name string, fn func(*process.Process)) {
	p := process.New(name)
	if err := p.Resolve(); err != nil {
		tui.Debug("Not updating %s: not installed", name)
		return
	}
	tui.Header("Updating " + name)
	fn(p)
}
