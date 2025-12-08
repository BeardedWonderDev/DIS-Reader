package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTarGz(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "asset-*.tar.gz")
	if err != nil {
		t.Fatalf("temp: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	gz := gzip.NewWriter(tmpFile)
	tw := tar.NewWriter(gz)
	content := []byte("hello")
	hdr := &tar.Header{Name: "dis-agent", Mode: 0o755, Size: int64(len(content))}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("write: %v", err)
	}
	tw.Close()
	gz.Close()
	tmpFile.Close()

	dest := t.TempDir()
	path, err := extractTarGz(tmpFile.Name(), dest)
	if err != nil {
		t.Fatalf("extractTarGz error: %v", err)
	}
	if path != filepath.Join(dest, "dis-agent") {
		t.Fatalf("unexpected path %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "hello" {
		t.Fatalf("expected extracted content, err=%v data=%q", err, string(data))
	}
}

func TestExtractZip(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "asset-*.zip")
	if err != nil {
		t.Fatalf("temp: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	zw := zip.NewWriter(tmpFile)
	w, err := zw.Create("dis-agent.exe")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := io.WriteString(w, "hello"); err != nil {
		t.Fatalf("write: %v", err)
	}
	zw.Close()
	tmpFile.Close()

	dest := t.TempDir()
	path, err := extractZip(tmpFile.Name(), dest)
	if err != nil {
		t.Fatalf("extractZip error: %v", err)
	}
	if path != filepath.Join(dest, "dis-agent.exe") {
		t.Fatalf("unexpected path %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "hello" {
		t.Fatalf("expected extracted content, err=%v data=%q", err, string(data))
	}
}

func TestArchSuffix(t *testing.T) {
	if archSuffix("arm64") != "arm64" {
		t.Fatalf("expected arm64 suffix")
	}
	if archSuffix("amd64") != "amd64" {
		t.Fatalf("expected passthrough")
	}
}
