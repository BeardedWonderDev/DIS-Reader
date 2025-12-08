package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"log/slog"
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
