package web

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/EtienneMR/tbx/internal/dirs"
	"github.com/EtienneMR/tbx/internal/tool"
	"github.com/EtienneMR/tbx/internal/tui"
)

func codiumServer(path string) (*exec.Cmd, string, string) {
	state_dir := filepath.Join(dirs.StateHome(), "codium")
	err := os.MkdirAll(state_dir, 0o755)
	tui.Check(err, "codiumServer: state_dir")

	cmd := tool.Codium.Exec().
		Command(
			"--server-data-dir",
			filepath.Join(state_dir, "server-data"),
			"--user-data-dir",
			filepath.Join(state_dir, "user-data"),
			"--extensions-dir",
			filepath.Join(state_dir, "extensions"),
			"--port",
			"0",
		)
	cmd.Dir = path

	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	tui.Check(err, "codiumServer: pipe stdout")

	tui.Check(cmd.Start(), "codiumServer: start")

	patterns, err := extractFirstRegex(stdout, regexp.MustCompile(`(http://localhost:\d+)([^\s]*)?`))
	tui.Check(err, "codiumServer: extract server url")

	go func() {
		io.Copy(io.Discard, stdout)
	}()

	return cmd, string(patterns[1]), string(patterns[2])
}

func cloudflareTunnel(url string) (*exec.Cmd, string) {
	cmd := tool.Cloudflared.Exec().Command("--url", url)

	stderr, err := cmd.StderrPipe()
	tui.Check(err, "cloudflareTunnel: pipe stderr")

	tui.Check(cmd.Start(), "cloudflareTunnel: start")

	patterns, err := extractFirstRegex(stderr, regexp.MustCompile(`https://[A-Za-z0-9._-]+\.trycloudflare\.com`))
	tui.Check(err, "cloudflareTunnel: extract tunnel url")

	go func() {
		io.Copy(io.Discard, stderr)
	}()

	time.Sleep(4 * time.Second) // let DNS settle

	return cmd, string(patterns[0])
}
