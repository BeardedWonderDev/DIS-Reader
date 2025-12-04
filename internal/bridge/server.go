package bridge

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultHeartbeatGrace = 90 * time.Second
)

// Server implements the AgentService gRPC handlers.
type Server struct {
	proto.UnimplementedAgentServiceServer

	auth     AgentAuthenticator
	registry AgentRegistry
	logger   *slog.Logger

	heartbeatGrace time.Duration
}

func NewServer(auth AgentAuthenticator, registry AgentRegistry, logger *slog.Logger) *Server {
	if registry == nil {
		registry = NewInMemoryRegistry()
	}
	return &Server{
		auth:           auth,
		registry:       registry,
		logger:         logger,
		heartbeatGrace: defaultHeartbeatGrace,
	}
}

// Connect is the bi-directional stream entrypoint for agents.
func (s *Server) Connect(stream proto.AgentService_ConnectServer) error {
	ctx := stream.Context()

	first, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "missing hello: %v", err)
	}
	hello := first.GetHello()
	if hello == nil {
		return status.Error(codes.Unauthenticated, "first message must be hello")
	}

	tenantID, agentID, err := s.auth.Authenticate(ctx, hello.ClientId, hello.ClientSecret, hello.TenantId)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "auth failed: %v", err)
	}
	if agentID == "" {
		agentID = hello.ClientId
	}

	conn := newStreamAgentConnection(stream, tenantID, agentID, s.logger)
	if err := s.registry.Register(ctx, tenantID, agentID, conn); err != nil {
		return status.Errorf(codes.Internal, "registry error: %v", err)
	}
	defer s.registry.Unregister(context.Background(), tenantID, agentID)

	if s.logger != nil {
		s.logger.Info("agent connected", slog.String("tenant_id", tenantID), slog.String("agent_id", agentID), slog.Any("labels", hello.Labels))
	}

	err = conn.run(ctx, s.heartbeatGrace)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("agent stream closed", slog.String("tenant_id", tenantID), slog.String("agent_id", agentID), slog.Any("err", err))
		}
	}
	return nil
}

// streamAgentConnection wraps the gRPC stream and fulfills AgentConnection.
type streamAgentConnection struct {
	tenantID string
	agentID  string
	stream   proto.AgentService_ConnectServer
	logger   *slog.Logger

	mu      sync.Mutex
	pending map[string]chan *proto.JobResult
	lastHB  time.Time
	closed  bool
}

func newStreamAgentConnection(stream proto.AgentService_ConnectServer, tenantID, agentID string, logger *slog.Logger) *streamAgentConnection {
	return &streamAgentConnection{
		tenantID: tenantID,
		agentID:  agentID,
		stream:   stream,
		logger:   logger,
		pending:  map[string]chan *proto.JobResult{},
		lastHB:   time.Now(),
	}
}

func (c *streamAgentConnection) TenantID() string { return c.tenantID }
func (c *streamAgentConnection) AgentID() string  { return c.agentID }

func (c *streamAgentConnection) SendJob(ctx context.Context, req *proto.JobRequest) (*proto.JobResult, error) {
	if req.JobId == "" {
		req.JobId = uuid.New().String()
	}

	resultCh := make(chan *proto.JobResult, 1)
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, errors.New("connection closed")
	}
	c.pending[req.JobId] = resultCh
	c.mu.Unlock()

	sendErr := c.stream.Send(&proto.ServerToAgent{
		Payload: &proto.ServerToAgent_JobRequest{JobRequest: req},
	})
	if sendErr != nil {
		c.cleanup(req.JobId)
		return nil, sendErr
	}

	select {
	case res := <-resultCh:
		return res, nil
	case <-ctx.Done():
		c.cleanup(req.JobId)
		return nil, ctx.Err()
	}
}

func (c *streamAgentConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}
	return nil
}

func (c *streamAgentConnection) run(ctx context.Context, grace time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := c.stream.Recv()
		if err != nil {
			return err
		}

		switch payload := msg.Payload.(type) {
		case *proto.AgentToServer_JobResult:
			c.dispatchResult(payload.JobResult)
		case *proto.AgentToServer_Heartbeat:
			c.lastHB = time.Now()
		case *proto.AgentToServer_Log:
			if c.logger != nil {
				c.logger.Info("agent log",
					slog.String("agent_id", c.agentID),
					slog.String("tenant_id", c.tenantID),
					slog.String("level", payload.Log.Level),
					slog.String("message", payload.Log.Message))
			}
		default:
			if c.logger != nil {
				c.logger.Debug("unhandled agent message", slog.Any("payload", fmt.Sprintf("%T", payload)))
			}
		}

		if time.Since(c.lastHB) > grace {
			return errors.New("heartbeat timeout")
		}
	}
}

func (c *streamAgentConnection) dispatchResult(res *proto.JobResult) {
	if res == nil {
		return
	}
	c.mu.Lock()
	ch, ok := c.pending[res.JobId]
	if ok {
		delete(c.pending, res.JobId)
	}
	c.mu.Unlock()
	if ok {
		ch <- res
		close(ch)
	}
}

func (c *streamAgentConnection) cleanup(jobID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ch, ok := c.pending[jobID]; ok {
		close(ch)
		delete(c.pending, jobID)
	}
}
