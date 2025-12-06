package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const repo = "BeardedWonderDev/DIS-Reader"

// InstallLatest downloads the latest GitHub release asset matching the current
// OS/arch, extracts the dis-agent binary, and installs it to destDir.
// Returns the installed binary path.
func InstallLatest(ctx context.Context, destDir string) (string, error) {
	if destDir == "" {
		switch runtime.GOOS {
		case "windows":
			destDir = filepath.Join(os.Getenv("ProgramFiles"), "DIS Agent")
		default:
			destDir = "/usr/local/bin"
		}
	}
	assetName, downloadURL, err := resolveAsset(ctx)
	if err != nil {
		return "", err
	}

	tmp, err := os.CreateTemp("", "dis-agent-asset")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())

	if err := download(ctx, downloadURL, tmp); err != nil {
		return "", fmt.Errorf("download asset: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return "", err
	}

	var binPath string
	if strings.HasSuffix(assetName, ".zip") {
		binPath, err = extractZip(tmp.Name(), destDir)
	} else {
		binPath, err = extractTarGz(tmp.Name(), destDir)
	}
	if err != nil {
		return "", fmt.Errorf("extract: %w", err)
	}
	return binPath, nil
}

func resolveAsset(ctx context.Context) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("github api: %s", resp.Status)
	}
	var payload struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", err
	}
	wantSuffix := fmt.Sprintf("_%s_%s", runtime.GOOS, archSuffix(runtime.GOARCH))
	for _, a := range payload.Assets {
		if strings.Contains(a.Name, wantSuffix) && strings.Contains(a.Name, "dis-agent") {
			return a.Name, a.BrowserDownloadURL, nil
		}
	}
	return "", "", errors.New("matching asset not found")
}

func archSuffix(goarch string) string {
	switch goarch {
	case "arm64":
		return "arm64"
	default:
		return goarch
	}
}

func download(ctx context.Context, url string, w io.Writer) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

func extractZip(path, destDir string) (string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	var bin string
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "dis-agent.exe") {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return "", err
			}
			outPath := filepath.Join(destDir, "dis-agent.exe")
			out, err := os.Create(outPath)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, rc); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			bin = outPath
			break
		}
	}
	if bin == "" {
		return "", errors.New("dis-agent.exe not found in archive")
	}
	return bin, nil
}

func extractTarGz(path, destDir string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	var bin string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		name := filepath.Base(hdr.Name)
		if hdr.Typeflag == tar.TypeReg && (name == "dis-agent" || name == "dis-agent.exe") {
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return "", err
			}
			outPath := filepath.Join(destDir, name)
			out, err := os.Create(outPath)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			if runtime.GOOS != "windows" {
				_ = os.Chmod(outPath, 0o755)
			}
			bin = outPath
			break
		}
	}
	if bin == "" {
		return "", errors.New("dis-agent binary not found in archive")
	}
	return bin, nil
}
