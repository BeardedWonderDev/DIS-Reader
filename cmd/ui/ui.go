package ui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/cmd/entity"
	authUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/auth"
	debugUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/debug"
	logUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/log"
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
	"github.com/BeardedWonderDev/DIS-Reader/disreader"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	App    *tview.Application
	WinMan *winman.Manager
	Layout *typesUI.ComponentLayout

	DIS   types.DISReaderService
	Auth  *authUI.Auth
	Debug *debugUI.Debug
	Log   *logUI.Log

	Theme *entity.Theme

	Logger *slog.Logger
}

// GetApp implements typesUI.UI.
func (u *UI) GetApp() *tview.Application {
	return u.App
}

// GetDIS implements typesUI.UI.
func (u *UI) GetDIS() types.DISReaderService {
	return u.DIS
}

// GetLayout implements typesUI.UI.
func (u *UI) GetLayout() *typesUI.ComponentLayout {
	return u.Layout
}

// GetTheme implements typesUI.UI.
func (u *UI) GetTheme() *entity.Theme {
	return u.Theme
}

// GetWinMan implements typesUI.UI.
func (u *UI) GetWinMan() *winman.Manager {
	return u.WinMan
}

func (u *UI) GetAuth() typesUI.Auth {
	return u.Auth
}

func (u *UI) GetLogger() *slog.Logger {
	return u.Logger
}

func (u *UI) SetFocus(p tview.Primitive) {
	go u.App.QueueUpdateDraw(func() {
		u.App.SetFocus(p)
	})
}

func (u *UI) Run() error {
	return u.App.Run()
}

func (u *UI) QuitApplication() {
	u.DIS.Shutdown()
	u.App.Stop()
}

type textViewWriter struct {
	app     *tview.Application
	view    *tview.TextView
	logChan chan string
}

func (tw *textViewWriter) Write(p []byte) (n int, err error) {
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

	tw.logChan <- formatted
	return len(p), nil
}

type TUIViewHandler struct {
	tw *textViewWriter
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

	h.tw.logChan <- formatted
	return nil
}

func (h *TUIViewHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TUIViewHandler) WithGroup(name string) slog.Handler {
	return h
}

func NewUI(cfg *types.Config) *UI {
	app := tview.NewApplication()

	logPanel := tview.NewTextView()
	logPanel.SetDynamicColors(true)
	logPanel.SetTitle(" 📃 Logs ")
	logPanel.SetBorder(true)
	logPanel.SetWordWrap(true)
	logPanel.SetBorderPadding(1, 1, 1, 1)

	tw := &textViewWriter{
		app:     app,
		view:    logPanel,
		logChan: make(chan string, 100),
	}

	go func() {
		for line := range tw.logChan {
			tw.app.QueueUpdate(func() {
				fmt.Fprint(tw.view, line)
				tw.view.ScrollToEnd()
			})
		}
	}()

	handler := &TUIViewHandler{tw: tw}

	logger := slog.New(handler)

	wm := winman.NewWindowManager()

	ui := UI{
		App:    app,
		WinMan: wm,
		DIS:    disreader.NewDISReaderService(cfg, logger),
		Theme:  &entity.TerminalTheme,
		Auth:   authUI.NewAuthService(),
		Debug:  debugUI.NewDebugService(),
		Log:    logUI.NewLogService(),
		Logger: logger,
	}

	ui.Auth.UI = &ui
	ui.Debug.UI = &ui
	ui.Log.UI = &ui

	logPanel.SetScrollable(true).SetChangedFunc(func() {
		ui.App.Draw()
	})

	ui.Layout = &typesUI.ComponentLayout{
		MainMenu:    ui.InitMainMenu(),
		SubMenuList: ui.initSubMenu(),
		LogList:     logPanel,
		OutputPanel: ui.InitOutputPanel(),
	}

	logPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTAB {
			app.SetFocus(ui.Layout.MainMenu.MenuList)
			return nil
		}
		return event
	})

	window := wm.NewWindow().
		Show().
		SetRoot(ui.setupAppLayout()).
		SetBorder(false)

	window.Maximize()

	ui.startupSequence()
	return &ui
}

// setupTitle sets up the header title of the application.
// containing the application name and version.
func setupAppTitle() *tview.TextView {
	title := tview.NewTextView()
	title.SetBorder(true)
	title.SetText(fmt.Sprintf("%s v%s", entity.APP_NAME, entity.APP_VERSION))
	title.SetTextAlign(tview.AlignCenter)

	return title
}

// setupAppLayout sets up the main grid layout of the application.
func (u *UI) setupAppLayout() *tview.Flex {

	// Setup the main layout
	splitSidebar := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(u.Layout.MainMenu.MenuList, 15, 1, true).
		AddItem(u.Layout.SubMenuList, 0, 1, false)

	splitMainPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(u.Layout.OutputPanel.Layout, 0, 3, false).
		AddItem(u.Layout.LogList, 0, 1, false)

	childLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(splitSidebar, 35, 1, true).
		AddItem(splitMainPanel, 0, 4, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(setupAppTitle(), 3, 1, false).
		AddItem(childLayout, 0, 1, true)

	return layout
}
