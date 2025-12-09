package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rivo/tview"
)

func TestRefreshAndHomeBuilder(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status" {
			_ = json.NewEncoder(w).Encode(statusResp{
				BridgeConnected: true,
				Version:         "v1",
				AgentID:         "agent",
				ServerURL:       "srv",
				LastHeartbeat:   "now",
				LastError:       "",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	ui := &uiState{api: ts.URL}
	app := tview.NewApplication()
	statusView := tview.NewTextView()

	ui.refresh(statusView, app)
	text := statusView.GetText(false)
	if !strings.Contains(text, "agent") || !strings.Contains(text, "srv") {
		t.Fatalf("status view not populated: %q", text)
	}

	pages := tview.NewPages()
	home := ui.home(app, pages)
	if home == nil {
		t.Fatalf("home builder returned nil")
	}
}

func TestConfigFormBuilder(t *testing.T) {
	cfg := effectiveConfig{}
	cfg.Config.ServerURL = "https://bridge"
	cfg.Config.ClientID = "cid"
	cfg.Config.ClientSecret = "secret"
	cfg.Config.DIS.JDBCConfig.JavaPath = "/java"
	cfg.Config.DIS.JDBCConfig.JDBCPort = "7777"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/config":
			_ = json.NewEncoder(w).Encode(cfg)
		case "/config/apply":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	ui := &uiState{api: ts.URL}
	app := tview.NewApplication()
	pages := tview.NewPages()
	form := ui.configForm(app, pages)
	if form == nil {
		t.Fatalf("config form nil")
	}
	if form.GetButtonCount() == 0 {
		t.Fatalf("expected buttons on form")
	}
}
