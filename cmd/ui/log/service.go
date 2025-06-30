package logUI

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
	"github.com/rivo/tview"
)

type LogPanelWriter struct {
	UI      typesUI.UI
	View    *tview.TextView
	Logger  *slog.Logger
	LogChan chan string
}

func InitLogPanel(u typesUI.UI) *LogPanelWriter {
	logPanel := &LogPanelWriter{
		UI:      u,
		LogChan: make(chan string, 100),
	}

	logPanel.View = logPanel.initLogPanelView()

	handler := &TUIViewHandler{PW: logPanel}
	logPanel.Logger = slog.New(handler)

	go func() {
		for line := range logPanel.LogChan {
			u.GetApp().QueueUpdate(func() {
				fmt.Fprint(logPanel.View, line)
				logPanel.View.ScrollToEnd()
			})
		}
	}()

	return logPanel
}

func (tw *LogPanelWriter) Write(p []byte) (n int, err error) {
	logLine := string(p)

	levelColor := "[white]"
	switch {
	case strings.Contains(logLine, "DEBUG"):
		levelColor = "[gray]"
	case strings.Contains(logLine, "INFO"):
		levelColor = "[green]"
	case strings.Contains(logLine, "WARN"):
		levelColor = "[yellow]"
	case strings.Contains(logLine, "ERROR"):
		levelColor = "[red]"
	}

	timestamp := time.Now().Format("15:04:05")
	formatted := fmt.Sprintf("%s[%s]%s %s\n", levelColor, timestamp, "[white]", strings.TrimSpace(logLine))

	tw.LogChan <- formatted
	return len(p), nil
}

type TUIViewHandler struct {
	PW *LogPanelWriter
}

func (h *TUIViewHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelDebug
}

func (h *TUIViewHandler) Handle(ctx context.Context, r slog.Record) error {
	var levelColor string
	switch r.Level {
	case slog.LevelDebug:
		levelColor = "[gray]"
	case slog.LevelInfo:
		levelColor = "[green]"
	case slog.LevelWarn:
		levelColor = "[yellow]"
	case slog.LevelError:
		levelColor = "[red]"
	default:
		levelColor = "[white]"
	}

	timestamp := time.Now().Format("15:04:05")
	msg := r.Message
	formatted := fmt.Sprintf("%s[%s]%s %s\n", levelColor, timestamp, "[white]", msg)

	h.PW.LogChan <- formatted
	return nil
}

func (h *TUIViewHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TUIViewHandler) WithGroup(name string) slog.Handler {
	return h
}
