package database

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAs400DateExprGuardsInvalidDates(t *testing.T) {
	col := "FILEC.IAH.AHPDTE"
	expr := As400DateExpr(col)
	requiredSubstrings := []string{
		"CASE ",
		"MOD(INTEGER",
		fmt.Sprintf("RIGHT('000000'||TRIM(CHAR(%s)),6)", col),
	}
	for _, substr := range requiredSubstrings {
		if !strings.Contains(expr, substr) {
			t.Fatalf("guarded expression missing %q: %s", substr, expr)
		}
	}
}

func TestAs400DateHookEpochMillis(t *testing.T) {
	hook := as400DateHook().(func(reflect.Type, reflect.Type, interface{}) (interface{}, error))
	from := reflect.TypeOf(float64(0))
	to := reflect.TypeOf(time.Time{})
	const ms = 1704355200000 // 2024-01-04T08:00:00Z
	want := time.UnixMilli(ms).UTC()
	got, err := hook(from, to, float64(ms))
	if err != nil {
		t.Fatalf("hook returned error: %v", err)
	}
	parsed := got.(time.Time)
	if !parsed.Equal(want) {
		t.Fatalf("expected %s, got %s", want, parsed)
	}
	ptrType := reflect.TypeOf(&time.Time{})
	gotPtr, err := hook(from, ptrType, float64(ms))
	if err != nil {
		t.Fatalf("hook pointer conversion error: %v", err)
	}
	timePtr, ok := gotPtr.(*time.Time)
	if !ok || timePtr == nil {
		t.Fatalf("expected *time.Time, got %#v", gotPtr)
	}
	if !timePtr.Equal(want) {
		t.Fatalf("expected pointer %s, got %s", want, timePtr)
	}
}

func TestAs400DateHookEpochSeconds(t *testing.T) {
	hook := as400DateHook().(func(reflect.Type, reflect.Type, interface{}) (interface{}, error))
	from := reflect.TypeOf(float64(0))
	const secs = 1700000000 // 2023-11-14T22:13:20Z
	want := time.Unix(secs, 0).UTC()
	got, err := hook(from, reflect.TypeOf(time.Time{}), float64(secs))
	if err != nil {
		t.Fatalf("hook returned error: %v", err)
	}
	if parsed := got.(time.Time); !parsed.Equal(want) {
		t.Fatalf("expected %s, got %s", want, parsed)
	}
}

type stubHandler struct{ records []slog.Record }

func (h *stubHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *stubHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *stubHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *stubHandler) WithGroup(name string) slog.Handler       { return h }

func TestLogStructuredLineParsesJSONAndPlain(t *testing.T) {
	h := &stubHandler{}
	logger := slog.New(h)

	logStructuredLine(logger, `{"status":"ok","message":"hello","requestId":"abc"}`, slog.LevelInfo)
	logStructuredLine(logger, `{"status":"err","message":"bad"}`, slog.LevelError)
	logStructuredLine(logger, "not-json", slog.LevelError)

	if len(h.records) < 3 {
		t.Fatalf("expected at least 3 records, got %d", len(h.records))
	}
	if h.records[0].Level != slog.LevelInfo || h.records[1].Level != slog.LevelError {
		t.Fatalf("unexpected levels: %v %v", h.records[0].Level, h.records[1].Level)
	}

	foundRequest := false
	h.records[0].Attrs(func(a slog.Attr) bool {
		if a.Key == "request_id" && a.Value.String() == "abc" {
			foundRequest = true
		}
		return true
	})
	if !foundRequest {
		t.Fatalf("expected request_id attribute in first record")
	}
}
