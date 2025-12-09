package controlapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/internal/agentcore"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type fakeServiceCtl struct{ calls []string }

func (f *fakeServiceCtl) Install(ctx context.Context) error {
	f.calls = append(f.calls, "install")
	return nil
}
func (f *fakeServiceCtl) Start(ctx context.Context) error {
	f.calls = append(f.calls, "start")
	return nil
}
func (f *fakeServiceCtl) Stop(ctx context.Context) error {
	f.calls = append(f.calls, "stop")
	return nil
}
func (f *fakeServiceCtl) Restart(ctx context.Context) error {
	f.calls = append(f.calls, "restart")
	return nil
}
func (f *fakeServiceCtl) Status(ctx context.Context) (string, error) {
	f.calls = append(f.calls, "status")
	return "active", nil
}

func TestControlAPIEndpoints(t *testing.T) {
	svc := &fakeServiceCtl{}
	opts := Options{
		Token:    "tok",
		StatusFn: func() Status { return Status{Running: true} },
		EffectiveFn: func() *agentcore.EffectiveConfig {
			return &agentcore.EffectiveConfig{Config: &agentcore.Config{ClientSecret: "secret"}}
		},
		ApplyFn:    func(cfg *agentcore.Config) error { return nil },
		ValidateFn: func(cfg *agentcore.Config) error { return nil },
		ServiceCtl: svc,
	}
	mux := http.NewServeMux()
	buildMux(mux, opts)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := srv.Client()

	req, _ := http.NewRequest("GET", srv.URL+"/status", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("status request failed: %v status=%d", err, resp.StatusCode)
	}

	req, _ = http.NewRequest("GET", srv.URL+"/config", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("config request failed: %v status=%d", err, resp.StatusCode)
	}

	req, _ = http.NewRequest("POST", srv.URL+"/config/apply", strings.NewReader(`{"serverURL":"x"}`))
	req.Header.Set("Authorization", "Bearer tok")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("apply request failed: %v status=%d", err, resp.StatusCode)
	}

	req, _ = http.NewRequest("POST", srv.URL+"/config/validate", strings.NewReader(`{"serverURL":"x"}`))
	req.Header.Set("Authorization", "Bearer tok")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusNoContent {
		t.Fatalf("validate request failed: %v status=%d", err, resp.StatusCode)
	}

	req, _ = http.NewRequest("POST", srv.URL+"/service/start", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("service start failed: %v status=%d", err, resp.StatusCode)
	}
	if len(svc.calls) == 0 || svc.calls[len(svc.calls)-1] != "start" {
		t.Fatalf("expected start call, got %v", svc.calls)
	}

	// unauthorized path
	resp, err = client.Get(srv.URL + "/status")
	if err != nil {
		t.Fatalf("unauth status err: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", resp.StatusCode)
	}
}

func TestControlAPIErrorPathsAndHelpers(t *testing.T) {
	// authorize should reject when token mismatches
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/status", nil)
	if authorize(w, r, "tok") {
		t.Fatalf("expected unauthorized without bearer header")
	}

	// Start should error when address missing
	if err := Start(context.Background(), Options{}); err == nil {
		t.Fatalf("expected error for missing addr")
	}

	// sanitizeConfig should mask secret fields
	cfg := &agentcore.Config{
		TenantID:     "tenant",
		ClientSecret: "secret",
		DIS:          types.DISConfig{Password: "pass"},
		Control:      agentcore.ControlConfig{Token: "tok"},
	}
	safe := sanitizeConfig(cfg)
	if safe.TenantID != "" || safe.ClientSecret != "***" || safe.DIS.Password != "***" || safe.Control.Token != "***" {
		t.Fatalf("expected masked config, got %+v", safe)
	}

	// mask helper
	if mask("") != "" || mask("x") != "***" {
		t.Fatalf("mask helper not behaving")
	}

	// buildMux negative paths: method not allowed + validation failure + apply failure + service error and unknown
	serviceErr := errors.New("svc fail")
	svc := &fakeServiceCtl{}
	opts := Options{
		Token: "tok",
		StatusFn: func() Status {
			return Status{Running: true}
		},
		EffectiveFn: func() *agentcore.EffectiveConfig {
			return &agentcore.EffectiveConfig{Config: &agentcore.Config{}}
		},
		ValidateFn: func(cfg *agentcore.Config) error { return errors.New("bad") },
		ApplyFn:    func(cfg *agentcore.Config) error { return errors.New("apply bad") },
		ServiceCtl: svc,
	}
	mux := http.NewServeMux()
	buildMux(mux, opts)

	// wrong method on /config
	req := httptest.NewRequest(http.MethodPost, "/config", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp := httptest.NewRecorder()
	mux.ServeHTTP(resp, req)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for config POST, got %d", resp.Code)
	}

	// validation failure
	req = httptest.NewRequest(http.MethodPost, "/config/validate", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer tok")
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for validation failure, got %d", resp.Code)
	}

	// apply failure
	req = httptest.NewRequest(http.MethodPost, "/config/apply", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer tok")
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for apply failure, got %d", resp.Code)
	}

	// service error
	badCtl := &fakeServiceCtl{}
	badOpts := opts
	badOpts.ServiceCtl = badCtl
	// override Start to fail
	badOpts.ServiceCtl = &fakeServiceCtlWithError{err: serviceErr}
	mux2 := http.NewServeMux()
	buildMux(mux2, badOpts)
	req = httptest.NewRequest(http.MethodPost, "/service/start", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp = httptest.NewRecorder()
	mux2.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for service failure, got %d", resp.Code)
	}

	// unknown action
	req = httptest.NewRequest(http.MethodPost, "/service/unknown", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown action, got %d", resp.Code)
	}
}

func TestControlAPIStartAndUnavailableHandlers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := Options{
		Addr:     "127.0.0.1:0",
		Token:    "",
		StatusFn: func() Status { return Status{} },
		EffectiveFn: func() *agentcore.EffectiveConfig {
			return nil
		},
	}
	if err := Start(ctx, opts); err != nil {
		t.Fatalf("expected start to succeed with addr: %v", err)
	}

	mux := http.NewServeMux()
	buildMux(mux, opts)

	// config unavailable when EffectiveFn returns nil
	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	resp := httptest.NewRecorder()
	mux.ServeHTTP(resp, req)
	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when config unavailable, got %d", resp.Code)
	}

	// service control unavailable returns 501
	req = httptest.NewRequest(http.MethodPost, "/service/start", nil)
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)
	if resp.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 when service control missing, got %d", resp.Code)
	}
}

type fakeServiceCtlWithError struct{ err error }

func (f *fakeServiceCtlWithError) Install(ctx context.Context) error { return f.err }
func (f *fakeServiceCtlWithError) Start(ctx context.Context) error   { return f.err }
func (f *fakeServiceCtlWithError) Stop(ctx context.Context) error    { return f.err }
func (f *fakeServiceCtlWithError) Restart(ctx context.Context) error { return f.err }
func (f *fakeServiceCtlWithError) Status(ctx context.Context) (string, error) {
	return "", f.err
}
