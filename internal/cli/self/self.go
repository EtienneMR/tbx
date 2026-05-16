package self

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/EtienneMR/tbx/internal/dirs"
	"github.com/EtienneMR/tbx/internal/tool"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "self",
	Short: "Manage tbx installation",
	Long:  `Install or update tbx.`,
}

func init() {
	Cmd.AddCommand(installCmd, updateCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install or update tbx and set up shell completion",
	Long: `Download and install the latest tbx release into
~/.local/share/tbx/tools/ and symlink the binary to ~/.local/bin/tbx.

Then, optionally install shell completion for bash, zsh, or fish.

  export EDITOR="tbx edit"   # use tbx as your $EDITOR afterwards`,
	Args: cobra.NoArgs,
	Run:  runInstallCmd,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all tbx-managed tools",
	Long:  `Update all installed tbx-managed tools, including itself.`,
	Args:  cobra.NoArgs,
	Run:   runUpdateCmd,
}

func runInstallCmd(cmd *cobra.Command, _ []string) {
	tui.Header("Installing tbx")

	tui.Step("Installing tbx")
	_, err := tool.Tbx.Update(true, false)
	tui.Check(err, "self install update")

	err = os.Symlink(tool.Tbx.BinaryPath(), filepath.Join(dirs.BinHome(), tool.Tbx.Name))
	if err != nil && !errors.Is(err, os.ErrExist) {
		tui.Check(err, "self install link")
	}

	tui.Step("Installing completion")

	shell := pickShell()
	if shell != "" {
		installCompletion(cmd.Root(), shell)
	}
}

func runUpdateCmd(cmd *cobra.Command, _ []string) {
	tui.Header("Updating tools")

	updated, err := tool.Tbx.Update(true, false)
	tui.Check(err, "self update")

	if updated {
		tui.Info("tbx has been updated, restarting command")
		tool.Tbx.Exec().Live(os.Args[1:]...)
		os.Exit(0)
	} else {
		for _, t := range tool.All {
			if t != tool.Tbx {
				t.Update(false, false)
			}
		}
	}
}

// pickShell presents a shell selector, pre-ordering by the detected $SHELL.
func pickShell() string {
	const noCompletion = "no completion"
	supported := []string{"bash", "zsh", "fish", noCompletion}

	detected := filepath.Base(os.Getenv("SHELL"))
	ordered := make([]string, 0, len(supported))

	// Put the detected shell first so it is pre-highlighted.
	for _, s := range supported {
		if strings.Contains(s, detected) {
			ordered = append([]string{s}, ordered...)
		} else {
			ordered = append(ordered, s)
		}
	}

	choice, err := tui.Select("Install completion ?", ordered)
	tui.Check(err, "install pickShell")
	if choice == noCompletion {
		return ""
	}
	return choice
}

// shellCfg describes where to write the completion script and what to print afterwards.
type shellCfg struct {
	// dest is the absolute path of the file to write.
	dest string
	// gen writes the completion script for this shell to w.
	gen func(w io.Writer) error
	// hint is printed after a successful install (empty = no hint needed).
	hint string
}

func installCompletion(root *cobra.Command, shell string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	cfg, err := shellConfig(root, shell, home)
	if err != nil {
		return err
	}

	tui.Step("Writing %s completion", shell)

	if err := os.MkdirAll(filepath.Dir(cfg.dest), 0o755); err != nil {
		return fmt.Errorf("create completion dir: %w", err)
	}

	f, err := os.Create(cfg.dest)
	if err != nil {
		return fmt.Errorf("create %s: %w", cfg.dest, err)
	}
	defer f.Close()

	if err := cfg.gen(f); err != nil {
		return fmt.Errorf("generate %s completion: %w", shell, err)
	}

	tui.Success("Completion written to %s", cfg.dest)

	if cfg.hint != "" {
		tui.Blank()
		tui.Info("%s", cfg.hint)
	}

	return nil
}

func shellConfig(root *cobra.Command, shell, home string) (shellCfg, error) {
	switch shell {

	case "bash":
		// bash-completion v2 auto-sources ~/.local/share/bash-completion/completions/
		// No .bashrc change needed when bash-completion is installed.
		dest := filepath.Join(home, ".local", "share", "bash-completion", "completions", "tbx")
		return shellCfg{
			dest: dest,
			gen:  func(w io.Writer) error { return root.GenBashCompletionV2(w, true) },
			hint: "Restart your shell or run:\n" +
				"  source " + dest,
		}, nil

	case "zsh":
		// ~/.local/share/zsh/site-functions follows the XDG convention and is
		// picked up automatically by many zsh setups (e.g. Homebrew, zimfw).
		// Users who manage $fpath manually should add the directory if missing.
		dest := filepath.Join(home, ".local", "share", "zsh", "site-functions", "_tbx")
		dir := filepath.Dir(dest)
		return shellCfg{
			dest: dest,
			gen:  func(w io.Writer) error { return root.GenZshCompletion(w) },
			hint: "If completion is not active, add to ~/.zshrc:\n" +
				"  fpath=(" + dir + " $fpath)\n" +
				"  autoload -U compinit && compinit",
		}, nil

	case "fish":
		// Fish auto-sources every file in ~/.config/fish/completions/ — no
		// shell config change needed.
		dest := filepath.Join(home, ".config", "fish", "completions", "tbx.fish")
		return shellCfg{
			dest: dest,
			gen:  func(w io.Writer) error { return root.GenFishCompletion(w, true) },
			hint: "",
		}, nil

	default:
		return shellCfg{}, fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish)", shell)
	}
}
