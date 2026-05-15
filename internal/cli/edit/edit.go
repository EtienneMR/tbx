// Package edit implements `tbx edit <path...>`.
//
// The command is designed to be set as $EDITOR.  It picks the best available
// editor for each path type and blocks when needed so callers such as git
// receive the edited content before continuing.
//
//	export EDITOR="tbx edit"
package edit

import (
	"os"
	"os/exec"

	"github.com/EtienneMR/tbx/internal/process"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "edit <path...>",
	Short: "Open files and directories in the best available editor",
	Long:  `Open one or more paths in the best available editor.`,

	Args: cobra.MinimumNArgs(1),
	Run:  run,
}

var vscode = process.New("codium", "code")
var nano = process.New("nano")
var dolphin = process.New("dolphin")

// paths groups the arguments by kind after stat-ing each one.
type paths struct {
	dirs  []string
	other []string
}

func classify(args []string) paths {
	var p paths
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil || !info.IsDir() {
			p.other = append(p.other, arg)
		} else {
			p.dirs = append(p.dirs, arg)
		}
	}
	return p
}

func run(_ *cobra.Command, args []string) {
	p := classify(args)

	if err := vscode.Resolve(); err == nil {
		openVSCode(args, len(p.other) > 0)
		return
	}

	openNano(p.other)
	openDolphin(p.dirs)
}

// openVSCode opens all paths in a single VS Code / VSCodium instance.
func openVSCode(paths []string, hasFiles bool) {
	argv := paths
	if hasFiles {
		argv = append([]string{"--wait"}, paths...)
		tui.Step("Opening in %s (blocking)", vscode.Resolved.Name)
	} else {
		tui.Step("Opening in %s", vscode.Resolved.Name)
	}
	vscode.Live(argv...)
}

// openNano opens each file sequentially in nano.
func openNano(files []string) {
	if len(files) == 0 {
		return
	}
	tui.Check(nano.Resolve(), "edit")

	for _, f := range files {
		file, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0)
		if err == nil {
			file.Close()
			tui.Step("Opening %s", f)
			nano.Live(f)
		} else {

			tui.Step("Opening %s (sudo)", f)
			nano.Sudo().Live(f)
		}
	}
}

// openDolphin opens each directory in a detached Dolphin window.
func openDolphin(dirs []string) {
	if len(dirs) == 0 {
		return
	}
	tui.Check(nano.Resolve(), "edit")

	for _, d := range dirs {

		tui.Step("Opening %s", d)

		cmd := dolphin.Command(d)
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Start(); err != nil {
			tui.Fatal("dolphin: %v", err)
		}

		go func(c *exec.Cmd) { _ = c.Wait() }(cmd)
	}
}
