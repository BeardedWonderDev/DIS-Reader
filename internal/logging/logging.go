package logging

import (
	"log/slog"
	"net/url"
	"strings"
	"time"
)

// CommonAttrs builds a consistent set of slog attributes used across agent/bridge logs.
func CommonAttrs(tenantID, agentID, jobID, phase, reqKind string) []any {
	attrs := make([]any, 0, 5)
	if tenantID != "" {
		attrs = append(attrs, slog.String("tenant_id", tenantID))
	}
	if agentID != "" {
		attrs = append(attrs, slog.String("agent_id", agentID))
	}
	if jobID != "" {
		attrs = append(attrs, slog.String("job_id", jobID))
	}
	if phase != "" {
		attrs = append(attrs, slog.String("phase", phase))
	}
	if reqKind != "" {
		attrs = append(attrs, slog.String("req_kind", reqKind))
	}
	return attrs
}

// DurationAttr records a millisecond duration under a consistent key.
func DurationAttr(d time.Duration) slog.Attr {
	return slog.Int64("duration_ms", d.Milliseconds())
}

// RowsAttrs records row counts and truncation state for query jobs.
func RowsAttrs(rows int, truncated bool) []any {
	attrs := []any{slog.Int("rows", rows)}
	if truncated {
		attrs = append(attrs, slog.Bool("truncated", true))
	}
	return attrs
}

// SQLHint returns a minimal, redacted hint about a SQL statement (first keyword only).
func SQLHint(sql string) string {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return ""
	}
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[0])
}

// RedactURLHost returns only the scheme://host:port portion, dropping path/query/credentials.
func RedactURLHost(raw string) string {
	if raw == "" {
		return ""
	}
	if u, err := url.Parse(raw); err == nil && u.Scheme != "" && u.Host != "" {
		host := u.Host
		if u.Port() == "" && strings.Contains(host, ":") {
			// already contains port or IPv6 literal; leave as-is
		}
		return u.Scheme + "://" + host
	}
	// Fallback: return raw minus any path/query if present
	if cut := strings.IndexAny(raw, "/?"); cut > 0 {
		return raw[:cut]
	}
	return raw
}

// MapLogLevel converts agent log strings into slog levels.
func MapLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	case "info", "":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}
