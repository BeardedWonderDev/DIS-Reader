package bridge

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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
	agentConfig           *proto.AgentConfig

	heartbeatGrace time.Duration

	mu    sync.RWMutex
	conns map[RegistryKey]*streamAgentConnection
}

type ServerOption func(*Server)

func WithAutoConnectOnRegister(enabled bool) ServerOption {
	return func(s *Server) {
		s.autoConnectOnRegister = enabled
	}
}

// WithAgentConfig sends the provided config to agents immediately after they register.
func WithAgentConfig(cfg *proto.AgentConfig) ServerOption {
	return func(s *Server) {
		s.agentConfig = cfg
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
		conns:                 map[RegistryKey]*streamAgentConnection{},
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

	if s.agentConfig != nil {
		s.mu.RLock()
		cfg := s.agentConfig
		s.mu.RUnlock()
		if err := stream.Send(&proto.ServerToAgent{Payload: &proto.ServerToAgent_Config{Config: cfg}}); err != nil {
			return status.Errorf(codes.Internal, "send agent config: %v", err)
		}
		if s.logger != nil {
			s.logger.Info("agent config sent", append(logging.CommonAttrs(tenantID, agentID, "", "bridge_connect", "config"),
				slog.Bool("has_loki", cfg.GetLoki() != nil),
				slog.Bool("has_runtime", cfg.GetRuntime() != nil))...)
		}
	}

	// Track live connection for targeted/broadcast config pushes.
	key := RegistryKey{Tenant: tenantID, Agent: agentID}
	s.mu.Lock()
	s.conns[key] = conn
	s.mu.Unlock()

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

	s.mu.Lock()
	delete(s.conns, key)
	s.mu.Unlock()
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("agent stream closed", append(logging.CommonAttrs(tenantID, agentID, "", "bridge_connect", "stream"), slog.Any("err", err))...)
		}
	}
	return nil
}

// SetAgentConfig updates the server-side AgentConfig and optionally broadcasts
// it to currently connected agents. Caller supplies full config payload.
func (s *Server) SetAgentConfig(cfg *proto.AgentConfig, broadcast bool, tenant string, clientID string) int {
	if cfg == nil {
		return 0
	}
	s.mu.Lock()
	s.agentConfig = cfg
	s.mu.Unlock()

	// If runtime includes a client secret, update authenticator for future connects.
	if rt := cfg.GetRuntime(); rt != nil && strings.TrimSpace(rt.GetClientSecret()) != "" {
		tID := strings.TrimSpace(rt.GetTenantId())
		if clientID == "" {
			clientID = tID
		}
		if ma, ok := s.auth.(MutableAuthenticator); ok && clientID != "" {
			ma.Upsert(clientID, StaticAgentSecret{ClientSecret: rt.GetClientSecret(), TenantID: tID, AgentID: clientID})
		}
	}

	if !broadcast {
		return 0
	}

	s.mu.RLock()
	conns := make(map[RegistryKey]*streamAgentConnection, len(s.conns))
	for k, v := range s.conns {
		conns[k] = v
	}
	s.mu.RUnlock()

	sent := 0
	for key, conn := range conns {
		if tenant != "" && key.Tenant != tenant {
			continue
		}
		if err := conn.SendConfig(cfg); err == nil {
			sent++
			if s.logger != nil {
				s.logger.Info("agent config broadcast", slog.String("tenant", key.Tenant), slog.String("agent", key.Agent))
			}
		} else if s.logger != nil {
			s.logger.Warn("agent config broadcast failed", slog.String("tenant", key.Tenant), slog.String("agent", key.Agent), slog.Any("err", err))
		}
	}
	return sent
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

func (c *streamAgentConnection) SendConfig(cfg *proto.AgentConfig) error {
	if cfg == nil {
		return nil
	}
	return c.stream.Send(&proto.ServerToAgent{Payload: &proto.ServerToAgent_Config{Config: cfg}})
}

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
	msgCh := make(chan *proto.AgentToServer)
	errCh := make(chan error, 1)

	go func() {
		for {
			msg, err := c.stream.Recv()
			if err != nil {
				errCh <- err
				return
			}
			msgCh <- msg
		}
	}()

	hbTimer := time.NewTimer(grace)
	defer hbTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		case <-hbTimer.C:
			if c.logger != nil {
				c.logger.Warn("heartbeat timeout", logging.CommonAttrs(c.tenantID, c.agentID, "", "heartbeat", "timeout")...)
			}
			return errors.New("heartbeat timeout")
		case msg := <-msgCh:
			switch payload := msg.Payload.(type) {
			case *proto.AgentToServer_JobResult:
				c.dispatchResult(payload.JobResult)
			case *proto.AgentToServer_Heartbeat:
				c.lastHB = time.Now()
				if !hbTimer.Stop() {
					<-hbTimer.C
				}
				hbTimer.Reset(grace)
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
