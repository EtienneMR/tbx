package dirs

import (
	"os"
	"path/filepath"
)

const appName = "tbx"

// StateHome returns $XDG_STATE_HOME/tbx (default: ~/.local/state/tbx).
func StateHome() string {
	return filepath.Join(xdgSubdir("XDG_STATE_HOME", ".local/state"), appName)
}

// DataHome returns $XDG_DATA_HOME/tbx (default: ~/.local/share/tbx).
func DataHome() string {
	return filepath.Join(xdgSubdir("XDG_DATA_HOME", ".local/share"), appName)
}

// BinHome returns the directory where user-installed binaries live.
func BinHome() string {
	return xdgSubdir("XDG_BIN_HOME", ".local/bin")
}

func xdgSubdir(envKey string, fallbackRel ...string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return filepath.Join(mustHome(), filepath.Join(fallbackRel...))
}

func mustHome() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}
	return h
}
