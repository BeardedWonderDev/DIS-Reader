package controlapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/agentcore"
	"github.com/BeardedWonderDev/DIS-Reader/internal/servicectl"
)

// Status represents high-level agent status for UIs.
type Status struct {
	Running         bool       `json:"running"`
	BridgeConnected bool       `json:"bridgeConnected"`
	LastHeartbeat   *time.Time `json:"lastHeartbeat,omitempty"`
	LastError       string     `json:"lastError,omitempty"`
	Version         string     `json:"version,omitempty"`
	AgentID         string     `json:"agentId,omitempty"`
	ServerURL       string     `json:"serverURL,omitempty"`
	ConfigPath      string     `json:"configPath,omitempty"`
	PendingRestart  bool       `json:"pendingRestart,omitempty"`
}

// Options control the control API server startup.
type Options struct {
	Addr        string
	Token       string
	StatusFn    func() Status
	EffectiveFn func() *agentcore.EffectiveConfig
	ApplyFn     func(*agentcore.Config) error
	ValidateFn  func(*agentcore.Config) error
	ServiceCtl  servicectl.Controller
}

// Start launches the control API server in a goroutine.
func Start(ctx context.Context, opts Options) error {
	if opts.Addr == "" {
		return errors.New("control api addr required")
	}
	mux := http.NewServeMux()

	buildMux(mux, opts)

	srv := &http.Server{Addr: opts.Addr, Handler: mux}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// server exits silently on error; caller can log
		}
	}()

	return nil
}

// buildMux wires handlers onto the provided mux; factored for testing.
func buildMux(mux *http.ServeMux, opts Options) {

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(w, r, opts.Token) {
			return
		}
		writeJSON(w, opts.StatusFn())
	})

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(w, r, opts.Token) {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		eff := opts.EffectiveFn()
		if eff == nil || eff.Config == nil {
			http.Error(w, "config unavailable", http.StatusInternalServerError)
			return
		}
		safe := sanitizeConfig(eff.Config)
		writeJSON(w, &agentcore.EffectiveConfig{Config: safe, Provenance: eff.Provenance})
	})

	mux.HandleFunc("/config/validate", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(w, r, opts.Token) {
			return
		}
		var cfg agentcore.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, fmt.Sprintf("invalid config: %v", err), http.StatusBadRequest)
			return
		}
		if opts.ValidateFn != nil {
			if err := opts.ValidateFn(&cfg); err != nil {
				http.Error(w, fmt.Sprintf("validation failed: %v", err), http.StatusBadRequest)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/config/apply", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(w, r, opts.Token) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var cfg agentcore.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, fmt.Sprintf("invalid config: %v", err), http.StatusBadRequest)
			return
		}
		if opts.ApplyFn != nil {
			if err := opts.ApplyFn(&cfg); err != nil {
				http.Error(w, fmt.Sprintf("apply failed: %v", err), http.StatusBadRequest)
				return
			}
		}
		writeJSON(w, map[string]string{"status": "applied"})
	})

	// Service controls
	mux.HandleFunc("/service/", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(w, r, opts.Token) {
			return
		}
		if opts.ServiceCtl == nil {
			http.Error(w, "service control unavailable", http.StatusNotImplemented)
			return
		}
		action := strings.TrimPrefix(r.URL.Path, "/service/")
		var err error
		switch action {
		case "install":
			err = opts.ServiceCtl.Install(r.Context())
		case "start":
			err = opts.ServiceCtl.Start(r.Context())
		case "stop":
			err = opts.ServiceCtl.Stop(r.Context())
		case "restart":
			err = opts.ServiceCtl.Restart(r.Context())
		case "status":
			var out string
			out, err = opts.ServiceCtl.Status(r.Context())
			if err == nil {
				writeJSON(w, map[string]string{"status": strings.TrimSpace(out)})
				return
			}
		default:
			http.Error(w, "unknown action", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("service %s failed: %v", action, err), http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]string{"status": action})
	})
}

func authorize(w http.ResponseWriter, r *http.Request, token string) bool {
	if token == "" {
		return true
	}
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") && strings.TrimPrefix(auth, "Bearer ") == token {
		return true
	}
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return false
}

func sanitizeConfig(cfg *agentcore.Config) *agentcore.Config {
	if cfg == nil {
		return nil
	}
	out := *cfg
	// Hide sensitive/non-surfaced fields
	out.TenantID = ""
	out.ClientSecret = mask(out.ClientSecret)
	if out.DIS.Password != "" {
		out.DIS.Password = "***"
	}
	if out.Control.Token != "" {
		out.Control.Token = "***"
	}
	// Observability runtime not exposed; AppliedLoki omitted by yaml tag already.
	return &out
}

func mask(s string) string {
	if s == "" {
		return ""
	}
	return "***"
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
