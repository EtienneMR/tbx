package tool

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EtienneMR/tbx/internal/dirs"
	"github.com/EtienneMR/tbx/internal/process"
	"github.com/EtienneMR/tbx/internal/tui"
)

// Tool describes a GitHub-released binary that tbx can install and manage.
type Tool struct {
	// Name is the short display name shown in logs and listings (e.g. "codium").
	Name string

	// Repo is the GitHub "owner/repo" slug (e.g. "VSCodium/vscodium").
	Repo string

	// BinRelPath is the path to the executable inside the tool's install dir.
	BinRelPath string

	// Download populates destDir with the tool binaries for the given version.
	// It is called with a clean staging directory; the VERSION file is written
	// automatically after Download returns.
	Download func(version, destDir string) error
}

// InstallDir returns the root install directory for this tool.
func (t *Tool) InstallDir() string {
	return filepath.Join(dirs.DataHome(), "tools", t.Repo)
}

// BinaryPath returns the absolute path to the tool's main executable.
func (t *Tool) BinaryPath() string {
	return filepath.Join(t.InstallDir(), t.BinRelPath)
}

// InstalledVersion reads the VERSION file written during installation.
// Returns an empty string when the tool is not installed.
func (t *Tool) InstalledVersion() string {
	data, err := os.ReadFile(filepath.Join(t.InstallDir(), "VERSION"))
	if err != nil {
		return ""
	}
	v := string(data)
	// trim trailing newline written by some editors
	for len(v) > 0 && (v[len(v)-1] == '\n' || v[len(v)-1] == '\r') {
		v = v[:len(v)-1]
	}
	return v
}

// LatestVersion queries the GitHub releases API for the newest tag.
func (t *Tool) LatestVersion() (string, error) {
	return latestRelease(t.Repo)
}

// Cmd returns a ready-to-use *process.Process pointing at the installed binary.
func (t *Tool) Exec() *process.Process {
	return process.NewResolved(t.Name, t.BinaryPath())
}

// Install downloads version into an atomic staging directory, then swaps it
// into place.  An existing install is preserved until the swap succeeds.
func (t *Tool) Install(version string) error {
	staging := t.InstallDir() + ".tmp"

	// Clean any leftover staging dir from a previous failed attempt.
	if err := os.RemoveAll(staging); err != nil {
		return fmt.Errorf("ttool: clean staging for %s: %s", t.Name, err)
	}
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return fmt.Errorf("ttool: mkdir staging for %s: %s", t.Name, err)
	}

	tui.Step("Downloading %s %s", t.Name, version)
	defer os.RemoveAll(staging)

	if err := t.Download(version, staging); err != nil {
		return fmt.Errorf("ttool: download %s %s: %w", t.Name, version, err)
	}

	if err := os.WriteFile(filepath.Join(staging, "VERSION"), []byte(version), 0o644); err != nil {
		return fmt.Errorf("ttool: write VERSION for %s: %w", t.Name, err)
	}

	if err := os.RemoveAll(t.InstallDir()); err != nil {
		return fmt.Errorf("ttool: remove old install of %s: %w", t.Name, err)
	}
	if err := os.Rename(staging, t.InstallDir()); err != nil {
		return fmt.Errorf("ttool: finalize install of %s: %w", t.Name, err)
	}

	tui.Success("Installed %s %s", t.Name, version)
	return nil
}

// Update checks whether a newer version is available, and installs it when so.
// It returns true when an installation was performed.
func (t *Tool) Update(install bool, prompt bool) (bool, error) {
	current := t.InstalledVersion()

	if current == "" && !install {
		tui.Debug("%s is not installed — skipping", t.Name)
		return false, nil
	}

	tui.Step("Checking %s", t.Name)
	latest, err := t.LatestVersion()
	if err != nil {
		return false, err
	}

	if current == latest {
		tui.Info("%s is up to date %s", t.Name, current)
		return false, nil
	}

	if current != "" {
		tui.Info("Update available for %s: %s → %s", t.Name, current, latest)
	}

	if prompt {
		label := fmt.Sprintf("Install %s %s?", t.Name, latest)
		ok, err := tui.Confirm(label, true)
		if err != nil || !ok {
			return false, err
		}
	}

	if err := t.Install(latest); err != nil {
		return true, err
	}

	return true, nil
}
