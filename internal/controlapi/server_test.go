package controlapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/internal/agentcore"
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
