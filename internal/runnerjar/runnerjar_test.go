package runnerjar

import (
	"os"
	"testing"
)

func TestExtractAndCleanup(t *testing.T) {
	extracted, err := Extract()
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if extracted.Dir == "" || extracted.JarPath == "" {
		t.Fatalf("expected non-empty dir and jar path")
	}
	data, err := os.ReadFile(extracted.JarPath)
	if err != nil {
		t.Fatalf("failed to read jar: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected jar bytes to be written")
	}

	if err := extracted.Cleanup(); err != nil {
		t.Fatalf("cleanup error: %v", err)
	}
	if _, err := os.Stat(extracted.Dir); !os.IsNotExist(err) {
		t.Fatalf("expected temp dir removed, got err=%v", err)
	}
}

func TestCleanupNoDirAndNil(t *testing.T) {
	var nilExtracted *Extracted
	if err := nilExtracted.Cleanup(); err != nil {
		t.Fatalf("nil cleanup returned error: %v", err)
	}

	empty := &Extracted{}
	if err := empty.Cleanup(); err != nil {
		t.Fatalf("empty dir cleanup returned error: %v", err)
	}
}

func TestExtractTempDirError(t *testing.T) {
	// Set TMPDIR to a file path so os.MkdirTemp fails.
	tmpFile, err := os.CreateTemp("", "not-a-dir")
	if err != nil {
		t.Fatalf("setup temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	t.Setenv("TMPDIR", tmpFile.Name())
	if _, err := Extract(); err == nil {
		t.Fatalf("expected Extract to fail when TMPDIR is not a directory")
	}
}
