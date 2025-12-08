package installer

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// mapTransport lets us stub http.DefaultClient calls per-URL.
type mapTransport map[string]func(*http.Request) (*http.Response, error)

func (m mapTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if fn, ok := m[r.URL.String()]; ok {
		return fn(r)
	}
	if fn, ok := m[r.URL.Path]; ok {
		return fn(r)
	}
	if fn, ok := m[r.URL.Host+r.URL.Path]; ok {
		return fn(r)
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("not found")),
	}, nil
}

func withStubHTTP(t *testing.T, tr http.RoundTripper) {
	t.Helper()
	orig := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: tr}
	t.Cleanup(func() { http.DefaultClient = orig })
}

func buildAssetName(version string) (name, url, ext string) {
	suffix := "_" + runtime.GOOS + "_" + archSuffix(runtime.GOARCH)
	ext = ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	name = "dis-agent_" + version + suffix + ext
	url = "https://github.com/" + repo + "/releases/download/" + version + "/" + name
	return
}

func makeArchive(payload string) ([]byte, error) {
	if runtime.GOOS == "windows" {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		w, err := zw.Create("dis-agent.exe")
		if err != nil {
			return nil, err
		}
		if _, err := io.WriteString(w, payload); err != nil {
			return nil, err
		}
		if err := zw.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	data := []byte(payload)
	if err := tw.WriteHeader(&tar.Header{
		Name: "dis-agent",
		Mode: 0o755,
		Size: int64(len(data)),
	}); err != nil {
		return nil, err
	}
	if _, err := tw.Write(data); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestInstallExplicitVersion(t *testing.T) {
	version := "v9.9.9"
	name, downloadURL, _ := buildAssetName(version)

	assetBytes, err := makeArchive("payload")
	require.NoError(t, err)

	withStubHTTP(t, mapTransport{
		downloadURL: func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(assetBytes)),
			}, nil
		},
	})

	dest := t.TempDir()
	binPath, err := Install(context.Background(), version, dest)
	require.NoError(t, err)

	contents, err := os.ReadFile(binPath)
	require.NoError(t, err)
	require.Equal(t, "payload", string(contents))

	if runtime.GOOS == "windows" {
		require.True(t, strings.HasSuffix(binPath, ".exe"))
	} else {
		require.Equal(t, filepath.Join(dest, "dis-agent"), binPath)
		info, statErr := os.Stat(binPath)
		require.NoError(t, statErr)
		require.NotZero(t, info.Mode()&0o111, "binary should be executable")
	}

	// sanity: computed name matches arch suffix expectation (indirect coverage)
	require.Contains(t, name, archSuffix(runtime.GOARCH))
}

func TestInstallLatestUsesMatchingAsset(t *testing.T) {
	name, downloadURL, ext := buildAssetName("v2.0.0")
	assetBytes, err := makeArchive("latest-payload")
	require.NoError(t, err)

	latestURL := "https://api.github.com/repos/" + repo + "/releases/latest"
	latestPayload := map[string]any{
		"tag_name": "v2.0.0",
		"assets": []map[string]string{
			{"name": name, "browser_download_url": downloadURL},
		},
	}
	jsonBytes, _ := json.Marshal(latestPayload)

	withStubHTTP(t, mapTransport{
		latestURL: func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
			}, nil
		},
		downloadURL: func(r *http.Request) (*http.Response, error) {
			// accept the resolved download URL
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(assetBytes)),
			}, nil
		},
	})

	dest := t.TempDir()
	binPath, err := Install(context.Background(), "latest", dest)
	require.NoError(t, err)
	require.Contains(t, binPath, "dis-agent")
	require.True(t, strings.HasSuffix(name, ext))

	binBytes, readErr := os.ReadFile(binPath)
	require.NoError(t, readErr)
	require.Equal(t, "latest-payload", string(binBytes))
}

func TestInstallLatestNoMatchingAsset(t *testing.T) {
	latestURL := "https://api.github.com/repos/" + repo + "/releases/latest"
	payload := map[string]any{
		"tag_name": "v1.0.0",
		"assets": []map[string]string{
			{"name": "unrelated_asset.tar.gz", "browser_download_url": "https://example.com/none"},
		},
	}
	jsonBytes, _ := json.Marshal(payload)

	withStubHTTP(t, mapTransport{
		latestURL: func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
			}, nil
		},
	})

	_, err := Install(context.Background(), "latest", t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "matching asset not found")
}

func TestInstallDefaultDestDirStillErrorsSafely(t *testing.T) {
	// destDir intentionally empty to cover default path selection without writing there.
	latestURL := "https://api.github.com/repos/" + repo + "/releases/latest"
	payload := map[string]any{
		"tag_name": "v1.0.0",
		"assets":   []map[string]string{}, // force resolveAsset failure
	}
	jsonBytes, _ := json.Marshal(payload)

	withStubHTTP(t, mapTransport{
		latestURL: func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
			}, nil
		},
	})

	_, err := Install(context.Background(), "latest", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "matching asset not found")
}

func TestDownloadHTTPError(t *testing.T) {
	url := "https://example.com/fail"
	withStubHTTP(t, mapTransport{
		url: func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Status:     "500 Internal Server Error",
				Body:       io.NopCloser(strings.NewReader("boom")),
			}, nil
		},
	})

	err := download(context.Background(), url, io.Discard)
	require.Error(t, err)
	require.Contains(t, err.Error(), "download failed")
}

func TestExtractArchivesMissingBinary(t *testing.T) {
	// zip without expected binary
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	_, err := zw.Create("other.exe")
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	zipPath := filepath.Join(t.TempDir(), "missing.zip")
	require.NoError(t, os.WriteFile(zipPath, zipBuf.Bytes(), 0o644))
	_, err = extractZip(zipPath, t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "dis-agent.exe")

	// tar.gz without expected binary
	var tarBuf bytes.Buffer
	gz := gzip.NewWriter(&tarBuf)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: "other", Mode: 0o644, Size: 0}))
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	tarPath := filepath.Join(t.TempDir(), "missing.tar.gz")
	require.NoError(t, os.WriteFile(tarPath, tarBuf.Bytes(), 0o644))
	_, err = extractTarGz(tarPath, t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "dis-agent")
}
