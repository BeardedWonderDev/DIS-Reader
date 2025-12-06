package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/BeardedWonderDev/DIS-Reader/internal/servicectl"
	"github.com/rivo/tview"
)

func main() {
	addrFlag := flag.String("addr", "http://127.0.0.1:7777", "control API address")
	tokenFlag := flag.String("token", "", "bearer token for control API")
	flag.Parse()

	app := tview.NewApplication()
	state := &uiState{api: *addrFlag, token: *tokenFlag}

	pages := tview.NewPages()
	pages.AddPage("home", state.home(app, pages), true, true)
	pages.AddPage("config", state.configForm(app, pages), true, false)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type uiState struct {
	api    string
	token  string
	status *statusResp
}

type statusResp struct {
	Running         bool   `json:"running"`
	BridgeConnected bool   `json:"bridgeConnected"`
	Version         string `json:"version"`
	AgentID         string `json:"agentId"`
	ServerURL       string `json:"serverURL"`
	LastHeartbeat   string `json:"lastHeartbeat"`
	LastError       string `json:"lastError"`
}

type effectiveConfig struct {
	Config struct {
		ServerURL          string `json:"serverURL"`
		ClientID           string `json:"clientID"`
		ClientSecret       string `json:"clientSecret"`
		AutoConnectOnStart bool   `json:"autoConnectOnStart"`
		DIS                struct {
			Host       string `json:"host"`
			User       string `json:"user"`
			Password   string `json:"password"`
			JDBCConfig struct {
				JavaPath string `json:"javaPath"`
				JDBCPort string `json:"jdbcPort"`
			} `json:"jdbcConfig"`
		} `json:"dis"`
		TLS struct {
			Enabled            bool `json:"enabled"`
			InsecureSkipVerify bool `json:"insecureSkipVerify"`
		} `json:"tls"`
	} `json:"config"`
}

// ----- UI builders -----

func (s *uiState) home(app *tview.Application, pages *tview.Pages) *tview.Flex {
	title := tview.NewTextView().SetTextAlign(tview.AlignLeft)
	title.SetText("DIS Agent Control (TUI)")

	statusView := tview.NewTextView().SetDynamicColors(true)

	buttons := tview.NewFlex().SetDirection(tview.FlexRow)
	buttons.AddItem(button("Refresh", func() { s.refresh(statusView, app) }), 1, 0, false)
	buttons.AddItem(button("Install Service", func() { s.service("install", statusView, app) }), 1, 0, false)
	buttons.AddItem(button("Start Service", func() { s.service("start", statusView, app) }), 1, 0, false)
	buttons.AddItem(button("Stop Service", func() { s.service("stop", statusView, app) }), 1, 0, false)
	buttons.AddItem(button("Restart Service", func() { s.service("restart", statusView, app) }), 1, 0, false)
	buttons.AddItem(button("Config", func() { pages.SwitchToPage("config") }), 1, 0, false)
	buttons.AddItem(button("Quit", func() { app.Stop() }), 1, 0, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 0, false).
		AddItem(buttons, 0, 1, true).
		AddItem(tview.NewTextView().SetText(""), 1, 0, false).
		AddItem(statusView, 0, 6, false)

	s.refresh(statusView, app)
	return layout
}

func (s *uiState) configForm(app *tview.Application, pages *tview.Pages) *tview.Form {
	cfg, _ := s.fetchConfig()
	form := tview.NewForm()
	form.AddInputField("Server URL", cfg.Config.ServerURL, 50, nil, func(text string) { cfg.Config.ServerURL = text })
	form.AddInputField("Client ID", cfg.Config.ClientID, 40, nil, func(text string) { cfg.Config.ClientID = text })
	form.AddPasswordField("Client Secret", cfg.Config.ClientSecret, 40, '*', func(text string) { cfg.Config.ClientSecret = text })
	form.AddInputField("DIS Host", cfg.Config.DIS.Host, 40, nil, func(text string) { cfg.Config.DIS.Host = text })
	form.AddInputField("DIS User", cfg.Config.DIS.User, 40, nil, func(text string) { cfg.Config.DIS.User = text })
	form.AddPasswordField("DIS Password", cfg.Config.DIS.Password, 40, '*', func(text string) { cfg.Config.DIS.Password = text })
	form.AddInputField("JDBC Port", cfg.Config.DIS.JDBCConfig.JDBCPort, 10, nil, func(text string) { cfg.Config.DIS.JDBCConfig.JDBCPort = text })
	form.AddInputField("Java Path", cfg.Config.DIS.JDBCConfig.JavaPath, 40, nil, func(text string) { cfg.Config.DIS.JDBCConfig.JavaPath = text })
	form.AddCheckbox("TLS Enabled", cfg.Config.TLS.Enabled, func(checked bool) { cfg.Config.TLS.Enabled = checked })
	form.AddCheckbox("TLS Insecure Skip Verify", cfg.Config.TLS.InsecureSkipVerify, func(checked bool) { cfg.Config.TLS.InsecureSkipVerify = checked })
	form.AddCheckbox("Auto Connect On Start", cfg.Config.AutoConnectOnStart, func(checked bool) { cfg.Config.AutoConnectOnStart = checked })

	form.AddButton("Save & Apply", func() {
		if err := s.applyConfig(cfg); err != nil {
			modal(app, pages, fmt.Sprintf("apply failed: %v", err))
			return
		}
		modal(app, pages, "config applied")
		pages.SwitchToPage("home")
	})
	form.AddButton("Cancel", func() { pages.SwitchToPage("home") })
	form.SetBorder(true).SetTitle("Configure Agent").SetTitleAlign(tview.AlignLeft)
	return form
}

// ----- Helpers -----

func button(label string, action func()) *tview.Button {
	return tview.NewButton(label).SetSelectedFunc(action)
}

func (s *uiState) refresh(tv *tview.TextView, app *tview.Application) {
	status, err := s.getStatus()
	if err != nil {
		tv.SetText(fmt.Sprintf("[red]error:[-] %v", err))
		return
	}
	s.status = status
	tv.SetText(fmt.Sprintf("Bridge: %v\nVersion: %s\nAgent: %s\nServer: %s\nLast HB: %s\nLast Error: %s",
		status.BridgeConnected, status.Version, status.AgentID, status.ServerURL, status.LastHeartbeat, status.LastError))
}

func (s *uiState) service(action string, tv *tview.TextView, app *tview.Application) {
	ctrl := servicectl.New()
	var err error
	ctx := context.Background()
	switch action {
	case "install":
		err = ctrl.Install(ctx)
	case "start":
		err = ctrl.Start(ctx)
	case "stop":
		err = ctrl.Stop(ctx)
	case "restart":
		err = ctrl.Restart(ctx)
	case "status":
		_, err = ctrl.Status(ctx)
	}
	if err != nil {
		// Fallback to control API if available
		url := fmt.Sprintf("%s/service/%s", strings.TrimRight(s.api, "/"), action)
		if postErr := s.postJSON(url, "{}"); postErr != nil {
			tv.SetText(fmt.Sprintf("[red]%s failed:[-] %v (local) ; %v (api)", action, err, postErr))
			return
		}
		tv.SetText(fmt.Sprintf("%s via API (local failed: %v)", action, err))
	} else {
		tv.SetText(fmt.Sprintf("%s ok (local)", action))
	}
	s.refresh(tv, app)
}

func (s *uiState) fetchConfig() (*effectiveConfig, error) {
	url := strings.TrimRight(s.api, "/") + "/config"
	var cfg effectiveConfig
	if err := s.getJSON(url, &cfg); err != nil {
		return &cfg, err
	}
	return &cfg, nil
}

func (s *uiState) applyConfig(cfg *effectiveConfig) error {
	url := strings.TrimRight(s.api, "/") + "/config/apply"
	b, _ := json.Marshal(cfg.Config)
	return s.postJSON(url, string(b))
}

func (s *uiState) getStatus() (*statusResp, error) {
	url := strings.TrimRight(s.api, "/") + "/status"
	var sr statusResp
	if err := s.getJSON(url, &sr); err != nil {
		return nil, err
	}
	return &sr, nil
}

func (s *uiState) getJSON(url string, dest interface{}) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	s.auth(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

func (s *uiState) postJSON(url, body string) error {
	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	s.auth(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", strings.TrimSpace(string(b)))
	}
	return nil
}

func (s *uiState) auth(req *http.Request) {
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
}

func modal(app *tview.Application, pages *tview.Pages, msg string) {
	modal := tview.NewModal().SetText(msg).AddButtons([]string{"OK"}).SetDoneFunc(func(i int, _ string) {
		pages.SwitchToPage("home")
	})
	pages.AddAndSwitchToPage("modal", modal, true)
	app.Draw()
}
