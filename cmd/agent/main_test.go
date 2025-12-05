package main

import (
	"context"
	"io"
	"sync/atomic"
	"testing"

	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"log/slog"
)

// stubDB implements database.DB with in-memory flags for lifecycle assertions.
type stubDB struct {
	startCalled bool
	stopCalled  bool
}

func (s *stubDB) StartJDBCRunner() error                 { s.startCalled = true; return nil }
func (s *stubDB) StopJDBCRunner() error                  { s.stopCalled = true; return nil }
func (s *stubDB) Connect(ctx context.Context) error      { return nil }
func (s *stubDB) Disconnect(ctx context.Context) error   { return nil }
func (s *stubDB) PingService(ctx context.Context) error  { return nil }
func (s *stubDB) PingDatabase(ctx context.Context) error { return nil }
func (s *stubDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}
func (s *stubDB) Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	return nil, nil
}
func (s *stubDB) QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error) {
	return nil, nil
}
func (s *stubDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}
func (s *stubDB) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	return nil, nil
}

func TestExecuteJob_StartStopJDBC(t *testing.T) {
	db := &stubDB{}
	ctx := context.Background()
	cfg := &AgentConfig{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var loggerVal atomic.Value
	loggerVal.Store(logger)

	startRes := executeJob(ctx, cfg, db, &bridgeproto.JobRequest{JobId: "1", Kind: bridgeproto.JobKind_JOB_KIND_START_JDBC}, &loggerVal)
	if startRes.Status != bridgeproto.Status_STATUS_OK || !db.startCalled {
		t.Fatalf("expected start to succeed and flag to be set, res=%v", startRes)
	}

	stopRes := executeJob(ctx, cfg, db, &bridgeproto.JobRequest{JobId: "2", Kind: bridgeproto.JobKind_JOB_KIND_STOP_JDBC}, &loggerVal)
	if stopRes.Status != bridgeproto.Status_STATUS_OK || !db.stopCalled {
		t.Fatalf("expected stop to succeed and flag to be set, res=%v", stopRes)
	}
}
