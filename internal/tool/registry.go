package tool

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var All = []*Tool{Tbx, Codium, Cloudflared}

var Tbx = &Tool{
	Name:       "tbx",
	Repo:       "EtienneMR/tbx",
	BinRelPath: "tbx",
	Download:   downloadTbx,
}

var Codium = &Tool{
	Name:       "codium",
	Repo:       "VSCodium/vscodium",
	BinRelPath: filepath.Join("bin", "codium-server"),
	Download:   downloadCodium,
}

var Cloudflared = &Tool{
	Name:       "cloudflared",
	Repo:       "cloudflare/cloudflared",
	BinRelPath: "cloudflared",
	Download:   downloadCloudflared,
}

func downloadTbx(version, dir string) error {
	filename := "tbx-" + runtime.GOOS
	url := ghRelease("EtienneMR/tbx", version, filename)
	return downloadBinary(url, dir, "tbx")
}

func downloadCodium(version, dir string) error {
	arch, err := vscodiumArch()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("vscodium-reh-web-%s-%s-%s.tar.gz", runtime.GOOS, arch, version)
	url := ghRelease("VSCodium/vscodium", version, filename)
	return httpDownloadExtract(url, dir)
}

func vscodiumArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "x64", nil
	case "arm64":
		return "arm64", nil
	case "arm":
		return "armhf", nil
	default:
		return "", fmt.Errorf("ttool: unsupported architecture for codium: %s", runtime.GOARCH)
	}
}

func downloadCloudflared(version, dir string) error {
	filename := fmt.Sprintf("cloudflared-%s-%s", runtime.GOOS, runtime.GOARCH)
	url := ghRelease("cloudflare/cloudflared", version, filename)
	return downloadBinary(url, dir, "cloudflared")
}
