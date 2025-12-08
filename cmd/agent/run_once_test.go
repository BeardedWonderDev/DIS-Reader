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
	"google.golang.org/grpc"
	"log/slog"

	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
)

// fakeDB counts calls for assertions.
type fakeDB struct {
	database.DB
	pingServiceCalled int32
}

func (f *fakeDB) PingService(ctx context.Context) error {
	atomic.AddInt32(&f.pingServiceCalled, 1)
	return nil
}

type recordingAgentService struct {
	bridgeproto.UnimplementedAgentServiceServer
	gotHello     bool
	gotJobResult bool
}

func (s *recordingAgentService) Connect(stream bridgeproto.AgentService_ConnectServer) error {
	// Expect hello first
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	if _, ok := msg.Payload.(*bridgeproto.AgentToServer_Hello); !ok {
		return errors.New("expected hello payload")
	}
	s.gotHello = true

	// Send a ping job
	if err := stream.Send(&bridgeproto.ServerToAgent{
		Payload: &bridgeproto.ServerToAgent_JobRequest{
			JobRequest: &bridgeproto.JobRequest{
				JobId: "job-1",
				Kind:  bridgeproto.JobKind_JOB_KIND_PING_SERVICE,
			},
		},
	}); err != nil {
		return err
	}

	// Receive job result from agent
	if _, err := stream.Recv(); err != nil {
		return err
	}
	s.gotJobResult = true

	return errors.New("terminate") // trigger stream error to exit runOnce
}

func TestRunOnceProcessesJobAndExits(t *testing.T) {
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
		ServerURL: lis.Addr().String(),
		ClientID:  "agent-1",
		TenantID:  "tenant-1",
		TLS: struct {
			Enabled            bool `mapstructure:"enabled" yaml:"enabled"`
			InsecureSkipVerify bool `mapstructure:"insecureSkipVerify" yaml:"insecureSkipVerify"`
		}{Enabled: false},
		DIS: types.DISConfig{LogLevel: 0},
	}
	db := &fakeDB{}
	baseHandler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})
	var loggerVal atomic.Value
	loggerVal.Store(slog.New(baseHandler))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = runOnce(ctx, cfg, db, baseHandler, &loggerVal, nil)
	if err == nil {
		t.Fatalf("expected runOnce to return error after stream closed")
	}
	if !rec.gotHello || !rec.gotJobResult {
		t.Fatalf("expected hello and job result exchange, gotHello=%v jobResult=%v", rec.gotHello, rec.gotJobResult)
	}
	if atomic.LoadInt32(&db.pingServiceCalled) != 1 {
		t.Fatalf("expected PingService to be called once, got %d", db.pingServiceCalled)
	}
}
