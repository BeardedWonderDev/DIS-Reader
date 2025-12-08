package logging

import (
	"log/slog"
	"testing"
	"time"
)

func TestCommonAttrs(t *testing.T) {
	attrs := CommonAttrs("tenant-a", "agent-1", "job-9", "phase", "query")
	if len(attrs) != 5 {
		t.Fatalf("expected 5 attrs, got %d", len(attrs))
	}
	attr := attrs[0].(slog.Attr)
	if attr.Key != "tenant_id" || attr.Value.String() != "tenant-a" {
		t.Fatalf("unexpected tenant attr: %#v", attr)
	}
}

func TestSQLHint(t *testing.T) {
	hint := SQLHint(" SELECT * FROM table WHERE id=1 ")
	if hint != "select" {
		t.Fatalf("expected select, got %q", hint)
	}
	if SQLHint("") != "" {
		t.Fatalf("expected empty for blank input")
	}
}

func TestSQLHintWhitespaceOnly(t *testing.T) {
	if SQLHint("   \n\t") != "" {
		t.Fatalf("expected empty string for whitespace input")
	}
}

func TestMapLogLevel(t *testing.T) {
	if lvl := MapLogLevel("warn"); lvl != slog.LevelWarn {
		t.Fatalf("expected warn, got %v", lvl)
	}
	if lvl := MapLogLevel("unknown"); lvl != slog.LevelInfo {
		t.Fatalf("expected default info, got %v", lvl)
	}
}

func TestMapLogLevelVariants(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"  DEBUG ", slog.LevelDebug},
		{"warning", slog.LevelWarn},
		{"err", slog.LevelError},
		{"", slog.LevelInfo},
		{"verbose", slog.LevelInfo},
	}
	for _, tc := range tests {
		if got := MapLogLevel(tc.input); got != tc.want {
			t.Fatalf("input %q: expected %v, got %v", tc.input, tc.want, got)
		}
	}
}

func TestRowsAttrs(t *testing.T) {
	attrs := RowsAttrs(5, true)
	if len(attrs) != 2 {
		t.Fatalf("expected rows+truncated, got %d", len(attrs))
	}
}

func TestDurationAttr(t *testing.T) {
	attr := DurationAttr(1500 * time.Millisecond)
	if attr.Key != "duration_ms" || attr.Value.Int64() != 1500 {
		t.Fatalf("unexpected duration attr: %#v", attr)
	}
}

func TestRedactURLHost(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"https://example.com:8443/api?x=1", "https://example.com:8443"},
		{"https://user:pass@example.com/path", "https://example.com"},
		{"example.com/path?x=1", "example.com"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := RedactURLHost(tc.raw); got != tc.want {
			t.Fatalf("raw %q: expected %q, got %q", tc.raw, tc.want, got)
		}
	}
}
