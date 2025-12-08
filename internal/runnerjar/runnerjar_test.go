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
