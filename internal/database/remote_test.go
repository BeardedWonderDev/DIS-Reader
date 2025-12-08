package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakeAgent struct {
	res     *bridgeproto.JobResult
	err     error
	lastReq *bridgeproto.JobRequest
}

func (f *fakeAgent) TenantID() string { return "t1" }
func (f *fakeAgent) AgentID() string  { return "a1" }
func (f *fakeAgent) SendJob(ctx context.Context, req *bridgeproto.JobRequest) (*bridgeproto.JobResult, error) {
	f.lastReq = req
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
func (r *fakeRegistry) Stats() bridge.RegistryStats {
	if r.err != nil {
		return bridge.RegistryStats{}
	}
	return bridge.RegistryStats{
		TotalAgents: 1,
		Tenants:     map[string]int{"t1": 1},
	}
}
func (r *fakeRegistry) ObserveQuery(tenantID, agentID string, latency time.Duration) {}

func TestRemoteDBQueryOk(t *testing.T) {
	agent := &fakeAgent{
		res: &bridgeproto.JobResult{
			Status: bridgeproto.Status_STATUS_OK,
			Rows:   []*bridgeproto.Row{{Fields: map[string]*structpb.Value{"x": structpb.NewNumberValue(1)}}},
		},
	}
	reg := &fakeRegistry{agent: agent}
	db := NewRemoteDB(reg, nil, 1000, 0)

	rows, err := db.Query(context.Background(), "select 1", "t1")
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
	db := NewRemoteDB(reg, nil, 1000, 0)

	_, err := db.Query(context.Background(), "select 1", "t1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoteDBZeroMaxRowsUnlimited(t *testing.T) {
	agent := &fakeAgent{
		res: &bridgeproto.JobResult{
			Status: bridgeproto.Status_STATUS_OK,
			Rows: []*bridgeproto.Row{
				{Fields: map[string]*structpb.Value{"x": structpb.NewNumberValue(1)}},
				{Fields: map[string]*structpb.Value{"x": structpb.NewNumberValue(2)}},
			},
		},
	}
	reg := &fakeRegistry{agent: agent}
	db := NewRemoteDB(reg, nil, 0, 0)

	rows, err := db.Query(context.Background(), "select 1", "t1")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected all rows when maxRows is 0, got %d", len(rows))
	}
}

func TestRemoteDBNoAgent(t *testing.T) {
	reg := &fakeRegistry{err: errors.New("no agent")}
	db := NewRemoteDB(reg, nil, 1000, 0)
	_, err := db.Query(context.Background(), "select 1", "t1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoteDBStartStopLifecycle(t *testing.T) {
	agent := &fakeAgent{
		res: &bridgeproto.JobResult{Status: bridgeproto.Status_STATUS_OK},
	}
	reg := &fakeRegistry{agent: agent}
	db := NewRemoteDB(reg, nil, 1000, 0)

	if err := db.StartJDBCRunner("t1"); err != nil {
		t.Fatalf("start err: %v", err)
	}
	if agent.lastReq == nil || agent.lastReq.Kind != bridgeproto.JobKind_JOB_KIND_START_JDBC {
		t.Fatalf("expected start jdbc job, got %v", agent.lastReq)
	}

	if err := db.StopJDBCRunner("t1"); err != nil {
		t.Fatalf("stop err: %v", err)
	}
	if agent.lastReq == nil || agent.lastReq.Kind != bridgeproto.JobKind_JOB_KIND_STOP_JDBC {
		t.Fatalf("expected stop jdbc job, got %v", agent.lastReq)
	}
}

func TestRemoteDBEnforceLimits(t *testing.T) {
	db := NewRemoteDB(&fakeRegistry{}, nil, 2, 0)
	rows := []types.ResultRow{
		{"a": 1}, {"a": 2}, {"a": 3},
	}
	limited := db.enforceLimits(rows)
	if len(limited) != 2 {
		t.Fatalf("expected max 2 rows, got %d", len(limited))
	}

	// Byte cap should stop before exceeding maxResultBytes
	byteCapped := NewRemoteDB(&fakeRegistry{}, nil, 0, 20)
	rows = []types.ResultRow{{"a": "short"}, {"a": "this is longer than cap"}}
	limited = byteCapped.enforceLimits(rows)
	if len(limited) != 1 {
		t.Fatalf("expected only first row to fit byte cap, got %d", len(limited))
	}
}

func TestRemoteDBQueryWithSourceIncludesFlagAndLimits(t *testing.T) {
	agent := &fakeAgent{
		res: &bridgeproto.JobResult{
			Status: bridgeproto.Status_STATUS_OK,
			Rows: []*bridgeproto.Row{
				{Fields: map[string]*structpb.Value{"x": structpb.NewNumberValue(1)}},
				{Fields: map[string]*structpb.Value{"x": structpb.NewNumberValue(2)}},
			},
		},
	}
	reg := &fakeRegistry{agent: agent}
	db := NewRemoteDB(reg, nil, 1, 0)

	rows, err := db.QueryWithSource(context.Background(), "select 1", "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.lastReq == nil || !agent.lastReq.IncludeSrc {
		t.Fatalf("expected IncludeSrc flag to be set on job request")
	}
	if len(rows) != 1 {
		t.Fatalf("expected rows to be limited to maxRows, got %d", len(rows))
	}
}
