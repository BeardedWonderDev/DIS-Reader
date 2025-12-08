package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUIStateGetAndPostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"ok": true})
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	s := &uiState{api: server.URL, token: "tok"}
	var out map[string]any
	if err := s.getJSON(server.URL, &out); err != nil {
		t.Fatalf("getJSON err: %v", err)
	}
	if _, ok := out["ok"]; !ok {
		t.Fatalf("expected ok key")
	}
	if err := s.postJSON(server.URL, "{}"); err != nil {
		t.Fatalf("postJSON err: %v", err)
	}

	// unauthorized
	s.token = ""
	if err := s.getJSON(server.URL, &out); err == nil {
		t.Fatalf("expected auth error when token missing")
	}
}
