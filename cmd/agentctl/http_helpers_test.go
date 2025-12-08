package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetStatusAndAuthHeaders(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status" {
			if r.Header.Get("Authorization") != "Bearer token123" {
				t.Fatalf("expected auth header propagated")
			}
			json.NewEncoder(w).Encode(statusResp{Version: "v1", AgentID: "agent"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer s.Close()

	ui := &uiState{api: s.URL, token: "token123"}
	st, err := ui.getStatus()
	if err != nil {
		t.Fatalf("getStatus error: %v", err)
	}
	if st.AgentID != "agent" || st.Version != "v1" {
		t.Fatalf("unexpected status resp: %+v", st)
	}
}

func TestPostJSONErrorPropagation(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad"))
	}))
	defer s.Close()

	ui := &uiState{api: s.URL}
	err := ui.postJSON(s.URL+"/apply", "{}")
	if err == nil || err.Error() != "bad" {
		t.Fatalf("expected error body returned, got %v", err)
	}
}

func TestFetchAndApplyConfigHelpers(t *testing.T) {
	cfg := effectiveConfig{}
	cfg.Config.ServerURL = "https://bridge"
	cfg.Config.ClientID = "cid"
	cfg.Config.ClientSecret = "secret"
	cfg.Config.DIS.JDBCConfig.JavaPath = "/java"

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/config":
			json.NewEncoder(w).Encode(cfg)
		case "/config/apply":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer s.Close()

	ui := &uiState{api: s.URL}
	got, err := ui.fetchConfig()
	if err != nil {
		t.Fatalf("fetchConfig err: %v", err)
	}
	if got.Config.ServerURL != "https://bridge" || got.Config.ClientID != "cid" {
		t.Fatalf("unexpected config: %+v", got.Config)
	}
	if err := ui.applyConfig(got); err != nil {
		t.Fatalf("applyConfig err: %v", err)
	}
}
