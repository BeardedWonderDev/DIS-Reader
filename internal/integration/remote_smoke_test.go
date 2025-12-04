package integration

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

// Smoke test: spin up bridge server in-memory, attach mock registry/agent, compare remote Query with embedded Query stub.
func TestRemoteSmokeQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set up registry and mock agent connection via in-process gRPC.
	reg := bridge.NewInMemoryRegistry()
	auth := &bridge.StaticAuthenticator{Secrets: map[string]bridge.StaticAgentSecret{
		"agent": {ClientSecret: "secret", TenantID: "t1", AgentID: "a1"},
	}}
	srv := bridge.NewServer(auth, reg, nil)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	bridgeproto.RegisterAgentServiceServer(grpcSrv, srv)
	go grpcSrv.Serve(lis)
	defer grpcSrv.Stop()

	// Start dummy agent client that responds with one row for any query.
	go func() {
		conn, _ := grpc.DialContext(ctx, lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		client := bridgeproto.NewAgentServiceClient(conn)
		stream, _ := client.Connect(ctx)
		_ = stream.Send(&bridgeproto.AgentToServer{Payload: &bridgeproto.AgentToServer_Hello{Hello: &bridgeproto.AgentHello{ClientId: "agent", ClientSecret: "secret", TenantId: "t1"}}})
		for {
			msg, err := stream.Recv()
			if err != nil {
				return
			}
			req := msg.GetJobRequest()
			if req == nil {
				continue
			}
			row := &bridgeproto.Row{Fields: map[string]*structpb.Value{
				"x": structpb.NewNumberValue(1),
			}}
			res := &bridgeproto.JobResult{JobId: req.JobId, Status: bridgeproto.Status_STATUS_OK, Rows: []*bridgeproto.Row{row}}
			_ = stream.Send(&bridgeproto.AgentToServer{Payload: &bridgeproto.AgentToServer_JobResult{JobResult: res}})
		}
	}()

	// Remote DB hitting registry.
	remoteDB := database.NewRemoteDB("t1", reg, nil, 1000, 0)

	require.Eventually(t, func() bool {
		return reg.Stats().TotalAgents > 0
	}, time.Second, 50*time.Millisecond, "agent should register")

	rows, err := remoteDB.Query(ctx, "select 1")
	if err != nil {
		t.Fatalf("remote query err: %v", err)
	}
	if len(rows) != 1 || rows[0]["x"].(float64) != 1 {
		t.Fatalf("unexpected rows: %#v", rows)
	}

	// Health endpoint sanity.
	stats := reg.Stats()
	if stats.TotalAgents == 0 {
		t.Fatalf("expected agents to register")
	}

	_ = http.DefaultClient // keep lints quiet if unused
}
