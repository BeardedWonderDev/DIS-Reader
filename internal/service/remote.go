package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/internal/invoices"
	"github.com/BeardedWonderDev/DIS-Reader/internal/parts"
	"github.com/BeardedWonderDev/DIS-Reader/internal/unit"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// RemoteService implements DISReaderRemote for multi-tenant remote mode.
type RemoteService struct {
	config   *types.DISConfig
	logger   *slog.Logger
	multiDB  database.MultiTenantDB
	registry bridge.AgentRegistry
	server   *bridge.Server

	defaultTenant string

	unitFactory    func(tenant string) types.UnitService
	invoiceFactory func(tenant string) types.InvoiceService
	partFactory    func(tenant string) types.PartService
}

func NewRemoteService(cfg *types.DISConfig, logger *slog.Logger, multiDB database.MultiTenantDB, registry bridge.AgentRegistry, server *bridge.Server) *RemoteService {
	return &RemoteService{
		config:   cfg,
		logger:   logger,
		multiDB:  multiDB,
		registry: registry,
		server:   server,
		unitFactory: func(tenant string) types.UnitService {
			return unit.NewUnitService(database.BindTenant(multiDB, tenant))
		},
		invoiceFactory: func(tenant string) types.InvoiceService {
			return invoices.NewInvoiceService(database.BindTenant(multiDB, tenant))
		},
		partFactory: func(tenant string) types.PartService {
			return parts.NewService(database.BindTenant(multiDB, tenant))
		},
	}
}

// WithDefaultTenant configures the fallback tenant used when a call omits one.
func (s *RemoteService) WithDefaultTenant(tenant string) {
	s.defaultTenant = tenant
}

// Bridge helpers
func (s *RemoteService) RegisterBridge(server *grpc.Server) {
	if s.server == nil || server == nil {
		return
	}
	proto.RegisterAgentServiceServer(server, s.server)
}

func (s *RemoteService) RegisterHealth(mux *http.ServeMux) {
	if s.registry == nil || mux == nil {
		return
	}
	mux.Handle("/healthz", bridge.HealthHandler(s.registry))
	mux.Handle("/metrics", bridge.MetricsHandler(s.registry))
	if s.config != nil && s.config.Bridge != nil && s.config.Bridge.PprofEnabled {
		base := s.config.Bridge.PprofPath
		if base == "" {
			base = "/debug/pprof/"
		}
		registerPprof(mux, base)
	}
}

// DISReaderRemote implementations
func (s *RemoteService) StartJDBCRunner(ctx context.Context, tenant string) error {
	t, err := s.resolveTenant(tenant)
	if err != nil {
		return err
	}
	return s.multiDB.StartJDBCRunner(t)
}

func (s *RemoteService) StopJDBCRunner(ctx context.Context, tenant string) error {
	t, err := s.resolveTenant(tenant)
	if err != nil {
		return err
	}
	return s.multiDB.StopJDBCRunner(t)
}

func (s *RemoteService) Connect(ctx context.Context, tenant string) error {
	t, err := s.resolveTenant(tenant)
	if err != nil {
		return err
	}
	return s.multiDB.Connect(ctx, t)
}
func (s *RemoteService) Disconnect(ctx context.Context, tenant string) error {
	t, err := s.resolveTenant(tenant)
	if err != nil {
		return err
	}
	return s.multiDB.Disconnect(ctx, t)
}
func (s *RemoteService) PingService(ctx context.Context, tenant string) error {
	t, err := s.resolveTenant(tenant)
	if err != nil {
		return err
	}
	return s.multiDB.PingService(ctx, t)
}
func (s *RemoteService) PingDatabase(ctx context.Context, tenant string) error {
	t, err := s.resolveTenant(tenant)
	if err != nil {
		return err
	}
	return s.multiDB.PingDatabase(ctx, t)
}

// PingBridge checks the bridge server/registry availability (tenant-agnostic)
func (s *RemoteService) PingBridge(ctx context.Context) error {
	if s.registry == nil {
		return fmt.Errorf("registry not initialized")
	}
	_ = ctx
	return nil
}

// PingAgent checks that an agent for the given tenant is available in the registry.
func (s *RemoteService) PingAgent(ctx context.Context, tenant string) error {
	t, err := s.resolveTenant(tenant)
	if err != nil {
		return err
	}
	_, err = s.registry.Pick(ctx, t)
	return err
}

func (s *RemoteService) UnitService(tenant string) types.UnitService {
	t, _ := s.resolveTenant(tenant) // ignore error for factory; caller should already have validated
	return s.unitFactory(t)
}

func (s *RemoteService) InvoiceService(tenant string) types.InvoiceService {
	t, _ := s.resolveTenant(tenant)
	return s.invoiceFactory(t)
}

func (s *RemoteService) PartService(tenant string) types.PartService {
	t, _ := s.resolveTenant(tenant)
	return s.partFactory(t)
}

// UpdateAgentConfig lets callers push a fresh AgentConfig to connected agents and
// update the default used for future agent registrations.
func (s *RemoteService) UpdateAgentConfig(ctx context.Context, cfg *proto.AgentConfig, tenant string, broadcast bool) error {
	if s.server == nil {
		return fmt.Errorf("bridge server not initialized")
	}
	_ = ctx // reserved for future cancellation support
	clientID := ""
	if s.config != nil && s.config.Bridge != nil {
		clientID = s.config.Bridge.ClientID
	}
	s.server.SetAgentConfig(cfg, broadcast, tenant, clientID)
	return nil
}

// ReadAgentConfig requests a sanitized snapshot of the agent's current configuration.
func (s *RemoteService) ReadAgentConfig(ctx context.Context, tenant string) (*proto.AgentConfigStatus, error) {
	t, err := s.resolveTenant(tenant)
	if err != nil {
		return nil, err
	}
	if s.registry == nil {
		return nil, fmt.Errorf("registry not initialized")
	}
	conn, err := s.registry.Pick(ctx, t)
	if err != nil {
		return nil, err
	}
	res, err := conn.SendJob(ctx, &proto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  proto.JobKind_JOB_KIND_READ_CONFIG,
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("nil job result")
	}
	if res.Status == proto.Status_STATUS_ERROR {
		return nil, fmt.Errorf("remote error: %s", res.Message)
	}
	if res.ConfigStatus == nil {
		return nil, fmt.Errorf("agent did not return config status")
	}
	return res.ConfigStatus, nil
}

// resolveTenant chooses the explicit tenant if provided, otherwise the default (if set).
// Returns an error when both are empty.
func (s *RemoteService) resolveTenant(tenant string) (string, error) {
	if tenant != "" {
		return tenant, nil
	}
	if s.defaultTenant != "" {
		return s.defaultTenant, nil
	}
	return "", fmt.Errorf("tenant is required")
}

// Utilities reused from embedded pprof registration
func registerPprof(mux *http.ServeMux, base string) {
	if base == "" {
		base = "/debug/pprof/"
	}
	mux.Handle(base, http.HandlerFunc(pprof.Index))
	mux.Handle(base+"cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle(base+"profile", http.HandlerFunc(pprof.Profile))
	mux.Handle(base+"symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle(base+"trace", http.HandlerFunc(pprof.Trace))
}
