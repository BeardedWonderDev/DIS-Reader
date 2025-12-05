package bridge

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/internal/logging"
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

	autoConnectOnRegister bool

	heartbeatGrace time.Duration
}

type ServerOption func(*Server)

func WithAutoConnectOnRegister(enabled bool) ServerOption {
	return func(s *Server) {
		s.autoConnectOnRegister = enabled
	}
}

func NewServer(auth AgentAuthenticator, registry AgentRegistry, logger *slog.Logger, opts ...ServerOption) *Server {
	if registry == nil {
		registry = NewInMemoryRegistry()
	}
	s := &Server{
		auth:                  auth,
		registry:              registry,
		logger:                logger,
		heartbeatGrace:        defaultHeartbeatGrace,
		autoConnectOnRegister: true,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
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

	errCh := make(chan error, 1)
	go func() {
		errCh <- conn.run(ctx, s.heartbeatGrace)
	}()

	connectStart := time.Now()
	if s.autoConnectOnRegister {
		ctxConnect, cancel := context.WithTimeout(ctx, 10*time.Second)
		if _, err := conn.SendJob(ctxConnect, &proto.JobRequest{
			JobId: uuid.New().String(),
			Kind:  proto.JobKind_JOB_KIND_CONNECT,
		}); err != nil && s.logger != nil {
			s.logger.Warn("agent auto-connect failed",
				append(logging.CommonAttrs(tenantID, agentID, "", "dis_connect", "connect"), slog.Any("err", err))...)
		}
		cancel()
	}

	if s.logger != nil {
		attrs := logging.CommonAttrs(tenantID, agentID, "", "bridge_connect", "connect")
		attrs = append(attrs,
			slog.Any("labels", hello.Labels),
			slog.String("version", hello.Version),
			logging.DurationAttr(time.Since(connectStart)),
		)
		s.logger.Info("agent connected", attrs...)
	}

	err = <-errCh
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("agent stream closed", append(logging.CommonAttrs(tenantID, agentID, "", "bridge_connect", "stream"), slog.Any("err", err))...)
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
				level := logging.MapLogLevel(payload.Log.Level)
				attrs := logging.CommonAttrs(c.tenantID, c.agentID, "", "agent_log", "log")
				for k, v := range payload.Log.Fields {
					attrs = append(attrs, slog.String(k, v))
				}
				attrs = append(attrs, slog.String("message", payload.Log.Message))
			c.logger.Log(ctx, level, "agent log", attrs...)
		}
		default:
			if c.logger != nil {
				c.logger.Debug("unhandled agent message", slog.Any("payload", fmt.Sprintf("%T", payload)))
			}
		}

		if time.Since(c.lastHB) > grace {
			if c.logger != nil {
				c.logger.Warn("heartbeat timeout", logging.CommonAttrs(c.tenantID, c.agentID, "", "heartbeat", "timeout")...)
			}
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
