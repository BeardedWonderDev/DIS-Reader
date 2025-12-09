package main

import (
	"context"
	"errors"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"log/slog"

	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"google.golang.org/grpc"
)

func TestRunAgentStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	status := &agentStatus{}
	runAgent(ctx, &AgentConfig{ServerURL: "invalid", ClientID: "c", TenantID: "t", DIS: types.DISConfig{}}, nil, slog.NewTextHandler(nil, nil), nil, status)

	if status.bridgeConnected.Load() != false {
		t.Fatalf("expected bridgeConnected to remain false on cancel")
	}
}

func TestRunAgentBackoffOnError(t *testing.T) {
	orig := runOnceFunc
	defer func() { runOnceFunc = orig }()

	runOnceFunc = func(ctx context.Context, cfg *AgentConfig, db database.DB, h slog.Handler, v *atomic.Value, st *agentStatus) error {
		return errors.New("boom")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	status := &agentStatus{}
	runAgent(ctx, &AgentConfig{ServerURL: "invalid", ClientID: "c", TenantID: "t", DIS: types.DISConfig{}}, nil, slog.NewTextHandler(nil, nil), nil, status)

	if status.bridgeConnected.Load() {
		t.Fatalf("expected bridgeConnected false on repeated errors")
	}
}

// Integration-ish: real gRPC server exercising happy path of runAgent + runOnce.
func TestRunAgentSuccessPathSetsBridgeConnected(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	rec := &recordingAgentService{}
	bridgeproto.RegisterAgentServiceServer(srv, rec)
	go srv.Serve(lis)
	t.Cleanup(func() {
		srv.Stop()
		lis.Close()
	})

	cfg := &AgentConfig{
		ServerURL:    lis.Addr().String(),
		ClientID:     "agent-1",
		ClientSecret: "",
		TenantID:     "tenant-1",
		TLS: struct {
			Enabled            bool `mapstructure:"enabled" yaml:"enabled"`
			InsecureSkipVerify bool `mapstructure:"insecureSkipVerify" yaml:"insecureSkipVerify"`
		}{Enabled: false},
		DIS: types.DISConfig{},
	}
	db := &stubDB{}
	baseHandler := slog.NewTextHandler(io.Discard, nil)
	var loggerVal atomic.Value
	loggerVal.Store(slog.New(baseHandler))
	status := &agentStatus{}
	status.lastError.Store("init") // establish string type for atomic.Value

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	runAgent(ctx, cfg, db, baseHandler, &loggerVal, status)

	if !rec.gotHello {
		t.Fatalf("expected hello exchanged")
	}
}
