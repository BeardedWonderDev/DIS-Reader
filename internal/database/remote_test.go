package database

import (
	"context"
	"errors"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakeAgent struct {
	res *bridgeproto.JobResult
	err error
}

func (f *fakeAgent) TenantID() string { return "t1" }
func (f *fakeAgent) AgentID() string  { return "a1" }
func (f *fakeAgent) SendJob(ctx context.Context, req *bridgeproto.JobRequest) (*bridgeproto.JobResult, error) {
	return f.res, f.err
}
func (f *fakeAgent) Close() error { return nil }

type fakeRegistry struct {
	agent bridge.AgentConnection
	err   error
}

func (r *fakeRegistry) Register(ctx context.Context, tenantID string, agentID string, conn bridge.AgentConnection) error {
	return nil
}
func (r *fakeRegistry) Unregister(ctx context.Context, tenantID string, agentID string) {}
func (r *fakeRegistry) Pick(ctx context.Context, tenantID string) (bridge.AgentConnection, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.agent, nil
}

func TestRemoteDBQueryOk(t *testing.T) {
	agent := &fakeAgent{
		res: &bridgeproto.JobResult{
			Status: bridgeproto.Status_STATUS_OK,
			Rows:   []*bridgeproto.Row{{Fields: map[string]*structpb.Value{"x": structpb.NewNumberValue(1)}}},
		},
	}
	reg := &fakeRegistry{agent: agent}
	db := NewRemoteDB("t1", reg, nil)

	rows, err := db.Query(context.Background(), "select 1")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}

func TestRemoteDBErrorStatus(t *testing.T) {
	agent := &fakeAgent{
		res: &bridgeproto.JobResult{
			Status:  bridgeproto.Status_STATUS_ERROR,
			Message: "boom",
		},
	}
	reg := &fakeRegistry{agent: agent}
	db := NewRemoteDB("t1", reg, nil)

	_, err := db.Query(context.Background(), "select 1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoteDBNoAgent(t *testing.T) {
	reg := &fakeRegistry{err: errors.New("no agent")}
	db := NewRemoteDB("t1", reg, nil)
	_, err := db.Query(context.Background(), "select 1")
	if err == nil {
		t.Fatal("expected error")
	}
}
