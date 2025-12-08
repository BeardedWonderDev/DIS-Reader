package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/types"

	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"google.golang.org/grpc/metadata"
)

type fakeConnectStream struct {
	sent int32
	fail bool
}

func (f *fakeConnectStream) Send(m *bridgeproto.AgentToServer) error {
	if f.fail {
		return errors.New("send-fail")
	}
	atomic.AddInt32(&f.sent, 1)
	return nil
}
func (f *fakeConnectStream) Recv() (*bridgeproto.ServerToAgent, error) { return nil, nil }
func (f *fakeConnectStream) Header() (metadata.MD, error)              { return nil, nil }
func (f *fakeConnectStream) Trailer() metadata.MD                      { return nil }
func (f *fakeConnectStream) CloseSend() error                          { return nil }
func (f *fakeConnectStream) Context() context.Context                  { return context.Background() }
func (f *fakeConnectStream) SendMsg(m interface{}) error               { return nil }
func (f *fakeConnectStream) RecvMsg(m interface{}) error               { return nil }

func TestSendHeartbeatsSendsAndUpdatesStatus(t *testing.T) {
	stream := &fakeConnectStream{}
	cfg := &AgentConfig{TenantID: "t", ClientID: "c", DIS: types.DISConfig{}}
	status := &agentStatus{}

	orig := heartbeatIntervalVar
	heartbeatIntervalVar = 10 * time.Millisecond
	defer func() { heartbeatIntervalVar = orig }()

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	sendHeartbeats(ctx, stream, cfg, nil, status)

	if atomic.LoadInt32(&stream.sent) == 0 {
		t.Fatalf("expected at least one heartbeat sent")
	}
	if status.lastHeartbeat.Load() == nil {
		t.Fatalf("expected lastHeartbeat to be set")
	}
}
