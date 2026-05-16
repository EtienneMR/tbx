package web

import (
	"os"
	"strings"

	"github.com/EtienneMR/tbx/internal/tool"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "web",
	Short: "Web-related helpers",
	Long:  "Helpers for opening local directories and URLs through Cloudflare tunnels.",
}

var codiumCmd = &cobra.Command{
	Use:               "codium [path]",
	Short:             "Open a directory in Codium and expose it over a tunnel",
	Long:              "Starts Codium for a local directory, then exposes it through a Cloudflare tunnel. If no path is provided, the current directory is used.",
	Example:           "tbx web codium\n tbx web codium ~/projects/my-app",
	Args:              cobra.RangeArgs(0, 1),
	ValidArgsFunction: completeDirArg,
	Run:               runCodiumCmd,
}

var tunnelCmd = &cobra.Command{
	Use:     "tunnel <url> [path]",
	Short:   "Expose an existing URL through a tunnel",
	Long:    "Starts a Cloudflare tunnel to an existing URL and prints the public URL. The optional path defaults to '/'.",
	Example: "tbx web tunnel http://localhost:3000\n tbx web tunnel http://localhost:3000 /app",
	Args:    cobra.RangeArgs(1, 2),
	Run:     runTunnelCmd,
}

func init() {
	Cmd.AddCommand(codiumCmd, tunnelCmd)
}

func completeDirArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveFilterDirs
}

func runCodiumCmd(cmd *cobra.Command, args []string) {
	path := "."
	if len(args) == 1 {
		path = args[0]
	}

	if info, err := os.Stat(path); err != nil {
		tui.Fatal("codium path %q: %s", path, err)
	} else if !info.IsDir() {
		tui.Fatal("codium path %q is not a directory", path)
	}

	tui.Info("Updating Codium and Cloudflared")
	if _, err := tool.Codium.Update(true, true); err != nil {
		tui.Fatal("updating Codium: %s", err)
	}
	if _, err := tool.Cloudflared.Update(true, true); err != nil {
		tui.Fatal("updating Cloudflared: %s", err)
	}

	tui.Info("Starting Codium for %s", path)
	codiumCommand, codiumHost, codiumPath := codiumServer(path)
	defer func() {
		if codiumCommand != nil && codiumCommand.Process != nil {
			_ = codiumCommand.Process.Kill()
		}
	}()

	tui.Info("Starting Cloudflared tunnel")
	cloudflareCommand, cloudflareHost := cloudflareTunnel(codiumHost)
	defer func() {
		if cloudflareCommand != nil && cloudflareCommand.Process != nil {
			_ = cloudflareCommand.Process.Kill()
		}
	}()

	tui.Blank()
	tui.Success("Editor available at %s%s", cloudflareHost, codiumPath)

	if _, err := codiumCommand.Process.Wait(); err != nil {
		tui.Fatal("codium server stopped: %s", err)
	}
}

func runTunnelCmd(cmd *cobra.Command, args []string) {
	url := args[0]
	path := "/"
	if len(args) > 1 && args[1] != "" {
		path = args[1]
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
	}

	tui.Info("Starting Cloudflared tunnel to %s", url)

	cloudflareCommand, cloudflareHost := cloudflareTunnel(url)
	defer func() {
		if cloudflareCommand != nil && cloudflareCommand.Process != nil {
			_ = cloudflareCommand.Process.Kill()
		}
	}()

	tui.Blank()
	tui.Success("Tunnel available at %s%s", cloudflareHost, path)

	if _, err := cloudflareCommand.Process.Wait(); err != nil {
		tui.Fatal("cloudflared tunnel stopped: %s", err)
	}
}
